package app

import (
	"testing"

	"cosmossdk.io/x/tx/signing"
	coinswapapi "github.com/Canto-Network/Canto/v7/api/canto/coinswap/v1"
	erc20v1 "github.com/Canto-Network/Canto/v7/api/canto/erc20/v1"
	coinswaptypes "github.com/Canto-Network/Canto/v7/x/coinswap/types"
	erc20types "github.com/Canto-Network/Canto/v7/x/erc20/types"
	"github.com/cosmos/cosmos-sdk/codec/address"
	sdk "github.com/cosmos/cosmos-sdk/types"
	evmv1 "github.com/evmos/ethermint/api/ethermint/evm/v1"
	evmtypes "github.com/evmos/ethermint/x/evm/types"
	"github.com/stretchr/testify/require"
	protov2 "google.golang.org/protobuf/proto"
)

func TestMsgSwapOrderSigners(t *testing.T) {

	addr := "canto13e9t6s6ra8caz5zzmy5w9v23dm2dr5nrr9sz03"

	signingOptions := signing.Options{
		AddressCodec: address.Bech32Codec{
			Bech32Prefix: sdk.GetConfig().GetBech32AccountAddrPrefix(),
		},
		ValidatorAddressCodec: address.Bech32Codec{
			Bech32Prefix: sdk.GetConfig().GetBech32ValidatorAddrPrefix(),
		},
	}
	signingOptions.DefineCustomGetSigners(protov2.MessageName(&coinswapapi.MsgSwapOrder{}), coinswaptypes.CreateGetSignersFromMsgSwapOrderV2(&signingOptions))
	signingOptions.DefineCustomGetSigners(protov2.MessageName(&evmv1.MsgEthereumTx{}), evmtypes.GetSignersFromMsgEthereumTxV2)
	signingOptions.DefineCustomGetSigners(protov2.MessageName(&erc20v1.MsgConvertERC20{}), erc20types.GetSignersFromMsgConvertERC20V2)

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
			want: [][]byte{[]byte(sdk.AccAddress(addr))},
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
