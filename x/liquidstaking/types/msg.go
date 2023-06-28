package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var (
	_ sdk.Msg = &MsgLiquidStake{}
	_ sdk.Msg = &MsgLiquidUnstake{}
	_ sdk.Msg = &MsgProvideInsurance{}
	_ sdk.Msg = &MsgCancelProvideInsurance{}
	_ sdk.Msg = &MsgDepositInsurance{}
	_ sdk.Msg = &MsgWithdrawInsurance{}
	_ sdk.Msg = &MsgWithdrawInsuranceCommission{}
	_ sdk.Msg = &MsgClaimDiscountedReward{}
)

const (
	TypeMsgLiquidStake                 = "liquid_stake"
	TypeMsgLiquidUnstake               = "liquid_unstake"
	TypeMsgProvideInsurance            = "provide_insurance"
	TypeMsgCancelProvideInsurance      = "cancel_provide_insurance"
	TypeMsgDepositInsurance            = "deposit_insurance"
	TypeMsgWithdrawInsurance           = "withdraw_insurance"
	TypeMsgWithdrawInsuranceCommission = "withdraw_insurance_commission"
	TypeMsgClaimDiscountedReward       = "claim_discounted_reward"
)

func NewMsgLiquidStake(delegatorAddress string, amount sdk.Coin) *MsgLiquidStake {
	return &MsgLiquidStake{
		DelegatorAddress: delegatorAddress,
		Amount:           amount,
	}
}
func (msg MsgLiquidStake) Route() string { return RouterKey }
func (msg MsgLiquidStake) Type() string  { return TypeMsgLiquidStake }
func (msg MsgLiquidStake) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.DelegatorAddress); err != nil {
		return sdkerrors.Wrapf(err, "invalid delegator address %s", msg.DelegatorAddress)
	}
	if ok := msg.Amount.IsZero(); ok {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "staking amount must not be zero")
	}
	if err := msg.Amount.Validate(); err != nil {
		return err
	}
	return nil
}
func (msg MsgLiquidStake) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}
func (msg MsgLiquidStake) GetSigners() []sdk.AccAddress {
	delegator := sdk.MustAccAddressFromBech32(msg.DelegatorAddress)
	return []sdk.AccAddress{delegator}
}

func (msg MsgLiquidStake) GetDelegator() sdk.AccAddress {
	addr := sdk.MustAccAddressFromBech32(msg.DelegatorAddress)
	return addr
}

func NewMsgLiquidUnstake(delegatorAddress string, amount sdk.Coin) *MsgLiquidUnstake {
	return &MsgLiquidUnstake{
		DelegatorAddress: delegatorAddress,
		Amount:           amount,
	}
}
func (msg MsgLiquidUnstake) Route() string { return RouterKey }
func (msg MsgLiquidUnstake) Type() string  { return TypeMsgLiquidUnstake }
func (msg MsgLiquidUnstake) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.DelegatorAddress); err != nil {
		return sdkerrors.Wrapf(err, "invalid delegator address %s", msg.DelegatorAddress)
	}
	if ok := msg.Amount.IsZero(); ok {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "unstaking amount must not be zero")
	}
	if err := msg.Amount.Validate(); err != nil {
		return err
	}
	return nil
}
func (msg MsgLiquidUnstake) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}
func (msg MsgLiquidUnstake) GetSigners() []sdk.AccAddress {
	delegator := sdk.MustAccAddressFromBech32(msg.DelegatorAddress)
	return []sdk.AccAddress{delegator}
}

func (msg MsgLiquidUnstake) GetDelegator() sdk.AccAddress {
	addr := sdk.MustAccAddressFromBech32(msg.DelegatorAddress)
	return addr
}

func NewMsgProvideInsurance(providerAddress, validatorAddress string, amount sdk.Coin, feeRate sdk.Dec) *MsgProvideInsurance {
	return &MsgProvideInsurance{
		ProviderAddress:  providerAddress,
		ValidatorAddress: validatorAddress,
		Amount:           amount,
		FeeRate:          feeRate,
	}
}
func (msg MsgProvideInsurance) Route() string { return RouterKey }
func (msg MsgProvideInsurance) Type() string  { return TypeMsgProvideInsurance }
func (msg MsgProvideInsurance) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.ProviderAddress); err != nil {
		return sdkerrors.Wrapf(err, "invalid provider address %s", msg.ProviderAddress)
	}
	if _, err := sdk.ValAddressFromBech32(msg.ValidatorAddress); err != nil {
		return sdkerrors.Wrapf(err, "invalid validator address %s", msg.ValidatorAddress)
	}
	if ok := msg.Amount.IsZero(); ok {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "collateral amount must not be zero")
	}
	if err := msg.Amount.Validate(); err != nil {
		return err
	}
	if msg.FeeRate.IsNil() {
		return ErrInvalidFeeRate
	}
	return nil
}
func (msg MsgProvideInsurance) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}
func (msg MsgProvideInsurance) GetSigners() []sdk.AccAddress {
	provider := sdk.MustAccAddressFromBech32(msg.ProviderAddress)
	return []sdk.AccAddress{provider}
}

func (msg MsgProvideInsurance) GetProvider() sdk.AccAddress {
	addr := sdk.MustAccAddressFromBech32(msg.ProviderAddress)
	return addr
}

func (msg MsgProvideInsurance) GetValidator() sdk.ValAddress {
	addr, _ := sdk.ValAddressFromBech32(msg.ValidatorAddress)
	return addr
}

func NewMsgCancelProvideInsurance(providerAddress string, insuranceId uint64) *MsgCancelProvideInsurance {
	return &MsgCancelProvideInsurance{
		ProviderAddress: providerAddress,
		Id:              insuranceId,
	}
}
func (msg MsgCancelProvideInsurance) Route() string { return RouterKey }
func (msg MsgCancelProvideInsurance) Type() string  { return TypeMsgCancelProvideInsurance }
func (msg MsgCancelProvideInsurance) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.ProviderAddress); err != nil {
		return sdkerrors.Wrapf(err, "invalid provider address %s", msg.ProviderAddress)
	}
	if msg.Id < 1 {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "invalid insurance id: %d", msg.Id)
	}
	return nil
}
func (msg MsgCancelProvideInsurance) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}
func (msg MsgCancelProvideInsurance) GetSigners() []sdk.AccAddress {
	provider := sdk.MustAccAddressFromBech32(msg.ProviderAddress)
	return []sdk.AccAddress{provider}
}
func (msg MsgCancelProvideInsurance) GetProvider() sdk.AccAddress {
	addr := sdk.MustAccAddressFromBech32(msg.ProviderAddress)
	return addr
}

func NewMsgDepositInsurance(providerAddress string, insuranceId uint64, amount sdk.Coin) *MsgDepositInsurance {
	return &MsgDepositInsurance{
		Id:              insuranceId,
		ProviderAddress: providerAddress,
		Amount:          amount,
	}
}
func (msg MsgDepositInsurance) Route() string { return RouterKey }
func (msg MsgDepositInsurance) Type() string  { return TypeMsgDepositInsurance }
func (msg MsgDepositInsurance) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.ProviderAddress); err != nil {
		return sdkerrors.Wrapf(err, "invalid provider address %s", msg.ProviderAddress)
	}
	if msg.Id < 1 {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "invalid insurance id: %d", msg.Id)
	}
	if ok := msg.Amount.IsZero(); ok {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "deposit amount must not be zero")
	}
	if err := msg.Amount.Validate(); err != nil {
		return err
	}
	return nil
}
func (msg MsgDepositInsurance) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}
func (msg MsgDepositInsurance) GetSigners() []sdk.AccAddress {
	provider := sdk.MustAccAddressFromBech32(msg.ProviderAddress)
	return []sdk.AccAddress{provider}
}

func (msg MsgDepositInsurance) GetProvider() sdk.AccAddress {
	addr := sdk.MustAccAddressFromBech32(msg.ProviderAddress)
	return addr
}

func NewMsgWithdrawInsurance(providerAddress string, insuranceId uint64) *MsgWithdrawInsurance {
	return &MsgWithdrawInsurance{
		ProviderAddress: providerAddress,
		Id:              insuranceId,
	}
}
func (msg MsgWithdrawInsurance) Route() string { return RouterKey }
func (msg MsgWithdrawInsurance) Type() string  { return TypeMsgWithdrawInsurance }
func (msg MsgWithdrawInsurance) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.ProviderAddress); err != nil {
		return sdkerrors.Wrapf(err, "invalid provider address %s", msg.ProviderAddress)
	}
	if msg.Id < 1 {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "invalid insurance id: %d", msg.Id)
	}
	return nil
}
func (msg MsgWithdrawInsurance) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}
func (msg MsgWithdrawInsurance) GetSigners() []sdk.AccAddress {
	provider := sdk.MustAccAddressFromBech32(msg.ProviderAddress)
	return []sdk.AccAddress{provider}
}

func (msg MsgWithdrawInsurance) GetProvider() sdk.AccAddress {
	addr := sdk.MustAccAddressFromBech32(msg.ProviderAddress)
	return addr
}

func NewMsgWithdrawInsuranceCommission(providerAddress string, insuranceId uint64) *MsgWithdrawInsuranceCommission {
	return &MsgWithdrawInsuranceCommission{
		ProviderAddress: providerAddress,
		Id:              insuranceId,
	}
}
func (msg MsgWithdrawInsuranceCommission) Route() string { return RouterKey }
func (msg MsgWithdrawInsuranceCommission) Type() string  { return TypeMsgWithdrawInsuranceCommission }
func (msg MsgWithdrawInsuranceCommission) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.ProviderAddress); err != nil {
		return sdkerrors.Wrapf(err, "invalid provider address %s", msg.ProviderAddress)
	}
	if msg.Id < 1 {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "invalid insurance id: %d", msg.Id)
	}
	return nil
}
func (msg MsgWithdrawInsuranceCommission) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}
func (msg MsgWithdrawInsuranceCommission) GetSigners() []sdk.AccAddress {
	provider := sdk.MustAccAddressFromBech32(msg.ProviderAddress)
	return []sdk.AccAddress{provider}
}

func (msg MsgWithdrawInsuranceCommission) GetProvider() sdk.AccAddress {
	addr := sdk.MustAccAddressFromBech32(msg.ProviderAddress)
	return addr
}

func NewMsgClaimDiscountedReward(requesterAddress string, amount sdk.Coin, minimumDiscountRate sdk.Dec) *MsgClaimDiscountedReward {
	return &MsgClaimDiscountedReward{
		RequesterAddress:    requesterAddress,
		Amount:              amount,
		MinimumDiscountRate: minimumDiscountRate,
	}
}
func (msg MsgClaimDiscountedReward) Route() string { return RouterKey }
func (msg MsgClaimDiscountedReward) Type() string  { return TypeMsgClaimDiscountedReward }
func (msg MsgClaimDiscountedReward) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.RequesterAddress); err != nil {
		return sdkerrors.Wrapf(err, "invalid requester address %s", msg.RequesterAddress)
	}
	if ok := msg.Amount.IsZero(); ok {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "maximum allowed ls tokens to pay must not be zero")
	}
	if err := msg.Amount.Validate(); err != nil {
		return err
	}
	if msg.MinimumDiscountRate.IsNil() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "minimum discount rate must not be nil")
	}
	if msg.MinimumDiscountRate.IsNegative() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "minimum discount rate must not be negative")
	}
	return nil
}
func (msg MsgClaimDiscountedReward) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}
func (msg MsgClaimDiscountedReward) GetSigners() []sdk.AccAddress {
	requester := sdk.MustAccAddressFromBech32(msg.RequesterAddress)
	return []sdk.AccAddress{requester}
}

func (msg MsgClaimDiscountedReward) GetRequestser() sdk.AccAddress {
	addr := sdk.MustAccAddressFromBech32(msg.RequesterAddress)
	return addr
}
