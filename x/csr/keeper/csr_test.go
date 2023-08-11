package keeper_test

import (
	"github.com/Canto-Network/Canto/v7/x/csr/types"
	"github.com/evmos/ethermint/tests"
)

// Sets a bunch of CSRs in the store and then get and ensure that each of them
// match up with what is stored on stack vs keeper
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
	}
}

// Creates a bunch of CSRs and checks if every single contract belongs the correct NFT
func (suite *KeeperTestSuite) TestGetNFTByContract() {
	csrs := GenerateCSRs(5)
	for _, csr := range csrs {
		suite.app.CSRKeeper.SetCSR(suite.ctx, csr)
	}

	// Iterate every smart contract in every csr and ensure it belongs to the correct NFT
	for _, csr := range csrs {
		contracts := csr.Contracts
		for _, contract := range contracts {
			result, found := suite.app.CSRKeeper.GetNFTByContract(suite.ctx, contract)

			// Validation
			suite.Require().True(found)
			suite.Require().Equal(csr.Id, result)
		}

		// Check if a newly generated address belongs to any NFT
		contract := tests.GenerateAddress().String()
		_, found := suite.app.CSRKeeper.GetNFTByContract(suite.ctx, contract)

		suite.Require().False(found)
	}
}

// Test the setter/getter for the Turnstile address
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

// Generates an array of CSRs with 4 randomly generated smart contracts per CSR for testing purposes.
// Will create `number` of CSRs where each NFT ID starts from 0 and increments to number - 1.
func GenerateCSRs(number int) []types.CSR {
	csrs := make([]types.CSR, 0)

	for index := 0; index < number; index++ {
		contracts := []string{tests.GenerateAddress().String(), tests.GenerateAddress().String(),
			tests.GenerateAddress().String(), tests.GenerateAddress().String()}
		id := uint64(index)

		csr := types.NewCSR(
			contracts,
			id,
		)
		csrs = append(csrs, csr)
	}
	return csrs
}
