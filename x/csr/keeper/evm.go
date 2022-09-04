package keeper

import (
	"github.com/Canto-Network/Canto/v2/contracts"
	"github.com/Canto-Network/Canto/v2/x/csr/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	evmtypes "github.com/evmos/ethermint/x/evm/types"
)

// mintCSR, this method is called in the process of CSR registration, mints a CSR to
// the account intending to register their CSR
func (k Keeper) MintCSR(
	ctx sdk.Context,
	receiver common.Address,
) (uint64, error) {
	// first retrieve CSR contract from state, fail if it hasn't been deployed
	csrNft, found := k.GetCSRNFT(ctx)
	if !found {
		return 0, sdkerrors.Wrapf(ErrCSRNFTNotDeployed, "CSR::Keeper: MintCSR: CSRNFT.sol not deployed")
	}
	resp, err := k.CallMethod(ctx, "mintCSR", contracts.CSRNFTContract, &csrNft, receiver)
	if err != nil {
		return 0, sdkerrors.Wrapf(ErrMethodCall, "CSR::Keeper: MintCSR: error calling Mint on CSR: %s", err.Error())
	}
	var ret types.CSRUint64Response

	// unwrap the resp data
	if err = contracts.CSRNFTContract.ABI.UnpackIntoInterface(&ret, "mintCSR", resp.Ret); err != nil {
		return 0, sdkerrors.Wrapf(ErrUnpackData, "CSR::Keeper: MintCSR: error unpacking data: %s", err.Error())
	}

	return ret.Value, nil
}

// deploy CSRNFT, this method is called in begin block, it takes as arguments the name and symbol of the CSRNFT
func (k Keeper) DeployCSRNFT(
	ctx sdk.Context,
	name,
	symbol string,
) (common.Address, error) {
	// deploy CSRNFT with name, symbol constructor args
	addr, err := k.DeployContract(ctx, contracts.CSRNFTContract, name, symbol)
	if err != nil {
		return common.Address{}, sdkerrors.Wrapf(ErrContractDeployments,
			"CSR::Keeper: DeployCSRNFT: error deploying Turnstile: %s", err.Error())
	}
	return addr, nil
}

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
	_, err = k.erc20Keeper.CallEVMWithData(ctx, types.ModuleAddress, nil, data, true)
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
	contractAddr *common.Address,
	args ...interface{},
) (*evmtypes.MsgEthereumTxResponse, error) {
	// pack method args

	methodArgs, err := contract.ABI.Pack(method, args...)

	if err != nil {
		return nil, sdkerrors.Wrapf(ErrContractDeployments, "CSR:Keeper::DeployContract: method call incorrect: %s", err.Error())
	}
	// call method
	resp, err := k.erc20Keeper.CallEVMWithData(ctx, types.ModuleAddress, contractAddr, methodArgs, true)
	if err != nil {
		return nil, sdkerrors.Wrapf(ErrContractDeployments, "CSR:Keeper: :CallMethod: error applying message: %s", err.Error())
	}

	return resp, nil
}
