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

	WAVAX         = common.HexToAddress("0xB31f66AA3C1e785363F0875A1B74E27b85FD66c7")
	BundleAddress = helpers.BundleAddress
	PrivateKey, _ = crypto.HexToECDSA(os.Getenv("MEV_PRIVATE"))
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
	cta *strategy.CrossTokenArbitrage
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

	arbbot.cta = strategy.NewCrossTokenArbitrage(arbbot)
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
		common.HexToAddress("0x264c1383ea520f73dd837f915ef3a732e204a493"), // BNB
		common.HexToAddress("0x7f041ce89a2079873693207653b24c15b5e6a293"), // LOOT
		common.HexToAddress("0x22d4002028f537599be9f666d1c4fa138522f9c8"), // PTP
		common.HexToAddress("0xb54f16fb19478766a268f172c9480f8da1a7c9c3"), // TIME
		common.HexToAddress("0x321e7092a180bb43555132ec53aaa65a5bf84251"), // gOHM
		common.HexToAddress("0x397bbd6a0e41bdf4c3f971731e180db8ad06ebc1"), // AVTX
		common.HexToAddress("0xb27c8941a7df8958a1778c0259f76d1f8b711c35"), // KLO
		common.HexToAddress("0xec3492a2508ddf4fdc0cd76f31f340b30d1793e6"), // CLY
		common.HexToAddress("0x65378b697853568da9ff8eab60c13e1ee9f4a654"), // HUSKY
		common.HexToAddress("0x490bf3abcab1fb5c88533d850f2a8d6d38298465"), // PLAYMATES
		common.HexToAddress("0x340fe1d898eccaad394e2ba0fc1f93d27c7b717a"), // ORBS
		common.HexToAddress("0x63a72806098Bd3D9520cC43356dD78afe5D386D9"), // AAVE.e
		common.HexToAddress("0x5817d4f0b62a59b17f75207da1848c2ce75e7af4"), // VTX

		common.HexToAddress("0x2f6f07cdcf3588944bf4c42ac74ff24bf56e7590"), // STG
		common.HexToAddress("0xf5ee578505f4d876fef288dfd9fd5e15e9ea1318"), // VOLT
		common.HexToAddress("0xea068fba19ce95f12d252ad8cb2939225c4ea02d"), // FIEF
		common.HexToAddress("0xeb8343d5284caec921f035207ca94db6baaacbcd"), // ECD
		common.HexToAddress("0xfb98b335551a418cd0737375a2ea0ded62ea213b"), // PENDLE
		common.HexToAddress("0x6d923f688c7ff287dc3a5943caeefc994f97b290"), // SMRT
		common.HexToAddress("0xf9a49321d3d34cf94c4abd1957c219572a646692"), // FAVAX
		common.HexToAddress("0x7c08413cbf02202a1c13643db173f2694e0f73f0"), // MAXI
		common.HexToAddress("0x83a283641c6b4df383bcddf807193284c84c5342"), // VPND
		common.HexToAddress("0x70928e5b188def72817b7775f0bf6325968e563b"), // LUNA WORMHOLE
		common.HexToAddress("0x7761e2338b35bceb6bda6ce477ef012bde7ae611"), // EGG
		common.HexToAddress("0x9f285507ea5b4f33822ca7abb5ec8953ce37a645"), // DEG
	}

	arb.tokens = tokens

	fmt.Println("Loading pairs")
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

	zeroAddress := common.HexToAddress("0x0000000000000000000000000000000000000000")

	joePairs := make([]*market.Pair, 0)
	pngPairs := make([]*market.Pair, 0)
	for i := 0; i < len(tokenPairs); i++ {

		joePair := joePairAddress[i]
		pngPair := pngPairAddresses[i]

		if joePair != zeroAddress {
			joePairs = append(joePairs, market.NewPair(joePair))
		}

		if pngPair != zeroAddress {
			pngPairs = append(pngPairs, market.NewPair(pngPair))
		}
	}

	fmt.Println("Setting up market")
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
