package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
)

var (
	amino = codec.NewLegacyAmino()

	// ModuleCdc references the global incentives module codec. Note, the codec
	// should ONLY be used in certain instances of tests and for JSON encoding.
	//
	// The actual codec used for serialization should be provided to
	// modules/incentives and defined at the application level.
	ModuleCdc = codec.NewProtoCodec(codectypes.NewInterfaceRegistry())

	// AminoCdc is a amino codec created to support amino JSON compatible msgs.
	AminoCdc = codec.NewAminoCodec(amino)
)

const (
	liquidStakeName                 = "liquidstaking/MsgLiquidStake"
	liquidUnstakeName               = "liquidstaking/MsgLiquidUnstake"
	ProvideInsuranceName            = "liquidstaking/MsgProvideInsurance"
	cancelProvideInsuranceName      = "liquidstaking/MsgCancelProvideInsurance"
	depositInsuranceName            = "liquidstaking/MsgDepositInsurance"
	withdrawInsuranceName           = "liquidstaking/MsgWithdrawInsurance"
	withdrawInsuranceCommissionName = "liquidstaking/MsgWithdrawInsuranceCommission"
)

func init() {
	RegisterLegacyAminoCodec(amino)
	amino.Seal()
}

// RegisterInterfaces register implementations
func RegisterInterfaces(registry codectypes.InterfaceRegistry) {
	registry.RegisterImplementations(
		(*sdk.Msg)(nil),
		&MsgLiquidStake{},
		&MsgLiquidUnstake{},
		&MsgProvideInsurance{},
		&MsgCancelProvideInsurance{},
		&MsgDepositInsurance{},
		&MsgWithdrawInsurance{},
		&MsgWithdrawInsuranceCommission{},
	)

	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}

// RegisterLegacyAminoCodec registers the necessary x/liquidstaking interfaces and
// concrete types on the provided LegacyAmino codec. These types are used for
// Amino JSON serialization.
func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&MsgLiquidStake{}, liquidStakeName, nil)
	cdc.RegisterConcrete(&MsgLiquidUnstake{}, liquidUnstakeName, nil)
	cdc.RegisterConcrete(&MsgProvideInsurance{}, ProvideInsuranceName, nil)
	cdc.RegisterConcrete(&MsgCancelProvideInsurance{}, cancelProvideInsuranceName, nil)
	cdc.RegisterConcrete(&MsgDepositInsurance{}, depositInsuranceName, nil)
	cdc.RegisterConcrete(&MsgWithdrawInsurance{}, withdrawInsuranceName, nil)
	cdc.RegisterConcrete(&MsgWithdrawInsuranceCommission{}, withdrawInsuranceCommissionName, nil)
}
