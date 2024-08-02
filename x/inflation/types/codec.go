package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
)

// ModuleCdc references the global incentives module codec. Note, the codec
// should ONLY be used in certain instances of tests and for JSON encoding.
//
// The actual codec used for serialization should be provided to
// modules/incentives and defined at the application level.
var (
	amino     = codec.NewLegacyAmino()
	ModuleCdc = codec.NewProtoCodec(codectypes.NewInterfaceRegistry())
)

// method required for x/inflation msg GetSignBytes methods
func init() {
	RegisterLegacyAminoCodec(amino)
	amino.Seal()
}

// RegisterInterfaces register implementations
func RegisterInterfaces(registry codectypes.InterfaceRegistry) {
	registry.RegisterImplementations(
		(*sdk.Msg)(nil),
		&MsgUpdateParams{},
	)

	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}

// register inflation msg types for Amino Codec in adherence to EIP-712 signing conventions
func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&MsgUpdateParams{}, "canto/x/inflation/MsgUpdateParams", nil)
	cdc.RegisterConcrete(&Params{}, "canto/x/inflation/Params", nil)
}
