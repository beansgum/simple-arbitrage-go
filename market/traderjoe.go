package market

import (
	"math/big"
	"sync"

	"github.com/c-ollins/simple-arbitrage-go/flashbundle"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

var WAVAX = common.HexToAddress("0xB31f66AA3C1e785363F0875A1B74E27b85FD66c7")

type Market struct {
	liqBalances map[common.Address]*Pair // All WAVAX Pairs
	balanceMu   sync.Mutex
	pairs       []*Pair
}

func NewMarket(pairs []*Pair, client *ethclient.Client) (*Market, error) {
	tj := &Market{
		liqBalances: make(map[common.Address]*Pair),
		pairs:       pairs,
	}

	for _, p := range pairs {
		err := p.UpdateAddresses(client)
		if err != nil {
			return nil, err
		}
	}

	return tj, nil
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
		pair.tokenBalances[pair.Token0] = reserve[0]
		pair.tokenBalances[pair.Token1] = reserve[1]

		if pair.Token0 == WAVAX {
			tj.liqBalances[pair.Token1] = pair
		} else {
			tj.liqBalances[pair.Token0] = pair
		}
	}
	tj.balanceMu.Unlock()

	return nil
}

func (tj *Market) GetTokensIn(tokenIn, tokenOut common.Address, amountOut *big.Int) *big.Int {
	tj.balanceMu.Lock()
	pair, ok := tj.liqBalances[tokenIn]
	if !ok {
		pair = tj.liqBalances[tokenOut]
	}
	tj.balanceMu.Unlock()

	reserveIn := pair.tokenBalances[tokenIn]
	reserveOut := pair.tokenBalances[tokenOut]

	return tj.getAmountIn(new(big.Int).Set(reserveIn), new(big.Int).Set(reserveOut), amountOut)
}

func (tj *Market) GetTokensOut(tokenIn, tokenOut common.Address, amountIn *big.Int) *big.Int {
	tj.balanceMu.Lock()
	pair, ok := tj.liqBalances[tokenIn]
	if !ok {
		pair = tj.liqBalances[tokenOut]
	}
	tj.balanceMu.Unlock()

	reserveIn := pair.tokenBalances[tokenIn]
	reserveOut := pair.tokenBalances[tokenOut]

	return tj.getAmountOut(new(big.Int).Set(reserveIn), new(big.Int).Set(reserveOut), amountIn)
}

func (tj *Market) getAmountIn(reserveIn, reserveOut, amountOut *big.Int) *big.Int {
	numerator := reserveIn.Mul(reserveIn, amountOut)
	numerator = numerator.Mul(numerator, big.NewInt(1000))

	denominator := reserveOut.Sub(reserveOut, amountOut)
	denominator = denominator.Mul(denominator, big.NewInt(997))

	amountIn := numerator.Div(numerator, denominator)

	return amountIn.Add(amountIn, big.NewInt(1))
}

func (tj *Market) getAmountOut(reserveIn, reserveOut, amountIn *big.Int) *big.Int {

	amountInWithFee := amountIn.Mul(amountIn, big.NewInt(997))
	numerator := big.NewInt(0).Mul(amountInWithFee, reserveOut)

	denominator := reserveIn.Mul(reserveIn, big.NewInt(1000))
	denominator = denominator.Add(denominator, amountInWithFee)

	return numerator.Div(numerator, denominator)
}
