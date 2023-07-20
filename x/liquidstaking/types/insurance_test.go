package types_test

import (
	"github.com/Canto-Network/Canto/v6/x/liquidstaking/types"
	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/stretchr/testify/suite"
	"github.com/tendermint/tendermint/crypto"
	"testing"
)

type insuranceTestSuite struct {
	suite.Suite
}

func TestInsuranceTestSuite(t *testing.T) {
	suite.Run(t, new(insuranceTestSuite))
}

func (suite *insuranceTestSuite) TestSortInsurances() {
	n := 3
	var val1, val2, val3 stakingtypes.Validator
	publicKeys := simapp.CreateTestPubKeys(n)
	var val1Addr, val2Addr, val3Addr sdk.ValAddress
	validatorMap := make(map[string]stakingtypes.Validator)

	val1Addr = sdk.ValAddress(publicKeys[0].Address())
	val2Addr = sdk.ValAddress(publicKeys[1].Address())
	val3Addr = sdk.ValAddress(publicKeys[2].Address())

	val1, err := stakingtypes.NewValidator(
		val1Addr,
		publicKeys[0],
		stakingtypes.Description{},
	)
	suite.NoError(err)
	fivePercent := sdk.NewDecWithPrec(5, 2)
	val1, err = val1.SetInitialCommission(stakingtypes.NewCommission(fivePercent, fivePercent, fivePercent))
	suite.NoError(err)
	validatorMap[val1Addr.String()] = val1

	val2, err = stakingtypes.NewValidator(
		val2Addr,
		publicKeys[1],
		stakingtypes.Description{},
	)
	suite.NoError(err)
	sevenPercent := sdk.NewDecWithPrec(7, 2)
	val2, err = val2.SetInitialCommission(stakingtypes.NewCommission(sevenPercent, sevenPercent, sevenPercent))
	suite.NoError(err)
	validatorMap[val2Addr.String()] = val2

	val3, err = stakingtypes.NewValidator(
		val3Addr,
		publicKeys[2],
		stakingtypes.Description{},
	)
	suite.NoError(err)
	threePercent := sdk.NewDecWithPrec(3, 2)
	val3, err = val3.SetInitialCommission(stakingtypes.NewCommission(threePercent, threePercent, threePercent))
	suite.NoError(err)
	validatorMap[val3Addr.String()] = val3

	sameValidatorSameInsuranceFeeLessId := func(validatorMap map[string]stakingtypes.Validator, a, b types.Insurance) bool {
		aValidator := validatorMap[a.ValidatorAddress]
		bValidator := validatorMap[b.ValidatorAddress]

		aFee := aValidator.Commission.Rate.Add(a.FeeRate)
		bFee := bValidator.Commission.Rate.Add(b.FeeRate)

		if !aFee.Equal(bFee) {
			return false
		}

		return a.Id < b.Id
	}

	sameValidatorLessInsuranceFee := func(validatorMap map[string]stakingtypes.Validator, a, b types.Insurance) bool {
		if a.ValidatorAddress != b.ValidatorAddress {
			return false
		}

		return a.FeeRate.LT(b.FeeRate)
	}

	lessTotalFee := func(validatorMap map[string]stakingtypes.Validator, a, b types.Insurance) bool {
		aValidator := validatorMap[a.ValidatorAddress]
		bValidator := validatorMap[b.ValidatorAddress]

		aFee := aValidator.Commission.Rate.Add(a.FeeRate)
		bFee := bValidator.Commission.Rate.Add(b.FeeRate)

		return aFee.LT(bFee)
	}

	cases := []struct {
		desc     string
		a, b     types.Insurance
		fn       func(validatorMap map[string]stakingtypes.Validator, a, b types.Insurance) bool
		expected bool
		descend  bool
	}{
		// ASCEND order
		{
			"same validator | same insurance fee | id a < b",
			types.NewInsurance(1, "", val1Addr.String(), fivePercent),
			types.NewInsurance(2, "", val1Addr.String(), fivePercent),
			sameValidatorSameInsuranceFeeLessId,
			true,
			false,
		},
		{
			"same validator | insurance fee a < b",
			types.NewInsurance(1, "", val1Addr.String(), threePercent),
			types.NewInsurance(2, "", val1Addr.String(), fivePercent),
			sameValidatorLessInsuranceFee,
			true,
			false,
		},
		{
			"same insurance fee | less validator fee a < b",
			types.NewInsurance(1, "", val3Addr.String(), threePercent),
			types.NewInsurance(2, "", val2Addr.String(), threePercent),
			lessTotalFee,
			true,
			false,
		},
		// DESCEND order
		{
			"same validator | same insurance fee | id b < a",
			types.NewInsurance(2, "", val1Addr.String(), fivePercent),
			types.NewInsurance(1, "", val1Addr.String(), fivePercent),
			sameValidatorSameInsuranceFeeLessId,
			true,
			true,
		},
		{
			"same validator | insurance fee b < a",
			types.NewInsurance(2, "", val1Addr.String(), fivePercent),
			types.NewInsurance(1, "", val1Addr.String(), threePercent),
			sameValidatorLessInsuranceFee,
			true,
			true,
		},
		{
			"same insurance fee | more validator fee",
			types.NewInsurance(2, "", val2Addr.String(), threePercent),
			types.NewInsurance(1, "", val3Addr.String(), threePercent),
			lessTotalFee,
			true,
			true,
		},
	}

	for _, tc := range cases {
		tc := tc
		suite.Run(tc.desc, func() {
			insurances := []types.Insurance{tc.b, tc.a}
			types.SortInsurances(validatorMap, insurances, tc.descend)
			if tc.descend {
				suite.Equal(
					tc.expected,
					tc.fn(validatorMap, insurances[1], insurances[0]),
				)
			} else {
				suite.Equal(
					tc.expected,
					tc.fn(validatorMap, insurances[0], insurances[1]),
				)
			}
		})
	}
}

func (suite *insuranceTestSuite) TestDerivedAddress() {
	i := types.NewInsurance(1, sdk.AccAddress("test").String(), sdk.ValAddress("testval").String(), sdk.NewDecWithPrec(5, 2))
	suite.Equal(
		sdk.AccAddress(crypto.AddressHash([]byte("liquidstakinginsurance1"))).String(),
		i.DerivedAddress().String(),
	)
	suite.Equal(
		"cosmos1p6qg4xu665ld3l8nr72z0vpsujf0s9ek9ln8gy",
		i.DerivedAddress().String(),
	)
}

func (suite *insuranceTestSuite) TestFeePoolAddress() {
	i := types.NewInsurance(1, sdk.AccAddress("test").String(), sdk.ValAddress("testval").String(), sdk.NewDecWithPrec(5, 2))
	suite.Equal(
		sdk.AccAddress(crypto.AddressHash([]byte("liquidstakinginsurancefee1"))).String(),
		i.FeePoolAddress().String(),
	)
	suite.Equal(
		"cosmos1fy0mcah0tcedpyqyz423mefdxh7zqz4gcfahxp",
		i.FeePoolAddress().String(),
	)
}

func (suite *insuranceTestSuite) TestGetProvider() {
	i := types.NewInsurance(1, sdk.AccAddress("test").String(), sdk.ValAddress("testval").String(), sdk.NewDecWithPrec(5, 2))
	suite.Equal(
		sdk.AccAddress("test").String(),
		i.GetProvider().String(),
	)
}

func (suite *insuranceTestSuite) TestGetValidator() {
	i := types.NewInsurance(1, sdk.AccAddress("test").String(), sdk.ValAddress("testval").String(), sdk.NewDecWithPrec(5, 2))
	suite.Equal(
		sdk.ValAddress("testval").String(),
		i.GetValidator().String(),
	)
}

func (suite *insuranceTestSuite) TestEqual() {
	i1 := types.NewInsurance(1, sdk.AccAddress("test").String(), sdk.ValAddress("testval").String(), sdk.NewDecWithPrec(5, 2))

	i2 := i1
	suite.True(i1.Equal(i2))
	i2.Id = 2
	suite.False(i1.Equal(i2))

	i2 = i1
	i2.ProviderAddress = sdk.AccAddress("test2").String()
	suite.False(i1.Equal(i2))

	i2 = i1
	i2.ValidatorAddress = sdk.ValAddress("testval2").String()
	suite.False(i1.Equal(i2))

	i2 = i1
	i2.FeeRate = sdk.NewDecWithPrec(6, 2)
	suite.False(i1.Equal(i2))
}

func (suite *insuranceTestSuite) TestSetStatus() {
	i := types.NewInsurance(1, sdk.AccAddress("test").String(), sdk.ValAddress("testval").String(), sdk.NewDecWithPrec(5, 2))
	suite.Equal(types.INSURANCE_STATUS_PAIRING, i.Status)
	i.SetStatus(types.INSURANCE_STATUS_UNPAIRING)
	suite.Equal(types.INSURANCE_STATUS_UNPAIRING, i.Status)
}

func (suite *insuranceTestSuite) TestValidate() {
	i := types.NewInsurance(3, sdk.AccAddress("test").String(), sdk.ValAddress("testval").String(), sdk.NewDecWithPrec(5, 2))
	suite.NoError(i.Validate(3))
	suite.Error(i.Validate(2))
	i.SetStatus(types.INSURANCE_STATUS_UNSPECIFIED)
	suite.Error(i.Validate(3))
}

func (suite *insuranceTestSuite) TestHasChunk() {
	i := types.NewInsurance(3, sdk.AccAddress("test").String(), sdk.ValAddress("testval").String(), sdk.NewDecWithPrec(5, 2))
	i.ChunkId = 1
	suite.True(i.HasChunk())

	i.EmptyChunk()
	suite.False(i.HasChunk())
}

func (suite *insuranceTestSuite) TestEmptyChunk() {
	i := types.NewInsurance(3, sdk.AccAddress("test").String(), sdk.ValAddress("testval").String(), sdk.NewDecWithPrec(5, 2))
	i.ChunkId = 1
	suite.True(i.HasChunk())

	i.EmptyChunk()
	suite.False(i.HasChunk())
	suite.Equal(types.Empty, i.ChunkId)
}
