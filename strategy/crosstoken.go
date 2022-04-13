package strategy

import (
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/c-ollins/simple-arbitrage-go/helpers"
	"github.com/c-ollins/simple-arbitrage-go/market"
	"github.com/c-ollins/simple-arbitrage-go/traderjoepair"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/shopspring/decimal"
)

type CrossToken struct {
	TokenIn  common.Address
	TokenOut common.Address
	Market   *market.Market

	InputSize  *big.Int
	OutputSize *big.Int
}

type CrossTokenArbitrage struct {
	flashBot FlashBot
}

func NewCrossTokenArbitrage(flashBot FlashBot) *CrossTokenArbitrage {
	return &CrossTokenArbitrage{flashBot: flashBot}
}

func (cta *CrossTokenArbitrage) EvaluateMarkets(markets []*market.Market, tokens []common.Address) error {
	fmt.Println("Evaluating with new function")
	// crossTokenPaths := make([][]*CrossToken, 0)

	testVolumes := []*big.Int{helpers.ToWei(0.01), helpers.ToWei(0.02), helpers.ToWei(0.1), helpers.ToWei(0.16), helpers.ToWei(0.5), helpers.ToWei(1.0), helpers.ToWei(2.0),
		helpers.ToWei(5.0), helpers.ToWei(10.0), helpers.ToWei(12.0), helpers.ToWei(16.0), helpers.ToWei(20.0), helpers.ToWei(30.0), helpers.ToWei(50.0)}
	for _, size := range testVolumes {
		for _, token := range tokens {

			firstMarket, tokenOutput, err := findBestMarket(markets, helpers.WAVAX, token, helpers.WrapBigInt(size))
			if err != nil {
				return err
			}

			firstCrossToken := &CrossToken{
				Market:   firstMarket,
				TokenIn:  helpers.WAVAX,
				TokenOut: token,

				InputSize:  size,
				OutputSize: tokenOutput,
			}

			lastToken := token
			lastOutput := tokenOutput

			crossTokens := make([]*CrossToken, 0)
			crossTokens = append(crossTokens, firstCrossToken)

			hopMarket, hopToken, hopOutput, err := cta.hopToken(markets, tokens, lastToken, lastOutput, size)
			if err != nil {
				// fmt.Println("error hoping token:", err)
				continue
			}

			crossTokens = append(crossTokens, &CrossToken{
				Market:   hopMarket,
				TokenIn:  lastToken,
				TokenOut: hopToken,

				InputSize:  lastOutput,
				OutputSize: hopOutput,
			})

			lastToken = hopToken
			lastOutput = hopOutput

			wavaxMarket, wavaxOutput, err := findBestMarket(markets, lastToken, helpers.WAVAX, lastOutput)
			if err != nil {
				return err
			}

			crossTokens = append(crossTokens, &CrossToken{
				Market:   wavaxMarket,
				TokenIn:  lastToken,
				TokenOut: helpers.WAVAX,

				InputSize:  lastOutput,
				OutputSize: wavaxOutput,
			})

			profit := big.NewInt(0).Sub(wavaxOutput, size)
			fmt.Printf("WAVAX input: %s, ouput: %s, profit: %s\n", helpers.ToDecimal(size), helpers.ToDecimal(wavaxOutput), helpers.ToDecimal(profit))
			if profit.Cmp(big.NewInt(0)) == 1 {
				// dumpPath(crossTokens)
				fmt.Printf("Path: WAVAX -(%s)> %s -(%s)> %s -(%s)> WAVAX\n", firstMarket.Name(), token, hopMarket.Name(), hopToken, wavaxMarket.Name())
				fmt.Printf("WAVAX input: %s, ouput: %s, profit: %s\n", helpers.ToDecimal(size), helpers.ToDecimal(wavaxOutput), helpers.ToDecimal(profit))
			}

		}
	}

	// os.Exit(1)
	return nil
}

func (cta *CrossTokenArbitrage) hopToken(markets []*market.Market, tokens []common.Address, tokenIn common.Address,
	amountIn, initialSize *big.Int) (*market.Market, common.Address, *big.Int, error) {

	var bestMarket *market.Market
	var bestToken common.Address
	var bestTokenOutput, bestProfit *big.Int

	// test WAVAX
	for _, tokenOut := range tokens {

		market, tokenOutput, err := findBestMarket(markets, tokenIn, tokenOut, helpers.WrapBigInt(amountIn))
		if err != nil {
			continue
		}

		pair := market.FindPair(tokenIn, tokenOut)
		pairIsLow, err := cta.pairBalanceIsLow(markets, pair)
		if err != nil {
			fmt.Println("Error getting pair balance:", err)
			continue
		}

		if pairIsLow {
			continue
		}

		_, wavaxOutput, err := findBestMarket(markets, tokenOut, helpers.WAVAX, tokenOutput)
		if err != nil {
			continue
		}

		profit := big.NewInt(0).Sub(wavaxOutput, initialSize)
		if (profit.Cmp(big.NewInt(0)) == 1) && (bestMarket == nil || profit.Cmp(bestProfit) == 1) {
			bestMarket = market
			bestToken = tokenOut
			bestTokenOutput = tokenOutput
			bestProfit = profit
		}

	}

	if bestMarket == nil {
		return nil, common.Address{}, nil, fmt.Errorf("no pair found for token resulting in profit")
	}

	return bestMarket, bestToken, bestTokenOutput, nil
}

// only for token pairs
func (cta *CrossTokenArbitrage) pairBalanceIsLow(markets []*market.Market, pair *market.Pair) (bool, error) {
	// TODO: check across markets

	_, firstTokenAvaxValue, err := findBestMarket(markets, pair.Token0, helpers.WAVAX, big.NewInt(0).Div(pair.TokenBalances[pair.Token0], big.NewInt(2)))
	if err != nil {
		return true, err
	}

	_, secondTokenAvaxValue, err := findBestMarket(markets, pair.Token1, helpers.WAVAX, big.NewInt(0).Div(pair.TokenBalances[pair.Token1], big.NewInt(2)))
	if err != nil {
		return true, err
	}

	fiftyAvax := helpers.ToWei(10.0)
	if firstTokenAvaxValue.Cmp(fiftyAvax) == -1 || secondTokenAvaxValue.Cmp(fiftyAvax) == -1 {
		// fmt.Printf("Pair %s has low balance; T1: %s, T2: %s\n", pair.Address, helpers.ToDecimal(firstTokenAvaxValue), helpers.ToDecimal(secondTokenAvaxValue))
		return true, nil
	}

	return false, nil
}

func (cta *CrossTokenArbitrage) evaluateMarkets(markets []*market.Market, tokens []common.Address) error {
	// max 1 hop

	crossTokenPaths := make([][]*CrossToken, 0)
	//test uniswap loss
	testVolumes := []*big.Int{helpers.ToWei(0.01), helpers.ToWei(0.02), helpers.ToWei(0.1), helpers.ToWei(0.16), helpers.ToWei(0.5), helpers.ToWei(1.0), helpers.ToWei(2.0),
		helpers.ToWei(5.0), helpers.ToWei(10.0), helpers.ToWei(12.0), helpers.ToWei(16.0), helpers.ToWei(20.0), helpers.ToWei(30.0), helpers.ToWei(50.0)}
	for _, size := range testVolumes {
		for _, token := range tokens {

			firstMarket, tokenOutput, err := findBestMarket(markets, helpers.WAVAX, token, helpers.WrapBigInt(size))
			if err != nil {
				return nil
			}

			firstCrossToken := &CrossToken{
				Market:   firstMarket,
				TokenIn:  helpers.WAVAX,
				TokenOut: token,
			}

			// efficiently hop

			for _, token2 := range tokens {
				secondMarket, token2Output, err := findBestMarket(markets, token, token2, helpers.WrapBigInt(tokenOutput))
				if err != nil {
					continue
				}

				pair := secondMarket.FindPair(token, token2)
				if helpers.ToDecimal(pair.TokenBalances[pair.Token0]).Cmp(decimal.NewFromInt(250)) != 1 || helpers.ToDecimal(pair.TokenBalances[pair.Token1]).Cmp(decimal.NewFromInt(500)) != 1 {
					continue
				}

				_, wavaxOutput, err := findBestMarket(markets, token2, helpers.WAVAX, helpers.WrapBigInt(token2Output))
				if err != nil {
					continue
				}

				profit := big.NewInt(0).Sub(wavaxOutput, size)

				if profit.Cmp(big.NewInt(0)) == 1 {
					// fmt.Printf("Path: WAVAX -(%s)> %s -(%s)> %s -(%s)> WAVAX\n", firstMarket.Name(), token, secondMarket.Name(), token2, thirdMarket.Name())
					fmt.Printf("WAVAX input: %s, ouput: %s, profit: %s\n", helpers.ToDecimal(size), helpers.ToDecimal(wavaxOutput), helpers.ToDecimal(profit))

					crossTokens := make([]*CrossToken, 0)
					crossTokens = append(crossTokens, firstCrossToken)
					crossTokens = append(crossTokens, &CrossToken{
						Market:   secondMarket,
						TokenIn:  token,
						TokenOut: token2,
					})

					crossTokens = append(crossTokens, &CrossToken{
						Market:   secondMarket,
						TokenIn:  token2,
						TokenOut: helpers.WAVAX,
					})

					crossTokenPaths = append(crossTokenPaths, crossTokens)
				}
			}
		}
	}

	fmt.Printf("Found %d cross tokens\n", len(crossTokenPaths))

	bestPath, amountIn, err := cta.GetBestPath(crossTokenPaths)
	if err != nil {
		fmt.Println("error getting paths:", err)
		return err
	} else if bestPath == nil {
		return nil
	}

	payloads := make([][]byte, 0)
	targets := make([]common.Address, 0)

	pair0 := bestPath[0].Market.FindPair(bestPath[0].TokenIn, bestPath[0].TokenOut)
	pair1 := bestPath[1].Market.FindPair(bestPath[1].TokenIn, bestPath[1].TokenOut)
	pair2 := bestPath[2].Market.FindPair(bestPath[2].TokenIn, bestPath[2].TokenOut)

	targets = append(targets, pair0.Address)
	payload0, amountOut0, err := cta.GeneratePayloads(bestPath[0], amountIn, pair1.Address)
	if err != nil {
		fmt.Println("error getting path payload:", err)
		return err
	}

	payloads = append(payloads, payload0)

	targets = append(targets, pair1.Address)
	payload1, amountOut1, err := cta.GeneratePayloads(bestPath[1], amountOut0, pair2.Address)
	if err != nil {
		fmt.Println("error getting path payload:", err)
		return err
	}

	payloads = append(payloads, payload1)

	targets = append(targets, pair2.Address)
	payload3, outputAvax, err := cta.GeneratePayloads(bestPath[2], amountOut1, helpers.BundleAddress)
	if err != nil {
		fmt.Println("error getting path payload:", err)
		return err
	}

	payloads = append(payloads, payload3)

	if outputAvax.Cmp(amountIn) != 1 {
		fmt.Printf("No profit after generating payloads: in: %s, out: %s\n", helpers.ToDecimal(amountIn), helpers.ToDecimal(outputAvax))
		return nil
	}

	fmt.Printf("Profit after generating payloads: in: %s, out: %s, profit: %s\n", helpers.ToDecimal(amountIn), helpers.ToDecimal(outputAvax),
		helpers.ToDecimal(big.NewInt(0).Sub(outputAvax, amountIn)))

	fmt.Println("Sending tx")

	tx, err := cta.flashBot.UniswapWeth(helpers.WrapBigInt(amountIn), targets, payloads)
	if err != nil {
		fmt.Println("error sending tx:", err)
		return err
	}

	fmt.Println(tx.Hash())
	time.Sleep(10 * time.Second)

	return nil
}

func (cta *CrossTokenArbitrage) GeneratePayloads(path *CrossToken, amountIn *big.Int, recipient common.Address) ([]byte, *big.Int, error) {
	outputAmount, err := path.Market.GetTokensOut(path.TokenIn, path.TokenOut, helpers.WrapBigInt(amountIn))
	if err != nil {
		return nil, nil, err
	}

	// outputAmount.Div(outputAmount, big.NewInt(100))
	// outputAmount.Mul(outputAmount, big.NewInt(98)) // expect 98%

	amount0Out := big.NewInt(0)
	amount1Out := big.NewInt(0)

	pair := path.Market.FindPair(path.TokenIn, path.TokenOut)
	fmt.Println("Balance0:", helpers.ToDecimal(pair.TokenBalances[pair.Token0]), helpers.ToDecimal(pair.TokenBalances[pair.Token1]))
	if pair.Token0 == path.TokenOut {
		amount0Out = outputAmount
	} else {
		amount1Out = outputAmount
	}

	contractAbi, _ := abi.JSON(strings.NewReader(traderjoepair.TraderjoepairABI))

	data, err := contractAbi.Pack("swap",
		amount0Out,
		amount1Out,
		recipient, []byte{})
	if err != nil {
		return nil, nil, err
	}

	return data, outputAmount, nil
}

func findBestMarket(markets []*market.Market, tokenIn, tokenOut common.Address, amountIn *big.Int) (*market.Market, *big.Int, error) {
	var bestmarket *market.Market
	var bestAmountOut *big.Int
	for _, market := range markets {

		if !market.HasPair(tokenIn, tokenOut) {
			continue
		}

		amountOut, err := market.GetTokensOut(tokenIn, tokenOut, helpers.WrapBigInt(amountIn))
		if err != nil {
			return nil, nil, err
		}

		if bestmarket == nil || amountOut.Cmp(bestAmountOut) == 1 {
			bestmarket = market
			bestAmountOut = amountOut
		}
	}

	if bestmarket == nil {
		return nil, nil, fmt.Errorf("token pair has no market available")
	}

	return bestmarket, bestAmountOut, nil
}

func (cta *CrossTokenArbitrage) GetBestPath(paths [][]*CrossToken) ([]*CrossToken, *big.Int, error) {
	fmt.Println("Finding best path")

	var amountIn *big.Int
	var bestCrossPath []*CrossToken
	var bestProfit *big.Int

	testVolumes := []*big.Int{helpers.ToWei(0.01), helpers.ToWei(0.02), helpers.ToWei(0.1), helpers.ToWei(0.16), helpers.ToWei(0.5),
		helpers.ToWei(1.0), helpers.ToWei(2.0), helpers.ToWei(5.0), helpers.ToWei(10.0), helpers.ToWei(12.0), helpers.ToWei(20.0)}

	for _, path := range paths {
		for _, testVolume := range testVolumes {

			amountOut1, err := path[0].Market.GetTokensOut(path[0].TokenIn, path[0].TokenOut, helpers.WrapBigInt(testVolume))
			if err != nil {
				return nil, nil, err
			}

			amountOut2, err := path[1].Market.GetTokensOut(path[1].TokenIn, path[1].TokenOut, helpers.WrapBigInt(amountOut1))
			if err != nil {
				return nil, nil, err
			}

			outputAvax, err := path[2].Market.GetTokensOut(path[2].TokenIn, path[2].TokenOut, helpers.WrapBigInt(amountOut2))
			if err != nil {
				return nil, nil, err
			}

			profit := big.NewInt(0).Sub(outputAvax, testVolume)
			if bestProfit != nil && profit.Cmp(bestProfit) == -1 {
				testVolume = big.NewInt(0).Div(amountIn, big.NewInt(2))

				amountOut1, err := path[0].Market.GetTokensOut(path[0].TokenIn, path[0].TokenOut, helpers.WrapBigInt(testVolume))
				if err != nil {
					return nil, nil, err
				}

				amountOut2, err := path[1].Market.GetTokensOut(path[1].TokenIn, path[1].TokenOut, helpers.WrapBigInt(amountOut1))
				if err != nil {
					return nil, nil, err
				}

				_, err = path[2].Market.GetTokensOut(path[2].TokenIn, path[2].TokenOut, helpers.WrapBigInt(amountOut2))
				if err != nil {
					return nil, nil, err
				}

				if profit.Cmp(bestProfit) == 1 {
					fmt.Println("Dividing the volume worked. ", helpers.ToDecimal(amountIn), helpers.ToDecimal(testVolume), helpers.ToDecimal(profit))
					bestCrossPath = path
					bestProfit = profit
					amountIn = helpers.WrapBigInt(testVolume)
					break
				}
			}
			// if path[0].TokenOut == common.HexToAddress("0x6e84a6216eA6dACC71eE8E6b0a5B7322EEbC0fDd") {
			// 	// fmt.Printf("Testing Path: WAVAX -(%s)> %s -(%s)> %s -(%s)> WAVAX, in %s, profit: %s\n", path[0].Market.Name(), path[0].TokenOut, path[1].Market.Name(),
			// 		// path[1].TokenOut, path[2].Market.Name(), helpers.ToDecimal(testVolume), helpers.ToDecimal(profit))
			// }

			if (bestCrossPath == nil || bestProfit.Cmp(profit) == -1) && profit.Cmp(helpers.ToWei(0.03)) == 1 {
				fmt.Printf("Path: WAVAX %s -(%s)> %s -(%s)> %s -(%s)> WAVAX %s, profit: %s\n", helpers.ToDecimal(testVolume), path[0].Market.Name(),
					path[0].TokenOut, path[1].Market.Name(), path[1].TokenOut, path[2].Market.Name(), helpers.ToDecimal(outputAvax), helpers.ToDecimal(profit))

				bestCrossPath = path
				bestProfit = profit
				amountIn = helpers.WrapBigInt(testVolume)
			}
		}
	}

	return bestCrossPath, amountIn, nil
}

func dumpPath(crossTokens []*CrossToken) {
	for i, token := range crossTokens {
		if i == len(crossTokens)-1 {
			fmt.Println(token.TokenOut, helpers.ToDecimal(token.OutputSize))
		} else {
			fmt.Printf("%s %s -(%s)> %s ", token.TokenIn, helpers.ToDecimal(token.InputSize), token.Market.Name(), helpers.ToDecimal(token.OutputSize))
		}

	}
}
