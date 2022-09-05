package types

import (
	"github.com/ethereum/go-ethereum/common"
	_ "github.com/ethereum/go-ethereum/core"
	_ "github.com/ethereum/go-ethereum/core/vm"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/types"

	"github.com/evmos/ethermint/x/evm/statedb"
	evmtypes "github.com/evmos/ethermint/x/evm/types"
)

// AccountKeeper defines the expected interface needed to retrieve account info.
type AccountKeeper interface {
	NewAccountWithAddress(ctx sdk.Context, addr sdk.AccAddress) types.AccountI
	GetModuleAddress(moduleName string) sdk.AccAddress
	GetSequence(ctx sdk.Context, addr sdk.AccAddress) (uint64, error)
	SetAccount(ctx sdk.Context, acc types.AccountI)
}

// BankKeeper defines the expected interface needed to retrieve account balances.
type BankKeeper interface {
	SendCoinsFromModuleToAccount(ctx sdk.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error
	SendCoins(ctx sdk.Context, fromAddr, toAddr sdk.AccAddress, amt sdk.Coins) error
	GetBalance(ctx sdk.Context, addr sdk.AccAddress, denom string) sdk.Coin
	GetAllBalances(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins
}

// EVMKeeper defines the expected EVM keeper interface used on erc20
type EVMKeeper interface {
	EVMConfig(ctx sdk.Context) (*evmtypes.EVMConfig, error)
	GetParams(ctx sdk.Context) evmtypes.Params
	GetAccount(ctx sdk.Context, addr common.Address) *statedb.Account
	GetAccountWithoutBalance(ctx sdk.Context, addr common.Address) *statedb.Account
}

type ERC20Keeper interface {
	CallEVMWithData(ctx sdk.Context, from common.Address,
		contract *common.Address, data []byte, commit bool) (*evmtypes.MsgEthereumTxResponse, error)
}
