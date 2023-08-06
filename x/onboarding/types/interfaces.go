package types

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"

	evmtypes "github.com/evmos/ethermint/x/evm/types"

	tmbytes "github.com/tendermint/tendermint/libs/bytes"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	transfertypes "github.com/cosmos/ibc-go/v3/modules/apps/transfer/types"
	clienttypes "github.com/cosmos/ibc-go/v3/modules/core/02-client/types"
	channeltypes "github.com/cosmos/ibc-go/v3/modules/core/04-channel/types"

	coinswaptypes "github.com/Canto-Network/Canto/v6/x/coinswap/types"
	erc20types "github.com/Canto-Network/Canto/v6/x/erc20/types"
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
	TradeInputForExactOutput(ctx sdk.Context, input coinswaptypes.Input, output coinswaptypes.Output) (sdk.Int, error)
	GetStandardDenom(ctx sdk.Context) string
}

// BankKeeper defines the banking keeper that must be fulfilled when
// creating a x/onboarding keeper.
type BankKeeper interface {
	GetBalance(ctx sdk.Context, addr sdk.AccAddress, denom string) sdk.Coin
	GetAllBalances(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins
	SendCoinsFromModuleToAccount(ctx sdk.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error
	SendCoinsFromModuleToModule(ctx sdk.Context, senderModule, recipientModule string, amt sdk.Coins) error
	MintCoins(ctx sdk.Context, name string, amt sdk.Coins) error
	SendCoins(ctx sdk.Context, fromAddr sdk.AccAddress, toAddr sdk.AccAddress, amt sdk.Coins) error
	BlockedAddr(addr sdk.AccAddress) bool
	SpendableCoins(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins
}

// AccountKeeper defines the expected account keeper
type AccountKeeper interface {
	GetAccount(sdk.Context, sdk.AccAddress) authtypes.AccountI
}

// TransferKeeper defines the expected IBC transfer keeper.
type TransferKeeper interface {
	GetDenomTrace(ctx sdk.Context, denomTraceHash tmbytes.HexBytes) (transfertypes.DenomTrace, bool)
	SendTransfer(
		ctx sdk.Context,
		sourcePort, sourceChannel string,
		token sdk.Coin,
		sender sdk.AccAddress, receiver string,
		timeoutHeight clienttypes.Height, timeoutTimestamp uint64,
	) error
}

// ChannelKeeper defines the expected IBC channel keeper.
type ChannelKeeper interface {
	GetChannel(ctx sdk.Context, srcPort, srcChan string) (channel channeltypes.Channel, found bool)
}
