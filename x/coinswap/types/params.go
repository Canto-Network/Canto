package types

import (
	"fmt"

	"gopkg.in/yaml.v2"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

const (
	UsdcIBCDenom = "ibc/17CD484EE7D9723B847D95015FA3EBD1572FD13BC84FB838F55B18A57450F25B"
	UsdtIBCDenom = "ibc/4F6A2DEFEA52CD8D90966ADCB2BD0593D3993AB0DF7F6AEB3EFD6167D79237B0"
	EthIBCDenom  = "ibc/DC186CA7A8C009B43774EBDC825C935CABA9743504CE6037507E6E5CCE12858A"
)

// Parameter store keys
var (
	KeyFee                    = []byte("Fee")                    // fee key
	KeyPoolCreationFee        = []byte("PoolCreationFee")        // fee key
	KeyTaxRate                = []byte("TaxRate")                // fee key
	KeyStandardDenom          = []byte("StandardDenom")          // standard token denom key
	KeyMaxStandardCoinPerPool = []byte("MaxStandardCoinPerPool") // max standard coin amount per pool
	KeyMaxSwapAmount          = []byte("MaxSwapAmount")          // whitelisted denoms

	DefaultFee                    = sdkmath.LegacyNewDecWithPrec(0, 0)
	DefaultPoolCreationFee        = sdk.NewInt64Coin(sdk.DefaultBondDenom, 0)
	DefaultTaxRate                = sdkmath.LegacyNewDecWithPrec(0, 0)
	DefaultMaxStandardCoinPerPool = sdkmath.NewIntWithDecimal(10000, 18)
	DefaultMaxSwapAmount          = sdk.NewCoins(
		sdk.NewCoin(UsdcIBCDenom, sdkmath.NewIntWithDecimal(10, 6)),
		sdk.NewCoin(UsdtIBCDenom, sdkmath.NewIntWithDecimal(10, 6)),
		sdk.NewCoin(EthIBCDenom, sdkmath.NewIntWithDecimal(1, 16)),
	)
)

// NewParams is the coinswap params constructor
func NewParams(fee, taxRate sdkmath.LegacyDec, poolCreationFee sdk.Coin, maxStandardCoinPerPool sdkmath.Int, maxSwapAmount sdk.Coins) Params {
	return Params{
		Fee:                    fee,
		TaxRate:                taxRate,
		PoolCreationFee:        poolCreationFee,
		MaxStandardCoinPerPool: maxStandardCoinPerPool,
		MaxSwapAmount:          maxSwapAmount,
	}
}

// ParamKeyTable returns the TypeTable for coinswap module
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

// ParamSetPairs implements paramtypes.KeyValuePairs
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(KeyFee, &p.Fee, validateFee),
		paramtypes.NewParamSetPair(KeyPoolCreationFee, &p.PoolCreationFee, validatePoolCreationFee),
		paramtypes.NewParamSetPair(KeyTaxRate, &p.TaxRate, validateTaxRate),
		paramtypes.NewParamSetPair(KeyMaxStandardCoinPerPool, &p.MaxStandardCoinPerPool, validateMaxStandardCoinPerPool),
		paramtypes.NewParamSetPair(KeyMaxSwapAmount, &p.MaxSwapAmount, validateMaxSwapAmount),
	}
}

// DefaultParams returns the default coinswap module parameters
func DefaultParams() Params {
	return Params{
		Fee:                    DefaultFee,
		PoolCreationFee:        DefaultPoolCreationFee,
		TaxRate:                DefaultTaxRate,
		MaxStandardCoinPerPool: DefaultMaxStandardCoinPerPool,
		MaxSwapAmount:          DefaultMaxSwapAmount,
	}
}

// String returns a human readable string representation of the parameters.
func (p Params) String() string {
	out, _ := yaml.Marshal(p)
	return string(out)
}

// Validate returns err if Params is invalid
func (p Params) Validate() error {
	if p.Fee.IsNegative() || !p.Fee.LT(sdkmath.LegacyOneDec()) {
		return fmt.Errorf("fee must be positive and less than 1: %s", p.Fee.String())
	}
	return nil
}

func validateFee(i interface{}) error {
	v, ok := i.(sdkmath.LegacyDec)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v.IsNegative() || !v.LT(sdkmath.LegacyOneDec()) {
		return fmt.Errorf("fee must be positive and less than 1: %s", v.String())
	}

	return nil
}

func validatePoolCreationFee(i interface{}) error {
	v, ok := i.(sdk.Coin)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v.IsNegative() {
		return fmt.Errorf("poolCreationFee must be positive: %s", v.String())
	}
	return nil
}

func validateTaxRate(i interface{}) error {
	v, ok := i.(sdkmath.LegacyDec)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v.IsNegative() || !v.LT(sdkmath.LegacyOneDec()) {
		return fmt.Errorf("fee must be positive and less than 1: %s", v.String())
	}
	return nil
}

func validateMaxStandardCoinPerPool(i interface{}) error {
	v, ok := i.(sdkmath.Int)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if !v.IsPositive() {
		return fmt.Errorf("maxStandardCoinPerPool must be positive: %s", v.String())
	}
	return nil
}

func validateMaxSwapAmount(i interface{}) error {
	v, ok := i.(sdk.Coins)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if err := v.Validate(); err != nil {
		return err
	}

	for _, coin := range v {
		// do something with the coin object, such as print its denomination and amount
		if err := sdk.ValidateDenom(coin.Denom); err != nil {
			return err
		}

		if coin.Amount.LT(sdkmath.ZeroInt()) {
			return fmt.Errorf("coin amount must be positive")
		}
	}

	return nil
}
