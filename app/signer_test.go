package app

import (
	"testing"

	"cosmossdk.io/x/tx/signing"
	coinswapapi "github.com/Canto-Network/Canto/v7/api/canto/coinswap/v1"
	coinswaptypes "github.com/Canto-Network/Canto/v7/x/coinswap/types"
	"github.com/cosmos/cosmos-sdk/codec/address"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	protov2 "google.golang.org/protobuf/proto"
)

func TestMsgSwapOrderSigners(t *testing.T) {
	sw := coinswapapi.MsgSwapOrder{
		Input: &coinswapapi.Input{Address: "signer"},
	}

	signingOptions := signing.Options{
		AddressCodec: address.Bech32Codec{
			Bech32Prefix: sdk.GetConfig().GetBech32AccountAddrPrefix(),
		},
		ValidatorAddressCodec: address.Bech32Codec{
			Bech32Prefix: sdk.GetConfig().GetBech32ValidatorAddrPrefix(),
		},
	}
	signingOptions.DefineCustomGetSigners(protov2.MessageName(&coinswapapi.MsgSwapOrder{}), coinswaptypes.GetSignersFromMsgSwapOrderV2)

	ctx, err := signing.NewContext(signingOptions)

	require.NoError(t, err)
	signers, err := ctx.GetSigners(&sw)

	require.NoError(t, err)
	assert.Equal(t, 1, len(signers))
	assert.Equal(t, []byte("signer"), signers[0])
}
