package main

import (
	"fmt"
	"math/big"
	"os"

	"github.com/c-ollins/simple-arbitrage-go/flashbundle"
	"github.com/c-ollins/simple-arbitrage-go/market"
	"github.com/c-ollins/simple-arbitrage-go/swaprouter"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

var (
	TraderJoeRouter = common.HexToAddress("0x60aE616a2155Ee3d9A68541Ba4544862310933d4")
	PangolinRouter  = common.HexToAddress("0xE54Ca86531e17Ef3616d22Ca28b0D458b6C89106")
	WAVAX           = common.HexToAddress("0xB31f66AA3C1e785363F0875A1B74E27b85FD66c7")

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

	pngRouter      *swaprouter.Swaprouter
	joeRouter      *swaprouter.Swaprouter
	bundleExecutor *flashbundle.Flashbundle

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
	joeRouter, err := swaprouter.NewSwaprouter(TraderJoeRouter, client)
	if err != nil {
		return nil, err
	}

	pngRouter, err := swaprouter.NewSwaprouter(PangolinRouter, client)
	if err != nil {
		return nil, err
	}

	bundleExecutor, err := flashbundle.NewFlashbundle(BundleAddress, client)
	if err != nil {
		return nil, err
	}

	arbbot := &arbBot{
		ethClient:      client,
		pngRouter:      pngRouter,
		joeRouter:      joeRouter,
		bundleExecutor: bundleExecutor,
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

	joePairs := []*market.Pair{
		market.NewPair(common.HexToAddress("0x454e67025631c065d3cfad6d71e6892f74487a15")), // JOE
		market.NewPair(common.HexToAddress("0xc71fa9d143ad905ee73b6edb4cd44df427df1fe7")), // FIRE
		market.NewPair(common.HexToAddress("0x140cac5f0e05cbec857e65353839fddd0d8482c1")), // CRA
		market.NewPair(common.HexToAddress("0xfe15c2695f1f920da45c30aae47d11de51007af9")), // WETH.e
		market.NewPair(common.HexToAddress("0x2774516897ac629ad3ed9dcac7e375dda78412b9")), // QI
		market.NewPair(common.HexToAddress("0x6f3a0c89f611ef5dc9d96650324ac633d02265d3")), // LINK.e
	}

	tj, err := market.NewMarket(joePairs, arb.ethClient)
	if err != nil {
		fmt.Println(err)
		return
	}

	arb.markets = append(arb.markets, tj)

	pngPairs := []*market.Pair{
		market.NewPair(common.HexToAddress("0x134ad631337e8bf7e01ba641fb650070a2e0efa8")), // JOE
		market.NewPair(common.HexToAddress("0x45324950c6ba08112ebf72754004a66a0a2b7721")), // FIRE
		market.NewPair(common.HexToAddress("0x960fa242468746c59bc32513e2e1e1c24fdfaf3f")), // CRA
		market.NewPair(common.HexToAddress("0x7c05d54fc5cb6e4ad87c6f5db3b807c94bb89c52")), // WETH.e
		market.NewPair(common.HexToAddress("0xe530dc2095ef5653205cf5ea79f8979a7028065c")), // QI
		market.NewPair(common.HexToAddress("0x5875c368cddd5fb9bf2f410666ca5aad236dabd4")), // LINK.e
	}

	png, err := market.NewMarket(pngPairs, arb.ethClient)
	if err != nil {
		fmt.Println(err)
		return
	}

	arb.markets = append(arb.markets, png)

	tokens := []common.Address{
		common.HexToAddress("0x6e84a6216ea6dacc71ee8e6b0a5b7322eebc0fdd"), // JOE
		common.HexToAddress("0xfcc6CE74f4cd7eDEF0C5429bB99d38A3608043a5"), // FIRE
		common.HexToAddress("0xA32608e873F9DdEF944B24798db69d80Bbb4d1ed"), // CRA
		common.HexToAddress("0x49d5c2bdffac6ce2bfdb6640f4f80f226bc10bab"), // WETH.e
		common.HexToAddress("0x8729438eb15e2c8b576fcc6aecda6a148776c0f5"), // QI
		common.HexToAddress("0x5947bb275c521040051d82396192181b413227a3"), // LINK.e
	}

	arb.tokens = tokens

	err = arb.blockNotifications()
	if err != nil {
		fmt.Println(err)
		return
	}
}
