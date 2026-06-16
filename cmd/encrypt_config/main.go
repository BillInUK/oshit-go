package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
	"oshit-go/common/utils"
)

const encPrefix = "ENC~"

// 需要加密的字段路径（dot notation，* 匹配数组下标）
var sensitiveFields = []string{
	// base-runtime.yaml
	"aws.access_key_id",
	"aws.secret_access_key",
	"mainnet_rpc.api_key",
	"rpc_endpoints.*.api_key",
	"rpc_endpoints.*.wss_api_key",
	// application.yaml
	"database.password",
	"database.user",
	"redis.password",
	"nacos.username",
	"nacos.password",
	"nacos.client_config.username",
	"nacos.client_config.password",
}

func main() {
	key := flag.String("key", "", "encryption key")
	file := flag.String("file", "", "YAML file to process")
	output := flag.String("output", "", "output file (default: overwrite input)")
	decrypt := flag.Bool("decrypt", false, "decrypt instead of encrypt")
	dryRun := flag.Bool("dry-run", false, "print result without writing")
	flag.Parse()

	if *key == "" || *file == "" {
		fmt.Fprintln(os.Stderr, "Usage: encrypt_config --key <key> --file <file> [--output <file>] [--decrypt] [--dry-run]")
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
	processNode(&root, *key, *decrypt, "", &changed)

	out, err := yaml.Marshal(&root)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error marshaling YAML: %v\n", err)
		os.Exit(1)
	}

	if *dryRun {
		fmt.Print(string(out))
		fmt.Fprintf(os.Stderr, "\n(%d fields processed)\n", changed)
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

	action := "Encrypted"
	if *decrypt {
		action = "Decrypted"
	}
	fmt.Printf("%s %d fields → %s\n", action, changed, outPath)
}

func processNode(node *yaml.Node, key string, decrypt bool, path string, changed *int) {
	switch node.Kind {
	case yaml.DocumentNode:
		for _, child := range node.Content {
			processNode(child, key, decrypt, path, changed)
		}
	case yaml.MappingNode:
		// Content 交替存放 key, value
		for i := 0; i+1 < len(node.Content); i += 2 {
			keyNode := node.Content[i]
			valNode := node.Content[i+1]
			current := keyNode.Value
			if path != "" {
				current = path + "." + keyNode.Value
			}
			if valNode.Kind == yaml.ScalarNode && isSensitive(current) {
				result, err := processValue(valNode.Value, key, decrypt)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Warning: %s: %v\n", current, err)
					continue
				}
				if result != valNode.Value {
					valNode.Value = result
					valNode.Tag = "!!str"
					*changed++
					fmt.Fprintf(os.Stderr, "  %s %s\n", actionSymbol(decrypt), current)
				}
			} else {
				processNode(valNode, key, decrypt, current, changed)
			}
		}
	case yaml.SequenceNode:
		for i, child := range node.Content {
			processNode(child, key, decrypt, fmt.Sprintf("%s.%d", path, i), changed)
		}
	}
}

func processValue(val, key string, decrypt bool) (string, error) {
	if decrypt {
		if !strings.HasPrefix(val, encPrefix) {
			return val, nil
		}
		plain, err := utils.JasyptDecrypt(val[len(encPrefix):], key, utils.JasyptDefaultAlgorithm)
		if err != nil {
			return "", err
		}
		return plain, nil
	}
	// encrypt
	if val == "" || strings.HasPrefix(val, encPrefix) {
		return val, nil
	}
	enc, err := utils.JasyptEncrypt(val, key, utils.JasyptDefaultAlgorithm)
	if err != nil {
		return "", err
	}
	return encPrefix + enc, nil
}

func isSensitive(fieldPath string) bool {
	pathParts := strings.Split(fieldPath, ".")
	for _, pattern := range sensitiveFields {
		patParts := strings.Split(pattern, ".")
		if matchParts(pathParts, patParts) {
			return true
		}
	}
	return false
}

func matchParts(path, pattern []string) bool {
	if len(path) != len(pattern) {
		return false
	}
	for i := range pattern {
		if pattern[i] == "*" {
			continue
		}
		if path[i] != pattern[i] {
			return false
		}
	}
	return true
}

func actionSymbol(decrypt bool) string {
	if decrypt {
		return "[D]"
	}
	return "[E]"
}
