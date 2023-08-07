package types_test

import (
	"testing"

	"github.com/Canto-Network/Canto/v7/x/csr/types"
	"github.com/stretchr/testify/suite"
)

type GensisStateSuite struct {
	suite.Suite
	params types.Params
}

func TestGenesisStateSuite(t *testing.T) {
	suite.Run(t, new(GensisStateSuite))
}

func (suite *GensisStateSuite) SetupTest() {
	suite.params = types.DefaultParams()
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
