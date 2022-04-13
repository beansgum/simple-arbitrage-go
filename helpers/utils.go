package helpers

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/shopspring/decimal"
)

var (
	WAVAX         = common.HexToAddress("0xB31f66AA3C1e785363F0875A1B74E27b85FD66c7")
	BundleAddress = common.HexToAddress("0x679232E1Dbb868aed553Bd333a95eeb35CC9B15E")
)

func ToDecimal(ivalue interface{}) decimal.Decimal {
	decimals := 18
	value := new(big.Int)
	switch v := ivalue.(type) {
	case string:
		value.SetString(v, 10)
	case *big.Int:
		value = v
	}

	mul := decimal.NewFromFloat(float64(10)).Pow(decimal.NewFromFloat(float64(decimals)))
	num, _ := decimal.NewFromString(value.String())
	result := num.Div(mul)

	return result
}

func ToWei(iamount interface{}) *big.Int {
	decimals := 18
	amount := decimal.NewFromFloat(0)
	switch v := iamount.(type) {
	case string:
		amount, _ = decimal.NewFromString(v)
	case float64:
		amount = decimal.NewFromFloat(v)
	case int64:
		amount = decimal.NewFromFloat(float64(v))
	case decimal.Decimal:
		amount = v
	case *decimal.Decimal:
		amount = *v
	}

	mul := decimal.NewFromFloat(float64(10)).Pow(decimal.NewFromFloat(float64(decimals)))
	result := amount.Mul(mul)

	wei := new(big.Int)
	wei.SetString(result.String(), 10)

	return wei
}

func WrapBigInt(b *big.Int) *big.Int {
	return big.NewInt(0).Set(b)
}
