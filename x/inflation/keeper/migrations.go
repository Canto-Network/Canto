package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	v7 "github.com/Canto-Network/Canto/v1/x/inflation/migrations/v7"

)


var _ module.MigrationHandler = Migrator{}.Migrate1to2

// Migrator is a struct for handling in-place store migrations.
type Migrator struct {
	keeper Keeper
}

// NewMigrator returns a new Migrator.
func NewMigrator(keeper Keeper) Migrator {
	return Migrator{
		keeper: keeper,
	}
}


func (m Migrator) Migrate1to2(ctx sdk.Context) error {
	return v7.UpdateParams(ctx, m.keeper)
}