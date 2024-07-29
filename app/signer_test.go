package app

import (
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
	protov2 "google.golang.org/protobuf/proto"

	"cosmossdk.io/x/tx/signing"
	"github.com/cosmos/cosmos-sdk/codec/address"
	sdk "github.com/cosmos/cosmos-sdk/types"

	coinswapv1 "github.com/Canto-Network/Canto/v8/api/canto/coinswap/v1"
	erc20v1 "github.com/Canto-Network/Canto/v8/api/canto/erc20/v1"
	coinswaptypes "github.com/Canto-Network/Canto/v8/x/coinswap/types"
	erc20types "github.com/Canto-Network/Canto/v8/x/erc20/types"
)

func TestDefineCustomGetSigners(t *testing.T) {
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
	signingOptions.DefineCustomGetSigners(protov2.MessageName(&erc20v1.MsgConvertERC20{}), erc20types.GetSignersFromMsgConvertERC20V2)
	signingOptions.DefineCustomGetSigners(protov2.MessageName(&coinswapv1.MsgSwapOrder{}), coinswaptypes.CreateGetSignersFromMsgSwapOrderV2(&signingOptions))

	ctx, err := signing.NewContext(signingOptions)
	require.NoError(t, err)

	tests := []struct {
		name    string
		msg     protov2.Message
		want    [][]byte
		wantErr bool
	}{
		{
			name: "MsgConvertERC20",
			msg: &erc20v1.MsgConvertERC20{
				Sender: common.BytesToAddress(accAddr.Bytes()).String(),
			},
			want: [][]byte{accAddr.Bytes()},
		},
		{
			name: "MsgSwapOrder",
			msg: &coinswapv1.MsgSwapOrder{
				Input: &coinswapv1.Input{Address: addr},
			},
			want: [][]byte{accAddr.Bytes()},
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
