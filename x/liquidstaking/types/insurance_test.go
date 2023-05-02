package types_test

import (
	"github.com/Canto-Network/Canto/v6/x/liquidstaking/types"
	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestSortInsurances(t *testing.T) {
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
	require.NoError(t, err)
	fivePercent := sdk.NewDecWithPrec(5, 2)
	val1, err = val1.SetInitialCommission(stakingtypes.NewCommission(fivePercent, fivePercent, fivePercent))
	validatorMap[val1Addr.String()] = val1

	val2, err = stakingtypes.NewValidator(
		val2Addr,
		publicKeys[1],
		stakingtypes.Description{},
	)
	require.NoError(t, err)
	sevenPercent := sdk.NewDecWithPrec(7, 2)
	val2, err = val2.SetInitialCommission(stakingtypes.NewCommission(sevenPercent, sevenPercent, sevenPercent))
	validatorMap[val2Addr.String()] = val2

	val3, err = stakingtypes.NewValidator(
		val3Addr,
		publicKeys[2],
		stakingtypes.Description{},
	)
	require.NoError(t, err)
	threePercent := sdk.NewDecWithPrec(3, 2)
	val3, err = val3.SetInitialCommission(stakingtypes.NewCommission(threePercent, threePercent, threePercent))
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
		t.Run(tc.desc, func(t *testing.T) {
			insurances := []types.Insurance{tc.b, tc.a}
			types.SortInsurances(validatorMap, insurances, tc.descend)
			if tc.descend {
				require.Equal(t, tc.expected, tc.fn(validatorMap, insurances[1], insurances[0]))
			} else {
				require.Equal(t, tc.expected, tc.fn(validatorMap, insurances[0], insurances[1]))
			}
		})
	}
}
