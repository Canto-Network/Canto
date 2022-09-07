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
		suite.Require().Equal(result.Contracts, csr.Contracts)
		suite.Require().Equal(result.Id, id)
		suite.Require().Equal(result.Beneficiary, csr.Beneficiary)
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

func (suite *KeeperTestSuite) TestSetGetTurnstile() {
	// Set Turnstile address and retrieve it
	addr := tests.GenerateAddress()
	suite.app.CSRKeeper.SetTurnstile(suite.ctx, addr)
	suite.Commit()
	// retrieve addr
	expectAddr, found := suite.app.CSRKeeper.GetTurnstile(suite.ctx)
	suite.Require().True(found)
	suite.Require().Equal(addr, expectAddr)
}

func (suite *KeeperTestSuite) TestSetGetCSRNFT() {
	addr := tests.GenerateAddress()
	suite.app.CSRKeeper.SetCSRNFT(suite.ctx, addr)
	// retrieve addr
	expectAddr, found := suite.app.CSRKeeper.GetCSRNFT(suite.ctx)
	suite.Require().True(found)
	suite.Require().Equal(expectAddr, addr)
}

// Generates an array of CSRs for testing purposes
func GenerateCSRs(number int) []types.CSR {
	csrs := make([]types.CSR, 0)

	for index := 0; index < number; index++ {
		owner := sdk.AccAddress(tests.GenerateAddress().Bytes())
		contracts := []string{tests.GenerateAddress().String(), tests.GenerateAddress().String(),
			tests.GenerateAddress().String(), tests.GenerateAddress().String()}
		id := uint64(index)
		account := s.app.CSRKeeper.CreateNewAccount(s.ctx)

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
