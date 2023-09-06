package keeper_test

import (
	"github.com/Canto-Network/Canto/v7/x/liquidstaking/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/simulation"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/evmos/ethermint/tests"
	"math/rand"
)

// Sets a bunch of insurances in the store and then get and ensure that each of them
// match up with what is stored on stack vs keeper
func (suite *KeeperTestSuite) TestInsuranceSetGet() {
	numberInsurances := 10
	insurances := GenerateInsurances(numberInsurances, false)
	for _, insurance := range insurances {
		suite.app.LiquidStakingKeeper.SetInsurance(suite.ctx, insurance)
	}

	for _, insurance := range insurances {
		id := insurance.Id
		status := insurance.Status
		chunkId := insurance.ChunkId
		// Get insurance from the store
		result, found := suite.app.LiquidStakingKeeper.GetInsurance(suite.ctx, id)

		// Validation
		suite.Require().True(found)
		suite.Require().Equal(result.Id, id)
		suite.Require().Equal(result.Status, status)
		suite.Require().Equal(result.ChunkId, chunkId)
	}
}

func (suite *KeeperTestSuite) TestDeleteInsurance() {
	numberInsurances := 10
	insurances := GenerateInsurances(numberInsurances, false)
	for _, insurance := range insurances {
		suite.app.LiquidStakingKeeper.SetInsurance(suite.ctx, insurance)
	}

	for _, insurance := range insurances {
		id := insurance.Id
		// Get insurance from the store
		result, found := suite.app.LiquidStakingKeeper.GetInsurance(suite.ctx, id)

		// Validation
		suite.Require().True(found)
		suite.Require().Equal(result.Id, id)

		// Delete insurance from the store
		suite.app.LiquidStakingKeeper.DeleteInsurance(suite.ctx, id)

		// Get insurance from the store
		result, found = suite.app.LiquidStakingKeeper.GetInsurance(suite.ctx, id)

		// Validation
		suite.Require().False(found)
		suite.Require().Equal(result.Id, uint64(0))
	}
}

func (suite *KeeperTestSuite) TestLastInsuranceIdSetGet() {
	// Set LastInsuranceId and retrieve it
	id := uint64(10)
	suite.app.LiquidStakingKeeper.SetLastInsuranceId(suite.ctx, id)

	result := suite.app.LiquidStakingKeeper.GetLastInsuranceId(suite.ctx)
	suite.Require().Equal(result, id)
}

func (suite *KeeperTestSuite) TestIterateAllInsurances() {
	numberInsurances := 10
	insurances := GenerateInsurances(numberInsurances, false)
	for _, insurance := range insurances {
		suite.app.LiquidStakingKeeper.SetInsurance(suite.ctx, insurance)
	}

	// Iterate all insurances
	var insuranceList []types.Insurance
	suite.app.LiquidStakingKeeper.IterateAllInsurances(suite.ctx, func(insurance types.Insurance) (stop bool) {
		insuranceList = append(insuranceList, insurance)
		return false
	})

	// Validation
	suite.Require().Equal(len(insuranceList), numberInsurances)
}

func (suite *KeeperTestSuite) TestGetAllInsurances() {
	numberInsurances := 10
	insurances := GenerateInsurances(numberInsurances, false)
	for _, insurance := range insurances {
		suite.app.LiquidStakingKeeper.SetInsurance(suite.ctx, insurance)
	}

	// Get all insurances
	insuranceList := suite.app.LiquidStakingKeeper.GetAllInsurances(suite.ctx)

	// Validation
	suite.Require().Equal(len(insuranceList), numberInsurances)
}

// Creates a bunch of insurances
func GenerateInsurances(number int, sameAddress bool) []types.Insurance {
	s := rand.NewSource(0)
	r := rand.New(s)

	insurances := make([]types.Insurance, number)
	for i := 0; i < number; i++ {
		var addr string
		if sameAddress {
			addr = authtypes.NewModuleAddress("test").String()
		} else {
			addr = sdk.AccAddress(tests.GenerateAddress().Bytes()).String()
		}

		randomFee := sdk.NewDecWithPrec(int64(simulation.RandIntBetween(r, 1, 99)), 2)
		insurances[i] = types.NewInsurance(uint64(i), addr, "", randomFee)
		insurances[i].ProviderAddress = addr
	}
	return insurances
}
