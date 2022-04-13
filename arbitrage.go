package main

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

func (arb *arbBot) UniswapWeth(_wethAmountToFirstMarket *big.Int, targets []common.Address, payloads [][]byte) (*types.Transaction, error) {
	txAuth, err := arb.txAuth()
	if err != nil {
		return nil, fmt.Errorf("error getting tx auth: %v", err)
	}

	tx, err := arb.bundleExecutor.UniswapWeth(txAuth, _wethAmountToFirstMarket, targets, payloads)
	if err != nil {
		return nil, fmt.Errorf("error sending tx: %v", err)
	}

	return tx, nil
}

func (arb *arbBot) FindPairTokens(pairAddresses []common.Address) ([][]common.Address, error) {
	return arb.bundleExecutor.FindPairTokens(&bind.CallOpts{}, pairAddresses)
}
