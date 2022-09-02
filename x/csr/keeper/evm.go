package keeper

import (
	"fmt"

	"github.com/Canto-Network/Canto/v2/x/csr/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	evmtypes "github.com/evmos/ethermint/x/evm/types"
)

// ABIEvents is used to unmarshal data from evm transactions
// EventHashes is used to match the event signature to the correct event for process

// Register the CSR in the store given the data from the evm transaction
func (k Keeper) RegisterCSREvent(data []byte) error {
	response, err := turnstileContract.Unpack(types.TurnstileEventRegisterCSR, data)
	if err != nil {
		return err
	}

	event := types.RegisterCSREvent{
		SmartContractAddress: response[0].(common.Address),
		Receiver:             response[1].(common.Address),
	}
	fmt.Println(event)

	// HANDLE LOGIC HERE

	return nil
}

// Update a CSR existing in the store given data from the evm transaction
func (k Keeper) UpdateCSREvent(data []byte) error {
	response, err := turnstileContract.Unpack(types.TurnstileEventUpdateCSR, data)
	if err != nil {
		return err
	}

	event := types.UpdateCSREvent{
		SmartContractAddress: response[0].(common.Address),
		Nft_id:               response[1].(uint64),
	}
	fmt.Println(event)

	// HANDLE LOGIC HERE

	return nil
}

// Retroactively register contracts that were previously deployed (non-CSR enabled smart contracts)
// and add the the CSR store
func (k Keeper) RetroactiveRegisterEvent() error {
	return nil
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
		return common.Address{}, sdkerrors.Wrapf(types.ErrContractDeployments,
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
			sdkerrors.Wrapf(types.ErrContractDeployments,
				"CSR:Keeper::DeployContract: error retrieving nonce: %s", err.Error())
	}

	// deploy contract using erc20 callEVMWithData, applies contract deployments to
	// current stateDb
	_, err = k.erc20Keeper.CallEVMWithData(ctx, types.ModuleAddress, nil, data, true)
	if err != nil {
		return common.Address{},
			sdkerrors.Wrapf(types.ErrAddressDerivation,
				"CSR:Keeper::DeployContract: error retrieving nonce: %s", err.Error())
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
	contractAddr *common.Address,
	args ...interface{},
) (*evmtypes.MsgEthereumTxResponse, error) {
	// pack method args

	methodArgs, err := contract.ABI.Pack(method, args...)

	if err != nil {
		return nil, sdkerrors.Wrapf(types.ErrContractDeployments, "CSR:Keeper::DeployContract: method call incorrect: %s", err.Error())
	}
	// call method
	resp, err := k.erc20Keeper.CallEVMWithData(ctx, types.ModuleAddress, contractAddr, methodArgs, true)
	if err != nil {
		return nil, sdkerrors.Wrapf(types.ErrContractDeployments, "CSR:Keeper: :CallMethod: error applying message: %s", err.Error())
	}

	return resp, nil
}
