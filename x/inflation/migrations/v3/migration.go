package v3

import (
	"github.com/Canto-Network/Canto/v2/x/inflation/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type InflationKeeper interface {
	// Get Params used to retrieve params for inflation calculation
	GetParams(ctx sdk.Context) types.Params
	// Used to rollback period
	SetPeriod(ctx sdk.Context, period uint64)
	BondedRatio(ctx sdk.Context) sdk.Dec
	// setting new Epoch Mint Provision to state
	SetEpochMintProvision(ctx sdk.Context, epochMinProvision sdk.Dec)
}

func UpdateParams(ctx sdk.Context, ik InflationKeeper) error{
	ctx.Logger().Info("Setting Inflation Provision per Epoch")
	params := ik.GetParams(ctx)
	// set Period to zero (no decay)
	ik.SetPeriod(ctx, 0)
	// set bonded ratio
	ratio := ik.BondedRatio(ctx)
	provision := types.CalculateEpochMintProvision(params, 0, 30, ratio)
	// set epoch mint provision to state
	ik.SetEpochMintProvision(ctx, provision)
	return nil
}