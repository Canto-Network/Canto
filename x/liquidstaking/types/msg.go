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
)

const (
	TypeMsgLiquidStake                 = "liquid_stake"
	TypeMsgLiquidUnstake               = "liquid_unstake"
	TypeMsgProvideInsurance            = "insurance_provide"
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

func NewMsgProvideInsurance(providerAddress string, amount sdk.Coin) *MsgProvideInsurance {
	return &MsgProvideInsurance{
		ProviderAddress: providerAddress,
		Amount:          amount,
	}
}
func (msg MsgProvideInsurance) Route() string { return RouterKey }
func (msg MsgProvideInsurance) Type() string  { return TypeMsgProvideInsurance }
func (msg MsgProvideInsurance) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.ProviderAddress); err != nil {
		return sdkerrors.Wrapf(err, "invalid provider address %s", msg.ProviderAddress)
	}
	if !msg.Amount.IsValid() {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidCoins, "invalid amount %s", msg.Amount)
	}
	return nil
}
func (msg MsgProvideInsurance) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}
func (msg MsgProvideInsurance) GetSigners() []sdk.AccAddress {
	funder := sdk.MustAccAddressFromBech32(msg.ProviderAddress)
	return []sdk.AccAddress{funder}
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
func (msg MsgCancelProvideInsurance) Type() string  { return TypeMsgCancelInsurance }
func (msg MsgCancelProvideInsurance) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.ProviderAddress); err != nil {
		return sdkerrors.Wrapf(err, "invalid provider address %s", msg.ProviderAddress)
	}
	return nil
}
func (msg MsgCancelProvideInsurance) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}
func (msg MsgCancelProvideInsurance) GetSigners() []sdk.AccAddress {
	funder := sdk.MustAccAddressFromBech32(msg.ProviderAddress)
	return []sdk.AccAddress{funder}
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
