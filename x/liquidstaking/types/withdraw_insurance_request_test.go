package types_test

import (
	"testing"

	"github.com/Canto-Network/Canto/v6/x/liquidstaking/types"
	"github.com/stretchr/testify/suite"
)

type withdrawInsuranceRequestTestSuite struct {
	suite.Suite
}

func TestWithdrawInsuranceRequestTestSuite(t *testing.T) {
	suite.Run(t, new(withdrawInsuranceRequestTestSuite))
}

func (suite *withdrawInsuranceRequestTestSuite) TestNewWithdrawInsuranceRequest() {
	wir := types.NewWithdrawInsuranceRequest(1)
	suite.Equal(uint64(1), wir.InsuranceId)
}

func (suite *withdrawInsuranceRequestTestSuite) TestValidate() {
	id := uint64(1)
	wir := types.WithdrawInsuranceRequest{
		InsuranceId: id,
	}

	insurance := types.Insurance{
		Id:     id,
		Status: types.INSURANCE_STATUS_PAIRED,
	}
	insuranceMap := map[uint64]types.Insurance{
		id: insurance,
	}
	suite.NoError(wir.Validate(insuranceMap))

	insuranceMap[id] = types.Insurance{
		Id:     id,
		Status: types.INSURANCE_STATUS_UNPAIRED,
	}
	suite.Error(wir.Validate(insuranceMap))

	delete(insuranceMap, id)
	suite.Error(wir.Validate(insuranceMap))
}

func (suite *withdrawInsuranceRequestTestSuite) TestEqual() {
	wir := types.WithdrawInsuranceRequest{
		InsuranceId: 1,
	}

	cpy := wir
	suite.True(cpy.Equal(wir))

	cpy.InsuranceId = 2
	suite.False(cpy.Equal(wir))
}
