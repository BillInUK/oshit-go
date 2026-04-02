package task

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/gagliardetto/solana-go"
	computebudget "github.com/gagliardetto/solana-go/programs/compute-budget"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/go-redsync/redsync/v4"
	"github.com/gofiber/fiber/v2/log"
	"github.com/mr-tron/base58"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"io/ioutil"
	"net/http"
	"oshit-go/common/pkg/dal/model"
	"oshit-go/common/pkg/entity"
	"oshit-go/common/utils"
	"time"
)

// 分布式锁 key
const (
	feeStatChainFeeLock  = "base:sol:fee:stat-chain-fee:lock"
	feeUpdatePerUnitLock = "base:sol:fee:update-per-unit:lock"
	feeUpdatePerTxLock   = "base:sol:fee:update-per-tx:lock"
	feeEstWeightAvgLock  = "base:sol:fee:est-weight-avg:lock"
)

// Redis 数据 key
const (
	feePriorityPerUnit  = "base:sol:fee:priority-per-unit"
	feePriorityPerTx    = "base:sol:fee:priority-per-tx"
	feeWeightAvgPerUnit = "base:sol:fee:weight-avg-per-unit"
	feeWeightAvgPerTx   = "base:sol:fee:weight-avg-per-tx"
)

// FeeTask 手续费统计任务
type FeeTask struct {
	db        *gorm.DB
	redis     redis.UniversalClient
	redSync   redsync.Redsync
	rpcClient *rpc.Client
	rpcURL    string
}

// NewFeeTask 创建手续费任务
func NewFeeTask(taskCtx *TaskContext) *FeeTask {
	rpcURL := taskCtx.ChainConfig.RPCURL
	rpcClient := rpc.New(rpcURL)
	return &FeeTask{
		db:        taskCtx.DB,
		redis:     taskCtx.Redis,
		redSync:   taskCtx.RedSync,
		rpcClient: rpcClient,
		rpcURL:    rpcURL,
	}
}

// Start 启动任务
func (t *FeeTask) Start() {
	go runPeriodic(&t.redSync, 5*time.Second, feeStatChainFeeLock, 2*time.Minute, t.readPriorityFee)
	go runPeriodic(&t.redSync, 5*time.Second, feeUpdatePerUnitLock, 2*time.Minute, t.updatePerUnitFee)
	go runPeriodic(&t.redSync, 5*time.Second, feeUpdatePerTxLock, 2*time.Minute, t.updatePerTxFee)
	go runPeriodic(&t.redSync, 5*time.Second, feeEstWeightAvgLock, 2*time.Minute, t.estimateWeightAvgFee)
}

// readPriorityFee 读取优先费用
func (t *FeeTask) readPriorityFee() {
	prefix := "solana 周期性获取链上手续费 -"
	ctx := context.Background()

	// 获取最大的slot防止重复插入
	var maxSlot uint64
	table := t.db.Table(model.TableNameFeeStatistics)
	// 使用 COALESCE 处理 NULL 值，表空时返回 0
	if err := table.Raw(`SELECT COALESCE(MAX(slot), 0) FROM t_fee_statistics`).Scan(&maxSlot).Error; err != nil {
		log.Errorf("%s 获取最大 Slot 出错: %v", prefix, err)
		return
	}

	currentSlot, err := t.rpcClient.GetSlot(ctx, rpc.CommitmentFinalized)
	if err != nil {
		log.Errorf("%s 获取最新的Slot错误: %v", prefix, err)
		return
	}
	//if currentSlot == uint64(maxSlot) {
	//	return
	//}
	//if currentSlot < uint64(maxSlot) {
	//	log.Warnf("%s 区块链最新的slot %d 小于数据库的最新slot %d", prefix, currentSlot, maxSlot)
	//	return
	//}

	// 构建请求的JSON数据
	params := entity.SOLRpcRequestBody{
		JsonRPC: "2.0",
		ID:      1,
		Method:  "getBlock",
		Params: []interface{}{
			currentSlot,
			map[string]interface{}{
				"encoding":                       "jsonParsed",
				"maxSupportedTransactionVersion": 0,
				"transactionDetails":             "full",
				"rewards":                        false,
			},
		},
	}
	// 将请求数据编码为JSON格式
	requestBody, _ := json.Marshal(params)

	// 发送HTTP请求
	resp, err := http.Post(t.rpcURL, "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		log.Errorf("%s，请求RPC错误: %v", prefix, err)
		return
	}
	defer resp.Body.Close()

	// 读取响应数据
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Errorf("%s 读取RPC回复错误: %v", prefix, err)
		return
	}

	// 解析响应的JSON数据
	var blockResponse entity.SOLRpcGetBlockResponse
	err = json.Unmarshal(body, &blockResponse)
	if err != nil {
		log.Errorf("%s 解析RPC回复错误: %v", prefix, err)
		return
	}

	log.Infof("%s 获取最新slot %d，当前区块交易数: %d", prefix, currentSlot, len(blockResponse.Result.Transactions))

	// 初始化SolFeeStatistic数组
	var statistics []model.FeeStatistics

	// 遍历所有交易
	for i, tx := range blockResponse.Result.Transactions {
		var computeUnitPrice float64 = 0
		var computeUnitLimit float64 = 0
		var computeUnitPriceInst *computebudget.SetComputeUnitPrice
		var computeUnitLimitInst *computebudget.SetComputeUnitLimit

		if tx.Meta.Err != nil {
			continue
		}

		// 遍历指令，找到ProgramID为ComputeBudget111111111111111111111111111111的交易数据
		for _, instr := range tx.Transaction.Message.Instructions {
			if instr.ProgramID == "ComputeBudget111111111111111111111111111111" {
				// 假设data字段表示的是ComputeUnitPrice
				accounts := []*solana.AccountMeta{}
				dataBytes, _ := base58.Decode(instr.Data)
				computeBudgetInst, _ := computebudget.DecodeInstruction(accounts, dataBytes)
				if computeBudgetInst == nil {
					continue
				}
				if computeBudgetInst.TypeID.Uint8() == computebudget.Instruction_SetComputeUnitPrice {
					computeUnitPriceInst, _ = computeBudgetInst.Impl.(*computebudget.SetComputeUnitPrice)
					computeUnitPrice = float64(computeUnitPriceInst.MicroLamports)
				}
				if computeBudgetInst.TypeID.Uint8() == computebudget.Instruction_SetComputeUnitLimit {
					computeUnitLimitInst, _ = computeBudgetInst.Impl.(*computebudget.SetComputeUnitLimit)
					computeUnitLimit = float64(computeUnitLimitInst.Units)
				}
				continue
			}
		}

		// 创建SolFeeStatistic实例并填充数据
		if computeUnitPrice != 0 {
			stat := model.FeeStatistics{
				Slot:             int64(currentSlot),
				TransactionIndex: int32(i),
				BlockHash:        blockResponse.Result.Blockhash,
				TransactionID:    tx.Transaction.Signatures[0],
				ComputeUnitPrice: computeUnitPrice,
				ComputeUnitLimit: computeUnitLimit,
				UnitsConsumed:    float64(tx.Meta.ComputeUnitsConsumed),
				Fee:              float64(tx.Meta.Fee),
				CreatedAt:        time.Now(),
				UpdatedAt:        time.Now(),
			}

			// 将统计数据添加到数组中
			statistics = append(statistics, stat)
		}
	}
	log.Infof("%s 获取最新slot %d，获取符合统计要求的交易数: %d", prefix, currentSlot, len(statistics))

	if len(statistics) > 0 {
		// 使用 Gorm 进行查询记录总数
		var recordCount int64
		if err := table.Model(&model.FeeStatistics{}).Count(&recordCount).Error; err != nil {
			log.Errorf("%s 查询记录总数失败: %v", prefix, err)
			return
		}

		// 如果记录数小于10000
		if recordCount <= 10000 {
			if err := table.Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "slot"}, {Name: "transaction_index"}},
				DoUpdates: clause.AssignmentColumns([]string{"compute_unit_price", "compute_unit_limit", "units_consumed", "fee", "updated_at"}),
			}).Create(&statistics).Error; err != nil {
				log.Errorf("%s 写入数据库错误: %v", prefix, err)
				return
			}
		} else {
			// 3. 如果总数超过10000，按 Slot + TransactionIndex 找出最旧的 n 条记录并更新
			var oldRecords []model.FeeStatistics

			// 查询 Slot 和 TransactionIndex 最小的 n 条记录
			table.Order("slot ASC, transaction_index ASC").Limit(len(statistics)).Find(&oldRecords)

			// 创建一个需要更新的 map 列表
			var updates []map[string]interface{}

			for i := range oldRecords {
				// 复制除 RecordId 外的字段
				updates = append(updates, map[string]interface{}{
					"slot":               statistics[i].Slot,
					"transaction_index":  statistics[i].TransactionIndex,
					"block_hash":         statistics[i].BlockHash,
					"transaction_id":     statistics[i].TransactionID,
					"compute_unit_price": statistics[i].ComputeUnitPrice,
					"compute_unit_limit": statistics[i].ComputeUnitLimit,
					"units_consumed":     statistics[i].UnitsConsumed,
					"fee":                statistics[i].Fee,
					"updated_at":         time.Now(),
				})
			}

			// 遍历更新列表并批量更新
			for i, update := range updates {
				if err := t.db.Model(&model.FeeStatistics{}).
					Where("record_id = ?", oldRecords[i].RecordID).
					Updates(update).Error; err != nil {
					log.Errorf("%s 更新数据库错误: %v", prefix, err)
					continue
				}
			}
		}
	}
}

// updatePerUnitFee 周期性获取最近100个区块的手续费，然后算出来5个级别的手续费的加权平均值
func (t *FeeTask) updatePerUnitFee() {
	prefix := "solana 周期性统计每个ComputeUnit所需优先手续费 -"
	ctx := context.Background()
	table := t.db.Table(model.TableNameFeeStatistics)

	var results []entity.SOLWeightAvgFee
	query := `
		WITH RankedData AS (
			SELECT
				*,
				NTILE(4) OVER (ORDER BY compute_unit_price) AS price_group
			FROM public.t_fee_statistics
		)
		SELECT
			price_group,
			ROUND(SUM(compute_unit_price * units_consumed)::NUMERIC / SUM(units_consumed))::NUMERIC AS weighted_avg_price
		FROM RankedData
		GROUP BY price_group
		ORDER BY price_group;
    `
	if err := table.Raw(query).Scan(&results).Error; err != nil {
		log.Errorf("%s 从数据库当中读取统计信息错误: %v", prefix, err)
		return
	}

	// 将查询结果映射到 FeeDetail 结构体
	feeDetail := entity.FeeDetail{}
	for _, result := range results {
		weightedAvgPrice := uint64(result.WeightedAvgPrice)
		switch result.PriceGroup {
		case 1:
			feeDetail.Low = weightedAvgPrice
		case 2:
			feeDetail.Medium = weightedAvgPrice
		case 3:
			feeDetail.High = weightedAvgPrice
		case 4:
			feeDetail.Extreme = weightedAvgPrice
		}
	}

	// 将 results 数组转换为 JSON 字符串
	jsonData, err := json.Marshal(feeDetail)
	if err != nil {
		log.Errorf("%s 序列化结构体错误: %v", prefix, err)
		return
	}
	// 将整个 JSON 数组存入 Redis
	t.redis.Set(ctx, feePriorityPerUnit, jsonData, 0).Err()
}

// updatePerTxFee 周期性获取最近100个区块的手续费，然后算出来5个级别的手续费的加权平均值
func (t *FeeTask) updatePerTxFee() {
	prefix := "solana 周期性统计每个Transaction所需优先手续费 -"
	ctx := context.Background()
	table := t.db.Table(model.TableNameFeeStatistics)

	var results []entity.SOLWeightAvgFee

	// 执行 GORM 查询
	query := `
	WITH FilteredData AS (
		SELECT
			*,
			PERCENT_RANK() OVER (ORDER BY compute_unit_price) AS rank
		FROM
			public.t_fee_statistics
	),
	RankedData AS (
		SELECT
			*,
			CASE
				WHEN rank <= 0.6 THEN 1 -- 最低的60%
				WHEN rank > 0.6 AND rank <= 0.8 THEN 2 -- 接下来的20%
				WHEN rank > 0.8 AND rank <= 0.9 THEN 3 -- 接下来的10%
				ELSE 4 -- 剩余的10%
			END AS price_group
		FROM
			FilteredData
	),
	GroupTransactionCount AS (
		SELECT
			price_group,
			COUNT(*) AS transaction_count
		FROM
			RankedData
		GROUP BY
			price_group
	)
	SELECT
		price_group,
		PERCENTILE_CONT(0.5) WITHIN GROUP (ORDER BY compute_unit_price) AS weighted_avg_price
	FROM
		RankedData
	GROUP BY
		price_group
	ORDER BY
		price_group;
	`

	if err := table.Raw(query).Scan(&results).Error; err != nil {
		log.Errorf("%s 从数据库当中读取统计信息错误: %v", prefix, err)
		return
	}

	// 将查询结果映射到 FeeDetail 结构体
	feeDetail := entity.FeeDetail{}
	for _, result := range results {
		weightedAvgPrice := uint64(result.WeightedAvgPrice)
		switch result.PriceGroup {
		case 1:
			feeDetail.Low = weightedAvgPrice
		case 2:
			feeDetail.Medium = weightedAvgPrice
		case 3:
			feeDetail.High = weightedAvgPrice
		case 4:
			feeDetail.Extreme = weightedAvgPrice
		}
	}

	// 将 results 数组转换为 JSON 字符串
	jsonData, err := json.Marshal(feeDetail)
	if err != nil {
		log.Errorf("%s 序列化结构体错误: %v", prefix, err)
		return
	}
	// 将整个 JSON 数组存入 Redis
	t.redis.Set(ctx, feePriorityPerTx, jsonData, 0).Err()
}

func (t *FeeTask) estimateWeightAvgFee() {
	url := "https://nameless-old-spring.solana-mainnet.quiknode.pro/a30ccaa45b9570bcfce344407013972849ae0484/"
	maxRecordNum := 20

	// 执行核心业务逻辑
	fee, err := utils.GetQnEstimatePriorityFees(url, 100, solana.TokenProgramID)
	if err != nil || fee == nil {
		log.Errorf("手续费获取失败: %v", err)
		return
	}

	// 计算加权平均值（反映Solana局部费用市场特性[1](@ref)）
	ps := fee.Result.PerComputeUnit.Percentiles
	slot := fee.Result.Context.Slot
	lowAvg := (ps.P50 + ps.P55 + ps.P60) / 3 // 基础费用优化策略[1](@ref)
	mediumAvg := (ps.P65 + ps.P70 + ps.P75 + ps.P80 + ps.P85) / 5
	highAvg := (ps.P90 + ps.P95) / 2 // 优先费用高区间统计[2](@ref)

	// 数据库写入（支持动态更新）
	record := model.QnFee{
		ID:        int32(slot%maxRecordNum) + 1,
		Slot:      int64(slot),
		LowAvg:    float64(lowAvg),
		MediumAvg: float64(mediumAvg),
		HighAvg:   float64(highAvg),
		UpdatedAt: time.Now(),
	}
	if err := t.db.Table(model.TableNameQnFee).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "id"}},
			DoUpdates: clause.AssignmentColumns([]string{"slot", "low_avg", "medium_avg", "high_avg", "updated_at"}),
		}).Create(&record).Error; err != nil {
		log.Errorf("数据库写入失败: %v", err)
		return
	}

	// 动态计算平均费用（基于最近20条记录）
	var feeDetail entity.FeeDetail
	if err := t.db.Raw(`
                SELECT
                    CAST(AVG(low_avg) AS BIGINT) AS low,
                    CAST(AVG(medium_avg) AS BIGINT) AS medium,
                    CAST(AVG(high_avg) AS BIGINT) AS high,
                    CAST(AVG(high_avg) AS BIGINT) AS extreme
                FROM (
                    SELECT low_avg, medium_avg, high_avg
                    FROM public.t_qn_fee
                    ORDER BY updated_at DESC
                    LIMIT 20
                ) AS recent_records
            `).Scan(&feeDetail).Error; err != nil {
		log.Errorf("费用统计失败: %v", err)
		return
	}

	// 缓存到Redis（双Key策略）
	if jsonData, err := json.Marshal(feeDetail); err == nil {
		ctx := context.Background()
		_ = t.redis.Set(ctx, feeWeightAvgPerUnit, jsonData, 0)
		_ = t.redis.Set(ctx, feeWeightAvgPerTx, jsonData, 0)
	} else {
		log.Errorf("序列化失败: %v", err)
		return
	}
}
