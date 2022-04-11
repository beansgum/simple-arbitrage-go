// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package flashbundle

import (
	"errors"
	"math/big"
	"strings"

	ethereum "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/event"
)

// Reference imports to suppress errors if they are not otherwise used.
var (
	_ = errors.New
	_ = big.NewInt
	_ = strings.NewReader
	_ = ethereum.NotFound
	_ = bind.Bind
	_ = common.Big1
	_ = types.BloomLookup
	_ = event.NewSubscription
)

// FlashbundleMetaData contains all meta data concerning the Flashbundle contract.
var FlashbundleMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_executor\",\"type\":\"address\"}],\"stateMutability\":\"payable\",\"type\":\"constructor\"},{\"inputs\":[{\"internalType\":\"addresspayable\",\"name\":\"_to\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"_value\",\"type\":\"uint256\"},{\"internalType\":\"bytes\",\"name\":\"_data\",\"type\":\"bytes\"}],\"name\":\"call\",\"outputs\":[{\"internalType\":\"bytes\",\"name\":\"\",\"type\":\"bytes\"}],\"stateMutability\":\"payable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"contractUniswapV2Factory\",\"name\":\"_uniswapFactory\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"_start\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"_stop\",\"type\":\"uint256\"}],\"name\":\"getPairsByIndexRange\",\"outputs\":[{\"internalType\":\"address[3][]\",\"name\":\"\",\"type\":\"address[3][]\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"contractIUniswapV2Pair[]\",\"name\":\"_pairs\",\"type\":\"address[]\"}],\"name\":\"getReservesByPairs\",\"outputs\":[{\"internalType\":\"uint256[3][]\",\"name\":\"\",\"type\":\"uint256[3][]\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"_wethAmountToFirstMarket\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"_ethAmountToCoinbase\",\"type\":\"uint256\"},{\"internalType\":\"address[]\",\"name\":\"_targets\",\"type\":\"address[]\"},{\"internalType\":\"bytes[]\",\"name\":\"_payloads\",\"type\":\"bytes[]\"}],\"name\":\"uniswapWeth\",\"outputs\":[],\"stateMutability\":\"payable\",\"type\":\"function\"},{\"stateMutability\":\"payable\",\"type\":\"receive\"}]",
}

// FlashbundleABI is the input ABI used to generate the binding from.
// Deprecated: Use FlashbundleMetaData.ABI instead.
var FlashbundleABI = FlashbundleMetaData.ABI

// Flashbundle is an auto generated Go binding around an Ethereum contract.
type Flashbundle struct {
	FlashbundleCaller     // Read-only binding to the contract
	FlashbundleTransactor // Write-only binding to the contract
	FlashbundleFilterer   // Log filterer for contract events
}

// FlashbundleCaller is an auto generated read-only Go binding around an Ethereum contract.
type FlashbundleCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// FlashbundleTransactor is an auto generated write-only Go binding around an Ethereum contract.
type FlashbundleTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// FlashbundleFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type FlashbundleFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// FlashbundleSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type FlashbundleSession struct {
	Contract     *Flashbundle      // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// FlashbundleCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type FlashbundleCallerSession struct {
	Contract *FlashbundleCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts      // Call options to use throughout this session
}

// FlashbundleTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type FlashbundleTransactorSession struct {
	Contract     *FlashbundleTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts      // Transaction auth options to use throughout this session
}

// FlashbundleRaw is an auto generated low-level Go binding around an Ethereum contract.
type FlashbundleRaw struct {
	Contract *Flashbundle // Generic contract binding to access the raw methods on
}

// FlashbundleCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type FlashbundleCallerRaw struct {
	Contract *FlashbundleCaller // Generic read-only contract binding to access the raw methods on
}

// FlashbundleTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type FlashbundleTransactorRaw struct {
	Contract *FlashbundleTransactor // Generic write-only contract binding to access the raw methods on
}

// NewFlashbundle creates a new instance of Flashbundle, bound to a specific deployed contract.
func NewFlashbundle(address common.Address, backend bind.ContractBackend) (*Flashbundle, error) {
	contract, err := bindFlashbundle(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &Flashbundle{FlashbundleCaller: FlashbundleCaller{contract: contract}, FlashbundleTransactor: FlashbundleTransactor{contract: contract}, FlashbundleFilterer: FlashbundleFilterer{contract: contract}}, nil
}

// NewFlashbundleCaller creates a new read-only instance of Flashbundle, bound to a specific deployed contract.
func NewFlashbundleCaller(address common.Address, caller bind.ContractCaller) (*FlashbundleCaller, error) {
	contract, err := bindFlashbundle(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &FlashbundleCaller{contract: contract}, nil
}

// NewFlashbundleTransactor creates a new write-only instance of Flashbundle, bound to a specific deployed contract.
func NewFlashbundleTransactor(address common.Address, transactor bind.ContractTransactor) (*FlashbundleTransactor, error) {
	contract, err := bindFlashbundle(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &FlashbundleTransactor{contract: contract}, nil
}

// NewFlashbundleFilterer creates a new log filterer instance of Flashbundle, bound to a specific deployed contract.
func NewFlashbundleFilterer(address common.Address, filterer bind.ContractFilterer) (*FlashbundleFilterer, error) {
	contract, err := bindFlashbundle(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &FlashbundleFilterer{contract: contract}, nil
}

// bindFlashbundle binds a generic wrapper to an already deployed contract.
func bindFlashbundle(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(FlashbundleABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Flashbundle *FlashbundleRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Flashbundle.Contract.FlashbundleCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Flashbundle *FlashbundleRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Flashbundle.Contract.FlashbundleTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Flashbundle *FlashbundleRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Flashbundle.Contract.FlashbundleTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Flashbundle *FlashbundleCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Flashbundle.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Flashbundle *FlashbundleTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Flashbundle.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Flashbundle *FlashbundleTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Flashbundle.Contract.contract.Transact(opts, method, params...)
}

// GetPairsByIndexRange is a free data retrieval call binding the contract method 0xab2217e4.
//
// Solidity: function getPairsByIndexRange(address _uniswapFactory, uint256 _start, uint256 _stop) view returns(address[3][])
func (_Flashbundle *FlashbundleCaller) GetPairsByIndexRange(opts *bind.CallOpts, _uniswapFactory common.Address, _start *big.Int, _stop *big.Int) ([][3]common.Address, error) {
	var out []interface{}
	err := _Flashbundle.contract.Call(opts, &out, "getPairsByIndexRange", _uniswapFactory, _start, _stop)

	if err != nil {
		return *new([][3]common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new([][3]common.Address)).(*[][3]common.Address)

	return out0, err

}

// GetPairsByIndexRange is a free data retrieval call binding the contract method 0xab2217e4.
//
// Solidity: function getPairsByIndexRange(address _uniswapFactory, uint256 _start, uint256 _stop) view returns(address[3][])
func (_Flashbundle *FlashbundleSession) GetPairsByIndexRange(_uniswapFactory common.Address, _start *big.Int, _stop *big.Int) ([][3]common.Address, error) {
	return _Flashbundle.Contract.GetPairsByIndexRange(&_Flashbundle.CallOpts, _uniswapFactory, _start, _stop)
}

// GetPairsByIndexRange is a free data retrieval call binding the contract method 0xab2217e4.
//
// Solidity: function getPairsByIndexRange(address _uniswapFactory, uint256 _start, uint256 _stop) view returns(address[3][])
func (_Flashbundle *FlashbundleCallerSession) GetPairsByIndexRange(_uniswapFactory common.Address, _start *big.Int, _stop *big.Int) ([][3]common.Address, error) {
	return _Flashbundle.Contract.GetPairsByIndexRange(&_Flashbundle.CallOpts, _uniswapFactory, _start, _stop)
}

// GetReservesByPairs is a free data retrieval call binding the contract method 0x4dbf0f39.
//
// Solidity: function getReservesByPairs(address[] _pairs) view returns(uint256[3][])
func (_Flashbundle *FlashbundleCaller) GetReservesByPairs(opts *bind.CallOpts, _pairs []common.Address) ([][3]*big.Int, error) {
	var out []interface{}
	err := _Flashbundle.contract.Call(opts, &out, "getReservesByPairs", _pairs)

	if err != nil {
		return *new([][3]*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new([][3]*big.Int)).(*[][3]*big.Int)

	return out0, err

}

// GetReservesByPairs is a free data retrieval call binding the contract method 0x4dbf0f39.
//
// Solidity: function getReservesByPairs(address[] _pairs) view returns(uint256[3][])
func (_Flashbundle *FlashbundleSession) GetReservesByPairs(_pairs []common.Address) ([][3]*big.Int, error) {
	return _Flashbundle.Contract.GetReservesByPairs(&_Flashbundle.CallOpts, _pairs)
}

// GetReservesByPairs is a free data retrieval call binding the contract method 0x4dbf0f39.
//
// Solidity: function getReservesByPairs(address[] _pairs) view returns(uint256[3][])
func (_Flashbundle *FlashbundleCallerSession) GetReservesByPairs(_pairs []common.Address) ([][3]*big.Int, error) {
	return _Flashbundle.Contract.GetReservesByPairs(&_Flashbundle.CallOpts, _pairs)
}

// Call is a paid mutator transaction binding the contract method 0x6dbf2fa0.
//
// Solidity: function call(address _to, uint256 _value, bytes _data) payable returns(bytes)
func (_Flashbundle *FlashbundleTransactor) Call(opts *bind.TransactOpts, _to common.Address, _value *big.Int, _data []byte) (*types.Transaction, error) {
	return _Flashbundle.contract.Transact(opts, "call", _to, _value, _data)
}

// Call is a paid mutator transaction binding the contract method 0x6dbf2fa0.
//
// Solidity: function call(address _to, uint256 _value, bytes _data) payable returns(bytes)
func (_Flashbundle *FlashbundleSession) Call(_to common.Address, _value *big.Int, _data []byte) (*types.Transaction, error) {
	return _Flashbundle.Contract.Call(&_Flashbundle.TransactOpts, _to, _value, _data)
}

// Call is a paid mutator transaction binding the contract method 0x6dbf2fa0.
//
// Solidity: function call(address _to, uint256 _value, bytes _data) payable returns(bytes)
func (_Flashbundle *FlashbundleTransactorSession) Call(_to common.Address, _value *big.Int, _data []byte) (*types.Transaction, error) {
	return _Flashbundle.Contract.Call(&_Flashbundle.TransactOpts, _to, _value, _data)
}

// UniswapWeth is a paid mutator transaction binding the contract method 0xecd494b3.
//
// Solidity: function uniswapWeth(uint256 _wethAmountToFirstMarket, uint256 _ethAmountToCoinbase, address[] _targets, bytes[] _payloads) payable returns()
func (_Flashbundle *FlashbundleTransactor) UniswapWeth(opts *bind.TransactOpts, _wethAmountToFirstMarket *big.Int, _ethAmountToCoinbase *big.Int, _targets []common.Address, _payloads [][]byte) (*types.Transaction, error) {
	return _Flashbundle.contract.Transact(opts, "uniswapWeth", _wethAmountToFirstMarket, _ethAmountToCoinbase, _targets, _payloads)
}

// UniswapWeth is a paid mutator transaction binding the contract method 0xecd494b3.
//
// Solidity: function uniswapWeth(uint256 _wethAmountToFirstMarket, uint256 _ethAmountToCoinbase, address[] _targets, bytes[] _payloads) payable returns()
func (_Flashbundle *FlashbundleSession) UniswapWeth(_wethAmountToFirstMarket *big.Int, _ethAmountToCoinbase *big.Int, _targets []common.Address, _payloads [][]byte) (*types.Transaction, error) {
	return _Flashbundle.Contract.UniswapWeth(&_Flashbundle.TransactOpts, _wethAmountToFirstMarket, _ethAmountToCoinbase, _targets, _payloads)
}

// UniswapWeth is a paid mutator transaction binding the contract method 0xecd494b3.
//
// Solidity: function uniswapWeth(uint256 _wethAmountToFirstMarket, uint256 _ethAmountToCoinbase, address[] _targets, bytes[] _payloads) payable returns()
func (_Flashbundle *FlashbundleTransactorSession) UniswapWeth(_wethAmountToFirstMarket *big.Int, _ethAmountToCoinbase *big.Int, _targets []common.Address, _payloads [][]byte) (*types.Transaction, error) {
	return _Flashbundle.Contract.UniswapWeth(&_Flashbundle.TransactOpts, _wethAmountToFirstMarket, _ethAmountToCoinbase, _targets, _payloads)
}

// Receive is a paid mutator transaction binding the contract receive function.
//
// Solidity: receive() payable returns()
func (_Flashbundle *FlashbundleTransactor) Receive(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Flashbundle.contract.RawTransact(opts, nil) // calldata is disallowed for receive function
}

// Receive is a paid mutator transaction binding the contract receive function.
//
// Solidity: receive() payable returns()
func (_Flashbundle *FlashbundleSession) Receive() (*types.Transaction, error) {
	return _Flashbundle.Contract.Receive(&_Flashbundle.TransactOpts)
}

// Receive is a paid mutator transaction binding the contract receive function.
//
// Solidity: receive() payable returns()
func (_Flashbundle *FlashbundleTransactorSession) Receive() (*types.Transaction, error) {
	return _Flashbundle.Contract.Receive(&_Flashbundle.TransactOpts)
}
