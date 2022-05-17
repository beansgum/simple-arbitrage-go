package main

import (
	"context"
	"fmt"

	"github.com/ethereum/go-ethereum/core/types"
)

func (arb *arbBot) blockNotifications() error {
	headers := make(chan *types.Header)
	sub, err := arb.ethClient.SubscribeNewHead(context.Background(), headers)
	if err != nil {
		return err
	}

	go func() {

		for {
			select {
			case err := <-sub.Err():
				fmt.Println(err)
			case <-headers:
				// Update all marjets reserves.
				for _, market := range arb.markets {
					err := market.UpdateReserves(arb.bundleExecutor)
					if err != nil {
						fmt.Println("error updating market reserve:", err)
					}
				}

				// fmt.Println("Updated reserves")

				// Find arb opportunites
				arb.cma.EvaluateMarkets(arb.markets, arb.tokens)

				// fmt.Println("Evaluated markets")
			}
		}
	}()
	return nil
}
