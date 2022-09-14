package keeper

import (
	"math/big"

	"github.com/Canto-Network/Canto/v2/contracts"
	"github.com/Canto-Network/Canto/v2/x/csr/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"

	evmtypes "github.com/evmos/ethermint/x/evm/types"
)

// Default gas limit for eth txs on the turnstile
var DefaultGasLimit uint64 = 25000000

// deploy Turnstile, this method is called in begin block, it takes as argument the, the deployment takes no arguments
func (k Keeper) DeployTurnstile(
	ctx sdk.Context,
) (common.Address, error) {
	// divert deployment call to DeployContract, no constructor arguments are needed for Turnstile
	addr, err := k.DeployContract(ctx, contracts.TurnstileContract)
	if err != nil {
		return common.Address{}, sdkerrors.Wrapf(ErrContractDeployments,
			"CSR::Keeper: DeployTurnstile: error deploying Turnstile: %s", err.Error())
	}
	// return deployed address of contract
	return addr, nil
}

// function to deploy an arbitrary smart-contract, takes as argument, the compiled
// contract object, as well as an arbitrary length array of arguments to the deployments
// deploys the contract from the module account
func (k Keeper) DeployContract(
	ctx sdk.Context,
	contract evmtypes.CompiledContract,
	args ...interface{},
) (common.Address, error) {
	// pack constructor arguments according to compiled contract's abi
	// method name is nil in this case, we are calling the constructor
	ctorArgs, err := contract.ABI.Pack("", args...)
	if err != nil {
		return common.Address{}, sdkerrors.Wrapf(ErrContractDeployments,
			"CSR::Keeper::DeployContract: error packing data: %s", err.Error())
	}
	// pack method data into byte string, enough for bin and constructor arguments
	data := make([]byte, len(contract.Bin)+len(ctorArgs))
	// copy bin into data, and append to data the constructor arguments
	copy(data[:len(contract.Bin)], contract.Bin)
	// copy constructor args last
	copy(data[len(contract.Bin):], ctorArgs)
	// retrieve sequence number first to derive address if not by CREATE2
	nonce, err := k.accountKeeper.GetSequence(ctx, types.ModuleAddress.Bytes())
	if err != nil {
		return common.Address{},
			sdkerrors.Wrapf(ErrContractDeployments,
				"CSR:Keeper::DeployContract: error retrieving nonce: %s", err.Error())
	}

	// deploy contract using erc20 callEVMWithData, applies contract deployments to
	// current stateDb
	amount := big.NewInt(0)
	_, err = k.CallEVM(ctx, types.ModuleAddress, nil, amount, data, true)
	if err != nil {
		return common.Address{},
			sdkerrors.Wrapf(ErrAddressDerivation,
				"CSR:Keeper::DeployContract: error deploying contract: %s", err.Error())
	}

	// determine how to derive contract address, is to be derived from nonce
	return crypto.CreateAddress(types.ModuleAddress, nonce), nil
}

// function to interact with a contract once it is deployed, requires function signature,
// as well as arguments, pass pointer of argument type to CallMethod, and returned value from call is returned
func (k Keeper) CallMethod(
	ctx sdk.Context,
	method string,
	contract evmtypes.CompiledContract,
	from common.Address,
	contractAddr *common.Address,
	amount *big.Int,
	args ...interface{},
) (*evmtypes.MsgEthereumTxResponse, error) {
	// pack method args

	data, err := contract.ABI.Pack(method, args...)

	if err != nil {
		return nil, sdkerrors.Wrapf(ErrContractDeployments, "CSR:Keeper::DeployContract: method call incorrect: %s", err.Error())
	}
	// call method
	resp, err := k.CallEVM(ctx, from, contractAddr, amount, data, true)
	if err != nil {
		return nil, sdkerrors.Wrapf(ErrContractDeployments, "CSR:Keeper: :CallMethod: error applying message: %s", err.Error())
	}

	return resp, nil
}

// CallEVM performs a smart contract method call using contract data and amount
func (k Keeper) CallEVM(
	ctx sdk.Context,
	from common.Address,
	contract *common.Address,
	amount *big.Int,
	data []byte,
	commit bool,
) (*evmtypes.MsgEthereumTxResponse, error) {
	nonce, err := k.accountKeeper.GetSequence(ctx, from.Bytes())
	if err != nil {
		return nil, err
	}

	gasLimit := DefaultGasLimit

	msg := ethtypes.NewMessage(
		from,
		contract,
		nonce,
		amount,        // amount
		gasLimit,      // gasLimit
		big.NewInt(0), // gasPrice
		big.NewInt(0), // gasFeeCap
		big.NewInt(0), // gasTipCap
		data,
		ethtypes.AccessList{}, // AccessList
		!commit,               // isFake
	)

	res, err := k.evmKeeper.ApplyMessage(ctx, msg, evmtypes.NewNoOpTracer(), commit)
	if err != nil {
		return nil, err
	}

	if res.Failed() {
		return nil, sdkerrors.Wrap(evmtypes.ErrVMExecution, res.VmError)
	}
	return res, nil
}
