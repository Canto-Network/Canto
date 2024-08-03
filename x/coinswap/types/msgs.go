package types

import (
	"fmt"

	sdkmath "cosmossdk.io/math"
	"cosmossdk.io/x/tx/signing"
	coinswapv1 "github.com/Canto-Network/Canto/v8/api/canto/coinswap/v1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	protov2 "google.golang.org/protobuf/proto"
)

var (
	_ sdk.Msg = &MsgSwapOrder{}
	_ sdk.Msg = &MsgAddLiquidity{}
	_ sdk.Msg = &MsgRemoveLiquidity{}
)

const (
	// LptTokenPrefix defines the prefix of liquidity token
	LptTokenPrefix = "lpt"
	// LptTokenFormat defines the name of liquidity token
	LptTokenFormat = "lpt-%d"
)

/* --------------------------------------------------------------------------- */
// MsgSwapOrder
/* --------------------------------------------------------------------------- */

// MsgSwapOrder - struct for swapping a coin
// Input and Output can either be exact or calculated.
// An exact coin has the senders desired buy or sell amount.
// A calculated coin has the desired denomination and bounded amount
// the sender is willing to buy or sell in this order.
//
// NewMsgSwapOrder creates a new MsgSwapOrder object.
func NewMsgSwapOrder(
	input Input,
	output Output,
	deadline int64,
	isBuyOrder bool,
) *MsgSwapOrder {
	return &MsgSwapOrder{
		Input:      input,
		Output:     output,
		Deadline:   deadline,
		IsBuyOrder: isBuyOrder,
	}
}

// NewMsgAddLiquidity creates a new MsgAddLiquidity object.
func NewMsgAddLiquidity(
	maxToken sdk.Coin,
	exactStandardAmt sdkmath.Int,
	minLiquidity sdkmath.Int,
	deadline int64,
	sender string,
) *MsgAddLiquidity {
	return &MsgAddLiquidity{
		MaxToken:         maxToken,
		ExactStandardAmt: exactStandardAmt,
		MinLiquidity:     minLiquidity,
		Deadline:         deadline,
		Sender:           sender,
	}
}

// NewMsgRemoveLiquidity creates a new MsgRemoveLiquidity object
func NewMsgRemoveLiquidity(
	minToken sdkmath.Int,
	withdrawLiquidity sdk.Coin,
	minStandardAmt sdkmath.Int,
	deadline int64,
	sender string,
) *MsgRemoveLiquidity {
	return &MsgRemoveLiquidity{
		MinToken:          minToken,
		WithdrawLiquidity: withdrawLiquidity,
		MinStandardAmt:    minStandardAmt,
		Deadline:          deadline,
		Sender:            sender,
	}
}

func CreateGetSignersFromMsgSwapOrderV2(options *signing.Options) func(msg protov2.Message) ([][]byte, error) {
	return func(msg protov2.Message) ([][]byte, error) {
		msgv2, ok := msg.(*coinswapv1.MsgSwapOrder)
		if !ok {
			return nil, fmt.Errorf("invalid x/coinswap/MsgSwapOrder msg v2: %v", msg)
		}

		addr, err := options.AddressCodec.StringToBytes(msgv2.Input.Address)
		if err != nil {
			return nil, err
		}

		return [][]byte{addr}, nil
	}
}
