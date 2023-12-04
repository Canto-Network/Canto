package keeper

import (
	"fmt"

	"cosmossdk.io/store/prefix"
	storetypes "cosmossdk.io/store/types"
	"github.com/ethereum/go-ethereum/common"

	"cosmossdk.io/log"
	"github.com/Canto-Network/Canto/v7/x/govshuttle/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

type (
	Keeper struct {
		storeKey   storetypes.StoreKey
		cdc        codec.BinaryCodec
		paramstore paramtypes.Subspace

		accKeeper   types.AccountKeeper
		erc20Keeper types.ERC20Keeper
		govKeeper   types.GovKeeper
	}
)

func NewKeeper(
	storeKey storetypes.StoreKey,
	cdc codec.BinaryCodec,
	ps paramtypes.Subspace,

	ak types.AccountKeeper,
	ek types.ERC20Keeper,
	gk types.GovKeeper,

) Keeper {
	// set KeyTable if it has not already been set
	if !ps.HasKeyTable() {
		ps = ps.WithKeyTable(types.ParamKeyTable())
	}

	return Keeper{

		cdc:         cdc,
		storeKey:    storeKey,
		paramstore:  ps,
		accKeeper:   ak,
		erc20Keeper: ek,
		govKeeper:   gk,
	}
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// retrieve the port address from state
func (k Keeper) GetPort(ctx sdk.Context) (common.Address, bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.PortKey)
	bz := store.Get(types.PortKey)
	// if not found return false
	if len(bz) == 0 {
		return common.Address{}, false
	}
	return common.BytesToAddress(bz), true
}

// commit the address of the current govShuttle mapcontract to state (Port.sol)
func (k Keeper) SetPort(ctx sdk.Context, portAddr common.Address) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.PortKey)
	store.Set(types.PortKey, portAddr.Bytes())
}
