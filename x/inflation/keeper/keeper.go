package keeper

import (
	"cosmossdk.io/core/store"
	"cosmossdk.io/log"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	distrkeeper "github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"

	"github.com/Canto-Network/Canto/v7/x/inflation/types"
)

// Keeper of the inflation store
type Keeper struct {
	storeService store.KVStoreService
	cdc          codec.BinaryCodec
	paramstore   paramtypes.Subspace

	accountKeeper    types.AccountKeeper
	bankKeeper       types.BankKeeper
	distrKeeper      distrkeeper.Keeper
	stakingKeeper    types.StakingKeeper
	feeCollectorName string

	authority string
}

// NewKeeper creates a new mint Keeper instance
func NewKeeper(
	storeService store.KVStoreService,
	cdc codec.BinaryCodec,
	ps paramtypes.Subspace,
	ak types.AccountKeeper,
	bk types.BankKeeper,
	dk distrkeeper.Keeper,
	sk types.StakingKeeper,
	feeCollectorName string,
	authority string,
) Keeper {
	// ensure mint module account is set
	if addr := ak.GetModuleAddress(types.ModuleName); addr == nil {
		panic("the mint module account has not been set")
	}

	// set KeyTable if it has not already been set
	if !ps.HasKeyTable() {
		ps = ps.WithKeyTable(types.ParamKeyTable())
	}

	return Keeper{
		storeService:     storeService,
		cdc:              cdc,
		paramstore:       ps,
		accountKeeper:    ak,
		bankKeeper:       bk,
		distrKeeper:      dk,
		stakingKeeper:    sk,
		feeCollectorName: feeCollectorName,
		authority:        authority,
	}
}

// GetAuthority returns the x/inflation module's authority.
func (k Keeper) GetAuthority() string {
	return k.authority
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", "x/"+types.ModuleName)
}

func (k Keeper) GetComPool(ctx sdk.Context) sdk.DecCoins {
	feePool, _ := k.distrKeeper.FeePool.Get(ctx)
	return feePool.CommunityPool
}
