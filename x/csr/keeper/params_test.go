package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	testkeeper "github.com/Canto-Network/Canto/v2/testutil/keeper"
	"github.com/Canto-Network/Canto/v2/x/csr/types"
)

func TestGetParams(t *testing.T) {
	k, ctx := testkeeper.CsrKeeper(t)
	params := types.DefaultParams()

	k.SetParams(ctx, params)

	require.EqualValues(t, params, k.GetParams(ctx))
}
