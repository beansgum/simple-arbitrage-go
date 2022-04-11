package main

import (
	"fmt"
	"math/big"
	"sort"

	"github.com/c-ollins/simple-arbitrage-go/market"
)

func (arb *arbBot) evaluateMarkets() {

	// temporary
	tj := arb.markets[0]
	png := arb.markets[1]

	crossedMarkets := make([]*market.CrossedMarket, 0)

	for _, token := range arb.tokens {
		tjBuyPrice := tj.GetTokensIn(token, WAVAX, ToWei(int64(10), 18))
		tjSellPrice := tj.GetTokensOut(WAVAX, token, ToWei(int64(10), 18))

		pngBuyPrice := png.GetTokensIn(token, WAVAX, ToWei(int64(10), 18))
		pngSellPrice := png.GetTokensOut(WAVAX, token, ToWei(int64(10), 18))

		if pngSellPrice.Cmp(tjBuyPrice) == 1 {
			cm := &market.CrossedMarket{
				Token:      token,
				BuyMarket:  tj,
				SellMarket: png,
				BuyPrice:   tjBuyPrice,
				SellPrice:  pngSellPrice,
			}

			crossedMarkets = append(crossedMarkets, cm)

			fmt.Printf("Buy %s in joe @ %s and sell in png @ %s for %s profit\n", token, ToDecimal(tjBuyPrice, 18), ToDecimal(pngSellPrice, 18), ToDecimal(cm.TradeProfit(big.NewInt(0)), 18))
		} else if tjSellPrice.Cmp(pngBuyPrice) == 1 {

			cm := &market.CrossedMarket{
				Token:      token,
				BuyMarket:  png,
				SellMarket: tj,
				BuyPrice:   pngBuyPrice,
				SellPrice:  tjSellPrice,
			}

			crossedMarkets = append(crossedMarkets, cm)

			fmt.Printf("Buy %s in png @ %s and sell in joe @ %s for %s profit\n", token, ToDecimal(tjBuyPrice, 18), ToDecimal(pngSellPrice, 18), ToDecimal(cm.TradeProfit(big.NewInt(0)), 18))
		} else {
			// fmt.Printf("No arb possible joe buy price: %s, png sell price: %s\n", ToDecimal(tjBuyPrice, 18), ToDecimal(pngSellPrice, 18))
		}
	}

	fmt.Printf("Found %d crossed markets\n", len(crossedMarkets))

	sort.Slice(crossedMarkets, func(i, j int) bool {
		return crossedMarkets[i].TradeProfit(big.NewInt(0)).Cmp(crossedMarkets[j].TradeProfit(big.NewInt(0))) == -1
	})

}
