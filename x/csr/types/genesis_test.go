package types_test

import (
	"testing"

	"github.com/Canto-Network/Canto/v2/x/csr/types"
	"github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/evmos/ethermint/tests"
)

type GensisStateSuite struct {
	suite.Suite
	params types.Params
	csrs   []*types.CSR
}

func TestGenesisStateSuite(t *testing.T) {
	suite.Run(t, new(GensisStateSuite))
}

func (suite *GensisStateSuite) SetupTest() {
	suite.params = types.DefaultParams()

	owner := sdk.AccAddress(tests.GenerateAddress().Bytes())
	contracts := []string{tests.GenerateAddress().String(), tests.GenerateAddress().String(),
		tests.GenerateAddress().String(), tests.GenerateAddress().String()}
	id := 0
	account := sdk.AccAddress(tests.GenerateAddress().Bytes())
	csr := types.NewCSR(
		owner,
		contracts,
		uint64(id),
		account,
	)
	suite.csrs = []*types.CSR{&csr}
}

// Test all of the genesis states, when empty and when not
func (suite *GensisStateSuite) TestGenesisStateValidate() {
	testCases := []struct {
		desc     string
		genState *types.GenesisState
		valid    bool
	}{
		{
			desc:     "Default genesis parameters are valid - pass",
			genState: types.DefaultGenesis(),
			valid:    true,
		},
		{
			desc: "Non-empty csrs - fail",
			genState: &types.GenesisState{
				Params: suite.params,
				Csrs:   suite.csrs,
			},
			valid: false,
		},
		{
			desc: "Eempty csrs - pass",
			genState: &types.GenesisState{
				Params: suite.params,
				Csrs:   []*types.CSR{},
			},
			valid: true,
		},
	}

	for _, tc := range testCases {
		err := tc.genState.Validate()

		if tc.valid {
			suite.Require().NoError(err, tc.desc)
		} else {
			suite.Require().Error(err, tc.desc)
		}
	}
}
