package utils

import (
	"math"
	"math/big"
)

// ToAmountRaw 将字面值金额转换为链上最小单位 (uint64)
// 例如: ToAmountRaw(1.5, 9)  => 1_500_000_000  (1.5 SOL → lamports)
//
//	ToAmountRaw(10.5, 6) => 10_500_000     (10.5 USDT → smallest unit)
func ToAmountRaw(amount float64, decimals uint8) uint64 {
	multiplier := math.Pow10(int(decimals))
	// 用 big.Float 避免浮点误差累积
	bf := new(big.Float).SetPrec(128)
	bf.SetFloat64(amount)
	bf.Mul(bf, new(big.Float).SetFloat64(multiplier))
	result, _ := bf.Uint64()
	return result
}

// ToAmountUI 将链上最小单位 (uint64) 转换为字面值金额 (float64)
// 例如: ToAmountUI(1_500_000_000, 9) => 1.5   (lamports → SOL)
//
//	ToAmountUI(10_500_000, 6)      => 10.5  (smallest unit → USDT)
func ToAmountUI(raw uint64, decimals uint8) float64 {
	return float64(raw) / math.Pow10(int(decimals))
}
