package market

import (
	"fmt"
	"math/big"
	"strings"
	"sync"

	"github.com/c-ollins/simple-arbitrage-go/flashbundle"
	"github.com/c-ollins/simple-arbitrage-go/helpers"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
)

type FlashBot interface {
	FindPairTokens(pairAddresses []common.Address) ([][]common.Address, error)
}

var WAVAX = helpers.WAVAX

type Market struct {
	liqBalances map[string]*Pair // All Token Pairs available in this market, key=non WAVAX token

	balanceMu     sync.Mutex
	pairs         []*Pair
	name          string
	marketAddress common.Address
}

func NewMarket(name string, marketAddress common.Address, pairs []*Pair, flashbot FlashBot) (*Market, error) {
	tj := &Market{
		name:          name,
		marketAddress: marketAddress,
		liqBalances:   make(map[string]*Pair),
		pairs:         pairs,
	}

	pairAddresses := make([]common.Address, 0)
	for _, p := range pairs {
		pairAddresses = append(pairAddresses, p.Address)
	}

	pairTokens, err := flashbot.FindPairTokens(pairAddresses)
	if err != nil {
		return nil, err
	}

	for i := range pairTokens {
		pairToken := pairTokens[i]
		pairs[i].Token0 = pairToken[0]
		pairs[i].Token1 = pairToken[1]
	}

	return tj, nil
}

func (tj *Market) Name() string {
	return tj.name
}

func (tj *Market) Address() common.Address {
	return tj.marketAddress
}

func (tj *Market) UpdateReserves(bundleExecutor *flashbundle.Flashbundle) error {
	addresses := make([]common.Address, 0)
	for _, pair := range tj.pairs {
		addresses = append(addresses, pair.Address)
	}

	reserves, err := bundleExecutor.GetReservesByPairs(&bind.CallOpts{}, addresses)
	if err != nil {
		return err
	}

	tj.balanceMu.Lock()
	for index, reserve := range reserves {
		pair := tj.pairs[index]
		pair.TokenBalances[pair.Token0] = reserve[0]
		pair.TokenBalances[pair.Token1] = reserve[1]

		pairKey := tokenAddress(pair.Token0, pair.Token1)
		tj.liqBalances[pairKey] = pair
	}
	tj.balanceMu.Unlock()

	return nil
}

func tokenAddress(token1, token2 common.Address) string {
	return strings.ToLower(token1.String()) + strings.ToLower(token2.String())
}

func (tj *Market) FindPair(tokenIn, tokenOut common.Address) *Pair {
	tj.balanceMu.Lock()
	defer tj.balanceMu.Unlock()

	pair, ok := tj.liqBalances[tokenAddress(tokenIn, tokenOut)]
	if !ok {
		pair = tj.liqBalances[tokenAddress(tokenOut, tokenIn)]
	}

	return pair
}

func (tj *Market) HasPair(tokenIn, tokenOut common.Address) bool {
	return tj.FindPair(tokenIn, tokenOut) != nil
}

func (tj *Market) GetTokensIn(tokenIn, tokenOut common.Address, amountOut *big.Int) (*big.Int, error) {
	pair := tj.FindPair(tokenIn, tokenOut)
	if pair == nil {
		return nil, fmt.Errorf("pair not found")
	}

	reserveIn := pair.TokenBalances[tokenIn]
	reserveOut := pair.TokenBalances[tokenOut]

	return tj.getAmountIn(new(big.Int).Set(reserveIn), new(big.Int).Set(reserveOut), amountOut), nil
}

func (tj *Market) GetTokensOut(tokenIn, tokenOut common.Address, amountIn *big.Int) (*big.Int, error) {
	pair := tj.FindPair(tokenIn, tokenOut)
	if pair == nil {
		return nil, fmt.Errorf("pair not found")
	}

	reserveIn := pair.TokenBalances[tokenIn]
	reserveOut := pair.TokenBalances[tokenOut]

	return tj.getAmountOut(new(big.Int).Set(reserveIn), new(big.Int).Set(reserveOut), amountIn), nil
}

func (tj *Market) getAmountIn(reserveIn, reserveOut, amountOut *big.Int) *big.Int {
	numerator := reserveIn.Mul(reserveIn, amountOut)
	numerator = numerator.Mul(numerator, big.NewInt(1000))

	denominator := reserveOut.Sub(reserveOut, amountOut)
	denominator = denominator.Mul(denominator, big.NewInt(997))

	amountIn := numerator.Div(numerator, denominator)
	amountIn = amountIn.Add(amountIn, big.NewInt(1))

	return amountIn
}

func (tj *Market) getAmountOut(reserveIn, reserveOut, amountIn *big.Int) *big.Int {

	amountInWithFee := amountIn.Mul(amountIn, big.NewInt(997))
	numerator := big.NewInt(0).Mul(amountInWithFee, reserveOut)

	denominator := reserveIn.Mul(reserveIn, big.NewInt(1000))
	denominator = denominator.Add(denominator, amountInWithFee)

	return numerator.Div(numerator, denominator)
}
