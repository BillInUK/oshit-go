package svc

import (
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/gagliardetto/solana-go"
	"github.com/gofiber/fiber/v2/log"
	"github.com/nacos-group/nacos-sdk-go/v2/clients"
	"github.com/nacos-group/nacos-sdk-go/v2/common/constant"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
	"gopkg.in/yaml.v3"

	"oshit-go/app/pos/api/internal/config"
	"oshit-go/common/pkg/dal/model"
	"oshit-go/common/utils"
)

const (
	defaultNacosGroup              = "oshit-go"
	defaultBaseRuntimeConfigDataID = "base-runtime.yaml"
	defaultPosRuntimeDataID        = "pos-runtime.yaml"
)

// ============ base-runtime.yaml 结构 ============

type baseRuntimeNacosConfig struct {
	System            runtimeSystemConfig  `yaml:"system"`
	Chain             runtimeChainConfig   `yaml:"chain"`
	RPCEndpoints      []runtimeRPCEndpoint `yaml:"rpc_endpoints"`
	Token             runtimeTokenConfig   `yaml:"token"`
	FeeTolerance      runtimeFeeTolerance  `yaml:"fee_tolerance"`
	LightHouseAddress string               `yaml:"lighthouse_address"`
}

type runtimeRPCEndpoint struct {
	Provider    string `yaml:"provider"`
	Endpoint    string `yaml:"endpoint"`
	APIKey      string `yaml:"api_key"`
	WssEndpoint string `yaml:"wss_endpoint"`
	WssAPIKey   string `yaml:"wss_api_key"`
	Weight      int    `yaml:"weight"`
}

type runtimeSystemConfig struct {
	Env int32 `yaml:"env"`
}

type runtimeChainConfig struct {
	ChainName string `yaml:"chain_name"`
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

// ============ pos-runtime.yaml 结构 ============

type posRuntimeNacosConfig struct {
	PosReward          posRewardConfig         `yaml:"pos_reward"`
	PosStarLevelRule   []posStarLevelRuleItem  `yaml:"pos_star_level_rule"`
	PosStarWhitelist   []posStarWhitelistItem  `yaml:"pos_star_whitelist"`
	StakeReward        stakeRewardConfig       `yaml:"stake_reward"`
	StakeLeaderReward  stakeLeaderRewardConfig `yaml:"stake_leader_reward"`
	StakeFixRate       []stakeFixRateItem      `yaml:"stake_fix_rate"`
	StakeInviteDist    stakeInviteDistConfig   `yaml:"stake_invite_dist"`
	StakeInviteRate    []stakeInviteRateItem   `yaml:"stake_invite_rate"`
	StakeStarLevelRule []stakeStarLevelItem    `yaml:"stake_star_level_rule"`
	StakeStarWhitelist []stakeStarWLItem       `yaml:"stake_star_whitelist"`
	StakeAmm           stakeAmmConfig          `yaml:"stake_amm"`
	StakeTotalLeaders  []stakeTotalLeaderItem  `yaml:"stake_total_leaders"`
	StakeTokenPool     []stakeTokenPoolItem    `yaml:"stake_token_pool"`
}

type posRewardConfig struct {
	RewardAccount    string  `yaml:"reward_account"`
	CostAccount      string  `yaml:"cost_account"`
	CostFeeRate      float64 `yaml:"cost_fee_rate"`
	MaxCostFee       float64 `yaml:"max_cost_fee"`
	QuoteTokenAmount float64 `yaml:"quote_token_amount"`
}

type posStarLevelRuleItem struct {
	StarLevel   int32   `yaml:"star_level"`
	Amount      float64 `yaml:"amount"`
	GroupAmount float64 `yaml:"group_amount"`
	Rate        float64 `yaml:"rate"`
}

type posStarWhitelistItem struct {
	NativeAccount string  `yaml:"native_account"`
	StarLevel     int32   `yaml:"star_level"`
	Rate          float64 `yaml:"rate"`
}

type stakeRewardConfig struct {
	ProgramID        string  `yaml:"program_id"`
	RewardAccount    string  `yaml:"reward_account"`
	CostAccount      string  `yaml:"cost_account"`
	QuoteTokenAmount float64 `yaml:"quote_token_amount"`
	CostFeeRate      int32   `yaml:"cost_fee_rate"`
}

type stakeLeaderRewardConfig struct {
	RewardAccount string `yaml:"reward_account"`
}

type stakeFixRateItem struct {
	StakeType      int32   `yaml:"stake_type"`
	RateTier       int32   `yaml:"rate_tier"`
	MinAmount      float64 `yaml:"min_amount"`
	FixRate        float64 `yaml:"fix_rate"`
	IndividualRate float64 `yaml:"individual_rate"`
}

type stakeInviteDistConfig struct {
	DistLevel int32 `yaml:"dist_level"`
}

type stakeInviteRateItem struct {
	DistLevel int32   `yaml:"dist_level"`
	Rate      float64 `yaml:"rate"`
}

type stakeStarLevelItem struct {
	StarLevel   int32   `yaml:"star_level"`
	Amount      float64 `yaml:"amount"`
	GroupAmount float64 `yaml:"group_amount"`
	Rate        float64 `yaml:"rate"`
}

type stakeStarWLItem struct {
	NativeAccount string  `yaml:"native_account"`
	StarLevel     int32   `yaml:"star_level"`
	Rate          float64 `yaml:"rate"`
}

type stakeAmmConfig struct {
	QuoteToken string `yaml:"quote_token"`
	PublicKey  string `yaml:"public_key"`
}

type stakeTotalLeaderItem struct {
	NativeAccount string  `yaml:"native_account"`
	StakeShare    float64 `yaml:"stake_share"`
}

type stakeTokenPoolItem struct {
	SourceAccount    string `yaml:"source_account"`
	FromTokenAccount string `yaml:"from_token_account"`
}

// ============ 初始化 & 加载 ============

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
		host := nacosCfg.Host
		if host == "" {
			return fmt.Errorf("nacos server config is empty")
		}
		port := nacosCfg.Port
		if port == 0 {
			port = 8848
		}
		serverConfigs = append(serverConfigs, constant.ServerConfig{
			IpAddr: host,
			Port:   port,
		})
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
	username := clientCfg.Username
	password := clientCfg.Password
	if username == "" {
		username = nacosCfg.Username
		password = nacosCfg.Password
	}

	configClient, err := clients.NewConfigClient(vo.NacosClientParam{
		ClientConfig: &constant.ClientConfig{
			NamespaceId:         namespace,
			TimeoutMs:           timeout,
			NotLoadCacheAtStart: clientCfg.NotLoadCacheAtStart,
			LogDir:              defaultString(clientCfg.LogDir, "./log/nacos"),
			CacheDir:            defaultString(clientCfg.CacheDir, "./cache/nacos"),
			LogLevel:            defaultString(clientCfg.LogLevel, "info"),
			Username:            username,
			Password:            password,
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
	if err := s.loadPosRuntimeFromNacos(); err != nil {
		log.Warnf("load pos runtime config from nacos failed, keep database configs: %v", err)
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

	posSub := s.posRuntimeSubscribeConfig()
	return s.NacosConfigClient.ListenConfig(vo.ConfigParam{
		DataId: posSub.DataId,
		Group:  posSub.Group,
		OnChange: func(namespace, group, dataId, data string) {
			if err := s.applyPosRuntimeContent(data); err != nil {
				log.Errorf("apply nacos pos runtime config [%s/%s] error: %v", group, dataId, err)
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

func (s *ServiceContext) loadPosRuntimeFromNacos() error {
	sub := s.posRuntimeSubscribeConfig()
	content, err := s.NacosConfigClient.GetConfig(vo.ConfigParam{DataId: sub.DataId, Group: sub.Group})
	if err != nil {
		return err
	}
	if strings.TrimSpace(content) == "" {
		return fmt.Errorf("empty nacos content")
	}
	return s.applyPosRuntimeContent(content)
}

// ============ apply 逻辑 ============

func (s *ServiceContext) applyBaseRuntimeContent(content string) error {
	var cfg baseRuntimeNacosConfig
	if err := yaml.Unmarshal([]byte(content), &cfg); err != nil {
		return err
	}
	if err := utils.JasyptDecode(&cfg, s.configDecryptKey, utils.JasyptDefaultAlgorithm); err != nil {
		return fmt.Errorf("decrypt nacos base runtime config: %w", err)
	}
	if cfg.Chain.ChainName == "" {
		return fmt.Errorf("chain.chain_name is required")
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

	// Build RPC client from endpoints
	if len(cfg.RPCEndpoints) > 0 {
		endpoints := make([]utils.RPCEndpointConfig, 0, len(cfg.RPCEndpoints))
		for _, ep := range cfg.RPCEndpoints {
			w := ep.Weight
			if w == 0 {
				w = 1
			}
			endpoints = append(endpoints, utils.RPCEndpointConfig{
				Provider: ep.Provider,
				Endpoint: ep.Endpoint,
				APIKey:   ep.APIKey,
				Weight:   w,
			})
		}
		pool, err := utils.NewRPCPool(endpoints)
		if err == nil {
			s.RpcClient = pool.First()
			s.RpcURL = pool.FirstURL()
		}
	}

	now := time.Now()
	s.ConfigMu.Lock()
	defer s.ConfigMu.Unlock()
	s.SystemConfig = &model.SystemConfig{Env: cfg.System.Env, CreatedAt: now, UpdatedAt: now}
	s.ChainConfig = &model.ChainConfig{ChainName: cfg.Chain.ChainName, Decimals: cfg.Chain.Decimals, Symbol: cfg.Chain.Symbol, CreatedAt: now, UpdatedAt: now}
	s.TokenConfig = &model.TokenConfig{TokenName: cfg.Token.TokenName, TokenSymbol: cfg.Token.TokenSymbol, Decimals: cfg.Token.Decimals, Mint: cfg.Token.Mint, CreatedAt: now, UpdatedAt: now}
	s.TokenDecimal = math.Pow(10, float64(cfg.Token.Decimals))
	s.FeeTolerance = &model.FeeTolerance{MaxLessRate: cfg.FeeTolerance.MaxLessRate, CreatedAt: now, UpdatedAt: now}
	s.LightHouseAddress = lighthouseAddr
	log.Infof("loaded base runtime config from nacos")
	return nil
}

func (s *ServiceContext) applyPosRuntimeContent(content string) error {
	var cfg posRuntimeNacosConfig
	if err := yaml.Unmarshal([]byte(content), &cfg); err != nil {
		return err
	}
	if err := utils.JasyptDecode(&cfg, s.configDecryptKey, utils.JasyptDefaultAlgorithm); err != nil {
		return fmt.Errorf("decrypt nacos pos runtime config: %w", err)
	}
	if cfg.PosReward.RewardAccount == "" || cfg.PosReward.CostAccount == "" {
		return fmt.Errorf("pos_reward.reward_account and pos_reward.cost_account are required")
	}
	if cfg.StakeReward.RewardAccount == "" || cfg.StakeReward.CostAccount == "" {
		return fmt.Errorf("stake_reward.reward_account and stake_reward.cost_account are required")
	}
	if cfg.StakeLeaderReward.RewardAccount == "" {
		return fmt.Errorf("stake_leader_reward.reward_account is required")
	}

	now := time.Now()

	// pos 配置
	posStarLevelRule := make(map[int32]model.PosStarLevelRule, len(cfg.PosStarLevelRule))
	for _, r := range cfg.PosStarLevelRule {
		posStarLevelRule[r.StarLevel] = model.PosStarLevelRule{
			StarLevel:   r.StarLevel,
			Amount:      r.Amount,
			GroupAmount: r.GroupAmount,
			Rate:        r.Rate,
			CreatedAt:   now,
			UpdatedAt:   now,
		}
	}
	posWhiteListMap := make(map[string]model.PosStarWhitelist, len(cfg.PosStarWhitelist))
	for _, w := range cfg.PosStarWhitelist {
		posWhiteListMap[w.NativeAccount] = model.PosStarWhitelist{
			NativeAccount: w.NativeAccount,
			StarLevel:     w.StarLevel,
			Rate:          w.Rate,
			CreatedAt:     now,
			UpdatedAt:     now,
		}
	}

	// stake 配置
	stakeFixConfig := make(map[int32]map[int32]model.StakeFixRateConfig)
	for _, c := range cfg.StakeFixRate {
		if stakeFixConfig[c.StakeType] == nil {
			stakeFixConfig[c.StakeType] = make(map[int32]model.StakeFixRateConfig)
		}
		stakeFixConfig[c.StakeType][c.RateTier] = model.StakeFixRateConfig{
			StakeType:      c.StakeType,
			RateTier:       c.RateTier,
			MinAmount:      c.MinAmount,
			FixRate:        c.FixRate,
			IndividualRate: c.IndividualRate,
			CreatedAt:      now,
			UpdatedAt:      now,
		}
	}
	stakeInviteRate := make(map[int32]model.StakeInviteRate, len(cfg.StakeInviteRate))
	for _, r := range cfg.StakeInviteRate {
		stakeInviteRate[r.DistLevel] = model.StakeInviteRate{
			DistLevel: r.DistLevel,
			Rate:      r.Rate,
		}
	}
	stakeStarLevelRule := make(map[int32]model.StakeStarLevelRule, len(cfg.StakeStarLevelRule))
	for _, r := range cfg.StakeStarLevelRule {
		stakeStarLevelRule[r.StarLevel] = model.StakeStarLevelRule{
			StarLevel:   r.StarLevel,
			Amount:      r.Amount,
			GroupAmount: r.GroupAmount,
			Rate:        r.Rate,
			CreatedAt:   now,
			UpdatedAt:   now,
		}
	}
	stakeStarWhitelist := make(map[string]model.StakeStarWhitelist, len(cfg.StakeStarWhitelist))
	for _, w := range cfg.StakeStarWhitelist {
		stakeStarWhitelist[w.NativeAccount] = model.StakeStarWhitelist{
			NativeAccount: w.NativeAccount,
			StarLevel:     w.StarLevel,
			Rate:          w.Rate,
			CreatedAt:     now,
			UpdatedAt:     now,
		}
	}
	totalAreaLeaders := make([]model.StakeTotalLeader, 0, len(cfg.StakeTotalLeaders))
	for _, l := range cfg.StakeTotalLeaders {
		totalAreaLeaders = append(totalAreaLeaders, model.StakeTotalLeader{
			NativeAccount: l.NativeAccount,
			StakeShare:    l.StakeShare,
			CreatedAt:     now,
			UpdatedAt:     now,
		})
	}
	stakeTokenPoolMap := make(map[string]model.StakeTokenPool, len(cfg.StakeTokenPool))
	for _, p := range cfg.StakeTokenPool {
		stakeTokenPoolMap[p.FromTokenAccount] = model.StakeTokenPool{
			SourceAccount:    p.SourceAccount,
			FromTokenAccount: p.FromTokenAccount,
		}
	}

	s.ConfigMu.Lock()
	defer s.ConfigMu.Unlock()

	// pos
	s.PosRewardConfig = &model.PosRewardConfig{
		RewardAccount:    cfg.PosReward.RewardAccount,
		CostAccount:      cfg.PosReward.CostAccount,
		CostFeeRate:      cfg.PosReward.CostFeeRate,
		MaxCostFee:       cfg.PosReward.MaxCostFee,
		QuoteTokenAmount: cfg.PosReward.QuoteTokenAmount,
		CreatedAt:        now,
		UpdatedAt:        now,
	}
	s.PosStarLevelRule = posStarLevelRule
	s.PosWhiteListMap = posWhiteListMap

	// stake
	s.StakeRewardConfig = &model.StakeRewardConfig{
		ProgramID:        cfg.StakeReward.ProgramID,
		RewardAccount:    cfg.StakeReward.RewardAccount,
		CostAccount:      cfg.StakeReward.CostAccount,
		QuoteTokenAmount: cfg.StakeReward.QuoteTokenAmount,
		CostFeeRate:      cfg.StakeReward.CostFeeRate,
		CreatedAt:        now,
		UpdatedAt:        now,
	}
	s.LeaderRewardConfig = &model.StakeLeaderRewardConfig{
		RewardAccount: cfg.StakeLeaderReward.RewardAccount,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	s.StakeFixConfig = stakeFixConfig
	s.StakeDistLevel = cfg.StakeInviteDist.DistLevel
	s.StakeInviteRate = stakeInviteRate
	s.StakeStarLevelRule = stakeStarLevelRule
	s.StakeStarWhitelist = stakeStarWhitelist
	s.StakeAmmConfig = &model.StakeAmmConfig{
		QuoteToken: cfg.StakeAmm.QuoteToken,
		PublicKey:  cfg.StakeAmm.PublicKey,
	}
	s.TotalAreaLeaders = totalAreaLeaders
	s.StakeTokenPoolMap = stakeTokenPoolMap

	log.Infof("loaded pos runtime config from nacos")
	return nil
}

// ============ subscribe config helpers ============

func (s *ServiceContext) baseRuntimeSubscribeConfig() config.NacosSubscribeConfig {
	sub := s.Config.Nacos.SubscribeConfigs.BaseRuntime
	if sub.DataId == "" {
		sub.DataId = defaultBaseRuntimeConfigDataID
	}
	if sub.Group == "" {
		sub.Group = defaultNacosGroup
	}
	return sub
}

func (s *ServiceContext) posRuntimeSubscribeConfig() config.NacosSubscribeConfig {
	sub := s.Config.Nacos.SubscribeConfigs.PosRuntime
	if sub.DataId == "" {
		sub.DataId = defaultPosRuntimeDataID
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
