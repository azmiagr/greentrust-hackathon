// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package greenpassportchain

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
	_ = abi.ConvertType
)

// GreenPassportRegistryMetaData contains all meta data concerning the GreenPassportRegistry contract.
var GreenPassportRegistryMetaData = &bind.MetaData{
	ABI: "[{\"type\":\"constructor\",\"inputs\":[{\"name\":\"initialOwner\",\"type\":\"address\",\"internalType\":\"address\"}],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"getPassport\",\"inputs\":[{\"name\":\"passportId\",\"type\":\"string\",\"internalType\":\"string\"}],\"outputs\":[{\"name\":\"profileId\",\"type\":\"string\",\"internalType\":\"string\"},{\"name\":\"storedPassportId\",\"type\":\"string\",\"internalType\":\"string\"},{\"name\":\"documentHashes\",\"type\":\"bytes32[]\",\"internalType\":\"bytes32[]\"},{\"name\":\"grsScore\",\"type\":\"uint16\",\"internalType\":\"uint16\"},{\"name\":\"issuedAt\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"issuer\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"exists\",\"type\":\"bool\",\"internalType\":\"bool\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"issuePassport\",\"inputs\":[{\"name\":\"profileId\",\"type\":\"string\",\"internalType\":\"string\"},{\"name\":\"passportId\",\"type\":\"string\",\"internalType\":\"string\"},{\"name\":\"documentHashes\",\"type\":\"bytes32[]\",\"internalType\":\"bytes32[]\"},{\"name\":\"grsScore\",\"type\":\"uint16\",\"internalType\":\"uint16\"},{\"name\":\"issuedAt\",\"type\":\"uint64\",\"internalType\":\"uint64\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"owner\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"}],\"stateMutability\":\"view\"},{\"type\":\"event\",\"name\":\"GreenPassportIssued\",\"inputs\":[{\"name\":\"passportId\",\"type\":\"string\",\"indexed\":true,\"internalType\":\"string\"},{\"name\":\"profileId\",\"type\":\"string\",\"indexed\":false,\"internalType\":\"string\"},{\"name\":\"grsScore\",\"type\":\"uint16\",\"indexed\":false,\"internalType\":\"uint16\"},{\"name\":\"documentCount\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"},{\"name\":\"issuedAt\",\"type\":\"uint64\",\"indexed\":false,\"internalType\":\"uint64\"},{\"name\":\"issuer\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"}],\"anonymous\":false},{\"type\":\"error\",\"name\":\"EmptyDocumentHashes\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"EmptyPassportId\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"EmptyProfileId\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"NotOwner\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"PassportAlreadyIssued\",\"inputs\":[]}]",
}

// GreenPassportRegistryABI is the input ABI used to generate the binding from.
// Deprecated: Use GreenPassportRegistryMetaData.ABI instead.
var GreenPassportRegistryABI = GreenPassportRegistryMetaData.ABI

// GreenPassportRegistry is an auto generated Go binding around an Ethereum contract.
type GreenPassportRegistry struct {
	GreenPassportRegistryCaller     // Read-only binding to the contract
	GreenPassportRegistryTransactor // Write-only binding to the contract
	GreenPassportRegistryFilterer   // Log filterer for contract events
}

// GreenPassportRegistryCaller is an auto generated read-only Go binding around an Ethereum contract.
type GreenPassportRegistryCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// GreenPassportRegistryTransactor is an auto generated write-only Go binding around an Ethereum contract.
type GreenPassportRegistryTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// GreenPassportRegistryFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type GreenPassportRegistryFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// GreenPassportRegistrySession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type GreenPassportRegistrySession struct {
	Contract     *GreenPassportRegistry // Generic contract binding to set the session for
	CallOpts     bind.CallOpts          // Call options to use throughout this session
	TransactOpts bind.TransactOpts      // Transaction auth options to use throughout this session
}

// GreenPassportRegistryCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type GreenPassportRegistryCallerSession struct {
	Contract *GreenPassportRegistryCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts                // Call options to use throughout this session
}

// GreenPassportRegistryTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type GreenPassportRegistryTransactorSession struct {
	Contract     *GreenPassportRegistryTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts                // Transaction auth options to use throughout this session
}

// GreenPassportRegistryRaw is an auto generated low-level Go binding around an Ethereum contract.
type GreenPassportRegistryRaw struct {
	Contract *GreenPassportRegistry // Generic contract binding to access the raw methods on
}

// GreenPassportRegistryCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type GreenPassportRegistryCallerRaw struct {
	Contract *GreenPassportRegistryCaller // Generic read-only contract binding to access the raw methods on
}

// GreenPassportRegistryTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type GreenPassportRegistryTransactorRaw struct {
	Contract *GreenPassportRegistryTransactor // Generic write-only contract binding to access the raw methods on
}

// NewGreenPassportRegistry creates a new instance of GreenPassportRegistry, bound to a specific deployed contract.
func NewGreenPassportRegistry(address common.Address, backend bind.ContractBackend) (*GreenPassportRegistry, error) {
	contract, err := bindGreenPassportRegistry(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &GreenPassportRegistry{GreenPassportRegistryCaller: GreenPassportRegistryCaller{contract: contract}, GreenPassportRegistryTransactor: GreenPassportRegistryTransactor{contract: contract}, GreenPassportRegistryFilterer: GreenPassportRegistryFilterer{contract: contract}}, nil
}

// NewGreenPassportRegistryCaller creates a new read-only instance of GreenPassportRegistry, bound to a specific deployed contract.
func NewGreenPassportRegistryCaller(address common.Address, caller bind.ContractCaller) (*GreenPassportRegistryCaller, error) {
	contract, err := bindGreenPassportRegistry(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &GreenPassportRegistryCaller{contract: contract}, nil
}

// NewGreenPassportRegistryTransactor creates a new write-only instance of GreenPassportRegistry, bound to a specific deployed contract.
func NewGreenPassportRegistryTransactor(address common.Address, transactor bind.ContractTransactor) (*GreenPassportRegistryTransactor, error) {
	contract, err := bindGreenPassportRegistry(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &GreenPassportRegistryTransactor{contract: contract}, nil
}

// NewGreenPassportRegistryFilterer creates a new log filterer instance of GreenPassportRegistry, bound to a specific deployed contract.
func NewGreenPassportRegistryFilterer(address common.Address, filterer bind.ContractFilterer) (*GreenPassportRegistryFilterer, error) {
	contract, err := bindGreenPassportRegistry(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &GreenPassportRegistryFilterer{contract: contract}, nil
}

// bindGreenPassportRegistry binds a generic wrapper to an already deployed contract.
func bindGreenPassportRegistry(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := GreenPassportRegistryMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_GreenPassportRegistry *GreenPassportRegistryRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _GreenPassportRegistry.Contract.GreenPassportRegistryCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_GreenPassportRegistry *GreenPassportRegistryRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _GreenPassportRegistry.Contract.GreenPassportRegistryTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_GreenPassportRegistry *GreenPassportRegistryRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _GreenPassportRegistry.Contract.GreenPassportRegistryTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_GreenPassportRegistry *GreenPassportRegistryCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _GreenPassportRegistry.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_GreenPassportRegistry *GreenPassportRegistryTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _GreenPassportRegistry.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_GreenPassportRegistry *GreenPassportRegistryTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _GreenPassportRegistry.Contract.contract.Transact(opts, method, params...)
}

// GetPassport is a free data retrieval call binding the contract method 0x904a6751.
//
// Solidity: function getPassport(string passportId) view returns(string profileId, string storedPassportId, bytes32[] documentHashes, uint16 grsScore, uint64 issuedAt, address issuer, bool exists)
func (_GreenPassportRegistry *GreenPassportRegistryCaller) GetPassport(opts *bind.CallOpts, passportId string) (struct {
	ProfileId        string
	StoredPassportId string
	DocumentHashes   [][32]byte
	GrsScore         uint16
	IssuedAt         uint64
	Issuer           common.Address
	Exists           bool
}, error) {
	var out []interface{}
	err := _GreenPassportRegistry.contract.Call(opts, &out, "getPassport", passportId)

	outstruct := new(struct {
		ProfileId        string
		StoredPassportId string
		DocumentHashes   [][32]byte
		GrsScore         uint16
		IssuedAt         uint64
		Issuer           common.Address
		Exists           bool
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.ProfileId = *abi.ConvertType(out[0], new(string)).(*string)
	outstruct.StoredPassportId = *abi.ConvertType(out[1], new(string)).(*string)
	outstruct.DocumentHashes = *abi.ConvertType(out[2], new([][32]byte)).(*[][32]byte)
	outstruct.GrsScore = *abi.ConvertType(out[3], new(uint16)).(*uint16)
	outstruct.IssuedAt = *abi.ConvertType(out[4], new(uint64)).(*uint64)
	outstruct.Issuer = *abi.ConvertType(out[5], new(common.Address)).(*common.Address)
	outstruct.Exists = *abi.ConvertType(out[6], new(bool)).(*bool)

	return *outstruct, err

}

// GetPassport is a free data retrieval call binding the contract method 0x904a6751.
//
// Solidity: function getPassport(string passportId) view returns(string profileId, string storedPassportId, bytes32[] documentHashes, uint16 grsScore, uint64 issuedAt, address issuer, bool exists)
func (_GreenPassportRegistry *GreenPassportRegistrySession) GetPassport(passportId string) (struct {
	ProfileId        string
	StoredPassportId string
	DocumentHashes   [][32]byte
	GrsScore         uint16
	IssuedAt         uint64
	Issuer           common.Address
	Exists           bool
}, error) {
	return _GreenPassportRegistry.Contract.GetPassport(&_GreenPassportRegistry.CallOpts, passportId)
}

// GetPassport is a free data retrieval call binding the contract method 0x904a6751.
//
// Solidity: function getPassport(string passportId) view returns(string profileId, string storedPassportId, bytes32[] documentHashes, uint16 grsScore, uint64 issuedAt, address issuer, bool exists)
func (_GreenPassportRegistry *GreenPassportRegistryCallerSession) GetPassport(passportId string) (struct {
	ProfileId        string
	StoredPassportId string
	DocumentHashes   [][32]byte
	GrsScore         uint16
	IssuedAt         uint64
	Issuer           common.Address
	Exists           bool
}, error) {
	return _GreenPassportRegistry.Contract.GetPassport(&_GreenPassportRegistry.CallOpts, passportId)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_GreenPassportRegistry *GreenPassportRegistryCaller) Owner(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _GreenPassportRegistry.contract.Call(opts, &out, "owner")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_GreenPassportRegistry *GreenPassportRegistrySession) Owner() (common.Address, error) {
	return _GreenPassportRegistry.Contract.Owner(&_GreenPassportRegistry.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_GreenPassportRegistry *GreenPassportRegistryCallerSession) Owner() (common.Address, error) {
	return _GreenPassportRegistry.Contract.Owner(&_GreenPassportRegistry.CallOpts)
}

// IssuePassport is a paid mutator transaction binding the contract method 0x7ea19325.
//
// Solidity: function issuePassport(string profileId, string passportId, bytes32[] documentHashes, uint16 grsScore, uint64 issuedAt) returns()
func (_GreenPassportRegistry *GreenPassportRegistryTransactor) IssuePassport(opts *bind.TransactOpts, profileId string, passportId string, documentHashes [][32]byte, grsScore uint16, issuedAt uint64) (*types.Transaction, error) {
	return _GreenPassportRegistry.contract.Transact(opts, "issuePassport", profileId, passportId, documentHashes, grsScore, issuedAt)
}

// IssuePassport is a paid mutator transaction binding the contract method 0x7ea19325.
//
// Solidity: function issuePassport(string profileId, string passportId, bytes32[] documentHashes, uint16 grsScore, uint64 issuedAt) returns()
func (_GreenPassportRegistry *GreenPassportRegistrySession) IssuePassport(profileId string, passportId string, documentHashes [][32]byte, grsScore uint16, issuedAt uint64) (*types.Transaction, error) {
	return _GreenPassportRegistry.Contract.IssuePassport(&_GreenPassportRegistry.TransactOpts, profileId, passportId, documentHashes, grsScore, issuedAt)
}

// IssuePassport is a paid mutator transaction binding the contract method 0x7ea19325.
//
// Solidity: function issuePassport(string profileId, string passportId, bytes32[] documentHashes, uint16 grsScore, uint64 issuedAt) returns()
func (_GreenPassportRegistry *GreenPassportRegistryTransactorSession) IssuePassport(profileId string, passportId string, documentHashes [][32]byte, grsScore uint16, issuedAt uint64) (*types.Transaction, error) {
	return _GreenPassportRegistry.Contract.IssuePassport(&_GreenPassportRegistry.TransactOpts, profileId, passportId, documentHashes, grsScore, issuedAt)
}

// GreenPassportRegistryGreenPassportIssuedIterator is returned from FilterGreenPassportIssued and is used to iterate over the raw logs and unpacked data for GreenPassportIssued events raised by the GreenPassportRegistry contract.
type GreenPassportRegistryGreenPassportIssuedIterator struct {
	Event *GreenPassportRegistryGreenPassportIssued // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *GreenPassportRegistryGreenPassportIssuedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(GreenPassportRegistryGreenPassportIssued)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(GreenPassportRegistryGreenPassportIssued)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *GreenPassportRegistryGreenPassportIssuedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *GreenPassportRegistryGreenPassportIssuedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// GreenPassportRegistryGreenPassportIssued represents a GreenPassportIssued event raised by the GreenPassportRegistry contract.
type GreenPassportRegistryGreenPassportIssued struct {
	PassportId    common.Hash
	ProfileId     string
	GrsScore      uint16
	DocumentCount *big.Int
	IssuedAt      uint64
	Issuer        common.Address
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterGreenPassportIssued is a free log retrieval operation binding the contract event 0x60157df590887888d8d4e76d351e978c5674f631fb00a58f5a54506f81198d71.
//
// Solidity: event GreenPassportIssued(string indexed passportId, string profileId, uint16 grsScore, uint256 documentCount, uint64 issuedAt, address indexed issuer)
func (_GreenPassportRegistry *GreenPassportRegistryFilterer) FilterGreenPassportIssued(opts *bind.FilterOpts, passportId []string, issuer []common.Address) (*GreenPassportRegistryGreenPassportIssuedIterator, error) {

	var passportIdRule []interface{}
	for _, passportIdItem := range passportId {
		passportIdRule = append(passportIdRule, passportIdItem)
	}

	var issuerRule []interface{}
	for _, issuerItem := range issuer {
		issuerRule = append(issuerRule, issuerItem)
	}

	logs, sub, err := _GreenPassportRegistry.contract.FilterLogs(opts, "GreenPassportIssued", passportIdRule, issuerRule)
	if err != nil {
		return nil, err
	}
	return &GreenPassportRegistryGreenPassportIssuedIterator{contract: _GreenPassportRegistry.contract, event: "GreenPassportIssued", logs: logs, sub: sub}, nil
}

// WatchGreenPassportIssued is a free log subscription operation binding the contract event 0x60157df590887888d8d4e76d351e978c5674f631fb00a58f5a54506f81198d71.
//
// Solidity: event GreenPassportIssued(string indexed passportId, string profileId, uint16 grsScore, uint256 documentCount, uint64 issuedAt, address indexed issuer)
func (_GreenPassportRegistry *GreenPassportRegistryFilterer) WatchGreenPassportIssued(opts *bind.WatchOpts, sink chan<- *GreenPassportRegistryGreenPassportIssued, passportId []string, issuer []common.Address) (event.Subscription, error) {

	var passportIdRule []interface{}
	for _, passportIdItem := range passportId {
		passportIdRule = append(passportIdRule, passportIdItem)
	}

	var issuerRule []interface{}
	for _, issuerItem := range issuer {
		issuerRule = append(issuerRule, issuerItem)
	}

	logs, sub, err := _GreenPassportRegistry.contract.WatchLogs(opts, "GreenPassportIssued", passportIdRule, issuerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(GreenPassportRegistryGreenPassportIssued)
				if err := _GreenPassportRegistry.contract.UnpackLog(event, "GreenPassportIssued", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseGreenPassportIssued is a log parse operation binding the contract event 0x60157df590887888d8d4e76d351e978c5674f631fb00a58f5a54506f81198d71.
//
// Solidity: event GreenPassportIssued(string indexed passportId, string profileId, uint16 grsScore, uint256 documentCount, uint64 issuedAt, address indexed issuer)
func (_GreenPassportRegistry *GreenPassportRegistryFilterer) ParseGreenPassportIssued(log types.Log) (*GreenPassportRegistryGreenPassportIssued, error) {
	event := new(GreenPassportRegistryGreenPassportIssued)
	if err := _GreenPassportRegistry.contract.UnpackLog(event, "GreenPassportIssued", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
