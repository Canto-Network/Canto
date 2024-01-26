package keeper

import (
	"fmt"

	"cosmossdk.io/core/store"
	"cosmossdk.io/log"
	"github.com/Canto-Network/Canto/v7/x/csr/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

type (
	Keeper struct {
		cdc          codec.BinaryCodec
		storeService store.KVStoreService
		paramstore   paramtypes.Subspace

		accountKeeper    types.AccountKeeper
		evmKeeper        types.EVMKeeper
		bankKeeper       types.BankKeeper
		FeeCollectorName string
	}
)

func NewKeeper(
	cdc codec.BinaryCodec,
	storeService store.KVStoreService,
	ps paramtypes.Subspace,
	accountKeeper types.AccountKeeper,
	evmKeeper types.EVMKeeper,
	bankKeeper types.BankKeeper,
	FeeCollectorName string,
) Keeper {
	// set KeyTable if it has not already been set
	if !ps.HasKeyTable() {
		ps = ps.WithKeyTable(types.ParamKeyTable())
	}

	return Keeper{
		storeService:     storeService,
		cdc:              cdc,
		paramstore:       ps,
		accountKeeper:    accountKeeper,
		evmKeeper:        evmKeeper,
		bankKeeper:       bankKeeper,
		FeeCollectorName: FeeCollectorName,
	}
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}
