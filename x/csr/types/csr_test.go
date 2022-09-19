package types

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/evmos/ethermint/tests"
	"github.com/stretchr/testify/suite"
)

type CSRTestSuite struct {
	suite.Suite
	contracts []string
	id        uint64
	account   string
}

func TestCSRSuite(t *testing.T) {
	suite.Run(t, new(CSRTestSuite))
}

func (suite *CSRTestSuite) SetupTest() {
	suite.contracts = []string{tests.GenerateAddress().String(), tests.GenerateAddress().String(),
		tests.GenerateAddress().String(), tests.GenerateAddress().String()}
	suite.id = 0
	suite.account = sdk.AccAddress(tests.GenerateAddress().Bytes()).String()
}

// TestCSR will do basic stateless validation testing on the CSR objects.
func (suite *CSRTestSuite) TestCSR() {
	testCases := []struct {
		msg        string
		csr        CSR
		expectPass bool
	}{
		{
			"Create CSR object - pass",
			CSR{
				Contracts: suite.contracts,
				Id:        suite.id,
			},
			true,
		},
		{
			"Create CSR object with 0 smart contracts - fail",
			CSR{
				Contracts: []string{},
				Id:        suite.id,
			},
			false,
		},
		{
			"Create CSR object with invalid smart contract addresses - fail",
			CSR{
				Contracts: append(suite.contracts, ""),
				Id:        suite.id,
			},
			false,
		},
		{
			"Create CSR object with duplicate smart contract addresses - fail",
			CSR{
				Contracts: append(suite.contracts, suite.contracts...),
				Id:        suite.id,
			},
			false,
		},
	}
	for _, tc := range testCases {
		suite.Run(tc.msg, func() {
			err := tc.csr.Validate()
			if tc.expectPass {
				suite.Require().NoError(err, tc.msg)
			} else {
				suite.Require().Error(err, tc.msg)
			}
		})
	}
}
