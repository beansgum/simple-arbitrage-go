package market

import (
	"math/big"

	"github.com/c-ollins/simple-arbitrage-go/traderjoepair"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

type CrossedMarket struct {
	Token common.Address

	BuyMarket  *Market
	SellMarket *Market

	BuyPrice  *big.Int
	SellPrice *big.Int
}

func (cm *CrossedMarket) TradeProfit(estGasFee *big.Int) *big.Int {
	profit := big.NewInt(0).Sub(cm.SellPrice, cm.BuyPrice)

	return profit.Sub(profit, estGasFee)
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
