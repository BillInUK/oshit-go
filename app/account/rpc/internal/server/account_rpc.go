package server

import (
	"context"
	"dubbo.apache.org/dubbo-go/v3"
	"dubbo.apache.org/dubbo-go/v3/protocol"
	"dubbo.apache.org/dubbo-go/v3/registry"
	"fmt"
	"github.com/gofiber/fiber/v2/log"
	account "oshit-go/app/account/rpc/pb"
	"oshit-go/app/account/rpc/svc"
	"strings"
)

// AccountRpcService 实现 AccountServiceHandler 接口
type AccountRpcService struct {
	// 这里可以添加你的依赖项，例如数据库连接、配置、日志等
	// db *gorm.DB
	// config *config.Config
}

// NewAccountRpcService 创建并返回一个新的处理器实例
func NewAccountRpcService() *AccountRpcService {
	return &AccountRpcService{
		// 在这里初始化你的依赖项
	}
}

// GetUser 实现获取用户信息的 RPC 方法
func (s *AccountRpcService) GetUser(ctx context.Context, req *account.GetUserRequest) (*account.User, error) {

	log.Infof("获取用户Id: %d", req.GetUserId())

	// 1. 参数验证
	if req.GetUserId() <= 0 {
		return nil, fmt.Errorf("无效的用户ID: %d", req.GetUserId())
	}

	// 2. 业务逻辑
	// 这里应该是你的数据库查询逻辑

	// 3. 返回示例数据
	return &account.User{
		Id:       req.GetUserId(),
		Username: fmt.Sprintf("用户%d", req.GetUserId()),
		Email:    fmt.Sprintf("user%d@example.com", req.GetUserId()),
	}, nil
}

// CreateUser 实现创建用户的 RPC 方法
func (s *AccountRpcService) CreateUser(ctx context.Context, req *account.CreateUserRequest) (*account.User, error) {
	// 1. 参数验证
	if strings.TrimSpace(req.GetUsername()) == "" {
		return nil, fmt.Errorf("用户名不能为空")
	}
	if strings.TrimSpace(req.GetEmail()) == "" {
		return nil, fmt.Errorf("邮箱不能为空")
	}
	if len(strings.TrimSpace(req.GetPassword())) < 6 {
		return nil, fmt.Errorf("密码长度至少为6位")
	}

	// 2. 创建用户逻辑
	// 这里应该是你的数据库创建逻辑

	// 3. 返回创建的用户信息
	return &account.User{
		Id:       1001, // 示例ID，应从数据库获取
		Username: req.GetUsername(),
		Email:    req.GetEmail(),
	}, nil
}

// ValidateToken 实现验证令牌的 RPC 方法
func (s *AccountRpcService) ValidateToken(ctx context.Context, req *account.ValidateTokenRequest) (*account.User, error) {
	// 1. 参数验证
	token := strings.TrimSpace(req.GetToken())
	if token == "" {
		return nil, fmt.Errorf("令牌不能为空")
	}

	// 2. 令牌验证逻辑
	if token != "valid_token_example" {
		return nil, fmt.Errorf("无效的令牌")
	}

	// 3. 返回用户信息
	return &account.User{
		Id:       2001,
		Username: "token_verified_user",
		Email:    "verified@example.com",
	}, nil
}

func StartServer(svcCtx *svc.ServiceContext) {
	nacosAddr := fmt.Sprintf("%s:%d", svcCtx.Config.Nacos.Host, svcCtx.Config.Nacos.Port)
	nacosGrpcPort := svcCtx.Config.Nacos.GrpcPort
	dubboName := svcCtx.Config.Dubbo.Protocol.Name
	dubboPort := svcCtx.Config.Dubbo.Protocol.Port

	ins, err := dubbo.NewInstance(
		dubbo.WithName(dubboName),
		dubbo.WithRegistry(
			registry.WithNacos(),
			registry.WithAddress(nacosAddr),
		),
		dubbo.WithProtocol(
			protocol.WithTriple(),
			protocol.WithPort(dubboPort),
		),
	)
	if err != nil {
		log.Errorf("new dubbo instance failed: %v", err)
		panic(err)
	}
	srv, err := ins.NewServer()
	if err != nil {
		log.Errorf("new server failed: %v", err)
		panic(err)
	}
	if err := account.RegisterAccountServiceHandler(srv, &AccountRpcService{}); err != nil {
		log.Errorf("register greet handler failed: %v", err)
		panic(err)
	}

	if err := srv.Serve(); err != nil {
		log.Errorf("server serve failed: %v", err)
		if strings.Contains(err.Error(), "client not connected") {
			log.Errorf("hint: Nacos client not connected (gRPC). Check %s is reachable and gRPC port %d is open (Nacos 2.x default).", nacosAddr, nacosGrpcPort)
		}
		panic(err)
	}
}
