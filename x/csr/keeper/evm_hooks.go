package keeper

import (
	"encoding/hex"
	"encoding/json"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/accounts/abi"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	evmtypes "github.com/evmos/ethermint/x/evm/types"
)

// Hooks wrapper struct for fees keeper
type Hooks struct {
	k Keeper
}

var _ evmtypes.EvmHooks = Hooks{}

// Hooks return the wrapper hooks struct for the Keeper
func (k Keeper) Hooks() Hooks {
	return Hooks{k}
}

// PostTxProcessing implements EvmHooks.PostTxProcessing. After each successful
// interaction with a registered contract, the contract deployer receives
// a share from the transaction fees paid by the user.
func (h Hooks) PostTxProcessing(ctx sdk.Context, msg core.Message, receipt *ethtypes.Receipt) error {
	bz := receipt.Logs[0].Data
	var dest []byte
	hex.Decode(dest, bz)
	fmt.Println(dest)
	decodeABI(bz)
	return nil
}

func decodeABI(data []byte) {
	var event abi.Event
	json.Unmarshal(jsonEventRandomEvent, &event)

	abi := abi.ABI{Events: map[string]abi.Event{"randomEvent": event}}

	response := EventRandomEvent{}

	abi.UnpackIntoInterface(&response, "randomEvent", data)
	fmt.Println("this is the response: ", response)
}

var jsonEventRandomEvent = []byte(`{
	"anonymous": false,
	"inputs": [
	  {
		"indexed": false, "name": "message", "type": "string"
	  }, {
		"indexed": false, "name": "sender", "type": "address"
	  }
	],
	"name": "RandomEvent",
	"type": "event"
  }`)

type EventRandomEvent struct {
	Message string
	Sender  common.Address
}
