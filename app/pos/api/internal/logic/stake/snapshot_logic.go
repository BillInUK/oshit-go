package stake

import (
	"context"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/go-redsync/redsync/v4"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
	posrpc "oshit-go/app/pos/api/internal/rpc"
	"oshit-go/app/pos/api/internal/svc"
	"oshit-go/app/pos/api/internal/task"
)

type StakeSnapShotLogic struct {
	prefix            string
	decimals          uint8
	ctx               context.Context
	srvCtx            *svc.ServiceContext
	db                *gorm.DB
	rd                redis.UniversalClient
	rs                redsync.Redsync
	baseClient        *posrpc.BaseClient
	rpcClient         *rpc.Client
	LightHouseAddress solana.PublicKey
}

func NewStakeSnapShotLogic(ctx context.Context, srvCtx *svc.ServiceContext) *StakeSnapShotLogic {
	return &StakeSnapShotLogic{
		prefix:            "Stake快照业务 -",
		ctx:               ctx,
		srvCtx:            srvCtx,
		db:                srvCtx.DB,
		rd:                srvCtx.Redis,
		rs:                srvCtx.RedSync,
		rpcClient:         srvCtx.RpcClient,
		baseClient:        srvCtx.BaseClient,
		decimals:          uint8(srvCtx.TokenConfig.Decimals),
		LightHouseAddress: srvCtx.LightHouseAddress,
	}
}

// GetConfig 获取奖励规则配置
func (l *StakeSnapShotLogic) GetConfig() (interface{}, error) {
	return nil, nil
}

// TakeStakeSnapShot 手动开启Stake快照
func (l *StakeSnapShotLogic) TakeStakeSnapShot() error {
	//if l.srvCtx.SystemConfig.Env == 0 {
	//	return response.FailWithMsg(c, "can not take snap shot on mainnet")
	//}
	taskCtx := &task.TaskContext{
		CoreContext:  l.srvCtx.CoreContext,
		RewardConfig: l.srvCtx.StakeRewardConfig,
	}
	snapShotTask := task.NewStakeSnapShotTask(taskCtx)
	snapShotTask.StartStakeSnapshotManually()
	return nil
}

// ResetStakeSnapShot 重置Stake快照
func (l *StakeSnapShotLogic) ResetStakeSnapShot() error {
	//if l.srvCtx.SystemConfig.Env == 0 {
	//	return response.FailWithMsg(c, "can not take snap shot on mainnet")
	//}

	// 使用 Exec 执行多条 SQL 语句
	err := l.db.Exec(`
		DELETE FROM t_stake_snap_shot; 
		DELETE FROM t_stake_reward;
		DELETE FROM t_stake_reward_claim_record;
	`).Error
	if err != nil {
		return err
	}
	return nil
}
