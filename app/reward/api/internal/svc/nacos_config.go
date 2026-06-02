package svc

import (
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/gofiber/fiber/v2/log"
	"github.com/nacos-group/nacos-sdk-go/v2/clients"
	"github.com/nacos-group/nacos-sdk-go/v2/common/constant"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
	"gopkg.in/yaml.v3"

	"oshit-go/app/reward/api/internal/config"
	"oshit-go/common/pkg/dal/model"
)

const (
	defaultNacosGroup              = "oshit-go"
	defaultBaseRuntimeConfigDataID = "base-runtime.yaml"
	defaultRewardRuntimeDataID     = "reward-runtime.yaml"
)

type baseRuntimeNacosConfig struct {
	System            runtimeSystemConfig `yaml:"system"`
	Chain             runtimeChainConfig  `yaml:"chain"`
	Token             runtimeTokenConfig  `yaml:"token"`
	FeeTolerance      runtimeFeeTolerance `yaml:"fee_tolerance"`
	LightHouseAddress string              `yaml:"lighthouse_address"`
}

type runtimeSystemConfig struct {
	Env int32 `yaml:"env"`
}

type runtimeChainConfig struct {
	ChainName string `yaml:"chain_name"`
	RPCURL    string `yaml:"rpc_url"`
	WssURL    string `yaml:"wss_url"`
	Decimals  int32  `yaml:"decimals"`
	Symbol    string `yaml:"symbol"`
}

type runtimeTokenConfig struct {
	TokenName   string `yaml:"token_name"`
	TokenSymbol string `yaml:"token_symbol"`
	Decimals    int32  `yaml:"decimals"`
	Mint        string `yaml:"mint"`
}

type runtimeFeeTolerance struct {
	MaxLessRate float64 `yaml:"max_less_rate"`
}

type rewardRuntimeNacosConfig struct {
	LevelDist           rewardLevelDistConfig     `yaml:"level_dist"`
	LevelRatio          []rewardLevelRatioConfig  `yaml:"level_ratio"`
	DiscountRate        rewardDiscountRateConfig  `yaml:"discount_rate"`
	TakeTokenConfig     rewardTakeTokenConfig     `yaml:"take_token"`
	GiveTokenConfig     rewardGiveTokenConfig     `yaml:"give_token"`
	CampaignQuoteConfig rewardCampaignQuoteConfig `yaml:"campaign_quote"`
	RewardCodeConfig    rewardRewardCodeConfig    `yaml:"reward_code"`
}

type rewardLevelDistConfig struct {
	DistLevel int32 `yaml:"dist_level"`
}

type rewardLevelRatioConfig struct {
	DistLevel int32   `yaml:"dist_level"`
	Ratio     float64 `yaml:"ratio"`
}

type rewardDiscountRateConfig struct {
	Rate float64 `yaml:"rate"`
}

type rewardTakeTokenConfig struct {
	InviteCode    string  `yaml:"invite_code"`
	RewardAccount string  `yaml:"reward_account"`
	CostAccount   string  `yaml:"cost_account"`
	Amount        float64 `yaml:"amount"`
	InviteAmount  float64 `yaml:"invite_amount"`
	CostFeeRate   float64 `yaml:"cost_fee_rate"`
	MaxCostFee    float64 `yaml:"max_cost_fee"`
	IsDefault     bool    `yaml:"is_default"`
	RewardInviter bool    `yaml:"reward_inviter"`
	Invited       bool    `yaml:"invited"`
}

type rewardGiveTokenConfig struct {
	RewardAccount  string  `yaml:"reward_account"`
	CostAccount    string  `yaml:"cost_account"`
	RewardRate     float64 `yaml:"reward_rate"`
	MaxReward      float64 `yaml:"max_reward"`
	MaxValidReward float64 `yaml:"max_valid_reward"`
	ValidRate      float64 `yaml:"valid_rate"`
}

type rewardCampaignQuoteConfig struct {
	RewardAccount string  `yaml:"reward_account"`
	CostAccount   string  `yaml:"cost_account"`
	QuoteRate     float64 `yaml:"quote_rate"`
	CostRate      float64 `yaml:"cost_rate"`
}

type rewardRewardCodeConfig struct {
	RewardAccount string `yaml:"reward_account"`
	CostAccount   string `yaml:"cost_account"`
}

func (s *ServiceContext) initNacosConfigClient() error {
	nacosCfg := s.Config.Nacos
	serverConfigs := make([]constant.ServerConfig, 0, len(nacosCfg.ServerConfig))
	for _, sc := range nacosCfg.ServerConfig {
		serverConfigs = append(serverConfigs, constant.ServerConfig{
			IpAddr:   sc.Host,
			Port:     sc.Port,
			GrpcPort: sc.GrpcPort,
		})
	}
	if len(serverConfigs) == 0 {
		return fmt.Errorf("nacos server config is empty")
	}

	clientCfg := nacosCfg.ClientConfig
	namespace := clientCfg.NamespaceId
	if namespace == "public" {
		namespace = ""
	}
	timeout := clientCfg.TimeoutMs
	if timeout == 0 {
		timeout = 5000
	}

	configClient, err := clients.NewConfigClient(vo.NacosClientParam{
		ClientConfig: &constant.ClientConfig{
			NamespaceId:         namespace,
			TimeoutMs:           timeout,
			NotLoadCacheAtStart: clientCfg.NotLoadCacheAtStart,
			LogDir:              defaultString(clientCfg.LogDir, "./log/nacos"),
			CacheDir:            defaultString(clientCfg.CacheDir, "./cache/nacos"),
			LogLevel:            defaultString(clientCfg.LogLevel, "info"),
			Username:            clientCfg.Username,
			Password:            clientCfg.Password,
		},
		ServerConfigs: serverConfigs,
	})
	if err != nil {
		return err
	}
	s.NacosConfigClient = configClient
	return nil
}

func (s *ServiceContext) initNacosRuntimeConfigs() error {
	if s.NacosConfigClient == nil {
		return nil
	}
	if err := s.loadBaseRuntimeFromNacos(); err != nil {
		log.Warnf("load base runtime config from nacos failed, keep database configs: %v", err)
	}
	if err := s.loadRewardRuntimeFromNacos(); err != nil {
		log.Warnf("load reward runtime config from nacos failed, keep database configs: %v", err)
	}
	return nil
}

func (s *ServiceContext) listenNacosConfigs() error {
	if s.NacosConfigClient == nil {
		return nil
	}
	baseSub := s.baseRuntimeSubscribeConfig()
	if err := s.NacosConfigClient.ListenConfig(vo.ConfigParam{
		DataId: baseSub.DataId,
		Group:  baseSub.Group,
		OnChange: func(namespace, group, dataId, data string) {
			if err := s.applyBaseRuntimeContent(data); err != nil {
				log.Errorf("apply nacos base runtime config [%s/%s] error: %v", group, dataId, err)
			}
		},
	}); err != nil {
		return err
	}

	rewardSub := s.rewardRuntimeSubscribeConfig()
	return s.NacosConfigClient.ListenConfig(vo.ConfigParam{
		DataId: rewardSub.DataId,
		Group:  rewardSub.Group,
		OnChange: func(namespace, group, dataId, data string) {
			if err := s.applyRewardRuntimeContent(data); err != nil {
				log.Errorf("apply nacos reward runtime config [%s/%s] error: %v", group, dataId, err)
			}
		},
	})
}

func (s *ServiceContext) loadBaseRuntimeFromNacos() error {
	sub := s.baseRuntimeSubscribeConfig()
	content, err := s.NacosConfigClient.GetConfig(vo.ConfigParam{DataId: sub.DataId, Group: sub.Group})
	if err != nil {
		return err
	}
	if strings.TrimSpace(content) == "" {
		return fmt.Errorf("empty nacos content")
	}
	return s.applyBaseRuntimeContent(content)
}

func (s *ServiceContext) loadRewardRuntimeFromNacos() error {
	sub := s.rewardRuntimeSubscribeConfig()
	content, err := s.NacosConfigClient.GetConfig(vo.ConfigParam{DataId: sub.DataId, Group: sub.Group})
	if err != nil {
		return err
	}
	if strings.TrimSpace(content) == "" {
		return fmt.Errorf("empty nacos content")
	}
	return s.applyRewardRuntimeContent(content)
}

func (s *ServiceContext) applyBaseRuntimeContent(content string) error {
	var cfg baseRuntimeNacosConfig
	if err := yaml.Unmarshal([]byte(content), &cfg); err != nil {
		return err
	}
	if cfg.Chain.ChainName == "" || cfg.Chain.RPCURL == "" {
		return fmt.Errorf("chain.chain_name and chain.rpc_url are required")
	}
	if cfg.Token.Mint == "" {
		return fmt.Errorf("token.mint is required")
	}
	lighthouse := cfg.LightHouseAddress
	if lighthouse == "" {
		lighthouse = s.LightHouseAddress.String()
	}
	lighthouseAddr, err := solana.PublicKeyFromBase58(lighthouse)
	if err != nil {
		return fmt.Errorf("invalid lighthouse address: %v", err)
	}

	now := time.Now()
	s.ConfigMu.Lock()
	defer s.ConfigMu.Unlock()
	s.SystemConfig = &model.SystemConfig{Env: cfg.System.Env, CreatedAt: now, UpdatedAt: now}
	s.ChainConfig = &model.ChainConfig{ChainName: cfg.Chain.ChainName, RPCURL: cfg.Chain.RPCURL, WssURL: cfg.Chain.WssURL, Decimals: cfg.Chain.Decimals, Symbol: cfg.Chain.Symbol, CreatedAt: now, UpdatedAt: now}
	s.TokenConfig = &model.TokenConfig{TokenName: cfg.Token.TokenName, TokenSymbol: cfg.Token.TokenSymbol, Decimals: cfg.Token.Decimals, Mint: cfg.Token.Mint, CreatedAt: now, UpdatedAt: now}
	s.TokenDecimal = math.Pow(10, float64(cfg.Token.Decimals))
	s.FeeTolerance = &model.FeeTolerance{MaxLessRate: cfg.FeeTolerance.MaxLessRate, CreatedAt: now, UpdatedAt: now}
	s.LightHouseAddress = lighthouseAddr
	if cfg.Chain.RPCURL != "" {
		s.RpcClient = rpc.New(cfg.Chain.RPCURL)
	}
	log.Infof("loaded base runtime config from nacos")
	return nil
}

func (s *ServiceContext) applyRewardRuntimeContent(content string) error {
	var cfg rewardRuntimeNacosConfig
	if err := yaml.Unmarshal([]byte(content), &cfg); err != nil {
		return err
	}
	if cfg.LevelDist.DistLevel <= 0 {
		return fmt.Errorf("level_dist.dist_level must be greater than 0")
	}
	if len(cfg.LevelRatio) == 0 {
		return fmt.Errorf("level_ratio is required")
	}
	if cfg.TakeTokenConfig.RewardAccount == "" || cfg.TakeTokenConfig.CostAccount == "" {
		return fmt.Errorf("take_token.reward_account and take_token.cost_account are required")
	}
	if cfg.GiveTokenConfig.RewardAccount == "" || cfg.GiveTokenConfig.CostAccount == "" {
		return fmt.Errorf("give_token.reward_account and give_token.cost_account are required")
	}
	if cfg.CampaignQuoteConfig.RewardAccount == "" || cfg.CampaignQuoteConfig.CostAccount == "" {
		return fmt.Errorf("campaign_quote.reward_account and campaign_quote.cost_account are required")
	}
	if cfg.RewardCodeConfig.RewardAccount == "" || cfg.RewardCodeConfig.CostAccount == "" {
		return fmt.Errorf("reward_code.reward_account and reward_code.cost_account are required")
	}

	levelRatio := make([]model.LevelRatio, 0, len(cfg.LevelRatio))
	levelRatioMap := make(map[int32]model.LevelRatio, len(cfg.LevelRatio))
	for _, item := range cfg.LevelRatio {
		if item.DistLevel <= 0 {
			return fmt.Errorf("level_ratio.dist_level must be greater than 0")
		}
		ratio := model.LevelRatio{DistLevel: item.DistLevel, Ratio: item.Ratio}
		levelRatio = append(levelRatio, ratio)
		levelRatioMap[item.DistLevel] = ratio
	}

	now := time.Now()
	s.ConfigMu.Lock()
	defer s.ConfigMu.Unlock()
	s.LevelDist = &model.LevelDist{DistLevel: cfg.LevelDist.DistLevel}
	s.LevelRatio = levelRatio
	s.LevelRatioMap = levelRatioMap
	s.DiscountRate = &model.DiscountRate{Rate: cfg.DiscountRate.Rate}
	s.TakeTokenConfig = &model.TakeTokenConfig{
		InviteCode:    cfg.TakeTokenConfig.InviteCode,
		RewardAccount: cfg.TakeTokenConfig.RewardAccount,
		CostAccount:   cfg.TakeTokenConfig.CostAccount,
		Amount:        cfg.TakeTokenConfig.Amount,
		InviteAmount:  cfg.TakeTokenConfig.InviteAmount,
		CostFeeRate:   cfg.TakeTokenConfig.CostFeeRate,
		MaxCostFee:    cfg.TakeTokenConfig.MaxCostFee,
		IsDefault:     cfg.TakeTokenConfig.IsDefault,
		RewardInviter: cfg.TakeTokenConfig.RewardInviter,
		Invited:       cfg.TakeTokenConfig.Invited,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	s.GiveTokenConfig = &model.GiveTokenConfig{
		RewardAccount:  cfg.GiveTokenConfig.RewardAccount,
		CostAccount:    cfg.GiveTokenConfig.CostAccount,
		RewardRate:     cfg.GiveTokenConfig.RewardRate,
		MaxReward:      cfg.GiveTokenConfig.MaxReward,
		MaxValidReward: cfg.GiveTokenConfig.MaxValidReward,
		ValidRate:      cfg.GiveTokenConfig.ValidRate,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	s.CampaignQuoteConfig = &model.CampaignQuoteConfig{
		RewardAccount: cfg.CampaignQuoteConfig.RewardAccount,
		CostAccount:   cfg.CampaignQuoteConfig.CostAccount,
		QuoteRate:     cfg.CampaignQuoteConfig.QuoteRate,
		CostRate:      cfg.CampaignQuoteConfig.CostRate,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	s.RewardCodeConfig = &model.RewardCodeConfig{
		RewardAccount: cfg.RewardCodeConfig.RewardAccount,
		CostAccount:   cfg.RewardCodeConfig.CostAccount,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	log.Infof("loaded reward runtime config from nacos")
	return nil
}

func (s *ServiceContext) baseRuntimeSubscribeConfig() config.NacosSubscribeConfig {
	sub := s.Config.Nacos.SubscribeConfigs.BaseRuntime
	if sub.DataId == "" {
		sub = s.Config.Nacos.SubscribeConfig
	}
	if sub.DataId == "" {
		sub.DataId = defaultBaseRuntimeConfigDataID
	}
	if sub.Group == "" {
		sub.Group = defaultNacosGroup
	}
	return sub
}

func (s *ServiceContext) rewardRuntimeSubscribeConfig() config.NacosSubscribeConfig {
	sub := s.Config.Nacos.SubscribeConfigs.RewardRuntime
	if sub.DataId == "" {
		sub.DataId = defaultRewardRuntimeDataID
	}
	if sub.Group == "" {
		sub.Group = defaultNacosGroup
	}
	return sub
}

func defaultString(value, fallback string) string {
	if value != "" {
		return value
	}
	return fallback
}
