package rpc

import (
	"fmt"
	"net/url"
	"oshit-go/common/pkg/entity"
	"strconv"
)

// CampaignClient - 封装 API 请求的客户端
type CampaignClient struct {
	baseURL string
	headers map[string]string
}

// NewCampaignClient - 创建 API 客户端实例
func NewCampaignClient(baseURL string, headers map[string]string) *CampaignClient {
	return &CampaignClient{
		baseURL: baseURL,
		headers: headers,
	}
}

// LoginTokenToUserID 通过 JWT token 获取用户 ID
func (c *CampaignClient) LoginTokenToUserID(jwtToken string) (uint64, error) {
	apiPath := "/social-media/internal/user/loginTokenToUserId"

	headers := map[string]string{
		"Content-Type": "application/json",
		"x-ac-jwt":     jwtToken, // 假设 jwtToken 是 CampaignClient 的字段
	}

	response, err := GetGenericRequest[uint64](
		c.baseURL+apiPath,
		nil, // 无查询参数
		headers,
	)
	if err != nil {
		return 0, err
	}

	if response.Code != 0 {
		return 0, fmt.Errorf("failed to get user ID: %s", response.Msg)
	}

	return response.Data, nil
}

// QueryScore 查询用户积分
func (c *CampaignClient) QueryScore(userId uint64) (*entity.ScoreData, error) {
	apiPath := "/score/internal/op/queryScore"

	queryParams := url.Values{}
	queryParams.Set("userId", strconv.FormatUint(userId, 10))

	headers := map[string]string{
		"Content-Type": "application/json",
	}

	response, err := GetGenericRequest[entity.ScoreData](
		c.baseURL+apiPath,
		queryParams,
		headers,
	)
	if err != nil {
		return nil, err
	}

	return &response.Data, nil
}

// FreezeScore 冻结积分
func (c *CampaignClient) FreezeScore(
	userId uint64,
	txId string,
	score uint64,
	sourceSys string,
	businessName string,
	reason string,
) (*entity.ScoreOperationResult, error) {
	apiPath := "/score/internal/op/freeze"

	requestBody := entity.FreezeRequest{
		UserID:       userId,
		TxID:         txId,
		Score:        score,
		SourceSys:    sourceSys,
		BusinessName: businessName,
		Reason:       reason,
	}

	response, err := PostGenericJsonRequest[entity.ScoreOperationResult](
		c.baseURL+apiPath,
		requestBody,
		c.headers,
	)
	if err != nil {
		return nil, err
	}

	return &response.Data, nil
}

// UnfreezeScore 解冻积分
func (c *CampaignClient) UnfreezeScore(
	userId uint64,
	txId string,
	flowId uint64,
	sourceSys string,
	businessName string,
	reason string,
) (*entity.ScoreOperationResult, error) {
	apiPath := "/score/internal/op/unfreeze"

	requestBody := entity.UnFreezeRequest{
		UserID:       userId,
		TxID:         txId,
		FlowID:       flowId,
		SourceSys:    sourceSys,
		BusinessName: businessName,
		Reason:       reason,
	}

	response, err := PostGenericJsonRequest[entity.ScoreOperationResult](
		c.baseURL+apiPath,
		requestBody,
		c.headers,
	)
	if err != nil {
		return nil, err
	}

	return &response.Data, nil
}

// ConsumeFrozenScore 消费冻结的积分
func (c *CampaignClient) ConsumeFrozenScore(
	userId uint64,
	txId string,
	flowId uint64,
	sourceSys string,
	businessName string,
	reason string) (*entity.ScoreOperationResult, error) {
	apiPath := "/score/internal/op/consumeFrozen"

	requestBody := entity.ConsumeFrozenRequest{
		UserID:       userId,
		TxID:         txId,
		FlowID:       flowId,
		SourceSys:    sourceSys,
		BusinessName: businessName,
		Reason:       reason,
	}

	response, err := PostGenericJsonRequest[entity.ScoreOperationResult](
		c.baseURL+apiPath,
		requestBody,
		c.headers,
	)
	if err != nil {
		return nil, err
	}

	return &response.Data, nil
}
