package app

import (
	"testing"

	"cosmossdk.io/x/tx/signing"
	coinswapapi "github.com/Canto-Network/Canto/v7/api/canto/coinswap/v1"
	coinswaptypes "github.com/Canto-Network/Canto/v7/x/coinswap/types"
	"github.com/cosmos/cosmos-sdk/codec/address"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	protov2 "google.golang.org/protobuf/proto"
)

func TestMsgSwapOrderSigners(t *testing.T) {

	addr := "canto13e9t6s6ra8caz5zzmy5w9v23dm2dr5nrr9sz03"
	accAddr, err := sdk.AccAddressFromBech32(addr)
	require.NoError(t, err)

	signingOptions := signing.Options{
		AddressCodec: address.Bech32Codec{
			Bech32Prefix: sdk.GetConfig().GetBech32AccountAddrPrefix(),
		},
		ValidatorAddressCodec: address.Bech32Codec{
			Bech32Prefix: sdk.GetConfig().GetBech32ValidatorAddrPrefix(),
		},
	}
	signingOptions.DefineCustomGetSigners(protov2.MessageName(&coinswapapi.MsgSwapOrder{}), coinswaptypes.CreateGetSignersFromMsgSwapOrderV2(&signingOptions))

	ctx, err := signing.NewContext(signingOptions)
	require.NoError(t, err)

	tests := []struct {
		name    string
		msg     protov2.Message
		want    [][]byte
		wantErr bool
	}{
		{
			name: "MsgSwapOrder",
			msg: &coinswapapi.MsgSwapOrder{
				Input: &coinswapapi.Input{Address: addr},
			},
			want: [][]byte{accAddr},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			signers, err := ctx.GetSigners(test.msg)
			if test.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			require.Equal(t, test.want, signers)
		})
	}
}
