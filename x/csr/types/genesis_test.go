package types_test

import (
	"testing"

	"github.com/Canto-Network/Canto/v2/x/csr/types"
	"github.com/evmos/ethermint/tests"
	"github.com/stretchr/testify/suite"
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

	contracts := []string{tests.GenerateAddress().String(), tests.GenerateAddress().String(), tests.GenerateAddress().String()}
	poolAddress := tests.GenerateAddress().String()
	nft1 := types.NewCSRNFT(0, poolAddress)
	nft2 := types.NewCSRNFT(1, poolAddress)

	csrPool := types.CSRPool{
		NftSupply:   2,
		PoolAddress: poolAddress,
		CsrNfts:     []*types.CSRNFT{&nft1, &nft2},
	}
	csr := types.CSR{
		Deployer:  tests.GenerateAddress().String(),
		Contracts: contracts,
		CsrPool:   &csrPool,
	}
	suite.csrs = []*types.CSR{&csr, &csr}

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
			desc: "Duplicate pool in genesis state - fail",
			genState: &types.GenesisState{
				Params: suite.params,
				Csrs:   suite.csrs,
			},
			valid: false,
		},
		{
			desc: "No duplicate pool in genesis state - pass",
			genState: &types.GenesisState{
				Params: suite.params,
				Csrs:   suite.csrs[:1],
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
