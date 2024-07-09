package types

import (
	"fmt"
	"strings"

	sdkmath "cosmossdk.io/math"
	erc20v1 "github.com/Canto-Network/Canto/v7/api/canto/erc20/v1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	ethermint "github.com/evmos/ethermint/types"
	protov2 "google.golang.org/protobuf/proto"

	"github.com/ethereum/go-ethereum/common"
)

var (
	_ sdk.Msg = &MsgConvertCoin{}
	_ sdk.Msg = &MsgConvertERC20{}
)

// NewMsgConvertCoin creates a new instance of MsgConvertCoin
func NewMsgConvertCoin(coin sdk.Coin, receiver common.Address, sender sdk.AccAddress) *MsgConvertCoin { // nolint: interfacer
	return &MsgConvertCoin{
		Coin:     coin,
		Receiver: receiver.Hex(),
		Sender:   sender.String(),
	}
}

// NewMsgConvertERC20 creates a new instance of MsgConvertERC20
func NewMsgConvertERC20(amount sdkmath.Int, receiver sdk.AccAddress, contract, sender common.Address) *MsgConvertERC20 { // nolint: interfacer
	return &MsgConvertERC20{
		ContractAddress: contract.String(),
		Amount:          amount,
		Receiver:        receiver.String(),
		Sender:          sender.Hex(),
	}
}

// GetSigners defines whose signature is required
func (msg MsgConvertERC20) GetSigners() []sdk.AccAddress {
	addr := common.HexToAddress(msg.Sender)
	return []sdk.AccAddress{addr.Bytes()}
}

func GetSignersFromMsgConvertERC20V2(msg protov2.Message) ([][]byte, error) {
	msgv2, ok := msg.(*erc20v1.MsgConvertERC20)
	if !ok {
		return nil, nil
	}

	msgv1 := MsgConvertERC20{
		Sender: msgv2.Sender,
	}

	signers := [][]byte{}
	for _, signer := range msgv1.GetSigners() {
		signers = append(signers, signer.Bytes())
	}

	return signers, nil
}

// ValidateErc20Denom checks if a denom is a valid erc20/
// denomination
func ValidateErc20Denom(denom string) error {
	denomSplit := strings.SplitN(denom, "/", 2)

	if len(denomSplit) != 2 || denomSplit[0] != ModuleName {
		return fmt.Errorf("invalid denom. %s denomination should be prefixed with the format 'erc20/", denom)
	}

	return ethermint.ValidateAddress(denomSplit[1])
}
