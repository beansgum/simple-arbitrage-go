package main

import (
	"fmt"
	"math/big"
	"os"
	"strings"

	"github.com/c-ollins/simple-arbitrage-go/erc20"
	"github.com/c-ollins/simple-arbitrage-go/flashbundle"
	"github.com/c-ollins/simple-arbitrage-go/helpers"
	"github.com/c-ollins/simple-arbitrage-go/market"
	"github.com/c-ollins/simple-arbitrage-go/strategy"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

var (
	TraderJoeFactory = common.HexToAddress("0x9Ad6C38BE94206cA50bb0d90783181662f0Cfa10")
	PangolinFactory  = common.HexToAddress("0xefa94DE7a4656D787667C749f7E1223D71E9FD88")
	TraderJoeRouter  = common.HexToAddress("0x60aE616a2155Ee3d9A68541Ba4544862310933d4")
	PangolinRouter   = common.HexToAddress("0xE54Ca86531e17Ef3616d22Ca28b0D458b6C89106")

	WAVAX         = helpers.WAVAX
	BundleAddress = helpers.BundleAddress
	ZeroAddress   = common.HexToAddress("0x0000000000000000000000000000000000000000")
	PrivateKey, _ = crypto.HexToECDSA(os.Getenv("MEV_PRIVATE")) // Mev wallet private key
)

func main() {
	arb, err := newArbBot()
	if err != nil {
		fmt.Println(err)
		return
	}

	arb.beginArbitrage()
	select {}

}

type arbBot struct {
	ethClient *ethclient.Client

	bundleExecutor *flashbundle.Flashbundle

	markets []*market.Market
	tokens  []common.Address

	cma *strategy.CrossedMarketArbitrage
}

func newArbBot() (*arbBot, error) {

	fmt.Println("Starting Up")
	client, err := ethclient.Dial("wss://speedy-nodes-nyc.moralis.io//avalanche/mainnet/ws")
	if err != nil {
		return nil, err
	}

	fmt.Println("connected")
	bundleExecutor, err := flashbundle.NewFlashbundle(BundleAddress, client)
	if err != nil {
		return nil, err
	}

	arbbot := &arbBot{
		ethClient:      client,
		bundleExecutor: bundleExecutor,
		markets:        make([]*market.Market, 0),
	}

	arbbot.cma = strategy.NewCrossMarketArbitrage(arbbot)

	return arbbot, nil
}

func (arb *arbBot) txAuth() (*bind.TransactOpts, error) {
	auth, err := bind.NewKeyedTransactorWithChainID(PrivateKey, big.NewInt(43114))
	if err != nil {
		return nil, fmt.Errorf("error creating transactor: %v", err)
	}

	auth.Value = big.NewInt(0)     // in wei
	auth.GasLimit = uint64(500000) // in units

	return auth, nil
}

func (arb *arbBot) beginArbitrage() {

	tokens := []common.Address{
		common.HexToAddress("0xb97ef9ef8734c71904d8002f8b6bc66dd9c48a6e"), // USDC
		common.HexToAddress("0xc7198437980c041c805a1edcba50c1ce5db95118"), // USDT.e
		common.HexToAddress("0xb599c3590f42f8f995ecfa0f85d2980b76862fc1"), // UST
		common.HexToAddress("0x130966628846bfd36ff31a822705796e8cb8c18d"), // MIM
		common.HexToAddress("0x449674B82F05d498E126Dd6615a1057A9c088f2C"), // LOST
		common.HexToAddress("0x6e84a6216ea6dacc71ee8e6b0a5b7322eebc0fdd"), // JOE
		common.HexToAddress("0xfcc6CE74f4cd7eDEF0C5429bB99d38A3608043a5"), // FIRE
		common.HexToAddress("0xA32608e873F9DdEF944B24798db69d80Bbb4d1ed"), // CRA
		common.HexToAddress("0x49d5c2bdffac6ce2bfdb6640f4f80f226bc10bab"), // WETH.e
		common.HexToAddress("0xb09fe1613fe03e7361319d2a43edc17422f36b09"), // BOG
		common.HexToAddress("0x5947bb275c521040051d82396192181b413227a3"), // LINK.e
		common.HexToAddress("0x8f47416cae600bccf9530e9f3aeaa06bdd1caa79"), // THOR
		common.HexToAddress("0x260Bbf5698121EB85e7a74f2E45E16Ce762EbE11"), // axlUST
		common.HexToAddress("0xd1c3f94de7e5b45fa4edbba472491a9f4b166fc4"), // XAVA
		common.HexToAddress("0x50b7545627a5162f82a992c33b87adc75187b218"), // WBTC.e
		common.HexToAddress("0x59414b3089ce2af0010e7523dea7e2b35d776ec7"), // YAK
		common.HexToAddress("0x120ad3e5a7c796349e591f1570d9f7980f4ea9cb"), // LUNA
		common.HexToAddress("0x4f60a160d8c2dddaafe16fcc57566db84d674bd6"), // JEWEL
		common.HexToAddress("0xce1bffbd5374dac86a2893119683f4911a2f7814"), // SPELL
		common.HexToAddress("0xf693248f96fe03422fea95ac0afbbbc4a8fdd172"), // TUS
	}

	arb.tokens = tokens

	fmt.Println("Loading pairs")
	// Find available pairs for all token combinations
	tokenPairs := make([][]common.Address, 0)
	for i, token := range tokens {
		wavaxPair := []common.Address{WAVAX, token}
		tokenPairs = append(tokenPairs, wavaxPair)

		for _, token2 := range tokens[i:] {
			tokenPair := []common.Address{token2, token}
			tokenPairs = append(tokenPairs, tokenPair)
		}
	}

	joePairAddress, err := arb.bundleExecutor.FindPairs(&bind.CallOpts{}, TraderJoeFactory, tokenPairs)
	if err != nil {
		fmt.Println("error getting pairs:", err)
		return
	}

	pngPairAddresses, err := arb.bundleExecutor.FindPairs(&bind.CallOpts{}, PangolinFactory, tokenPairs)
	if err != nil {
		fmt.Println("error getting pairs:", err)
		return
	}

	joePairs := make([]*market.Pair, 0)
	pngPairs := make([]*market.Pair, 0)
	for i := 0; i < len(tokenPairs); i++ {

		joePair := joePairAddress[i]
		pngPair := pngPairAddresses[i]

		if joePair != ZeroAddress {
			joePairs = append(joePairs, market.NewPair(joePair))
		}

		if pngPair != ZeroAddress {
			pngPairs = append(pngPairs, market.NewPair(pngPair))
		}
	}

	fmt.Println("Setting up markets")
	tj, err := market.NewMarket("JOE", TraderJoeRouter, joePairs, arb)
	if err != nil {
		fmt.Println(err)
		return
	}

	arb.markets = append(arb.markets, tj)

	png, err := market.NewMarket("PNG", PangolinRouter, pngPairs, arb)
	if err != nil {
		fmt.Println(err)
		return
	}

	arb.markets = append(arb.markets, png)

	fmt.Println("Ready")

	err = arb.blockNotifications()
	if err != nil {
		fmt.Println(err)
		return
	}
}

func (arb *arbBot) transferFunds(amount *big.Int, address common.Address) error {
	contractAbi, _ := abi.JSON(strings.NewReader(erc20.Erc20ABI))

	data, err := contractAbi.Pack("transfer",
		address,
		amount)
	if err != nil {
		return err
	}

	txAuth, err := arb.txAuth()
	if err != nil {
		return err
	}

	tx, err := arb.bundleExecutor.Call(txAuth, WAVAX, big.NewInt(0), data)
	if err != nil {
		fmt.Println("error sending tx:", err)
		return err
	}

	fmt.Println("ERC20 transfer done:", tx.Hash())
	return nil
}
