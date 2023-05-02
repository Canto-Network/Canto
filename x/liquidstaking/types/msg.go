package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var (
	_ sdk.Msg = &MsgLiquidStake{}
	_ sdk.Msg = &MsgLiquidUnstake{}
	_ sdk.Msg = &MsgInsuranceProvide{}
	_ sdk.Msg = &MsgCancelInsuranceProvide{}
	_ sdk.Msg = &MsgDepositInsurance{}
	_ sdk.Msg = &MsgWithdrawInsurance{}
	_ sdk.Msg = &MsgWithdrawInsuranceCommission{}
)

const (
	TypeMsgLiquidStake                 = "liquid_stake"
	TypeMsgLiquidUnstake               = "liquid_unstake"
	TypeMsgInsuranceProvide            = "insurance_provide"
	TypeMsgCancelInsurance             = "cancel_insurance"
	TypeMsgDepositInsurance            = "deposit_insurance"
	TypeMsgWithdrawInsurance           = "withdraw_insurance"
	TypeMsgWithdrawInsuranceCommission = "withdraw_insurance_commission"
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
	if !msg.Amount.IsValid() {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidCoins, "invalid amount %s", msg.Amount)
	}
	return nil
}
func (msg MsgLiquidStake) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}
func (msg MsgLiquidStake) GetSigners() []sdk.AccAddress {
	funder := sdk.MustAccAddressFromBech32(msg.DelegatorAddress)
	return []sdk.AccAddress{funder}
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
	if !msg.Amount.IsValid() {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidCoins, "invalid amount %s", msg.Amount)
	}
	return nil
}
func (msg MsgLiquidUnstake) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}
func (msg MsgLiquidUnstake) GetSigners() []sdk.AccAddress {
	funder := sdk.MustAccAddressFromBech32(msg.DelegatorAddress)
	return []sdk.AccAddress{funder}
}

func (msg MsgLiquidUnstake) GetDelegator() sdk.AccAddress {
	addr := sdk.MustAccAddressFromBech32(msg.DelegatorAddress)
	return addr
}

func NewMsgInsuranceProvide(providerAddress string, amount sdk.Coin) *MsgInsuranceProvide {
	return &MsgInsuranceProvide{
		ProviderAddress: providerAddress,
		Amount:          amount,
	}
}
func (msg MsgInsuranceProvide) Route() string { return RouterKey }
func (msg MsgInsuranceProvide) Type() string  { return TypeMsgInsuranceProvide }
func (msg MsgInsuranceProvide) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.ProviderAddress); err != nil {
		return sdkerrors.Wrapf(err, "invalid provider address %s", msg.ProviderAddress)
	}
	if !msg.Amount.IsValid() {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidCoins, "invalid amount %s", msg.Amount)
	}
	return nil
}
func (msg MsgInsuranceProvide) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}
func (msg MsgInsuranceProvide) GetSigners() []sdk.AccAddress {
	funder := sdk.MustAccAddressFromBech32(msg.ProviderAddress)
	return []sdk.AccAddress{funder}
}

func (msg MsgInsuranceProvide) GetProvider() sdk.AccAddress {
	addr := sdk.MustAccAddressFromBech32(msg.ProviderAddress)
	return addr
}

func (msg MsgInsuranceProvide) GetValidator() sdk.ValAddress {
	addr, _ := sdk.ValAddressFromBech32(msg.ValidatorAddress)
	return addr
}

func NewMsgCancelInsuranceProvide(providerAddress string, insuranceId uint64) *MsgCancelInsuranceProvide {
	return &MsgCancelInsuranceProvide{
		ProviderAddress: providerAddress,
		Id:              insuranceId,
	}
}
func (msg MsgCancelInsuranceProvide) Route() string { return RouterKey }
func (msg MsgCancelInsuranceProvide) Type() string  { return TypeMsgCancelInsurance }
func (msg MsgCancelInsuranceProvide) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.ProviderAddress); err != nil {
		return sdkerrors.Wrapf(err, "invalid provider address %s", msg.ProviderAddress)
	}
	return nil
}
func (msg MsgCancelInsuranceProvide) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}
func (msg MsgCancelInsuranceProvide) GetSigners() []sdk.AccAddress {
	funder := sdk.MustAccAddressFromBech32(msg.ProviderAddress)
	return []sdk.AccAddress{funder}
}

func (msg MsgCancelInsuranceProvide) GetProvider() sdk.AccAddress {
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
	if !msg.Amount.IsValid() {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidCoins, "invalid amount %s", msg.Amount)
	}
	return nil
}
func (msg MsgDepositInsurance) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}
func (msg MsgDepositInsurance) GetSigners() []sdk.AccAddress {
	funder := sdk.MustAccAddressFromBech32(msg.ProviderAddress)
	return []sdk.AccAddress{funder}
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
	return nil
}
func (msg MsgWithdrawInsurance) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}
func (msg MsgWithdrawInsurance) GetSigners() []sdk.AccAddress {
	funder := sdk.MustAccAddressFromBech32(msg.ProviderAddress)
	return []sdk.AccAddress{funder}
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
	return nil
}
func (msg MsgWithdrawInsuranceCommission) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}
func (msg MsgWithdrawInsuranceCommission) GetSigners() []sdk.AccAddress {
	funder := sdk.MustAccAddressFromBech32(msg.ProviderAddress)
	return []sdk.AccAddress{funder}
}

func (msg MsgWithdrawInsuranceCommission) GetProvider() sdk.AccAddress {
	addr := sdk.MustAccAddressFromBech32(msg.ProviderAddress)
	return addr
}
