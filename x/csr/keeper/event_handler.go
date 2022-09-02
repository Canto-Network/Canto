package keeper

import (
	"github.com/Canto-Network/Canto/v2/x/csr/types"
	"github.com/ethereum/go-ethereum/common"
)

// ABIEvents is used to unmarshal data from evm transactions
// EventHashes is used to match the event signature to the correct event for process

// Register the CSR in the store given the data from the evm transaction
func (k Keeper) RegisterCSREvent(data []byte) error {
	response, err := turnstileContract.Unpack(types.TurnstileEventRegisterCSR, data)
	if err != nil {
		return err
	}

	event := types.RegisterCSREvent{
		SmartContractAddress: response[0].(common.Address),
		Receiver:             response[1].(common.Address),
	}

	// HANDLE LOGIC HERE

	return nil
}

// Update a CSR existing in the store given data from the evm transaction
func (k Keeper) UpdateCSREvent(data []byte) error {
	response, err := turnstileContract.Unpack(types.TurnstileEventUpdateCSR, data)
	if err != nil {
		return err
	}

	event := types.UpdateCSREvent{
		SmartContractAddress: response[0].(common.Address),
		Nft_id:               response[1].(uint64),
	}

	// HANDLE LOGIC HERE

	return nil
}

// Retroactively register contracts that were previously deployed (non-CSR enabled smart contracts)
// and add the the CSR store
func (k Keeper) RetroactiveRegisterEvent() error {
	return nil
}
