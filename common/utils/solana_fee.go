package utils

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/gagliardetto/solana-go"
	"github.com/mr-tron/base58"
)

type QnEstimatePriorityFeesParam struct {
	LastNBlocks int    `json:"last_n_blocks"`
	Account     string `json:"account"`
}

type QnRequestBody struct {
	JSONRPC string                      `json:"jsonrpc"`
	ID      int                         `json:"id"`
	Method  string                      `json:"method"`
	Params  QnEstimatePriorityFeesParam `json:"params"`
}

// QnEstimatePriorityFee ResponseBody 用于存储响应体
type QnEstimatePriorityFee struct {
	JSONRPC string `json:"jsonrpc"`
	Result  struct {
		Context struct {
			Slot int `json:"slot"`
		} `json:"context"`
		PerComputeUnit struct {
			Extreme     int `json:"extreme"`
			High        int `json:"high"`
			Low         int `json:"low"`
			Medium      int `json:"medium"`
			Percentiles struct {
				P50  int `json:"50"`
				P55  int `json:"55"`
				P60  int `json:"60"`
				P65  int `json:"65"`
				P70  int `json:"70"`
				P75  int `json:"75"`
				P80  int `json:"80"`
				P85  int `json:"85"`
				P90  int `json:"90"`
				P95  int `json:"95"`
				P100 int `json:"100"`
			}
		} `json:"per_compute_unit"`

		PerTransaction struct {
			Extreme     int `json:"extreme"`
			High        int `json:"high"`
			Low         int `json:"low"`
			Medium      int `json:"medium"`
			Percentiles struct {
				P50  int `json:"50"`
				P55  int `json:"55"`
				P60  int `json:"60"`
				P65  int `json:"65"`
				P70  int `json:"70"`
				P75  int `json:"75"`
				P80  int `json:"80"`
				P85  int `json:"85"`
				P90  int `json:"90"`
				P95  int `json:"95"`
				P100 int `json:"100"`
			}
		} `json:"per_transaction"`
	} `json:"result"`
	ID int `json:"id"`
}

func GetQnEstimatePriorityFees(url string, lastBlocks int, programId solana.PublicKey) (*QnEstimatePriorityFee, error) {
	params := QnEstimatePriorityFeesParam{
		LastNBlocks: lastBlocks,
		Account:     programId.String(),
	}
	requestBody := QnRequestBody{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "qn_estimatePriorityFees",
		Params:  params,
	}
	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("error marshalling request body: %v", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("error creating HTTP request: %v", err)
	}

	req.Header.Set("x-qn-api-version", "1")
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error sending HTTP request: %v", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %v", err)
	}

	var estimatePriorityFee QnEstimatePriorityFee
	if err := json.Unmarshal(body, &estimatePriorityFee); err != nil {
		return nil, fmt.Errorf("error unmarshalling response body: %v", err)
	}
	return &estimatePriorityFee, nil
}

func CalcGasFee(sigNum, lamPerSig, computeUnitPrice, computeUnitLimit uint64, computeUnitSet bool) uint64 {
	var unitLimit uint64 = 0
	if computeUnitSet {
		unitLimit = computeUnitLimit
	} else {
		unitLimit = DefaultComputeUnitLimit
	}
	sigFee := sigNum * lamPerSig
	computeUnitsDFee := (float64(computeUnitPrice) * 1000 / float64(solana.LAMPORTS_PER_SOL)) * float64(unitLimit)
	return uint64(math.Ceil(float64(sigFee) + computeUnitsDFee))
}

// GetInstructionDiscriminator 计算指令的 discriminator
func GetInstructionDiscriminator(instructionName string) []byte {
	hash := sha256.Sum256([]byte("global:" + instructionName))
	return hash[:8]
}

// FindPDA 查找PDA账户
func FindPDA(programID solana.PublicKey, seeds []byte, additionalSeeds ...[]byte) (solana.PublicKey, uint8, error) {
	allSeeds := [][]byte{seeds}
	allSeeds = append(allSeeds, additionalSeeds...)
	return solana.FindProgramAddress(allSeeds, programID)
}

// GenerateRandomSolanaTxID 生成随机的 Solana 交易 ID（Signature）
func GenerateRandomSolanaTxID() (solana.Signature, error) {
	signatureBytes := make([]byte, 64)
	_, err := rand.Read(signatureBytes)
	if err != nil {
		return solana.Signature{}, fmt.Errorf("failed to generate random bytes: %v", err)
	}
	signature := solana.SignatureFromBytes(signatureBytes)
	return signature, nil
}

// GenerateRandomSolanaTxIDString 生成随机的 Solana 交易 ID 字符串
func GenerateRandomSolanaTxIDString() (string, error) {
	signature, err := GenerateRandomSolanaTxID()
	if err != nil {
		return "", err
	}
	return signature.String(), nil
}

func SaveSolanaPrivateKeyToJSON(base58PrivateKey, filename, filePath string) error {
	privateKeyBytes, err := base58.Decode(base58PrivateKey)
	if err != nil {
		return fmt.Errorf("invalid Base58 private key: %v", err)
	}

	if err := os.MkdirAll(filePath, 0700); err != nil {
		return fmt.Errorf("failed to create directory: %v", err)
	}

	fullPath := filepath.Join(filePath, filename)

	file, err := os.Create(fullPath)
	if err != nil {
		return fmt.Errorf("failed to create file: %v", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")

	if err := encoder.Encode(privateKeyBytes); err != nil {
		return fmt.Errorf("JSON encoding failed: %v", err)
	}

	fmt.Printf("Private key successfully saved to: %s\n", fullPath)
	return nil
}

// FilterAndTranslateSOLError 过滤并翻译SOL错误信息为英文提示
func FilterAndTranslateSOLError(err error) string {
	if err == nil {
		return ""
	}

	errStr := err.Error()

	if strings.Contains(errStr, "status code: 50") {
		if strings.Contains(errStr, "status code: 502") {
			return "Network service error, please try again later"
		} else if strings.Contains(errStr, "status code: 503") {
			return "Service temporarily unavailable, please try again later"
		} else if strings.Contains(errStr, "status code: 504") {
			return "Request timeout, please try again later"
		}
		return "Network connection error, please try again later"
	}

	if strings.Contains(errStr, "context deadline exceeded") ||
		strings.Contains(errStr, "Client.Timeout exceeded") {
		return "Request timeout, please check your network connection"
	}

	if strings.Contains(errStr, "Transaction simulation failed") {
		if strings.Contains(strings.ToLower(errStr), "insufficient funds") {
			return "Insufficient account balance, please ensure you have enough SOL for transaction fees and rent"
		}
		if strings.Contains(errStr, "Blockhash not found") {
			return "Transaction expired, please refresh and try again"
		}
		if strings.Contains(errStr, "already been processed") {
			return "Transaction already processed successfully, no need to resubmit"
		}
		if strings.Contains(errStr, "custom program error") {
			re := regexp.MustCompile(`custom program error:\s*(0x[0-9a-fA-F]+|\d+)`)
			if match := re.FindStringSubmatch(errStr); len(match) > 1 {
				errorCode := match[1]
				switch errorCode {
				case "0x1900", "6400":
					return "Wallet compatibility issue. Please try using a different wallet or contact our support team for assistance."
				case "0x0":
					return "Wallet compatibility issue. Please try using a different wallet or contact our support team for assistance."
				default:
					return fmt.Sprintf("Wallet compatibility issue. Please try using a different wallet or contact our support team for assistance. (Error code: %s)", errorCode)
				}
			}
			return "Wallet compatibility issue. Please try using a different wallet or contact our support team for assistance."
		}
		return "Transaction validation failed, please check transaction data"
	}

	if strings.Contains(errStr, "rpc call sendTransaction") {
		if strings.Contains(errStr, "Post") && strings.Contains(errStr, "EOF") {
			return "Network connection interrupted, please try again later"
		}
		return "Network request failed, please check your connection"
	}

	return "Failed to send transaction, please try again later"
}
