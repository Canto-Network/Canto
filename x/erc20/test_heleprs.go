package erc20

import (
	"encoding/json"
	"fmt"
	"math/big"

	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/evmos/ethermint/server/config"
	evmkeeper "github.com/evmos/ethermint/x/evm/keeper"
	evm "github.com/evmos/ethermint/x/evm/types"
	feemarketkeeper "github.com/evmos/ethermint/x/feemarket/keeper"

	"github.com/Canto-Network/Canto/v7/contracts"
)

func DeployContract(ctx sdk.Context,
	evmKeeper evmkeeper.Keeper, feemarketKeeper feemarketkeeper.Keeper,
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

	erc20DeployTx := evm.NewTxContract(
		chainID,
		nonce,
		nil,     // amount
		res.Gas, // gasLimit
		nil,     // gasPrice
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
