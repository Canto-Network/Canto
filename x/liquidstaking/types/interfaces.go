package types

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	abci "github.com/tendermint/tendermint/abci/types"
)

// AccountKeeper defines the expected interface needed to retrieve account info.
type AccountKeeper interface {
	GetModuleAddress(moduleName string) sdk.AccAddress
	GetModuleAccount(ctx sdk.Context, moduleName string) authtypes.ModuleAccountI
	GetAccount(ctx sdk.Context, addr sdk.AccAddress) authtypes.AccountI
	SetModuleAccount(ctx sdk.Context, macc authtypes.ModuleAccountI)
}

// BankKeeper defines the expected interface needed to retrieve account balances.
type BankKeeper interface {
	SendCoinsFromModuleToAccount(ctx sdk.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error
	SendCoinsFromModuleToModule(ctx sdk.Context, senderModule, recipientModule string, amt sdk.Coins) error
	SendCoinsFromAccountToModule(ctx sdk.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error
	SendCoins(ctx sdk.Context, fromAddr sdk.AccAddress, toAddr sdk.AccAddress, amt sdk.Coins) error
	InputOutputCoins(ctx sdk.Context, inputs []banktypes.Input, outputs []banktypes.Output) error
	MintCoins(ctx sdk.Context, moduleName string, amt sdk.Coins) error
	BurnCoins(ctx sdk.Context, moduleName string, amt sdk.Coins) error
	IsSendEnabledCoin(ctx sdk.Context, coin sdk.Coin) bool
	GetDenomMetaData(ctx sdk.Context, denom string) (banktypes.Metadata, bool)
	SetDenomMetaData(ctx sdk.Context, denomMetaData banktypes.Metadata)
	GetAllBalances(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins
	GetBalance(ctx sdk.Context, addr sdk.AccAddress, denom string) sdk.Coin
	HasBalance(ctx sdk.Context, addr sdk.AccAddress, amt sdk.Coin) bool
	GetSupply(ctx sdk.Context, denom string) sdk.Coin
}

// BankKeeper defines the expected interface needed to retrieve delegation info.
type StakingKeeper interface {
	Validator(sdk.Context, sdk.ValAddress) stakingtypes.ValidatorI
	ValidatorByConsAddr(sdk.Context, sdk.ConsAddress) stakingtypes.ValidatorI
	GetValidator(ctx sdk.Context, addr sdk.ValAddress) (validator stakingtypes.Validator, found bool)

	GetAllValidators(ctx sdk.Context) (validators []stakingtypes.Validator)
	GetBondedValidatorsByPower(ctx sdk.Context) []stakingtypes.Validator

	GetLastTotalPower(ctx sdk.Context) sdk.Int
	GetLastValidatorPower(ctx sdk.Context, valAddr sdk.ValAddress) int64

	Delegation(sdk.Context, sdk.AccAddress, sdk.ValAddress) stakingtypes.DelegationI
	GetDelegation(ctx sdk.Context,
		delAddr sdk.AccAddress, valAddr sdk.ValAddress) (delegation stakingtypes.Delegation, found bool)
	IterateDelegations(ctx sdk.Context, delegator sdk.AccAddress,
		fn func(index int64, delegation stakingtypes.DelegationI) (stop bool))
	IterateUnbondingDelegations(ctx sdk.Context, fn func(index int64, ubd stakingtypes.UnbondingDelegation) (stop bool))
	IterateDelegatorUnbondingDelegations(ctx sdk.Context, delegator sdk.AccAddress, cb func(ubd stakingtypes.UnbondingDelegation) (stop bool))
	Delegate(
		ctx sdk.Context, delAddr sdk.AccAddress, bondAmt sdk.Int, tokenSrc stakingtypes.BondStatus,
		validator stakingtypes.Validator, subtractAccount bool,
	) (newShares sdk.Dec, err error)

	BondDenom(ctx sdk.Context) (res string)
	Undelegate(ctx sdk.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress, sharesAmount sdk.Dec) (time.Time, error)
	Unbond(ctx sdk.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress, shares sdk.Dec) (amount sdk.Int, err error)
	UnbondingTime(ctx sdk.Context) (res time.Duration)
	SetUnbondingDelegationEntry(
		ctx sdk.Context, delegatorAddr sdk.AccAddress, validatorAddr sdk.ValAddress,
		creationHeight int64, minTime time.Time, balance sdk.Int,
	) stakingtypes.UnbondingDelegation
	InsertUBDQueue(ctx sdk.Context, ubd stakingtypes.UnbondingDelegation,
		completionTime time.Time)
	ValidateUnbondAmount(
		ctx sdk.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress, amt sdk.Int,
	) (shares sdk.Dec, err error)
	GetAllUnbondingDelegations(ctx sdk.Context, delegator sdk.AccAddress) []stakingtypes.UnbondingDelegation
	GetUnbondingDelegation(
		ctx sdk.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress,
	) (ubd stakingtypes.UnbondingDelegation, found bool)
	BeginRedelegation(
		ctx sdk.Context, delAddr sdk.AccAddress, valSrcAddr, valDstAddr sdk.ValAddress, sharesAmount sdk.Dec,
	) (completionTime time.Time, err error)
	GetAllRedelegations(
		ctx sdk.Context, delegator sdk.AccAddress, srcValAddress, dstValAddress sdk.ValAddress,
	) []stakingtypes.Redelegation
	HasReceivingRedelegation(ctx sdk.Context, delAddr sdk.AccAddress, valDstAddr sdk.ValAddress) bool
	BlockValidatorUpdates(ctx sdk.Context) []abci.ValidatorUpdate
	HasMaxUnbondingDelegationEntries(ctx sdk.Context,
		delegatorAddr sdk.AccAddress, validatorAddr sdk.ValAddress) bool
}

// DistributionKeeper defines the expected interface needed to retrieve distribution info.
type DistributionKeeper interface {
	IncrementValidatorPeriod(ctx sdk.Context, val stakingtypes.ValidatorI) uint64
	CalculateDelegationRewards(ctx sdk.Context, val stakingtypes.ValidatorI, del stakingtypes.DelegationI, endingPeriod uint64) (rewards sdk.DecCoins)
	WithdrawDelegationRewards(ctx sdk.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) (sdk.Coins, error)
}

// SlashingKeeper expected slashing keeper (noalias)
type SlashingKeeper interface {
	IsTombstoned(ctx sdk.Context, consAddr sdk.ConsAddress) bool
}

// StakingHooks event hooks for staking validator object (noalias)
type StakingHooks interface {
	AfterDelegationModified(_ sdk.Context, _ sdk.AccAddress, _ sdk.ValAddress)
}
