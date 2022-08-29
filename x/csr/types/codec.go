package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
)

var (
	amino   = codec.NewLegacyAmino()
	ModuleCdc = codec.NewProtoCodec(codectypes.NewInterfaceRegistry())
	AminoCdc = codec.NewAminoCodec(amino)
)

const (
	// amino names
)

// method required for x/csr msg GetSignBytes methods
func init() {
	RegisterLegacyAminoCodec(amino)
	amino.Seal()
}

// register interfaces for the AminoCodec
func RegisterInterfaces(registry codectypes.InterfaceRegistry) {

	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}

// register csr msg types for Amino Codec in adherence to EIP-712 signing conventions
func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
}
