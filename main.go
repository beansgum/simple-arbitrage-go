package main

import (
	"fmt"
	"math/big"
	"os"

	"github.com/c-ollins/simple-arbitrage-go/factory"
	"github.com/c-ollins/simple-arbitrage-go/flashbundle"
	"github.com/c-ollins/simple-arbitrage-go/market"
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

	WAVAX = common.HexToAddress("0xB31f66AA3C1e785363F0875A1B74E27b85FD66c7")

	BundleAddress = common.HexToAddress("0xbf3Bf019D4Abc6D3B813795e9a3e3FB2A3f4E19e")

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

	joeFactory *factory.Factory
	pngFactory *factory.Factory

	markets []*market.Market
	tokens  []common.Address
}

func newArbBot() (*arbBot, error) {

	fmt.Println("Starting Up")
	client, err := ethclient.Dial("wss://speedy-nodes-nyc.moralis.io//avalanche/mainnet/ws")
	if err != nil {
		return nil, err
	}

	fmt.Println("connected")
	joeFactory, err := factory.NewFactory(TraderJoeFactory, client)
	if err != nil {
		return nil, err
	}

	pngFactory, err := factory.NewFactory(PangolinFactory, client)
	if err != nil {
		return nil, err
	}

	bundleExecutor, err := flashbundle.NewFlashbundle(BundleAddress, client)
	if err != nil {
		return nil, err
	}

	arbbot := &arbBot{
		ethClient:      client,
		bundleExecutor: bundleExecutor,
		joeFactory:     joeFactory,
		pngFactory:     pngFactory,
		markets:        make([]*market.Market, 0),
	}

	return arbbot, nil
}

func (arb *arbBot) txAuth() (*bind.TransactOpts, error) {
	auth, err := bind.NewKeyedTransactorWithChainID(PrivateKey, big.NewInt(43114))
	if err != nil {
		return nil, fmt.Errorf("error creating transactor: %v", err)
	}

	auth.Value = big.NewInt(0)     // in wei
	auth.GasLimit = uint64(200000) // in units

	return auth, nil
}

func (arb *arbBot) beginArbitrage() {

	tokens := []common.Address{
		common.HexToAddress("0xb97ef9ef8734c71904d8002f8b6bc66dd9c48a6e"), // USDC
		common.HexToAddress("0xc7198437980c041c805a1edcba50c1ce5db95118"), // USDT.e
		common.HexToAddress("0xb599c3590f42f8f995ecfa0f85d2980b76862fc1"), // UST
		common.HexToAddress("0xd24c2ad096400b6fbcd2ad8b24e7acbc21a1da64"), // Frax
		common.HexToAddress("0x449674B82F05d498E126Dd6615a1057A9c088f2C"), // LOST
		common.HexToAddress("0x6e84a6216ea6dacc71ee8e6b0a5b7322eebc0fdd"), // JOE
		common.HexToAddress("0xfcc6CE74f4cd7eDEF0C5429bB99d38A3608043a5"), // FIRE
		common.HexToAddress("0xA32608e873F9DdEF944B24798db69d80Bbb4d1ed"), // CRA
		common.HexToAddress("0x49d5c2bdffac6ce2bfdb6640f4f80f226bc10bab"), // WETH.e
		common.HexToAddress("0x8729438eb15e2c8b576fcc6aecda6a148776c0f5"), // QI
		common.HexToAddress("0x5947bb275c521040051d82396192181b413227a3"), // LINK.e
		common.HexToAddress("0x2b2c81e08f1af8835a78bb2a90ae924ace0ea4be"), // sAvax
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
		common.HexToAddress("0x027dbca046ca156de9622cd1e2d907d375e53aa7"), // AMPL
		common.HexToAddress("0xb54f16fb19478766a268f172c9480f8da1a7c9c3"), // TIME
		common.HexToAddress("0xb2a85c5ecea99187a977ac34303b80acbddfa208"), // ROCO
		common.HexToAddress("0x397bbd6a0e41bdf4c3f971731e180db8ad06ebc1"), // AVTX
		common.HexToAddress("0xb27c8941a7df8958a1778c0259f76d1f8b711c35"), // KLO
		common.HexToAddress("0xec3492a2508ddf4fdc0cd76f31f340b30d1793e6"), // CLY
		common.HexToAddress("0x65378b697853568da9ff8eab60c13e1ee9f4a654"), // HUSKY
		common.HexToAddress("0xc38f41a296a4493ff429f1238e030924a1542e50"), // SNOB
		common.HexToAddress("0x340fe1d898eccaad394e2ba0fc1f93d27c7b717a"), // ORBS
	}

	arb.tokens = tokens

	joePairs := make([]*market.Pair, 0)
	pngPairs := make([]*market.Pair, 0)

	fmt.Println("Loading pairs")
	for _, token := range tokens {
		joePair, err := arb.joeFactory.GetPair(&bind.CallOpts{}, WAVAX, token)
		if err != nil {
			fmt.Println("error getting pair:", err)
			return
		}

		pngPair, err := arb.pngFactory.GetPair(&bind.CallOpts{}, WAVAX, token)
		if err != nil {
			fmt.Println("error getting pair:", err)
			return
		}

		joePairs = append(joePairs, market.NewPair(joePair))
		pngPairs = append(pngPairs, market.NewPair(pngPair))
	}

	fmt.Println("Setting up market")
	tj, err := market.NewMarket("JOE", TraderJoeRouter, joePairs, arb.ethClient)
	if err != nil {
		fmt.Println(err)
		return
	}

	arb.markets = append(arb.markets, tj)

	png, err := market.NewMarket("PNG", PangolinRouter, pngPairs, arb.ethClient)
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
