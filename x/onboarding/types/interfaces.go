package types

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"

	evmtypes "github.com/evmos/ethermint/x/evm/types"

	tmbytes "github.com/cometbft/cometbft/libs/bytes"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	transfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v8/modules/core/04-channel/types"

	coinswaptypes "github.com/Canto-Network/Canto/v7/x/coinswap/types"
	erc20types "github.com/Canto-Network/Canto/v7/x/erc20/types"
)

type Erc20Keeper interface {
	ConvertCoin(
		goCtx context.Context,
		msg *erc20types.MsgConvertCoin,
	) (*erc20types.MsgConvertCoinResponse, error)
	GetTokenPairID(ctx sdk.Context, token string) []byte
	GetTokenPair(ctx sdk.Context, id []byte) (erc20types.TokenPair, bool)
	BalanceOf(
		ctx sdk.Context,
		abi abi.ABI,
		contract, account common.Address,
	) *big.Int
	CallEVM(
		ctx sdk.Context,
		abi abi.ABI,
		from, contract common.Address,
		commit bool,
		method string,
		args ...interface{},
	) (*evmtypes.MsgEthereumTxResponse, error)
}

type CoinwapKeeper interface {
	TradeInputForExactOutput(ctx sdk.Context, input coinswaptypes.Input, output coinswaptypes.Output) (sdkmath.Int, error)
	GetStandardDenom(ctx sdk.Context) (string, error)
}

// BankKeeper defines the banking keeper that must be fulfilled when
// creating a x/onboarding keeper.
type BankKeeper interface {
	GetBalance(ctx context.Context, addr sdk.AccAddress, denom string) sdk.Coin
	GetAllBalances(ctx context.Context, addr sdk.AccAddress) sdk.Coins
	SendCoinsFromModuleToAccount(ctx context.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error
	SendCoinsFromModuleToModule(ctx context.Context, senderModule, recipientModule string, amt sdk.Coins) error
	MintCoins(ctx context.Context, name string, amt sdk.Coins) error
	SendCoins(ctx context.Context, fromAddr sdk.AccAddress, toAddr sdk.AccAddress, amt sdk.Coins) error
	BlockedAddr(addr sdk.AccAddress) bool
	SpendableCoins(ctx context.Context, addr sdk.AccAddress) sdk.Coins
}

// AccountKeeper defines the expected account keeper
type AccountKeeper interface {
	GetAccount(context.Context, sdk.AccAddress) sdk.AccountI
}

// TransferKeeper defines the expected IBC transfer keeper.
type TransferKeeper interface {
	GetDenomTrace(ctx sdk.Context, denomTraceHash tmbytes.HexBytes) (transfertypes.DenomTrace, bool)
}

// ChannelKeeper defines the expected IBC channel keeper.
type ChannelKeeper interface {
	GetChannel(ctx sdk.Context, srcPort, srcChan string) (channel channeltypes.Channel, found bool)
}
