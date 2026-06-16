package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
	"oshit-go/common/utils"
)

// 用新密码重新加密 Solana 私钥
//
// 适用于：
//   - base-service-registry.yaml 中的 services.*.encrypted_key
//   - 数据库 t_service_key 表中的 encrypted_key（输出 SQL）
//
// 用法：
//
//	reencrypt_keys --old-key <旧密码> --new-key <新密码> --file base-service-registry.yaml [--dry-run]
//	reencrypt_keys --old-key <旧密码> --new-key <新密码> --value <单个密文> (输出新密文)
const algo = utils.JasyptDefaultAlgorithm

func main() {
	oldKey := flag.String("old-key", "", "old jasypt password (for re-encrypt mode)")
	newKey := flag.String("new-key", "", "new jasypt password (from KMS)")
	file := flag.String("file", "", "YAML file containing services.*.encrypted_key")
	output := flag.String("output", "", "output file (default: overwrite input)")
	value := flag.String("value", "", "single encrypted value to re-encrypt (prints result)")
	plaintext := flag.String("plaintext", "", "plaintext private key to encrypt (prints result)")
	dryRun := flag.Bool("dry-run", false, "print result without writing")
	flag.Parse()

	// 明文加密模式：用于加密新的 Solana 私钥
	if *plaintext != "" {
		if *newKey == "" {
			fmt.Fprintln(os.Stderr, "Usage: reencrypt_keys --new-key <key> --plaintext <private-key>")
			os.Exit(1)
		}
		enc, err := utils.JasyptEncrypt(*plaintext, *newKey, algo)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(enc)
		return
	}

	if *oldKey == "" || *newKey == "" {
		fmt.Fprintln(os.Stderr, "Usage:")
		fmt.Fprintln(os.Stderr, "  加密新私钥:  reencrypt_keys --new-key <key> --plaintext <private-key>")
		fmt.Fprintln(os.Stderr, "  迁移旧密文:  reencrypt_keys --old-key <old> --new-key <new> --value <ciphertext>")
		fmt.Fprintln(os.Stderr, "  迁移YAML:    reencrypt_keys --old-key <old> --new-key <new> --file <yaml>")
		os.Exit(1)
	}

	// 单值模式：用于手动迁移数据库记录
	if *value != "" {
		result, err := reencryptValue(*value, *oldKey, *newKey)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(result)
		return
	}

	if *file == "" {
		fmt.Fprintln(os.Stderr, "Error: --file or --value is required")
		os.Exit(1)
	}

	data, err := os.ReadFile(*file)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading %s: %v\n", *file, err)
		os.Exit(1)
	}

	var root yaml.Node
	if err := yaml.Unmarshal(data, &root); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing YAML: %v\n", err)
		os.Exit(1)
	}

	changed := 0
	if err := processServicesNode(&root, *oldKey, *newKey, &changed); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	out, err := yaml.Marshal(&root)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error marshaling YAML: %v\n", err)
		os.Exit(1)
	}

	if *dryRun {
		fmt.Print(string(out))
		fmt.Fprintf(os.Stderr, "\n(%d keys re-encrypted)\n", changed)
		return
	}

	outPath := *file
	if *output != "" {
		outPath = *output
	}
	if err := os.WriteFile(outPath, out, 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing %s: %v\n", outPath, err)
		os.Exit(1)
	}

	fmt.Printf("Re-encrypted %d keys → %s\n", changed, outPath)

	// 输出数据库迁移提示
	fmt.Fprintln(os.Stderr, "\n=== 数据库 t_service_key 迁移 ===")
	fmt.Fprintln(os.Stderr, "对每条记录执行:")
	fmt.Fprintln(os.Stderr, "  reencrypt_keys --old-key <old> --new-key <new> --value <encrypted_key>")
	fmt.Fprintln(os.Stderr, "然后用输出的新密文更新数据库:")
	fmt.Fprintln(os.Stderr, "  UPDATE t_service_key SET encrypted_key = '<新密文>' WHERE service = '...' AND sub_service = '...';")
}

// processServicesNode 在 yaml.Node 树中找到 services 列表，遍历每个 item 的 encrypted_key 并重新加密
func processServicesNode(node *yaml.Node, oldKey, newKey string, changed *int) error {
	// 找到 document 根节点下的 mapping
	rootMap := node
	if node.Kind == yaml.DocumentNode && len(node.Content) > 0 {
		rootMap = node.Content[0]
	}
	if rootMap.Kind != yaml.MappingNode {
		return fmt.Errorf("YAML root is not a map")
	}

	// 找到 "services" key
	var servicesNode *yaml.Node
	for i := 0; i+1 < len(rootMap.Content); i += 2 {
		if rootMap.Content[i].Value == "services" {
			servicesNode = rootMap.Content[i+1]
			break
		}
	}
	if servicesNode == nil {
		return fmt.Errorf("no 'services' key in YAML")
	}
	if servicesNode.Kind != yaml.SequenceNode {
		return fmt.Errorf("'services' is not a list")
	}

	for idx, itemNode := range servicesNode.Content {
		if itemNode.Kind != yaml.MappingNode {
			continue
		}
		var serviceName, subService string
		var encKeyNode *yaml.Node
		for i := 0; i+1 < len(itemNode.Content); i += 2 {
			k := itemNode.Content[i].Value
			v := itemNode.Content[i+1]
			switch k {
			case "service":
				serviceName = v.Value
			case "sub_service":
				subService = v.Value
			case "encrypted_key":
				encKeyNode = v
			}
		}
		if encKeyNode == nil || strings.TrimSpace(encKeyNode.Value) == "" {
			continue
		}

		newEncKey, err := reencryptValue(encKeyNode.Value, oldKey, newKey)
		if err != nil {
			return fmt.Errorf("service[%d] %s/%s: %w", idx, serviceName, subService, err)
		}
		encKeyNode.Value = newEncKey
		encKeyNode.Tag = "!!str"
		*changed++
		fmt.Fprintf(os.Stderr, "  [R] %s/%s\n", serviceName, subService)
	}

	return nil
}

func reencryptValue(ciphertext, oldKey, newKey string) (string, error) {
	// 用旧密码解密
	plaintext, err := utils.JasyptDecrypt(ciphertext, oldKey, algo)
	if err != nil {
		return "", fmt.Errorf("decrypt with old key: %w", err)
	}

	// 用新密码加密
	newCiphertext, err := utils.JasyptEncrypt(plaintext, newKey, algo)
	if err != nil {
		return "", fmt.Errorf("encrypt with new key: %w", err)
	}

	return newCiphertext, nil
}
