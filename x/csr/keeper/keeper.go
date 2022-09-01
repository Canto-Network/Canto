package keeper

import (
	"fmt"

	"github.com/tendermint/tendermint/libs/log"

	"github.com/Canto-Network/Canto/v2/x/csr/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

type (
	Keeper struct {
		cdc        	  codec.BinaryCodec
		storeKey   	  sdk.StoreKey
		paramstore 	  paramtypes.Subspace
		evmKeeper  	  types.EVMKeeper
		bankKeeper 	  types.BankKeeper
		accountKeeper types.AccountKeeper
		erc20Keeper   types.ERC20Keeper
	}
)

func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey sdk.StoreKey,
	ps paramtypes.Subspace,
	evmKeeper types.EVMKeeper,
	bankKeeper types.BankKeeper,
	acccountKeeper types.AccountKeeper,
	erc20Keeper types.ERC20Keeper,
) Keeper {
	// set KeyTable if it has not already been set
	if !ps.HasKeyTable() {
		ps = ps.WithKeyTable(types.ParamKeyTable())
	}

	return Keeper{
		storeKey:   storeKey,
		cdc:        cdc,
		paramstore: ps,
		evmKeeper:  evmKeeper,
		bankKeeper: bankKeeper,
		accountKeeper: acccountKeeper,
		erc20Keeper: erc20Keeper,
	}
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}
