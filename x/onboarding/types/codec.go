package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
)

var (
	amino     = codec.NewLegacyAmino()
	ModuleCdc = codec.NewProtoCodec(codectypes.NewInterfaceRegistry())
)

// method required for x/csr msg GetSignBytes methods
func init() {
	RegisterLegacyAminoCodec(amino)
	amino.Seal()
}

// register interfaces for the AminoCodec
func RegisterInterfaces(registry codectypes.InterfaceRegistry) {
	registry.RegisterImplementations(
		(*sdk.Msg)(nil),
		&MsgUpdateParams{},
	)

	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}

// register csr msg types for Amino Codec in adherence to EIP-712 signing conventions
func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&MsgUpdateParams{}, "canto/x/onboarding/MsgUpdateParams", nil)
	cdc.RegisterConcrete(&Params{}, "canto/x/onboarding/Params", nil)
}
