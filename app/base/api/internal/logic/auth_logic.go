package logic

import (
	"context"
	"fmt"
	"github.com/gagliardetto/solana-go"
	"github.com/go-redis/redis/v8"
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"oshit-go/app/base/api/internal/svc"
	"oshit-go/app/base/api/internal/types"
	"oshit-go/app/base/dal/model"
	"oshit-go/common/utils"
	"time"
)

const (
	SOLRegisterSignMsg               = "I am register %s for token %s with my address %s with nonce %d"
	SOLRegisterWithInviteCodeSignMsg = "I am register %s for token %s with my address %s with nonce %d inviteCode %s"
	SOLLoginSignMsg                  = "I am login %s for token %s with my address %s with nonce %d"
	SOLLoginWithInviteCodeSignMsg    = "I am login %s for token %s with my address %s with nonce %d inviteCode %s"
)

type AuthLogic struct {
	ctx    context.Context
	srvCtx *svc.ServiceContext
	db     *gorm.DB
	rd     *redis.Client
}

func NewAuthLogic(ctx context.Context, srvCtx *svc.ServiceContext) *AuthLogic {
	return &AuthLogic{
		ctx:    ctx,
		srvCtx: srvCtx,
		db:     srvCtx.DB,
		rd:     srvCtx.Redis,
	}
}

func (l *AuthLogic) Login(req *types.LoginReq) (*types.LoginRsp, error) {
	// 校验native account
	nativeAccount, err := solana.PublicKeyFromBase58(req.Account)
	if err != nil {
		return nil, errors.New("malformed native account")
	}
	// 校验sign
	sign, err := solana.SignatureFromBase58(req.Sign)
	if err != nil {
		return nil, errors.New("malformed signature")
	}
	// 拼接签名的消息
	msg := ""
	if req.InviteCode == "" {
		msg = fmt.Sprintf(SOLLoginSignMsg, req.Brand, req.Symbol, req.Account, req.Nonce)
	} else {
		msg = fmt.Sprintf(SOLLoginWithInviteCodeSignMsg, req.Brand, req.Symbol, req.Account, req.Nonce, req.InviteCode)
	}
	// 验证签名
	result := nativeAccount.Verify([]byte(msg), sign)
	if !result {
		return nil, errors.New("verify sign failed")
	}
	// 查询地址信息
	record, err := l.QueryNativeAccountInfo(req.Brand, req.Symbol, req.Account)
	if err != nil {
		return nil, errors.New("query native account info error")
	}
	// 如果地址信息为空，则注册
	if record == nil {
		record, err = l.RegisterNativeAccount(req.Brand, req.Symbol, req.Account, req.InviteCode)
		if err != nil {
			return nil, errors.New("register address failed")
		}
	}
	// 颁发jwt token
	tokens, err := utils.GenerateNewTokens(record.RecordID, map[string]interface{}{
		"brand":   req.Brand,
		"symbol":  req.Symbol,
		"account": req.Account,
		"nonce":   req.Nonce,
	})
	if err != nil {
		return nil, errors.New("generate new tokens error")
	}
	// 返回response
	return &types.LoginRsp{
		Token: *tokens,
	}, nil
}

// QueryNativeAccountInfo 查询原生账户信息
func (l *AuthLogic) QueryNativeAccountInfo(brand, symbol, nativeAccount string) (*model.NativeAccountInfo, error) {
	var record model.NativeAccountInfo

	if err := l.db.Model(&record).Where("native_account = ?", nativeAccount).First(&record).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return &record, nil
}

// QueryNativeAccountInfoByInviteCode 根据邀请码查询原生账户信息
func (l *AuthLogic) QueryNativeAccountInfoByInviteCode(inviteCode string) (*model.NativeAccountInfo, error) {
	var record model.NativeAccountInfo

	if err := l.db.Model(&record).Where("invite_code = ?", inviteCode).First(&record).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return &record, nil
}

// RegisterNativeAccount 注册原生账户
func (l *AuthLogic) RegisterNativeAccount(brand, symbol, nativeAccount, inviteCode string) (*model.NativeAccountInfo, error) {
	// 找到token配置
	var tokenConfig model.TokenConfig
	if err := l.db.Model(&tokenConfig).Where("name = ? AND symbol = ?", brand, symbol).First(&tokenConfig).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("can not find token config")
		}
		return nil, err
	}

	// 查看地址是否已经注册
	var count int64 = 0
	if err := l.db.Model(&model.NativeAccountInfo{}).Where("native_account = ?", nativeAccount).Count(&count).Error; err != nil {
		return nil, err
	}
	if count != 0 {
		return nil, errors.New(fmt.Sprintf("native account [%s] already registered", nativeAccount))
	}

	// 生成token account
	tokenMintAccount, err := solana.PublicKeyFromBase58(tokenConfig.Mint)
	if err != nil {
		return nil, errors.New("invalid token mint account")
	}
	nativeAccountPubKey, err := solana.PublicKeyFromBase58(nativeAccount)
	if err != nil {
		return nil, errors.New("invalid native account")
	}
	tokenAccount, _, _ := solana.FindAssociatedTokenAddress(nativeAccountPubKey, tokenMintAccount)

	// 生成新的邀请码
	newInviteCode := utils.GenerateRandomString(8)
	nativeAccountInfo := model.NativeAccountInfo{
		NativeAccount: nativeAccount,
		TokenAccount:  tokenAccount.String(),
		InviteCode:    newInviteCode,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	if err := l.db.Create(&nativeAccountInfo).Error; err != nil {
		return nil, err
	}

	// 处理邀请关系
	if inviteCode != "" {
		// 根据邀请码找到邀请人
		var inviterRecord model.NativeAccountInfo
		if err := l.db.Model(&inviterRecord).Where("invite_code = ?", inviteCode).First(&inviterRecord).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, errors.New("can not find inviter by invite code")
			}
			return nil, err
		}

		// 如果已经存在邀请关系，则不再确定邀请关系
		var inviteCount int64 = 0
		if err := l.db.Model(&model.InviteRelation{}).Where("invitee = ?", nativeAccount).Count(&inviteCount).Error; err != nil {
			return nil, err
		}
		if inviteCount != 0 {
			return &nativeAccountInfo, nil
		}

		// 查看邀请人在邀请关系表里面的信息
		var inviterDetermineRecord model.InviteRelation
		if err := l.db.Model(&inviterDetermineRecord).Where("invitee = ?", inviterRecord.NativeAccount).First(&inviterDetermineRecord).Error; err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("find inviter determine record error")
		}

		// 确定邀请关系
		level := 1
		if err == nil {
			level = int(inviterDetermineRecord.Level) + 1
		}

		inviteRelation := model.InviteRelation{
			Inviter:   inviterRecord.NativeAccount,
			Invitee:   nativeAccount,
			Channel:   "InviteCode",
			Level:     int32(level),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		if err := l.db.Create(&inviteRelation).Error; err != nil {
			return nil, err
		}
	}

	return &nativeAccountInfo, nil
}
