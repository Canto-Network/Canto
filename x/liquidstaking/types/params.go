package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

var (
	KeyDynamicFeeRate = []byte("DynamicFeeRate")

	DefaultR0       = sdk.ZeroDec()
	DefaultUSoftCap = sdk.MustNewDecFromStr("0.05")
	DefaultUHardCap = sdk.MustNewDecFromStr("0.1")
	DefaultUOptimal = sdk.MustNewDecFromStr("0.09")
	DefaultSlope1   = sdk.MustNewDecFromStr("0.1")
	DefaultSlope2   = sdk.MustNewDecFromStr("0.4")
	DefaultMaxFee   = sdk.MustNewDecFromStr("0.5")
)

var _ paramtypes.ParamSet = &Params{}

// ParamKeyTable returns the parameter key table.
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

// NewParams creates a new Params object
func NewParams(
	dynamicFeeRate DynamicFeeRate,
	// r0, uSoftCap, uHardCap, uOptimal, slope1, slope2, maxFeeRate sdk.Dec,
) Params {
	return Params{
		dynamicFeeRate,
	}
}

func DefaultParams() Params {
	return NewParams(
		DynamicFeeRate{
			R0:         DefaultR0,
			USoftCap:   DefaultUSoftCap,
			UHardCap:   DefaultUHardCap,
			UOptimal:   DefaultUOptimal,
			Slope1:     DefaultSlope1,
			Slope2:     DefaultSlope2,
			MaxFeeRate: DefaultMaxFee,
		},
	)
}

// ParamSetPairs returns the parameter set pairs.
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(KeyDynamicFeeRate, &p.DynamicFeeRate, validateDynamicFeeRate),
	}
}

func (p Params) Validate() error {
	for _, v := range []struct {
		value     interface{}
		validator func(interface{}) error
	}{
		{
			p.DynamicFeeRate, validateDynamicFeeRate,
		},
	} {
		if err := v.validator(v.value); err != nil {
			return err
		}
	}
	return nil
}

// TODO: Write test codes for it right now!!
func validateR0(i interface{}) error {
	v, ok := i.(sdk.Dec)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v.IsNil() {
		return fmt.Errorf("r0 should not be nil")
	}

	if v.IsNegative() {
		return fmt.Errorf("r0 should not be negative")
	}

	if v.GT(sdk.OneDec()) {
		return fmt.Errorf("r0 should not be greater than 1")
	}

	return nil
}

func validateUSoftCap(i interface{}) error {
	v, ok := i.(sdk.Dec)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v.IsNil() {
		return fmt.Errorf("uSoftCap should not be nil")
	}

	if v.IsNegative() {
		return fmt.Errorf("uSoftCap should not be negative")
	}

	if v.GT(sdk.OneDec()) {
		return fmt.Errorf("uSoftCap should not be greater than 1")
	}

	return nil
}

func validateUHardCap(i interface{}) error {
	v, ok := i.(sdk.Dec)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v.IsNil() {
		return fmt.Errorf("uHardCap should not be nil")
	}

	if v.IsNegative() {
		return fmt.Errorf("uHardCap should not be negative")
	}

	if v.GT(sdk.OneDec()) {
		return fmt.Errorf("uHardCap should not be greater than 1")
	}

	return nil
}

func validateUOptimal(i interface{}) error {
	v, ok := i.(sdk.Dec)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v.IsNil() {
		return fmt.Errorf("uOptimal should not be nil")
	}

	if v.IsNegative() {
		return fmt.Errorf("uOptimal should not be negative")
	}

	if v.GT(sdk.OneDec()) {
		return fmt.Errorf("uOptimal should not be greater than 1")
	}

	return nil
}

func validateSlope1(i interface{}) error {
	v, ok := i.(sdk.Dec)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v.IsNil() {
		return fmt.Errorf("slope1 should not be nil")
	}

	if v.IsNegative() {
		return fmt.Errorf("slope1 should not be negative")
	}

	return nil
}

func validateSlope2(i interface{}) error {
	v, ok := i.(sdk.Dec)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v.IsNil() {
		return fmt.Errorf("slope2 should not be nil")
	}

	if v.IsNegative() {
		return fmt.Errorf("slope2 should not be negative")
	}

	return nil
}

func validateMaxFeeRate(i interface{}) error {
	v, ok := i.(sdk.Dec)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v.IsNil() {
		return fmt.Errorf("maxFeeRate should not be nil")
	}

	if v.IsNegative() {
		return fmt.Errorf("maxFeeRate should not be negative")
	}

	if v.GT(sdk.OneDec()) {
		return fmt.Errorf("maxFeeRate should not be greater than 1")
	}

	return nil
}

func validateDynamicFeeRate(i interface{}) (err error) {
	v, ok := i.(DynamicFeeRate)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if err = validateR0(v.R0); err != nil {
		return err
	}
	if err = validateUSoftCap(v.USoftCap); err != nil {
		return err
	}
	if err = validateUHardCap(v.UHardCap); err != nil {
		return err
	}
	if err = validateUOptimal(v.UOptimal); err != nil {
		return err
	}
	if err = validateSlope1(v.Slope1); err != nil {
		return err
	}
	if err = validateSlope2(v.Slope2); err != nil {
		return err
	}
	if err = validateMaxFeeRate(v.MaxFeeRate); err != nil {
		return err
	}

	// validate dynamic fee model
	if !v.USoftCap.LT(v.UOptimal) {
		return fmt.Errorf("uSoftCap should be less than uOptimal")
	}
	if !v.UOptimal.LT(v.UHardCap) {
		return fmt.Errorf("uOptimal should be less than uHardCap")
	}
	if !v.R0.Add(v.Slope1).Add(v.Slope2).LTE(v.MaxFeeRate) {
		return fmt.Errorf("r0 + slope1 + slope2 should not exceeds maxFeeRate")
	}
	return nil
}
