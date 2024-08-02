package types

import (
	errorsmod "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

var (
	DefaultEnableCSR = false
	DefaultCSRShares = sdkmath.LegacyNewDecWithPrec(20, 2)

	ParamStoreKeyEnableCSR = []byte("EnableCSR")
	ParamStoreKeyCSRShares = []byte("CSRShares")
)

// ParamKeyTable the param key table
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

// NewParams creates a new Params instance
func NewParams(enableCSR bool, csrShares sdkmath.LegacyDec) Params {
	return Params{
		EnableCsr: enableCSR,
		CsrShares: csrShares,
	}
}

// DefaultParams returns a default set of parameters
func DefaultParams() Params {
	return NewParams(DefaultEnableCSR, DefaultCSRShares)
}

// ParamSetPairs get the params.ParamSet
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(ParamStoreKeyEnableCSR, &p.EnableCsr, ValidateEnableCSR),
		paramtypes.NewParamSetPair(ParamStoreKeyCSRShares, &p.CsrShares, ValidateShares),
	}
}

// Validates the boolean which enables CSR
func ValidateEnableCSR(i interface{}) error {
	_, ok := i.(bool)
	if !ok {
		return errorsmod.Wrapf(ErrInvalidParams, "Params::Validate::ValidateEnableCSR enableCSR must be a bool")
	}

	return nil
}

// Validates the CSR share dec that is inputted
func ValidateShares(i interface{}) error {
	v, ok := i.(sdkmath.LegacyDec)

	if !ok {
		return errorsmod.Wrapf(ErrInvalidParams, "Params::Validate::ValidateShares CSRShares must be of type sdkmath.LegacyDec")
	}

	if v.IsNil() {
		return errorsmod.Wrapf(ErrInvalidParams, "Params::Validate::ValidateShares cannot have entry that is nil for CSRShares")
	}

	if v.IsNegative() {
		return errorsmod.Wrapf(ErrInvalidParams, "Params::Validate::ValidateShares CSRShares cannot be negative")
	}

	if v.GT(sdkmath.LegacyOneDec()) {
		return errorsmod.Wrapf(ErrInvalidParams, "Params::Validate::ValidateShares CSRShares cannot be greater than 1")
	}

	return nil
}

func (p Params) Validate() error {
	if err := ValidateEnableCSR(p.EnableCsr); err != nil {
		return err
	}
	return ValidateShares(p.CsrShares)
}
