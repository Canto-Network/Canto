package keeper

import (
	"github.com/Canto-Network/Canto/v2/x/csr/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// GetParams get all parameters as types.Params
func (k Keeper) GetParams(ctx sdk.Context) types.Params {
	return types.NewParams(types.DefaultEnableCSR, types.DefaultCSRShares, types.DefaultAddressDerivationCostCreate)
}

// SetParams set the params
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) {
	k.paramstore.SetParamSet(ctx, &params)
}
