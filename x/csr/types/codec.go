package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
)

var (
	amino     = codec.NewLegacyAmino()
	ModuleCdc = codec.NewProtoCodec(codectypes.NewInterfaceRegistry())
	AminoCdc  = codec.NewAminoCodec(amino)
)

// method required for x/csr msg GetSignBytes methods
func init() {
	RegisterLegacyAminoCodec(amino)
	amino.Seal()
}

// register interfaces for the AminoCodec
func RegisterInterfaces(registry codectypes.InterfaceRegistry) {
}

// register csr msg types for Amino Codec in adherence to EIP-712 signing conventions
func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
}
