package lottery

import (
	"math/rand"

	"github.com/pkg/errors"
)

type Range struct {
	Min    int
	Max    int
	Weight int
}

type ProbabilityConfig struct {
	Ranges []Range
}

// generateAmountByConfig 根据概率配置选取区间，再在区间内均匀随机取值
func generateAmountByConfig(config ProbabilityConfig) int {
	totalWeight := 0
	for _, r := range config.Ranges {
		totalWeight += r.Weight
	}

	pick := rand.Intn(totalWeight)
	cumulative := 0
	for _, r := range config.Ranges {
		cumulative += r.Weight
		if pick < cumulative {
			return r.Min + rand.Intn(r.Max-r.Min+1)
		}
	}
	// 保底返回最后一个区间的最小值
	last := config.Ranges[len(config.Ranges)-1]
	return last.Min
}

// GenerateWeightedLotteryAmount 根据抽奖次数生成带权重的随机金额
func GenerateWeightedLotteryAmount(count int) (int, error) {
	var config ProbabilityConfig

	switch {
	case count >= 5 && count < 10:
		// 前5次：大概率获得500-1000
		config = ProbabilityConfig{
			Ranges: []Range{
				{Min: 1000, Max: 1500, Weight: 9050}, // 90.5% 概率
				{Min: 2001, Max: 3000, Weight: 950},  // 0.95% 概率
			},
		}
	case count >= 10 && count < 20:
		// 6-10次：大概率获得1001-3500
		config = ProbabilityConfig{
			Ranges: []Range{
				{Min: 1000, Max: 1500, Weight: 8000}, // 80% 概率
				{Min: 2001, Max: 3000, Weight: 1905}, // 19.05% 概率
				{Min: 3001, Max: 5000, Weight: 95},   // 0.95% 概率
			},
		}
	case count >= 20:
		// 11-20次：大概率获得3501-5000
		config = ProbabilityConfig{
			Ranges: []Range{
				{Min: 1000, Max: 1500, Weight: 7500}, // 75% 概率
				{Min: 2000, Max: 3000, Weight: 2300}, // 23% 概率
				{Min: 3001, Max: 5000, Weight: 185},  // 1.85% 概率
				{Min: 5001, Max: 10000, Weight: 15},  // 0.15% 概率
			},
		}
	default:
		return 0, errors.New("count must be 5,10,20")
	}

	return generateAmountByConfig(config), nil
}
