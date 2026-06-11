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

	"oshit-go/app/base/api/internal/config"
	core_context "oshit-go/app/base/api/internal/context"
	"oshit-go/app/base/api/internal/task"
	"oshit-go/common/pkg/dal/model"
	"oshit-go/common/utils"
)

const (
	defaultNacosGroup                  = "oshit-go"
	defaultBaseRuntimeConfigDataID     = "base-runtime.yaml"
	defaultServiceRegistryConfigDataID = "base-service-registry.yaml"
)

type baseRuntimeNacosConfig struct {
	System            runtimeSystemConfig      `yaml:"system"`
	Chain             runtimeChainConfig       `yaml:"chain"`
	RPCEndpoints      []runtimeRPCEndpoint     `yaml:"rpc_endpoints"`
	MainnetRPC        runtimeMainnetRPCConfig  `yaml:"mainnet_rpc"`
	Token             runtimeTokenConfig       `yaml:"token"`
	FeeTolerance      runtimeFeeTolerance      `yaml:"fee_tolerance"`
	AWS               runtimeAWSConfig         `yaml:"aws"`
	LightHouseAddress string                   `yaml:"lighthouse_address"`
}

type runtimeSystemConfig struct {
	Env int32 `yaml:"env"`
}

type runtimeChainConfig struct {
	ChainName string `yaml:"chain_name"`
	Decimals  int32  `yaml:"decimals"`
	Symbol    string `yaml:"symbol"`
}

type runtimeRPCEndpoint struct {
	Provider    string `yaml:"provider"`
	Endpoint    string `yaml:"endpoint"`
	APIKey      string `yaml:"api_key"`
	WssEndpoint string `yaml:"wss_endpoint"`
	WssAPIKey   string `yaml:"wss_api_key"`
	Weight      int    `yaml:"weight"`
}

type runtimeMainnetRPCConfig struct {
	Provider string `yaml:"provider"`
	Endpoint string `yaml:"endpoint"`
	APIKey   string `yaml:"api_key"`
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

type runtimeAWSConfig struct {
	AccessKeyID     string `yaml:"access_key_id"`
	SecretAccessKey string `yaml:"secret_access_key"`
	Region          string `yaml:"region"`
}

type serviceRegistryNacosConfig struct {
	Services []serviceRegistryItem `yaml:"services"`
}

type serviceRegistryItem struct {
	Service      string           `yaml:"service"`
	SubService   string           `yaml:"sub_service"`
	Enabled      *bool            `yaml:"enabled"`
	Address      string           `yaml:"address"`
	Webhook      string           `yaml:"webhook"`
	MqGroup      string           `yaml:"mq_group"`
	MqTopic      string           `yaml:"mq_topic"`
	HookType     int32            `yaml:"hook_type"`
	TxSource     int32            `yaml:"tx_source"`
	Confirm      bool             `yaml:"confirm"`
	MultiSign    bool             `yaml:"multi_sign"`
	EncryptedKey string           `yaml:"encrypted_key"`
	Scan         scanRegistryItem `yaml:"scan"`
}

type scanRegistryItem struct {
	Enabled          *bool  `yaml:"enabled"`
	NativeAccount    string `yaml:"native_account"`
	PdaAccount       string `yaml:"pda_account"`
	InitialUntilTxID string `yaml:"initial_until_tx_id"`
	InitialSlot      uint64 `yaml:"initial_slot"`
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

func (s *ServiceContext) initNacosRuntimeAndRegistry() ([]task.ScanConfig, error) {
	if s.NacosConfigClient == nil {
		return nil, nil
	}

	if err := s.loadRuntimeConfigFromNacos(); err != nil {
		log.Warnf("load base runtime config from nacos failed, keep database configs: %v", err)
	}

	scanConfigs, err := s.loadServiceRegistryFromNacos()
	if err != nil {
		log.Warnf("load base service registry from nacos failed, keep database service configs: %v", err)
		return nil, nil
	}
	return scanConfigs, nil
}

func (s *ServiceContext) listenNacosConfigs() error {
	if s.NacosConfigClient == nil {
		return nil
	}
	runtimeSub := s.baseRuntimeSubscribeConfig()
	if err := s.NacosConfigClient.ListenConfig(vo.ConfigParam{
		DataId: runtimeSub.DataId,
		Group:  runtimeSub.Group,
		OnChange: func(namespace, group, dataId, data string) {
			if err := s.applyRuntimeConfigContent(data); err != nil {
				log.Errorf("apply nacos runtime config [%s/%s] error: %v", group, dataId, err)
			}
		},
	}); err != nil {
		return err
	}

	registrySub := s.serviceRegistrySubscribeConfig()
	return s.NacosConfigClient.ListenConfig(vo.ConfigParam{
		DataId: registrySub.DataId,
		Group:  registrySub.Group,
		OnChange: func(namespace, group, dataId, data string) {
			scanConfigs, err := s.applyServiceRegistryContent(data)
			if err != nil {
				log.Errorf("apply nacos service registry [%s/%s] error: %v", group, dataId, err)
				return
			}
			if s.TaskMgr != nil {
				if err := s.TaskMgr.ReconcileScanConfigs(scanConfigs); err != nil {
					log.Errorf("reconcile scan configs from nacos error: %v", err)
				}
			}
		},
	})
}

func (s *ServiceContext) loadRuntimeConfigFromNacos() error {
	sub := s.baseRuntimeSubscribeConfig()
	content, err := s.NacosConfigClient.GetConfig(vo.ConfigParam{DataId: sub.DataId, Group: sub.Group})
	if err != nil {
		return err
	}
	if strings.TrimSpace(content) == "" {
		return fmt.Errorf("empty nacos content")
	}
	return s.applyRuntimeConfigContent(content)
}

func (s *ServiceContext) loadServiceRegistryFromNacos() ([]task.ScanConfig, error) {
	sub := s.serviceRegistrySubscribeConfig()
	content, err := s.NacosConfigClient.GetConfig(vo.ConfigParam{DataId: sub.DataId, Group: sub.Group})
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(content) == "" {
		return nil, fmt.Errorf("empty nacos content")
	}
	return s.applyServiceRegistryContent(content)
}

func (s *ServiceContext) applyRuntimeConfigContent(content string) error {
	var cfg baseRuntimeNacosConfig
	if err := yaml.Unmarshal([]byte(content), &cfg); err != nil {
		return err
	}
	if err := utils.JasyptDecode(&cfg, s.serviceKeyDecryptPwd, serviceKeyDecryptAlgo); err != nil {
		return fmt.Errorf("decrypt nacos runtime config: %w", err)
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

	// Build RPC pool from endpoints
	if len(cfg.RPCEndpoints) > 0 {
		endpoints := make([]utils.RPCEndpointConfig, 0, len(cfg.RPCEndpoints))
		for _, ep := range cfg.RPCEndpoints {
			w := ep.Weight
			if w == 0 {
				w = 1
			}
			endpoints = append(endpoints, utils.RPCEndpointConfig{
				Provider:    ep.Provider,
				Endpoint:    ep.Endpoint,
				APIKey:      ep.APIKey,
				WssEndpoint: ep.WssEndpoint,
				WssAPIKey:   ep.WssAPIKey,
				Weight:      w,
			})
		}
		pool, err := utils.NewRPCPool(endpoints)
		if err != nil {
			return fmt.Errorf("create rpc pool: %v", err)
		}
		s.RpcPool = pool
		s.RpcClient = pool.First()
	}

	// Build mainnet RPC client
	if cfg.MainnetRPC.Endpoint != "" {
		mainnetURL := utils.BuildRPCURL(cfg.MainnetRPC.Provider, cfg.MainnetRPC.Endpoint, cfg.MainnetRPC.APIKey)
		s.MainnetRpcClient = rpc.New(mainnetURL)
		s.MainnetRpcURL = mainnetURL
		if cfg.MainnetRPC.Provider == "helius" && cfg.MainnetRPC.APIKey != "" {
			s.HeliusAPIKey = cfg.MainnetRPC.APIKey
		}
	}

	now := time.Now()
	s.ConfigMu.Lock()
	defer s.ConfigMu.Unlock()
	s.SystemConfig = &model.SystemConfig{Env: cfg.System.Env, CreatedAt: now, UpdatedAt: now}
	s.ChainConfig = &model.ChainConfig{ChainName: cfg.Chain.ChainName, Decimals: cfg.Chain.Decimals, Symbol: cfg.Chain.Symbol, CreatedAt: now, UpdatedAt: now}
	s.TokenConfig = &model.TokenConfig{TokenName: cfg.Token.TokenName, TokenSymbol: cfg.Token.TokenSymbol, Decimals: cfg.Token.Decimals, Mint: cfg.Token.Mint, CreatedAt: now, UpdatedAt: now}
	s.TokenDecimal = math.Pow(10, float64(cfg.Token.Decimals))
	s.FeeTolerance = model.FeeTolerance{MaxLessRate: cfg.FeeTolerance.MaxLessRate, CreatedAt: now, UpdatedAt: now}
	s.AwsConfig = &model.AwsConfig{AccessKeyID: cfg.AWS.AccessKeyID, SecretAccessKey: cfg.AWS.SecretAccessKey, Region: cfg.AWS.Region, CreatedAt: now, UpdatedAt: now}
	s.LightHouseAddress = lighthouseAddr
	fmt.Println("========== Nacos Runtime Config Loaded ==========")
	fmt.Printf("  system.env          = %d\n", cfg.System.Env)
	fmt.Printf("  chain.chain_name    = %s\n", cfg.Chain.ChainName)
	fmt.Printf("  chain.decimals      = %d\n", cfg.Chain.Decimals)
	fmt.Printf("  chain.symbol        = %s\n", cfg.Chain.Symbol)
	fmt.Printf("  rpc_endpoints       = %d configured\n", len(cfg.RPCEndpoints))
	for i, ep := range cfg.RPCEndpoints {
		fmt.Printf("    [%d] provider=%s endpoint=%s weight=%d\n", i, ep.Provider, ep.Endpoint, ep.Weight)
	}
	if cfg.MainnetRPC.Endpoint != "" {
		fmt.Printf("  mainnet_rpc         = %s (%s)\n", cfg.MainnetRPC.Endpoint, cfg.MainnetRPC.Provider)
	}
	fmt.Printf("  token.name          = %s\n", cfg.Token.TokenName)
	fmt.Printf("  token.symbol        = %s\n", cfg.Token.TokenSymbol)
	fmt.Printf("  token.decimals      = %d\n", cfg.Token.Decimals)
	fmt.Printf("  token.mint          = %s\n", cfg.Token.Mint)
	fmt.Printf("  fee_tolerance.rate  = %f\n", cfg.FeeTolerance.MaxLessRate)
	fmt.Printf("  lighthouse_address  = %s\n", lighthouseAddr.String())
	fmt.Printf("  aws.region          = %s\n", cfg.AWS.Region)
	fmt.Println("==================================================")
	return nil
}

func (s *ServiceContext) applyServiceRegistryContent(content string) ([]task.ScanConfig, error) {
	var cfg serviceRegistryNacosConfig
	if err := yaml.Unmarshal([]byte(content), &cfg); err != nil {
		return nil, err
	}
	if err := utils.JasyptDecode(&cfg, s.serviceKeyDecryptPwd, serviceKeyDecryptAlgo); err != nil {
		return nil, fmt.Errorf("decrypt nacos service registry config: %w", err)
	}
	serviceInfoMap := make(map[string]map[string]model.ServiceInfo)
	serviceKeyMap := make(core_context.ServiceKey)
	scanConfigs := make([]task.ScanConfig, 0, len(cfg.Services))
	now := time.Now()

	for _, item := range cfg.Services {
		if item.Service == "" || item.SubService == "" {
			return nil, fmt.Errorf("service and sub_service are required")
		}
		enabled := true
		if item.Enabled != nil {
			enabled = *item.Enabled
		}
		if serviceInfoMap[item.Service] == nil {
			serviceInfoMap[item.Service] = make(map[string]model.ServiceInfo)
		}
		serviceInfoMap[item.Service][item.SubService] = model.ServiceInfo{
			Service:    item.Service,
			SubService: item.SubService,
			Address:    item.Address,
			Webhook:    item.Webhook,
			MqGroup:    item.MqGroup,
			MqTopic:    item.MqTopic,
			HookType:   item.HookType,
			TxSource:   item.TxSource,
			Confirm:    item.Confirm,
			MultiSign:  item.MultiSign,
			CreatedAt:  now,
			UpdatedAt:  now,
		}
		if strings.TrimSpace(item.EncryptedKey) != "" {
			plainKey, err := utils.JasyptDecrypt(item.EncryptedKey, s.serviceKeyDecryptPwd, serviceKeyDecryptAlgo)
			if err != nil {
				return nil, fmt.Errorf("decrypt service key [%s/%s] error: %v", item.Service, item.SubService, err)
			}
			privateKey, err := solana.PrivateKeyFromBase58(plainKey)
			if err != nil {
				return nil, fmt.Errorf("malformed service key [%s/%s]: %v", item.Service, item.SubService, err)
			}
			if serviceKeyMap[item.Service] == nil {
				serviceKeyMap[item.Service] = make(map[string]solana.PrivateKey)
			}
			serviceKeyMap[item.Service][item.SubService] = privateKey
		}

		if item.Scan.PdaAccount != "" {
			scanEnabled := enabled
			if item.Scan.Enabled != nil {
				scanEnabled = *item.Scan.Enabled
			}
			scanConfigs = append(scanConfigs, task.ScanConfig{
				Service:          item.Service,
				SubService:       item.SubService,
				Enabled:          scanEnabled,
				NativeAccount:    item.Scan.NativeAccount,
				PdaAccount:       item.Scan.PdaAccount,
				InitialUntilTxID: item.Scan.InitialUntilTxID,
				InitialSlot:      item.Scan.InitialSlot,
			})
		}
	}

	s.ConfigMu.Lock()
	s.ServiceInfoMap = serviceInfoMap
	s.ServiceKeyMap = serviceKeyMap
	s.ConfigMu.Unlock()
	fmt.Println("========== Nacos Service Registry Loaded ==========")
	fmt.Printf("  total services: %d\n", len(cfg.Services))
	for i, item := range cfg.Services {
		enabled := true
		if item.Enabled != nil {
			enabled = *item.Enabled
		}
		scanEnabled := false
		if item.Scan.PdaAccount != "" {
			scanEnabled = enabled
			if item.Scan.Enabled != nil {
				scanEnabled = *item.Scan.Enabled
			}
		}
		fmt.Printf("  [%d] service=%s sub_service=%s enabled=%v address=%s hook_type=%d scan_enabled=%v\n",
			i, item.Service, item.SubService, enabled, item.Address, item.HookType, scanEnabled)
	}
	fmt.Println("====================================================")
	return scanConfigs, nil
}

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

func (s *ServiceContext) serviceRegistrySubscribeConfig() config.NacosSubscribeConfig {
	sub := s.Config.Nacos.SubscribeConfigs.ServiceRegistry
	if sub.DataId == "" {
		sub.DataId = defaultServiceRegistryConfigDataID
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
