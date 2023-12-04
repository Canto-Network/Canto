package types

import (
	"fmt"

	sdkmath "cosmossdk.io/math"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

// Parameter store key
var (
	ParamStoreKeyEnableOnboarding     = []byte("EnableOnboarding")
	ParamStoreKeyAutoSwapThreshold    = []byte("AutoSwapThreshold")
	ParamsStoreKeyWhitelistedChannels = []byte("WhitelistedChannels")
	DefaultAutoSwapThreshold          = sdkmath.NewIntWithDecimal(4, 18) // 4 Canto
	DefaultWhitelistedChannels        = []string{"channel-0"}
)

var _ paramtypes.ParamSet = &Params{}

// ParamKeyTable returns the parameter key table.
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

// NewParams creates a new Params instance
func NewParams(
	enableOnboarding bool,
	autoSwapThreshold sdkmath.Int,
	whitelistedChannels []string,
) Params {
	return Params{
		EnableOnboarding:    enableOnboarding,
		AutoSwapThreshold:   autoSwapThreshold,
		WhitelistedChannels: whitelistedChannels,
	}
}

// DefaultParams defines the default params for the onboarding module
func DefaultParams() Params {
	return Params{
		EnableOnboarding:    true,
		AutoSwapThreshold:   DefaultAutoSwapThreshold,
		WhitelistedChannels: DefaultWhitelistedChannels,
	}
}

// ParamSetPairs returns the parameter set pairs.
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(ParamStoreKeyEnableOnboarding, &p.EnableOnboarding, validateBool),
		paramtypes.NewParamSetPair(ParamStoreKeyAutoSwapThreshold, &p.AutoSwapThreshold, validateAutoSwapThreshold),
		paramtypes.NewParamSetPair(ParamsStoreKeyWhitelistedChannels, &p.WhitelistedChannels, validateWhitelistedChannels),
	}
}

func validateWhitelistedChannels(i interface{}) error {
	_, ok := i.([]string)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	return nil
}

func validateBool(i interface{}) error {
	_, ok := i.(bool)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	return nil
}

func validateAutoSwapThreshold(i interface{}) error {
	v, ok := i.(sdkmath.Int)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v.IsNegative() {
		return fmt.Errorf("auto swap threshold must be positive: %s", v.String())
	}

	return nil
}

// Validate checks that the fields have valid values
func (p Params) Validate() error {
	if err := validateBool(p.EnableOnboarding); err != nil {
		return err
	}
	if err := validateAutoSwapThreshold(p.AutoSwapThreshold); err != nil {
		return err
	}
	return nil
}
