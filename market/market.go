package market

import (
	"fmt"
	"math/big"
	"strings"

	"github.com/c-ollins/simple-arbitrage-go/traderjoepair"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

type PricedMarket struct {
	Market    *Market
	BuyPrice  *big.Int
	SellPrice *big.Int
}

type CrossedMarket struct {
	Token common.Address // non WAVAX of the pair

	BuyMarket  *Market
	SellMarket *Market

	BuyPrice  *big.Int
	SellPrice *big.Int
}

func (cm *CrossedMarket) TradeProfit(estGasFee *big.Int) *big.Int {
	profit := big.NewInt(0).Sub(cm.SellPrice, cm.BuyPrice)

	return profit.Sub(profit, estGasFee)
}

func (cm *CrossedMarket) Profit() *big.Int {
	profit := big.NewInt(0).Sub(cm.SellPrice, cm.BuyPrice)

	return profit
}

func (cm *CrossedMarket) BuyCallData(amountIn *big.Int) (common.Address, []byte, *big.Int, error) {

	amount0Out := big.NewInt(0)
	amount1Out := big.NewInt(0)
	outputAmount := big.NewInt(0)

	var tokenIn common.Address
	tokenOut := cm.Token
	pair := cm.BuyMarket.liqBalances[cm.Token]
	sellPair := cm.SellMarket.liqBalances[cm.Token]

	if tokenOut == pair.Token0 {
		tokenIn = pair.Token1
		amount0Out = cm.BuyMarket.GetTokensOut(tokenIn, tokenOut, amountIn)
		outputAmount = amount0Out
	} else if tokenOut == pair.Token1 {
		tokenIn = pair.Token0
		amount1Out = cm.BuyMarket.GetTokensOut(tokenIn, tokenOut, amountIn)
		outputAmount = amount1Out
	}

	fmt.Printf("Buy Token in: %s, token out: %s\n", tokenIn, tokenOut)
	fmt.Printf("Amount %s, %s\n", amount0Out, amount1Out)
	contractAbi, _ := abi.JSON(strings.NewReader(traderjoepair.TraderjoepairABI))

	data, err := contractAbi.Pack("swap",
		amount0Out,
		amount1Out,
		sellPair.Address, []byte{})
	if err != nil {
		return pair.Address, nil, nil, err
	}

	// target, calldata
	return pair.Address, data, outputAmount, nil
}

func (cm *CrossedMarket) SellCallData(amountIn *big.Int, recipient common.Address) (common.Address, []byte, *big.Int, error) {

	amount0Out := big.NewInt(0)
	amount1Out := big.NewInt(0)
	outputAmount := big.NewInt(0)

	var tokenIn common.Address
	tokenOut := WAVAX
	pair := cm.SellMarket.liqBalances[cm.Token]

	if tokenOut == pair.Token0 {
		tokenIn = pair.Token1
		amount0Out = cm.SellMarket.GetTokensOut(tokenIn, tokenOut, amountIn)

		// amount0Out.Div(amount0Out, big.NewInt(100))
		// amount0Out.Mul(amount0Out, big.NewInt(100)) // expect 98%

		outputAmount = amount0Out
	} else if tokenOut == pair.Token1 {
		tokenIn = pair.Token0
		amount1Out = cm.SellMarket.GetTokensOut(tokenIn, tokenOut, amountIn)

		// amount1Out.Div(amount1Out, big.NewInt(100))
		// amount1Out.Mul(amount1Out, big.NewInt(100)) // expect 98%

		outputAmount = amount1Out
	}

	contractAbi, _ := abi.JSON(strings.NewReader(traderjoepair.TraderjoepairABI))
	fmt.Printf("sell Token in: %s, token out: %s\n", tokenIn, tokenOut)
	fmt.Printf("Amount %s, %s\n", amount0Out, amount1Out)
	data, err := contractAbi.Pack("swap",
		amount0Out,
		amount1Out,
		recipient, []byte{})
	if err != nil {
		return pair.Address, nil, nil, err
	}

	// target, calldata
	return pair.Address, data, outputAmount, nil
}

type Pair struct {
	Address common.Address
	Token0  common.Address
	Token1  common.Address

	tokenBalances map[common.Address]*big.Int
}

func NewPair(address common.Address) *Pair {
	return &Pair{Address: address, tokenBalances: make(map[common.Address]*big.Int)}
}

func (p *Pair) UpdateAddresses(client *ethclient.Client) error {
	joeLP, err := traderjoepair.NewTraderjoepair(p.Address, client)
	if err != nil {
		return err
	}

	p.Token0, err = joeLP.Token0(&bind.CallOpts{})
	if err != nil {
		return err
	}

	p.Token1, err = joeLP.Token1(&bind.CallOpts{})
	if err != nil {
		return err
	}

	return nil
}
