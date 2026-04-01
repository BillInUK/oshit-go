package utils

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/redis/go-redis/v9"
	"io"
	"net/http"
	"net/url"
	"oshit-go/common/pkg/response"
	"strings"
	"time"
)

// ApiResponse - API 通用返回结构
type ApiResponse[T any] struct {
	Code  int    `json:"code"`
	Msg   string `json:"msg"`
	Count int    `json:"count"`
	Data  T      `json:"data"`
}

// GenericApiResponse - 用于message字段的API
type GenericApiResponse[T any] struct {
	Code  int    `json:"code"`
	Msg   string `json:"msg"`
	Count int    `json:"count"`
	Data  T      `json:"data"`
}

// 定义错误类型
var (
	ErrNetwork  = errors.New("network error")  // 网络层错误
	ErrHTTP     = errors.New("http error")     // HTTP 错误
	ErrBusiness = errors.New("business error") // 业务错误
)

// AcquireDistributedRateLimit 分布式限流函数
func AcquireDistributedRateLimit(rdb redis.UniversalClient, key string, limit int) bool {
	now := time.Now().Unix() // 当前时间戳（秒级）

	// 计算 Redis key（每秒一个）
	rateKey := fmt.Sprintf("%s:%d", key, now)

	// 使用 Redis INCR 递增请求数
	count, err := rdb.Incr(context.Background(), rateKey).Result()
	if err != nil {
		fmt.Println("Redis 连接失败:", err)
		return false
	}

	// 设置 1 秒后过期，防止 key 无限制增长
	if count == 1 {
		rdb.Expire(context.Background(), rateKey, time.Second)
	}

	// 如果请求数超过限制，则返回 false
	return count <= int64(limit)
}

// IsRpcRateLimitedError 判断是否为Rpc请求限制错误
func IsRpcRateLimitedError(err error) bool {
	if err == nil {
		return false
	}

	// 如果错误包含 "request limit reached" 字符串，说明是请求限制错误
	if strings.Contains(err.Error(), "request limit reached") {
		return true
	}

	return false
}

// NetworkError 网络层错误
type NetworkError struct {
	Err error
}

func (e *NetworkError) Error() string {
	return fmt.Sprintf("network error: %v", e.Err)
}

func (e *NetworkError) Unwrap() error {
	return e.Err
}

// HTTPError HTTP 错误
type HTTPError struct {
	StatusCode int
	Message    string
	Response   string
}

func (e *HTTPError) Error() string {
	return fmt.Sprintf("http error: status %d, message: %s, response: %s",
		e.StatusCode, e.Message, e.Response)
}

// BusinessError 业务错误
type BusinessError struct {
	Code int
	Msg  string
}

func (e *BusinessError) Error() string {
	return fmt.Sprintf("business error: code=%d, message=%s", e.Code, e.Msg)
}

// PostFormRequest 发送 POST 请求，使用 x-www-form-urlencoded 格式
func PostFormRequest[T any](reqUrl string, formData url.Values, headers map[string]string) (*ApiResponse[T], error) {
	// 创建 HTTP 请求
	req, err := http.NewRequest("POST", reqUrl, strings.NewReader(formData.Encode()))
	if err != nil {
		return nil, &NetworkError{Err: err} // 网络层错误
	}

	// 设置 Content-Type
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// 额外 Headers 处理
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	// 创建 HTTP 客户端，并设置超时时间
	client := &http.Client{Timeout: 1 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, &NetworkError{Err: err} // 网络层错误
	}
	defer resp.Body.Close()

	// 读取响应体
	bodyBytes, _ := io.ReadAll(resp.Body)
	responseBody := string(bodyBytes)

	// 处理非 2xx 状态码
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var errResp struct {
			Error string `json:"error"`
		}
		_ = json.Unmarshal(bodyBytes, &errResp) // 尝试解析 JSON 错误信息

		return nil, &HTTPError{
			StatusCode: resp.StatusCode,
			Message:    errResp.Error,
			Response:   responseBody,
		} // HTTP 错误
	}

	// 解析 JSON 响应
	var apiResponse ApiResponse[T]
	err = json.Unmarshal(bodyBytes, &apiResponse)
	if err != nil {
		return nil, fmt.Errorf("error decoding JSON response: %v, response: %s", err, responseBody)
	}

	// 检查业务层错误（假设 code != 0 表示业务错误）
	if apiResponse.Code != response.SUCCESS {
		return nil, &BusinessError{
			Code: apiResponse.Code,
			Msg:  apiResponse.Msg,
		} // 业务错误
	}

	return &apiResponse, nil
}

// GetGenericRequest 发送 GET 请求
func GetGenericRequest[T any](reqUrl string, queryParams url.Values, headers map[string]string) (*GenericApiResponse[T], error) {
	// 构造完整URL
	parsedURL, err := url.Parse(reqUrl)
	if err != nil {
		return nil, &NetworkError{Err: err}
	}
	parsedURL.RawQuery = queryParams.Encode()
	finalURL := parsedURL.String()

	// 创建请求
	req, err := http.NewRequest("GET", finalURL, nil)
	if err != nil {
		return nil, &NetworkError{Err: err}
	}

	// 设置headers
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	// 发送请求
	client := &http.Client{Timeout: 1 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, &NetworkError{Err: err}
	}
	defer resp.Body.Close()

	// 读取响应
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, &NetworkError{Err: err}
	}
	responseBody := string(bodyBytes)

	// 处理非2xx状态码
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var errResp struct {
			Error string `json:"error"`
		}
		_ = json.Unmarshal(bodyBytes, &errResp)
		return nil, &HTTPError{
			StatusCode: resp.StatusCode,
			Message:    errResp.Error,
			Response:   responseBody,
		}
	}

	// 解析响应
	var apiResponse GenericApiResponse[T]
	if err := json.Unmarshal(bodyBytes, &apiResponse); err != nil {
		return nil, fmt.Errorf("error decoding JSON response: %v, response: %s", err, responseBody)
	}

	// 处理业务错误
	if apiResponse.Code != response.SUCCESS { // 假设response.SUCCESS是0
		return nil, &BusinessError{
			Code: apiResponse.Code,
			Msg:  apiResponse.Msg,
		}
	}

	return &apiResponse, nil
}

// PostGenericFormRequest 发送 POST 请求，使用 x-www-form-urlencoded 格式
func PostGenericFormRequest[T any](reqUrl string, formData url.Values, headers map[string]string) (*GenericApiResponse[T], error) {
	// 创建 HTTP 请求
	req, err := http.NewRequest("POST", reqUrl, strings.NewReader(formData.Encode()))
	if err != nil {
		return nil, &NetworkError{Err: err} // 网络层错误
	}

	// 设置 Content-Type
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// 额外 Headers 处理
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	// 创建 HTTP 客户端，并设置超时时间
	client := &http.Client{Timeout: 1 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, &NetworkError{Err: err} // 网络层错误
	}
	defer resp.Body.Close()

	// 读取响应体
	bodyBytes, _ := io.ReadAll(resp.Body)
	responseBody := string(bodyBytes)

	// 处理非 2xx 状态码
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var errResp struct {
			Error string `json:"error"`
		}
		_ = json.Unmarshal(bodyBytes, &errResp) // 尝试解析 JSON 错误信息

		return nil, &HTTPError{
			StatusCode: resp.StatusCode,
			Message:    errResp.Error,
			Response:   responseBody,
		} // HTTP 错误
	}

	// 解析 JSON 响应
	var apiResponse GenericApiResponse[T]
	err = json.Unmarshal(bodyBytes, &apiResponse)
	if err != nil {
		return nil, fmt.Errorf("error decoding JSON response: %v, response: %s", err, responseBody)
	}

	// 检查业务层错误（假设 code != 0 表示业务错误）
	if apiResponse.Code != response.SUCCESS {
		return nil, &BusinessError{
			Code: apiResponse.Code,
			Msg:  apiResponse.Msg,
		} // 业务错误
	}

	return &apiResponse, nil
}

// PostGenericJsonRequest 发送 POST 请求，使用 x-www-form-urlencoded 格式
func PostGenericJsonRequest[T any](reqUrl string, jsonData interface{}, headers map[string]string) (*GenericApiResponse[T], error) {
	// 序列化 JSON 数据
	jsonBytes, err := json.Marshal(jsonData)
	if err != nil {
		return nil, &NetworkError{
			Err: fmt.Errorf("json marshal error: %w", err),
		}
	}

	// 创建 HTTP 请求
	req, err := http.NewRequest("POST", reqUrl, bytes.NewReader(jsonBytes))
	if err != nil {
		return nil, &NetworkError{Err: err} // 网络层错误
	}

	// 设置 Content-Type
	req.Header.Set("Content-Type", "application/json")

	// 额外 Headers 处理
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	// 创建 HTTP 客户端，并设置超时时间
	client := &http.Client{Timeout: 1 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, &NetworkError{Err: err} // 网络层错误
	}
	defer resp.Body.Close()

	// 读取响应体
	bodyBytes, _ := io.ReadAll(resp.Body)
	responseBody := string(bodyBytes)

	// 处理非 2xx 状态码
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var errResp struct {
			Error string `json:"error"`
		}
		_ = json.Unmarshal(bodyBytes, &errResp) // 尝试解析 JSON 错误信息

		return nil, &HTTPError{
			StatusCode: resp.StatusCode,
			Message:    errResp.Error,
			Response:   responseBody,
		} // HTTP 错误
	}

	// 解析 JSON 响应
	var apiResponse GenericApiResponse[T]
	if err := json.Unmarshal(bodyBytes, &apiResponse); err != nil {
		return nil, fmt.Errorf("error decoding JSON response: %v, response: %s", err, responseBody)
	}

	// 检查业务层错误（假设 code != 0 表示业务错误）
	if apiResponse.Code != response.SUCCESS {
		return nil, &BusinessError{
			Code: apiResponse.Code,
			Msg:  apiResponse.Msg,
		} // 业务错误
	}

	return &apiResponse, nil
}
