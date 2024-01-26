package keeper

import (
	"fmt"

	"cosmossdk.io/core/store"
	"cosmossdk.io/log"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"

	"github.com/Canto-Network/Canto/v7/x/erc20/types"
)

// Keeper of this module maintains collections of erc20.
type Keeper struct {
	storeService store.KVStoreService
	cdc          codec.BinaryCodec
	paramstore   paramtypes.Subspace

	accountKeeper types.AccountKeeper
	bankKeeper    types.BankKeeper
	evmKeeper     types.EVMKeeper

	// the address capable of executing a MsgUpdateParams message. Typically, this
	// should be the x/gov module account.
	authority string
}

// NewKeeper creates new instances of the erc20 Keeper
func NewKeeper(
	storeService store.KVStoreService,
	cdc codec.BinaryCodec,
	ps paramtypes.Subspace,
	ak types.AccountKeeper,
	bk types.BankKeeper,
	evmKeeper types.EVMKeeper,
	authority string,

) Keeper {
	// set KeyTable if it has not already been set
	if !ps.HasKeyTable() {
		ps = ps.WithKeyTable(types.ParamKeyTable())
	}

	return Keeper{
		storeService:  storeService,
		cdc:           cdc,
		paramstore:    ps,
		accountKeeper: ak,
		bankKeeper:    bk,
		evmKeeper:     evmKeeper,
		authority:     authority,
	}
}

// GetAuthority returns the x/erc20 module's authority.
func (k Keeper) GetAuthority() string {
	return k.authority
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}
