package main

import (
	"fmt"
	"math/big"

	"github.com/c-ollins/simple-arbitrage-go/market"
	"github.com/ethereum/go-ethereum/common"
)

func (arb *arbBot) evaluateMarkets() {

	pricedMarkets := make(map[common.Address][]*market.PricedMarket)
	for _, token := range arb.tokens {
		for _, m := range arb.markets {
			buyPrice := m.GetTokensIn(token, WAVAX, ToWei(0.01))   // get amount of avax 0.01token can get
			sellPrice := m.GetTokensOut(WAVAX, token, ToWei(0.01)) //

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

				if big.NewInt(0).Sub(m2.SellPrice, m1.BuyPrice).Cmp(ToWei(0.000000002)) == 1 {
					fmt.Printf("Token %s m2(%s) buy %s m1(%s) sell %s, profit: %v \n", token, m2.Market.Name(), ToDecimal(m1.BuyPrice), m1.Market.Name(), ToDecimal(m2.SellPrice), ToDecimal(big.NewInt(0).Sub(m2.SellPrice, m1.BuyPrice)))

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

	crossedMarket := arb.getBestCrossedMarket(crossedMarkets)
	if crossedMarket == nil {
		return
	}

	if crossedMarket.BuyPrice.Cmp(ToWei(23.02)) >= 0 {
		fmt.Printf("Rejecting %s, buy price high: %s\n", crossedMarket.Token, ToDecimal(crossedMarket.BuyPrice))
		return
	}

	buyTarget, buyCallData, outputTokens, err := crossedMarket.BuyCallData(big.NewInt(0).Set(crossedMarket.BuyPrice))
	if err != nil {
		fmt.Println("error getting buy calldata:", err)
		return
	}

	sellTarget, sellCallData, outputAvax, err := crossedMarket.SellCallData(big.NewInt(0).Set(outputTokens), BundleAddress)
	if err != nil {
		fmt.Println("error getting buy calldata:", err)
		return
	}
	fmt.Printf("input: %s, output token: %s, output avax: %s  cmp: %d\n", ToDecimal(crossedMarket.BuyPrice), ToDecimal(outputTokens), ToDecimal(outputAvax), outputAvax.Cmp(crossedMarket.BuyPrice))

	if outputAvax.Cmp(crossedMarket.BuyPrice) != 1 {
		fmt.Printf("Would not procced, output(%s) is less than input(%s)\n", ToDecimal(outputAvax), ToDecimal(crossedMarket.BuyPrice))
		return
	}

	profit := big.NewInt(0).Sub(outputAvax, crossedMarket.BuyPrice)
	fmt.Printf("Swap %s/WAVAX = %savax, profit: %s\n\n", crossedMarket.Token, ToDecimal(outputAvax), ToDecimal(profit))
	txAuth, err := arb.txAuth()
	if err != nil {
		fmt.Println("error getting tx auth:", err)
		return
	}

	tx, err := arb.bundleExecutor.UniswapWeth(txAuth, big.NewInt(0).Set(crossedMarket.BuyPrice), []common.Address{buyTarget, sellTarget}, [][]byte{buyCallData, sellCallData})
	if err != nil {
		fmt.Println("error sending tx:", err)
		return
	}

	fmt.Println(tx.Hash())
	// os.Exit(1)

}

func (arb *arbBot) getBestCrossedMarket(crossedMarkets []*market.CrossedMarket) *market.CrossedMarket {

	var bestCrossedMarket *market.CrossedMarket

	for _, crossedMarket := range crossedMarkets {
		testVolumes := []*big.Int{crossedMarket.BuyPrice, ToWei(0.01), ToWei(0.1), ToWei(0.16), ToWei(0.25), ToWei(0.5), ToWei(1.0), ToWei(2.0), ToWei(5.0), ToWei(10.0), ToWei(12.0), ToWei(20.0)}
		buyMarket := crossedMarket.BuyMarket
		sellMarket := crossedMarket.SellMarket

		for _, testVolume := range testVolumes {
			expectedBuyTokens := buyMarket.GetTokensOut(WAVAX, crossedMarket.Token, WrapBigInt(testVolume))

			proceedsFromSellTokens := sellMarket.GetTokensOut(crossedMarket.Token, WAVAX, WrapBigInt(expectedBuyTokens))
			profit := big.NewInt(0).Sub(proceedsFromSellTokens, testVolume)
			// fmt.Printf("Vol: %s Expected buy: %s, Expected output: %s\n", ToDecimal(testVolume), ToDecimal(expectedBuyTokens), ToDecimal(proceedsFromSellTokens))
			// fmt.Printf("%s buy %s from %s and sell %s in %s for %s profit\n", crossedMarket.Token, ToDecimal(testVolume), buyMarket.Name(), ToDecimal(proceedsFromSellTokens), sellMarket.Name(), ToDecimal(profit))
			if (bestCrossedMarket == nil || bestCrossedMarket.Profit().Cmp(profit) == -1) && profit.Cmp(ToWei(0.015)) == 1 {
				fmt.Printf("Found new best %s buy %s sell %s profit %s\n", crossedMarket.Token, ToDecimal(testVolume), ToDecimal(proceedsFromSellTokens), ToDecimal(profit))
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

	return bestCrossedMarket
}

func WrapBigInt(b *big.Int) *big.Int {
	return big.NewInt(0).Set(b)
}
