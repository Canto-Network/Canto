package keeper_test

import (
	"github.com/Canto-Network/Canto/v2/x/csr/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/evmos/ethermint/tests"
)

// Sets a bunch of CSRs in the store and then
func (suite *KeeperTestSuite) TestCSRSetGet() {
	numberCSRs := 10
	csrs := GenerateCSRs(numberCSRs)
	for _, csr := range csrs {
		suite.app.CSRKeeper.SetCSR(suite.ctx, csr)
	}

	for _, csr := range csrs {
		id := csr.Id
		// Get CSR from the store
		result, found := suite.app.CSRKeeper.GetCSR(suite.ctx, id)

		// Validation
		suite.Require().True(found)
		suite.Require().Equal(result.Owner, csr.Owner)
		suite.Require().Equal(result.Contracts, csr.Contracts)
		suite.Require().Equal(result.Id, id)
		suite.Require().Equal(result.Account, csr.Account)
	}
}

// Creates a bunch of CSRs and checks if every single contract belongs to the correct pool
func (suite *KeeperTestSuite) TestGetNFTByContract() {
	csrs := GenerateCSRs(5)
	for _, csr := range csrs {
		suite.app.CSRKeeper.SetCSR(suite.ctx, csr)
	}

	for _, csr := range csrs {
		contracts := csr.Contracts
		for _, contract := range contracts {
			result, found := suite.app.CSRKeeper.GetNFTByContract(suite.ctx, contract)

			suite.Require().True(found)
			suite.Require().Equal(csr.Id, result)
		}

		// Check if a newly generated address belongs to any pool
		contract := tests.GenerateAddress().String()
		_, found := suite.app.CSRKeeper.GetNFTByContract(suite.ctx, contract)

		suite.Require().False(found)
	}
}

// Creates a bunch of CSRs and then assigns ownership of some to a single account
func (suite *KeeperTestSuite) TestGetCSRsByOwner() {
	owner := sdk.AccAddress(tests.GenerateAddress().Bytes()).String()
	csrs := GenerateCSRs(10)

	csrs[0].Owner = owner
	csrs[3].Owner = owner
	csrs[4].Owner = owner
	csrs[6].Owner = owner
	csrs[7].Owner = owner

	for _, csr := range csrs {
		suite.app.CSRKeeper.SetCSR(suite.ctx, csr)
	}

	csrsForOwner := suite.app.CSRKeeper.GetCSRsByOwner(suite.ctx, owner)
	suite.Require().Equal(5, len(csrsForOwner))
	suite.Require().Equal([]uint64{0, 3, 4, 6, 7}, csrsForOwner)
}

// Generates an array of CSRs for testing purposes
func GenerateCSRs(number int) []types.CSR {
	csrs := make([]types.CSR, 0)

	for index := 0; index < number; index++ {
		owner := sdk.AccAddress(tests.GenerateAddress().Bytes())
		contracts := []string{tests.GenerateAddress().String(), tests.GenerateAddress().String(),
			tests.GenerateAddress().String(), tests.GenerateAddress().String()}
		id := uint64(index)
		account := sdk.AccAddress(tests.GenerateAddress().Bytes())

		csr := types.NewCSR(
			owner,
			contracts,
			id,
			account,
		)
		csrs = append(csrs, csr)
	}
	return csrs
}
