package stake

import (
	"context"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/go-redsync/redsync/v4"
	"github.com/gofiber/fiber/v2/log"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
	posrpc "oshit-go/app/pos/api/internal/rpc"
	"oshit-go/app/pos/api/internal/svc"
	"oshit-go/app/pos/api/types"
	"oshit-go/common/pkg/dal/model"
	"time"
)

type StakeRewardLogic struct {
	prefix   string
	decimals uint8
	ctx      context.Context
	srvCtx   *svc.ServiceContext
	db       *gorm.DB
	rd       redis.UniversalClient
	rs       redsync.Redsync

	rpcClient         *rpc.Client
	LightHouseAddress solana.PublicKey
	baseClient        *posrpc.BaseClient

	// 业务相关参数
	distLevel     int32                               // 向上递归奖励的层级
	serviceConfig *model.StakeRewardConfig            // 下发奖励配置
	starWhiteList map[string]model.StakeStarWhitelist // 星级用户白名单
	starLevelRule map[int32]model.StakeStarLevelRule  // 星级评定规则

	// 快照逻辑
	snapShotLogic *StakeSnapShotLogic
}

func NewStakeRewardLogic(ctx context.Context, srvCtx *svc.ServiceContext) *StakeRewardLogic {
	return &StakeRewardLogic{
		prefix:            "Stake奖励业务 -",
		ctx:               ctx,
		srvCtx:            srvCtx,
		db:                srvCtx.DB,
		rd:                srvCtx.Redis,
		rs:                srvCtx.RedSync,
		rpcClient:         srvCtx.RpcClient,
		baseClient:        srvCtx.BaseClient,
		distLevel:         srvCtx.StakeDistLevel,
		starWhiteList:     srvCtx.StakeStarWhitelist,
		starLevelRule:     srvCtx.StakeStarLevelRule,
		decimals:          uint8(srvCtx.TokenConfig.Decimals),
		LightHouseAddress: srvCtx.LightHouseAddress,
		snapShotLogic:     NewStakeSnapShotLogic(ctx, srvCtx),
	}
}

// GetConfig 获取奖励规则配置
func (l *StakeRewardLogic) GetConfig() (*model.StakeRewardConfig, error) {
	return l.serviceConfig, nil
}

// GetClaimRecord 根据交易id获取领取stake奖励记录
func (l *StakeRewardLogic) GetClaimRecord(txId string) (*model.StakeRewardClaimRecord, error) {
	var err error
	var record model.StakeRewardClaimRecord
	table := l.db.Table(model.TableNameStakeRewardClaimRecord)
	if err = table.Where("tx_id = ?", txId).First(&record).Error; err != nil {
		return nil, err
	}
	return &record, nil
}

// GetRewardRecord 根据地址获取未领取的stake奖励
func (l *StakeRewardLogic) GetRewardRecord(nativeAccount string) ([]model.StakeReward, error) {
	var rewards []model.StakeReward

	table := l.db.Table(model.TableNameStakeReward)
	query := table.Where("native_account = ? and state = ? and pending  = ?", nativeAccount, 0, false)

	// 查询数据并获取总数
	if err := query.Order("created_at desc").Find(&rewards).Error; err != nil {
		return nil, err
	}
	// 返回结果和总数
	return rewards, nil
}

func (l *StakeRewardLogic) GetStakeSnapShot(nativeAccount string, snapShotDay time.Time) ([]model.StakeSnapShot, error) {
	var snapShots []model.StakeSnapShot
	if err := l.db.Table(model.TableNameStakeSnapShot).
		Where("native_account = ? and snap_day = ?", nativeAccount, snapShotDay).
		Scan(&snapShots).Error; err != nil {
		log.Errorf("%s 根据日期查询快照失败: %v", l.prefix, err)
		return nil, err
	}
	return snapShots, nil
}

// GetTxInfo 根据交易id获取领取stake奖励记录
func (l *StakeRewardLogic) GetTxInfo(nativeAccount string) (*types.ClaimStakeRewardTxInfo, error) {
	return nil, nil
}
