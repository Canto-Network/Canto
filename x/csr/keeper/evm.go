package keeper

import (
	"math/big"

	errorsmod "cosmossdk.io/errors"
	"github.com/Canto-Network/Canto/v7/contracts"
	"github.com/Canto-Network/Canto/v7/x/csr/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"

	evmtypes "github.com/evmos/ethermint/x/evm/types"
)

// Default gas limit for eth txs from the module account
var DefaultGasLimit uint64 = 30000000

// DeployTurnstile will deploy the Turnstile smart contract from the csr module account. This will allow the
// CSR module to interact with the CSR NFTs in a permissionless way.
func (k Keeper) DeployTurnstile(
	ctx sdk.Context,
) (common.Address, error) {
	addr, err := k.DeployContract(ctx, contracts.TurnstileContract)
	if err != nil {
		return common.Address{}, errorsmod.Wrapf(ErrContractDeployments,
			"EVM::DeployTurnstile error deploying Turnstile: %s", err.Error())
	}
	return addr, nil
}

// DeployContract will deploy an arbitrary smart-contract. It takes the compiled contract object as
// well as an arbitrary number of arguments which will be supplied to the contructor. All contracts deployed
// are deployed by the module account.
func (k Keeper) DeployContract(
	ctx sdk.Context,
	contract evmtypes.CompiledContract,
	args ...interface{},
) (common.Address, error) {
	// pack constructor arguments according to compiled contract's abi
	// method name is nil in this case, we are calling the constructor
	ctorArgs, err := contract.ABI.Pack("", args...)
	if err != nil {
		return common.Address{}, errorsmod.Wrapf(ErrContractDeployments,
			"EVM::DeployContract error packing data: %s", err.Error())
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
			errorsmod.Wrapf(ErrContractDeployments,
				"EVM::DeployContract error retrieving nonce: %s", err.Error())
	}

	amount := big.NewInt(0)
	_, err = k.CallEVM(ctx, types.ModuleAddress, nil, amount, data, true)
	if err != nil {
		return common.Address{},
			errorsmod.Wrapf(ErrContractDeployments,
				"EVM::DeployContract error deploying contract: %s", err.Error())
	}

	// Derive the newly created module smart contract using the module address and nonce
	return crypto.CreateAddress(types.ModuleAddress, nonce), nil
}

// CallMethod is a function to interact with a contract once it is deployed. It inputs the method name on the
// smart contract, the compiled smart contract, the address from which the tx will be made, the contract address,
// the amount to be supplied in msg.value, and finally an arbitrary number of arguments that should be supplied to the
// function method.
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
		return nil, errorsmod.Wrapf(ErrMethodCall, "EVM::CallMethod there was an issue packing the arguments into the method signature: %s", err.Error())
	}

	// call method
	resp, err := k.CallEVM(ctx, from, contractAddr, amount, data, true)
	if err != nil {
		return nil, errorsmod.Wrapf(ErrMethodCall, "EVM::CallMethod error applying message: %s", err.Error())
	}

	return resp, nil
}

// CallEVM performs a EVM transaction given the from address, the to address, amount to be sent, data and
// whether to commit the tx in the EVM keeper.
func (k Keeper) CallEVM(
	ctx sdk.Context,
	from common.Address,
	to *common.Address,
	amount *big.Int,
	data []byte,
	commit bool,
) (*evmtypes.MsgEthereumTxResponse, error) {
	nonce, err := k.accountKeeper.GetSequence(ctx, from.Bytes())
	if err != nil {
		return nil, err
	}

	// As evmKeeper.ApplyMessage does not directly increment the gas meter, any transaction
	// completed through the CSR module account will technically be 'free'. As such, we can
	// set the gas limit to some arbitrarily high enough number such that every transaction
	// from the module account will always go through.
	// see: https://github.com/evmos/ethermint/blob/35850e620d2825327a175f46ec3e8c60af84208d/x/evm/keeper/state_transition.go#L466
	gasLimit := DefaultGasLimit

	// Create the EVM msg
	msg := ethtypes.NewMessage(
		from,
		to,
		nonce,
		amount,        // amount
		gasLimit,      // gasLimit
		big.NewInt(0), // gasPrice
		big.NewInt(0), // gasFeeCap
		big.NewInt(0), // gasTipCap
		data,
		ethtypes.AccessList{}, // AccessList
		!commit,
	)

	// Apply the msg to the EVM keeper
	res, err := k.evmKeeper.ApplyMessage(ctx, msg, evmtypes.NewNoOpTracer(), commit)
	if err != nil {
		return nil, err
	}

	if res.Failed() {
		return nil, errorsmod.Wrap(evmtypes.ErrVMExecution, res.VmError)
	}
	return res, nil
}
