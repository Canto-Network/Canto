package keeper

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/tendermint/tendermint/crypto/ed25519"

	"github.com/Canto-Network/Canto/v2/x/csr/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// ABIEvents is used to unmarshal data from evm transactions
// EventHashes is used to match the event signature to the correct event for process

// Register the CSR in the store given the data from the evm transaction
func (k Keeper) RegisterCSREvent(ctx sdk.Context, data []byte) error {
	var event types.RegisterCSREvent
	err := turnstileContract.UnpackIntoInterface(&event, types.TurnstileEventRegisterCSR, data)
	if err != nil {
		return err
	}

	// Check if the smart contract is already registered -> prevent double registration
	nftID, found := k.GetNFTByContract(ctx, event.SmartContractAddress.String())
	if found {
		return sdkerrors.Wrapf(ErrPrevRegisteredSmartContract,
			"EventHandler::RegisterCSREvent this smart contract is already registered to an existing NFT: %d", nftID)
	}

	// Check if the user is attempting to register a non-smart contract address
	account := k.evmKeeper.GetAccount(ctx, event.SmartContractAddress)
	if !account.IsContract() {
		return sdkerrors.Wrapf(ErrRegisterEOA, "EventHandler::RegisterCSREvent user is attempting to register a non-smart contract address")
	}

	// **** Mint a new NFT ****

	// Create a new beneficiary account
	privateKey := ed25519.GenPrivKey().PubKey()
	address := sdk.AccAddress(privateKey.Address())
	beneficiary := k.accountKeeper.NewAccountWithAddress(ctx, address)
	k.accountKeeper.SetAccount(ctx, beneficiary)

	// Create CSR object and validate
	csr := types.NewCSR(
		sdk.AccAddress(event.Receiver.Bytes()),
		[]string{event.SmartContractAddress.String()},
		0, // update this to the new nft id
		address,
	)
	if err := csr.Validate(); err != nil {
		return err
	}

	// Set the CSR in the store
	k.SetCSR(ctx, csr)

	return nil
}

// Update a CSR existing in the store given data from the evm transaction
func (k Keeper) UpdateCSREvent(ctx sdk.Context, data []byte) error {
	_, err := turnstileContract.Unpack(types.TurnstileEventUpdateCSR, data)
	if err != nil {
		return err
	}

	// event := types.UpdateCSREvent{
	// 	SmartContractAddress: response[0].(common.Address),
	// 	Nft_id:               response[1].(uint64),
	// }

	// HANDLE LOGIC HERE

	return nil
}

// Retroactively register contracts that were previously deployed (non-CSR enabled smart contracts)
// and add the the CSR store
func (k Keeper) RetroactiveRegisterEvent() error {
	return nil
}
