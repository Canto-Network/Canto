package csr_test

import (
	"testing"

	keepertest "github.com/Canto-Network/Canto/v2/testutil/keeper"
	"github.com/Canto-Network/Canto/v2/x/csr"
	"github.com/Canto-Network/Canto/v2/x/csr/types"
	"github.com/stretchr/testify/require"
)

func TestGenesis(t *testing.T) {
	genesisState := types.GenesisState{
		Params: types.DefaultParams(),

		// this line is used by starport scaffolding # genesis/test/state
	}

	k, ctx := keepertest.CsrKeeper(t)
	csr.InitGenesis(ctx, *k, genesisState)
	got := csr.ExportGenesis(ctx, *k)
	require.NotNil(t, got)

	// this line is used by starport scaffolding # genesis/test/assert
}
