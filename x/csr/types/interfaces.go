package types

import (
	context "context"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/vm"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/evmos/ethermint/x/evm/statedb"
	evmtypes "github.com/evmos/ethermint/x/evm/types"
)

// AccountKeeper defines the expected interface needed to retrieve account info.
type AccountKeeper interface {
	GetSequence(ctx context.Context, addr sdk.AccAddress) (uint64, error)
}

// BankKeeper defines the expected interface needed to send tx fees from the fee collector module to the CSR module for distribution.
type BankKeeper interface {
	SendCoinsFromModuleToModule(ctx context.Context, senderModule string, recipientModule string, amt sdk.Coins) error
	BurnCoins(ctx context.Context, moduleName string, amt sdk.Coins) error
}

// EVMKeeper defines the expected EVM keeper interface used to make EVM deployments and txs from the module account.
type EVMKeeper interface {
	ApplyMessage(ctx sdk.Context, msg core.Message, tracer vm.EVMLogger, commit bool) (*evmtypes.MsgEthereumTxResponse, error)
	EstimateGas(c context.Context, req *evmtypes.EthCallRequest) (*evmtypes.EstimateGasResponse, error)
	GetAccount(ctx sdk.Context, addr common.Address) *statedb.Account
	GetParams(ctx sdk.Context) evmtypes.Params
}
