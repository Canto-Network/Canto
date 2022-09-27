package keeper

import (
	"encoding/json"
	"fmt"
	"math/big"

	"github.com/Canto-Network/Canto/v2/contracts"
	"github.com/Canto-Network/Canto/v2/x/csr/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"

	"github.com/evmos/ethermint/server/config"
	evmtypes "github.com/evmos/ethermint/x/evm/types"
)

// Default gas limit for eth txs on the turnstile
var DefaultGasLimit uint64 = 100000

// DeployTurnstile will deploy the Turnstile smart contract from the csr module account. This will allow the
// CSR module to interact with the CSR NFTs in a permissionless way.
func (k Keeper) DeployTurnstile(
	ctx sdk.Context,
) (common.Address, error) {
	addr, err := k.DeployContract(ctx, contracts.TurnstileContract)
	if err != nil {
		return common.Address{}, sdkerrors.Wrapf(ErrContractDeployments,
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
		return common.Address{}, sdkerrors.Wrapf(ErrContractDeployments,
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
			sdkerrors.Wrapf(ErrContractDeployments,
				"EVM::DeployContract error retrieving nonce: %s", err.Error())
	}

	amount := big.NewInt(0)
	_, err = k.CallEVM(ctx, types.ModuleAddress, nil, amount, data, true)
	if err != nil {
		return common.Address{},
			sdkerrors.Wrapf(ErrContractDeployments,
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
		return nil, sdkerrors.Wrapf(ErrMethodCall, "EVM::CallMethod there was an issue packing the arguments into the method signature: %s", err.Error())
	}

	// call method
	resp, err := k.CallEVM(ctx, from, contractAddr, amount, data, true)
	if err != nil {
		return nil, sdkerrors.Wrapf(ErrMethodCall, "EVM::CallMethod error applying message: %s", err.Error())
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

	// Default the gas limit to const
	gasLimit := DefaultGasLimit
	if commit {
		args, err := json.Marshal(evmtypes.TransactionArgs{
			From: &from,
			To:   to,
			Data: (*hexutil.Bytes)(&data),
		})
		if err != nil {
			return nil, sdkerrors.Wrapf(sdkerrors.ErrJSONMarshal, "failed to marshal tx args: %s", err.Error())
		}

		gasRes, err := k.evmKeeper.EstimateGas(sdk.WrapSDKContext(ctx), &evmtypes.EthCallRequest{
			Args:   args,
			GasCap: config.DefaultGasCap,
		})
		// If no error then we set the gas res
		if err != nil {
			return nil, err
		}
		gasLimit = gasRes.Gas
	}
	info := "The number of units of gas consumed is " + fmt.Sprint(gasLimit)
	k.Logger(ctx).Info(info)

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
		return nil, sdkerrors.Wrap(evmtypes.ErrVMExecution, res.VmError)
	}
	return res, nil
}
