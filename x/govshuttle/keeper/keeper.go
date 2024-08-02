package keeper

import (
	"fmt"

	"cosmossdk.io/core/store"
	"cosmossdk.io/store/prefix"
	"github.com/ethereum/go-ethereum/common"

	"cosmossdk.io/log"
	"github.com/Canto-Network/Canto/v7/x/govshuttle/types"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/runtime"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govkeeper "github.com/cosmos/cosmos-sdk/x/gov/keeper"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

type (
	Keeper struct {
		storeService store.KVStoreService
		cdc          codec.BinaryCodec
		paramstore   paramtypes.Subspace

		accKeeper   types.AccountKeeper
		erc20Keeper types.ERC20Keeper
		govKeeper   *govkeeper.Keeper
		authority   string
	}
)

func NewKeeper(
	storeService store.KVStoreService,
	cdc codec.BinaryCodec,
	ps paramtypes.Subspace,
	ak types.AccountKeeper,
	ek types.ERC20Keeper,
	gk *govkeeper.Keeper,
	authority string,

) Keeper {
	// set KeyTable if it has not already been set
	if !ps.HasKeyTable() {
		ps = ps.WithKeyTable(types.ParamKeyTable())
	}

	return Keeper{
		cdc:          cdc,
		storeService: storeService,
		paramstore:   ps,
		accKeeper:    ak,
		erc20Keeper:  ek,
		govKeeper:    gk,
		authority:    authority,
	}
}

// GetAuthority returns the x/govshuttle module's authority.
func (k Keeper) GetAuthority() string {
	return k.authority
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// retrieve the port address from state
func (k Keeper) GetPort(ctx sdk.Context) (common.Address, bool) {
	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	prefixStore := prefix.NewStore(store, types.PortKey)
	bz := prefixStore.Get(types.PortKey)
	// if not found return false
	if len(bz) == 0 {
		return common.Address{}, false
	}
	return common.BytesToAddress(bz), true
}

// commit the address of the current govShuttle mapcontract to state (Port.sol)
func (k Keeper) SetPort(ctx sdk.Context, portAddr common.Address) {
	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	prefixStore := prefix.NewStore(store, types.PortKey)
	prefixStore.Set(types.PortKey, portAddr.Bytes())
}
