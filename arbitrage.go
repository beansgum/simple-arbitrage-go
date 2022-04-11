package main

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"os"
	"sort"

	"github.com/c-ollins/simple-arbitrage-go/market"
	"github.com/ethereum/go-ethereum/common"
)

func (arb *arbBot) evaluateMarkets() {

	// temporary
	tj := arb.markets[0]
	png := arb.markets[1]

	crossedMarkets := make([]*market.CrossedMarket, 0)

	for _, token := range arb.tokens {
		tjBuyPrice := tj.GetTokensIn(token, WAVAX, ToWei(0.01))
		tjSellPrice := tj.GetTokensOut(WAVAX, token, ToWei(0.01))

		pngBuyPrice := png.GetTokensIn(token, WAVAX, ToWei(0.01))
		pngSellPrice := png.GetTokensOut(WAVAX, token, ToWei(0.01))

		if pngSellPrice.Cmp(tjBuyPrice) == 1 {
			cm := &market.CrossedMarket{
				Token:      token,
				BuyMarket:  tj,
				SellMarket: png,
				BuyPrice:   tjBuyPrice,
				SellPrice:  pngSellPrice,
			}

			crossedMarkets = append(crossedMarkets, cm)

			fmt.Printf("Buy %s in joe @ %s and sell in png @ %s for %s profit\n", token, ToDecimal(tjBuyPrice), ToDecimal(pngSellPrice), ToDecimal(cm.TradeProfit(big.NewInt(0))))
		} else if tjSellPrice.Cmp(pngBuyPrice) == 1 {

			cm := &market.CrossedMarket{
				Token:      token,
				BuyMarket:  png,
				SellMarket: tj,
				BuyPrice:   pngBuyPrice,
				SellPrice:  tjSellPrice,
			}

			crossedMarkets = append(crossedMarkets, cm)

			fmt.Printf("Buy %s in png @ %s and sell in joe @ %s for %s profit\n", token, ToDecimal(pngBuyPrice), ToDecimal(tjSellPrice), ToDecimal(cm.TradeProfit(big.NewInt(0))))
		} else {
			// fmt.Printf("No arb possible joe buy price: %s, png sell price: %s\n", ToDecimal(tjBuyPrice, 18), ToDecimal(pngSellPrice, 18))
		}
	}

	fmt.Printf("Found %d crossed markets\n", len(crossedMarkets))

	sort.Slice(crossedMarkets, func(i, j int) bool {
		return crossedMarkets[i].TradeProfit(big.NewInt(0)).Cmp(crossedMarkets[j].TradeProfit(big.NewInt(0))) == -1
	})

	for _, crossedMarket := range crossedMarkets {

		if crossedMarket.BuyPrice.Cmp(ToWei(5.02)) >= 0 {
			fmt.Printf("Rejecting %s, buy price high: %s\n", crossedMarket.Token, ToDecimal(crossedMarket.BuyPrice))
			continue
		}

		if crossedMarket.TradeProfit(big.NewInt(0)).Cmp(ToWei(0.02)) == -1 {
			fmt.Printf("Rejecting %s, profit too small\n", crossedMarket.Token)
			continue
		}

		// big.NewInt(0).Set(crossedMarket.BuyPrice)
		buyTarget, buyCallData, outputAmount, err := crossedMarket.BuyCallData(big.NewInt(0).Set(crossedMarket.BuyPrice))
		if err != nil {
			fmt.Println("error getting buy calldata:", err)
			return
		}

		fmt.Println("Swap input:", ToDecimal(crossedMarket.BuyPrice))
		fmt.Printf("Swap WAVAX/%s = %s\n", crossedMarket.Token, ToDecimal(outputAmount))

		sellTarget, sellCallData, outputAvax, err := crossedMarket.SellCallData(big.NewInt(0).Set(outputAmount), BundleAddress)
		if err != nil {
			fmt.Println("error getting buy calldata:", err)
			return
		}

		fmt.Printf("Swap %s/WAVAX = %savax\n\n", crossedMarket.Token, ToDecimal(outputAvax))
		txAuth, err := arb.txAuth()
		if err != nil {
			fmt.Println("error getting tx auth:", err)
			return
		}

		fmt.Println(buyTarget, hex.EncodeToString(buyCallData))
		fmt.Println(sellTarget, hex.EncodeToString(sellCallData))

		// _ = txAuth
		tx, err := arb.bundleExecutor.UniswapLoss(txAuth, big.NewInt(0).Set(crossedMarket.BuyPrice), []common.Address{buyTarget, sellTarget}, [][]byte{buyCallData, sellCallData})
		if err != nil {
			fmt.Println("error sending tx:", err)
			return
		}

		fmt.Println(tx.Hash())
		os.Exit(1)
	}

}
