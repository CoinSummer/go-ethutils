package go_ethutils

import (
	"github.com/shopspring/decimal"
	"math"
	"math/big"
	"strconv"
)

func ParserStrAmount(amountStr string, decimalLen int) *big.Int {
	amount, err := strconv.ParseFloat(amountStr, 10)
	if err != nil {
		return nil
	}
	return ParserFloatAmount(amount, decimalLen)
}

func ParserFloatAmount(amount float64, decimalLen int) *big.Int {
	value := decimal.NewFromFloat(math.Pow10(decimalLen)).Mul(decimal.NewFromFloat(amount))
	return value.BigInt()
}

func FormatAmount(amount *big.Int, decimal int) float64 {
	tenDecimal := big.NewFloat(math.Pow(10, float64(decimal)))
	value, _ := new(big.Float).Quo(new(big.Float).SetInt(amount), tenDecimal).Float64()
	return value
}

func BigPercent(value *big.Int, per int64) *big.Int {
	a := big.NewInt(0).Mul(value, big.NewInt(per))
	return big.NewInt(0).Div(a, big.NewInt(10000))
}
