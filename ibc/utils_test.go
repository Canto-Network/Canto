package ibc

import (
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"

	transfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v8/modules/core/04-channel/types"
	ibctesting "github.com/cosmos/ibc-go/v8/testing"
)

func init() {
	cfg := sdk.GetConfig()
	cfg.SetBech32PrefixForAccount("canto", "cantopub")
}

func TestGetTransferSenderRecipient(t *testing.T) {
	testCases := []struct {
		name         string
		packet       channeltypes.Packet
		expSender    string
		expRecipient string
		expError     bool
	}{
		{
			"empty packet",
			channeltypes.Packet{},
			"", "",
			true,
		},
		{
			"invalid packet data",
			channeltypes.Packet{
				Data: ibctesting.MockFailPacketData,
			},
			"", "",
			true,
		},
		{
			"empty FungibleTokenPacketData",
			channeltypes.Packet{
				Data: transfertypes.ModuleCdc.MustMarshalJSON(
					&transfertypes.FungibleTokenPacketData{},
				),
			},
			"", "",
			true,
		},
		{
			"invalid sender",
			channeltypes.Packet{
				Data: transfertypes.ModuleCdc.MustMarshalJSON(
					&transfertypes.FungibleTokenPacketData{
						Sender:   "cosmos1",
						Receiver: "canto1x2w87cvt5mqjncav4lxy8yfreynn273xrrp06v",
						Amount:   "123456",
					},
				),
			},
			"", "",
			true,
		},
		{
			"invalid recipient",
			channeltypes.Packet{
				Data: transfertypes.ModuleCdc.MustMarshalJSON(
					&transfertypes.FungibleTokenPacketData{
						Sender:   "cosmos1qql8ag4cluz6r4dz28p3w00dnc9w8ueulg2gmc",
						Receiver: "canto1",
						Amount:   "123456",
					},
				),
			},
			"", "",
			true,
		},
		{
			"valid - cosmos sender, canto recipient",
			channeltypes.Packet{
				Data: transfertypes.ModuleCdc.MustMarshalJSON(
					&transfertypes.FungibleTokenPacketData{
						Sender:   "cosmos1qql8ag4cluz6r4dz28p3w00dnc9w8ueulg2gmc",
						Receiver: "canto1x2w87cvt5mqjncav4lxy8yfreynn273xrrp06v",
						Amount:   "123456",
					},
				),
			},
			"canto1qql8ag4cluz6r4dz28p3w00dnc9w8ueud7tc0s",
			"canto1x2w87cvt5mqjncav4lxy8yfreynn273xrrp06v",
			false,
		},
		{
			"valid - canto sender, cosmos recipient",
			channeltypes.Packet{
				Data: transfertypes.ModuleCdc.MustMarshalJSON(
					&transfertypes.FungibleTokenPacketData{
						Sender:   "canto1x2w87cvt5mqjncav4lxy8yfreynn273xrrp06v",
						Receiver: "cosmos1qql8ag4cluz6r4dz28p3w00dnc9w8ueulg2gmc",
						Amount:   "123456",
					},
				),
			},
			"canto1x2w87cvt5mqjncav4lxy8yfreynn273xrrp06v",
			"canto1qql8ag4cluz6r4dz28p3w00dnc9w8ueud7tc0s",
			false,
		},
		{
			"valid - osmosis sender, canto recipient",
			channeltypes.Packet{
				Data: transfertypes.ModuleCdc.MustMarshalJSON(
					&transfertypes.FungibleTokenPacketData{
						Sender:   "osmo1qql8ag4cluz6r4dz28p3w00dnc9w8ueuhnecd2",
						Receiver: "canto1x2w87cvt5mqjncav4lxy8yfreynn273xrrp06v",
						Amount:   "123456",
					},
				),
			},
			"canto1qql8ag4cluz6r4dz28p3w00dnc9w8ueud7tc0s",
			"canto1x2w87cvt5mqjncav4lxy8yfreynn273xrrp06v",
			false,
		},
	}

	for _, tc := range testCases {
		sender, recipient, _, _, err := GetTransferSenderRecipient(tc.packet)

		if tc.expError {
			require.Error(t, err, tc.name)
		} else {
			require.NoError(t, err, tc.name)
			require.Equal(t, tc.expSender, sender.String())
			require.Equal(t, tc.expRecipient, recipient.String())
		}
	}
}

func TestGetTransferAmount(t *testing.T) {
	testCases := []struct {
		name      string
		packet    channeltypes.Packet
		expAmount string
		expError  bool
	}{
		{
			"empty packet",
			channeltypes.Packet{},
			"",
			true,
		},
		{
			"invalid packet data",
			channeltypes.Packet{
				Data: ibctesting.MockFailPacketData,
			},
			"",
			true,
		},
		{
			"invalid amount - empty",
			channeltypes.Packet{
				Data: transfertypes.ModuleCdc.MustMarshalJSON(
					&transfertypes.FungibleTokenPacketData{
						Sender:   "cosmos1qql8ag4cluz6r4dz28p3w00dnc9w8ueulg2gmc",
						Receiver: "canto1x2w87cvt5mqjncav4lxy8yfreynn273xrrp06v",
						Amount:   "",
					},
				),
			},
			"",
			true,
		},
		{
			"invalid amount - non-int",
			channeltypes.Packet{
				Data: transfertypes.ModuleCdc.MustMarshalJSON(
					&transfertypes.FungibleTokenPacketData{
						Sender:   "cosmos1qql8ag4cluz6r4dz28p3w00dnc9w8ueulg2gmc",
						Receiver: "canto1x2w87cvt5mqjncav4lxy8yfreynn273xrrp06v",
						Amount:   "test",
					},
				),
			},
			"test",
			true,
		},
		{
			"valid",
			channeltypes.Packet{
				Data: transfertypes.ModuleCdc.MustMarshalJSON(
					&transfertypes.FungibleTokenPacketData{
						Sender:   "cosmos1qql8ag4cluz6r4dz28p3w00dnc9w8ueulg2gmc",
						Receiver: "canto1x2w87cvt5mqjncav4lxy8yfreynn273xrrp06v",
						Amount:   "10000",
					},
				),
			},
			"10000",
			false,
		},
	}

	for _, tc := range testCases {
		amt, err := GetTransferAmount(tc.packet)
		if tc.expError {
			require.Error(t, err, tc.name)
		} else {
			require.NoError(t, err, tc.name)
			require.Equal(t, tc.expAmount, amt)
		}
	}
}
