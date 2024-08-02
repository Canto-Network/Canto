package types

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// BankKeeper defines the expected bank keeper
type BankKeeper interface {
	SendCoins(ctx context.Context, fromAddr sdk.AccAddress, toAddr sdk.AccAddress, amt sdk.Coins) error
	GetAllBalances(ctx context.Context, addr sdk.AccAddress) sdk.Coins
	GetBalance(ctx context.Context, addr sdk.AccAddress, denom string) sdk.Coin

	GetSupply(ctx context.Context, denom string) sdk.Coin
	IterateTotalSupply(ctx context.Context, cb func(sdk.Coin) bool)
	SendCoinsFromModuleToAccount(ctx context.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error
	SendCoinsFromAccountToModule(ctx context.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error
	SendCoinsFromModuleToModule(ctx context.Context, senderModule, recipientModule string, amt sdk.Coins) error
	BurnCoins(ctx context.Context, name string, amt sdk.Coins) error
	MintCoins(ctx context.Context, name string, amt sdk.Coins) error
	SpendableCoins(ctx context.Context, addr sdk.AccAddress) sdk.Coins
}

// AccountKeeper defines the expected account keeper
type AccountKeeper interface {
	GetAccount(ctx context.Context, addr sdk.AccAddress) sdk.AccountI
	GetModuleAddress(name string) sdk.AccAddress
	IterateAccounts(ctx context.Context, cb func(account sdk.AccountI) (stop bool))
}
