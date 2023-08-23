package keeper

import (
	"encoding/json"
	"fmt"
	"math/big"

	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/evmos/ethermint/server/config"
	evm "github.com/evmos/ethermint/x/evm/types"

	"github.com/Canto-Network/Canto/v7/contracts"
	"github.com/Canto-Network/Canto/v7/x/erc20/types"
)

func DeployContract(ctx sdk.Context,
	evmKeeper types.EVMKeeper, feemarketKeeper types.FeeMarketKeeper,
	address common.Address, signer keyring.Signer,
	name, symbol string, decimals uint8) (common.Address, error) {
	c := sdk.WrapSDKContext(ctx)
	chainID := evmKeeper.ChainID()

	ctorArgs, err := contracts.ERC20MinterBurnerDecimalsContract.ABI.Pack("", name, symbol, decimals)
	if err != nil {
		return common.Address{}, err
	}

	data := append(contracts.ERC20MinterBurnerDecimalsContract.Bin, ctorArgs...)
	args, err := json.Marshal(&evm.TransactionArgs{
		From: &address,
		Data: (*hexutil.Bytes)(&data),
	})
	if err != nil {
		return common.Address{}, err
	}

	res, err := evmKeeper.EstimateGas(c, &evm.EthCallRequest{
		Args:   args,
		GasCap: uint64(config.DefaultGasCap),
	})
	if err != nil {
		return common.Address{}, err
	}

	nonce := evmKeeper.GetNonce(ctx, address)
	minGasPrice := feemarketKeeper.GetParams(ctx).MinGasPrice.BigInt()
	minGasPrice = new(big.Int).Add(minGasPrice, big.NewInt(1))
	erc20DeployTx := evm.NewTxContract(
		chainID,
		nonce,
		nil,         // amount
		res.Gas,     // gasLimit
		minGasPrice, // gasPrice
		feemarketKeeper.GetBaseFee(ctx),
		big.NewInt(1),
		data,                   // input
		&ethtypes.AccessList{}, // accesses
	)

	erc20DeployTx.From = address.Hex()
	err = erc20DeployTx.Sign(ethtypes.LatestSignerForChainID(chainID), signer)
	if err != nil {
		return common.Address{}, err
	}

	rsp, err := evmKeeper.EthereumTx(c, erc20DeployTx)
	if err != nil {
		return common.Address{}, err
	}

	if rsp.VmError != "" {
		return common.Address{}, fmt.Errorf("failed to deploy contract: %s", rsp.VmError)
	}
	return crypto.CreateAddress(address, nonce), nil
}

func DeployERC20Contract(
	ctx sdk.Context,
	k Keeper,
	ak types.AccountKeeper,
	coinMetadata banktypes.Metadata,
) (common.Address, error) {
	decimals := uint8(0)
	if len(coinMetadata.DenomUnits) > 0 {
		decimalsIdx := len(coinMetadata.DenomUnits) - 1
		decimals = uint8(coinMetadata.DenomUnits[decimalsIdx].Exponent)
	}
	ctorArgs, err := contracts.ERC20MinterBurnerDecimalsContract.ABI.Pack(
		"",
		coinMetadata.Name,
		coinMetadata.Symbol,
		decimals,
	)
	if err != nil {
		return common.Address{}, sdkerrors.Wrapf(types.ErrABIPack, "coin metadata is invalid %s: %s", coinMetadata.Name, err.Error())
	}

	data := make([]byte, len(contracts.ERC20MinterBurnerDecimalsContract.Bin)+len(ctorArgs))
	copy(data[:len(contracts.ERC20MinterBurnerDecimalsContract.Bin)], contracts.ERC20MinterBurnerDecimalsContract.Bin)
	copy(data[len(contracts.ERC20MinterBurnerDecimalsContract.Bin):], ctorArgs)

	nonce, err := ak.GetSequence(ctx, types.ModuleAddress.Bytes())
	if err != nil {
		return common.Address{}, err
	}

	contractAddr := crypto.CreateAddress(types.ModuleAddress, nonce)
	_, err = k.CallEVMWithData(ctx, types.ModuleAddress, nil, data, true)
	if err != nil {
		return common.Address{}, sdkerrors.Wrapf(err, "failed to deploy contract for %s", coinMetadata.Name)
	}

	return contractAddr, nil
}
