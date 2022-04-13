package strategy

import (
	"fmt"
	"math/big"
	"time"

	"github.com/c-ollins/simple-arbitrage-go/helpers"
	"github.com/c-ollins/simple-arbitrage-go/market"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

type FlashBot interface {
	UniswapWeth(_wethAmountToFirstMarket *big.Int, targets []common.Address, payloads [][]byte) (*types.Transaction, error)
}

type CrossedMarketArbitrage struct {
	flashBot FlashBot
}

func NewCrossMarketArbitrage(flashBot FlashBot) *CrossedMarketArbitrage {
	return &CrossedMarketArbitrage{flashBot: flashBot}
}

func (cma *CrossedMarketArbitrage) EvaluateMarkets(markets []*market.Market, tokens []common.Address) {
	pricedMarkets := make(map[common.Address][]*market.PricedMarket)
	for _, token := range tokens {
		for _, m := range markets {
			buyPrice, err := m.GetTokensIn(token, helpers.WAVAX, helpers.ToWei(0.01)) // get amount of avax 0.01token can get
			if err != nil {
				fmt.Println("buy token not found in market")
				continue
			}

			sellPrice, err := m.GetTokensOut(helpers.WAVAX, token, helpers.ToWei(0.01)) //
			if err != nil {
				fmt.Println("sell token not found in market")
				continue
			}

			pricedMarket := &market.PricedMarket{
				Market:    m,
				BuyPrice:  buyPrice,
				SellPrice: sellPrice,
			}

			pricedMarkets[token] = append(pricedMarkets[token], pricedMarket)
		}
	}

	crossedMarkets := make([]*market.CrossedMarket, 0)
	for token, tokenMarkets := range pricedMarkets {
		for _, m1 := range tokenMarkets {
			for _, m2 := range tokenMarkets {
				if m1.Market.Address() == m2.Market.Address() {
					continue
				}

				if big.NewInt(0).Sub(m2.SellPrice, m1.BuyPrice).Cmp(helpers.ToWei(0.000000002)) == 1 {
					fmt.Printf("Token %s m2(%s) buy %s m1(%s) sell %s, profit: %v \n", token, m2.Market.Name(),
						helpers.ToDecimal(m1.BuyPrice), m1.Market.Name(), helpers.ToDecimal(m2.SellPrice), helpers.ToDecimal(big.NewInt(0).Sub(m2.SellPrice, m1.BuyPrice)))

					cm := &market.CrossedMarket{
						Token:      token,
						BuyMarket:  m2.Market, // DO NOT TOUCH
						SellMarket: m1.Market,

						BuyPrice:  m1.BuyPrice,
						SellPrice: m2.SellPrice,
					}

					crossedMarkets = append(crossedMarkets, cm)
				}
			}
		}
	}

	fmt.Printf("Found %d crossed markets\n", len(crossedMarkets))

	crossedMarket, err := cma.getBestCrossedMarket(crossedMarkets)
	if err != nil {
		fmt.Println("error getting best crossed market:", err)
		return
	} else if crossedMarket == nil {
		return
	}

	if crossedMarket.BuyPrice.Cmp(helpers.ToWei(23.02)) >= 0 {
		fmt.Printf("Rejecting %s, buy price high: %s\n", crossedMarket.Token, helpers.ToDecimal(crossedMarket.BuyPrice))
		return
	}

	buyTarget, buyCallData, outputTokens, err := crossedMarket.BuyCallData(big.NewInt(0).Set(crossedMarket.BuyPrice))
	if err != nil {
		fmt.Println("error getting buy calldata:", err)
		return
	}

	sellTarget, sellCallData, outputAvax, err := crossedMarket.SellCallData(big.NewInt(0).Set(outputTokens), helpers.BundleAddress)
	if err != nil {
		fmt.Println("error getting buy calldata:", err)
		return
	}
	fmt.Printf("input: %s, output token: %s, output avax: %s  cmp: %d\n", helpers.ToDecimal(crossedMarket.BuyPrice), helpers.ToDecimal(outputTokens), helpers.ToDecimal(outputAvax), outputAvax.Cmp(crossedMarket.BuyPrice))

	if outputAvax.Cmp(crossedMarket.BuyPrice) != 1 {
		fmt.Printf("Would not procced, output(%s) is less than input(%s)\n", helpers.ToDecimal(outputAvax), helpers.ToDecimal(crossedMarket.BuyPrice))
		return
	}

	profit := big.NewInt(0).Sub(outputAvax, crossedMarket.BuyPrice)
	fmt.Printf("Swap %s/WAVAX = %savax, profit: %s\n\n", crossedMarket.Token, helpers.ToDecimal(outputAvax), helpers.ToDecimal(profit))
	tx, err := cma.flashBot.UniswapWeth(big.NewInt(0).Set(crossedMarket.BuyPrice), []common.Address{buyTarget, sellTarget}, [][]byte{buyCallData, sellCallData})
	if err != nil {
		fmt.Println("error sending tx:", err)
		return
	}

	fmt.Println(tx.Hash())
	time.Sleep(10000) // wait for block
	// os.Exit(1)
}

func (cma *CrossedMarketArbitrage) getBestCrossedMarket(crossedMarkets []*market.CrossedMarket) (*market.CrossedMarket, error) {

	var bestCrossedMarket *market.CrossedMarket

	for _, crossedMarket := range crossedMarkets {
		testVolumes := []*big.Int{crossedMarket.BuyPrice, helpers.ToWei(0.01), helpers.ToWei(0.1), helpers.ToWei(0.16),
			helpers.ToWei(0.25), helpers.ToWei(0.5), helpers.ToWei(1.0), helpers.ToWei(2.0), helpers.ToWei(5.0), helpers.ToWei(10.0), helpers.ToWei(12.0), helpers.ToWei(20.0)}
		buyMarket := crossedMarket.BuyMarket
		sellMarket := crossedMarket.SellMarket

		for _, testVolume := range testVolumes {
			expectedBuyTokens, err := buyMarket.GetTokensOut(helpers.WAVAX, crossedMarket.Token, helpers.WrapBigInt(testVolume))
			if err != nil {
				return nil, err
			}

			proceedsFromSellTokens, err := sellMarket.GetTokensOut(crossedMarket.Token, helpers.WAVAX, helpers.WrapBigInt(expectedBuyTokens))
			if err != nil {
				return nil, err
			}

			profit := big.NewInt(0).Sub(proceedsFromSellTokens, testVolume)
			// fmt.Printf("Vol: %s Expected buy: %s, Expected output: %s\n", ToDecimal(testVolume), ToDecimal(expectedBuyTokens), ToDecimal(proceedsFromSellTokens))
			// fmt.Printf("%s buy %s from %s and sell %s in %s for %s profit\n", crossedMarket.Token, ToDecimal(testVolume), buyMarket.Name(), ToDecimal(proceedsFromSellTokens), sellMarket.Name(), ToDecimal(profit))
			if (bestCrossedMarket == nil || bestCrossedMarket.Profit().Cmp(profit) == -1) && profit.Cmp(helpers.ToWei(0.015)) == 1 {
				fmt.Printf("Found new best %s buy %s sell %s profit %s\n", crossedMarket.Token, helpers.ToDecimal(testVolume), helpers.ToDecimal(proceedsFromSellTokens), helpers.ToDecimal(profit))

				bestCrossedMarket = &market.CrossedMarket{
					Token:      crossedMarket.Token,
					BuyMarket:  buyMarket,
					SellMarket: sellMarket,
					BuyPrice:   testVolume,
					SellPrice:  proceedsFromSellTokens,
				}
			}
		}
	}

	return bestCrossedMarket, nil
}
