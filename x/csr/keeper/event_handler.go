package keeper

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/ethereum/go-ethereum/common"
	"github.com/tendermint/tendermint/crypto/ed25519"

	"github.com/Canto-Network/Canto/v2/x/csr/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// ABIEvents is used to unmarshal data from evm transactions
// EventHashes is used to match the event signature to the correct event for process

// Register the CSR in the store given the data from the evm transaction
func (k Keeper) RegisterCSREvent(ctx sdk.Context, data []byte) error {
	var event types.RegisterCSREvent
	// Unpack the data
	err := TurnstileContract.UnpackIntoInterface(&event, types.TurnstileEventRegisterCSR, data)
	if err != nil {
		return err
	}

	// Validate that the contract entered can be registered
	err = k.ValidateContract(ctx, event.SmartContractAddress)
	if err != nil {
		return err
	}

	// Create a new beneficiary account
	address := k.CreateNewAccount(ctx)

	// Create CSR object and validate
	csr := types.NewCSR(
		sdk.AccAddress(event.Receiver.Bytes()),
		[]string{event.SmartContractAddress.String()},
		0, // Init the NFT to 0 before validation
		address,
	)
	if err := csr.Validate(); err != nil {
		return err
	}

	// Mint the NFT after all validation has been done
	nft, err := k.MintCSR(ctx, event.Receiver)
	if err != nil {
		return err
	}
	csr.Id = nft

	// Set the CSR in the store
	k.SetCSR(ctx, csr)

	return nil
}

// Update a CSR existing in the store given data from the evm transaction
func (k Keeper) UpdateCSREvent(ctx sdk.Context, data []byte) error {
	var event types.UpdateCSREvent
	// Unpack the data
	err := TurnstileContract.UnpackIntoInterface(&event, types.TurnstileEventUpdateCSR, data)
	if err != nil {
		return err
	}

	// Validate that the contract entered can be registered
	err = k.ValidateContract(ctx, event.SmartContractAddress)
	if err != nil {
		return err
	}

	// Check if the NFT that is updated exists
	csr, found := k.GetCSR(ctx, event.Id)
	if !found {
		return sdkerrors.Wrapf(ErrNFTNotFound, "EventHandler::UpdateCSREvent the nft entered does not currently exist")
	}

	// Add the new smart contract to the CSR NFT and validate
	csr.Contracts = append(csr.Contracts, event.SmartContractAddress.String())
	err = csr.Validate()
	if err != nil {
		return err
	}
	k.SetCSR(ctx, *csr)

	return nil
}

// Retroactively register contracts that were previously deployed (non-CSR enabled smart contracts)
// and add the the CSR store
func (k Keeper) RetroactiveRegisterEvent() error {
	return nil
}

// ValidateContract checks if the smart contract can be registered to a CSR. It checks
// if the address is a smart contract address and whether it has been registered to an
// existing csr
func (k Keeper) ValidateContract(ctx sdk.Context, contract common.Address) error {
	// Check if the smart contract is already registered -> prevent double registration
	nftID, found := k.GetNFTByContract(ctx, contract.String())
	if found {
		return sdkerrors.Wrapf(ErrPrevRegisteredSmartContract,
			"EventHandler::RegisterCSREvent this smart contract is already registered to an existing NFT: %d", nftID)
	}

	// Check if the user is attempting to register a non-smart contract address
	account := k.evmKeeper.GetAccount(ctx, contract)
	if !account.IsContract() {
		return sdkerrors.Wrapf(ErrRegisterEOA,
			"EventHandler::RegisterCSREvent user is attempting to register a non-smart contract address")
	}
	return nil
}

// Creates a new account. Primarily used for the creation of the beneficiary account
func (k Keeper) CreateNewAccount(ctx sdk.Context) sdk.AccAddress {
	pubKey := ed25519.GenPrivKey().PubKey()
	address := sdk.AccAddress(pubKey.Address())
	beneficiary := k.accountKeeper.NewAccountWithAddress(ctx, address)
	k.accountKeeper.SetAccount(ctx, beneficiary)
	return address
}
