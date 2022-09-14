package keeper

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/ethereum/go-ethereum/common"
	"github.com/tendermint/tendermint/crypto/ed25519"

	"github.com/Canto-Network/Canto/v2/x/csr/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Register the CSR in the store given the data from the evm event
func (k Keeper) RegisterEvent(ctx sdk.Context, data []byte) error {
	var event types.RegisterCSREvent
	// Unpack the data
	err := TurnstileContract.UnpackIntoInterface(&event, types.TurnstileEventRegister, data)
	if err != nil {
		return err
	}

	// Validate that the contract entered can be registered
	err = k.ValidateContract(ctx, event.SmartContractAddress)
	if err != nil {
		return err
	}

	// Check that the receiver account  exists
	if acct := k.evmKeeper.GetAccount(ctx, event.Receiver); acct == nil {
		return sdkerrors.Wrapf(ErrNonexistentAcct, "EventHandler::RegisterEvent: account does not exist: %s", event.Receiver)
	}

	// Create CSR object and validate
	csr := types.NewCSR(
		[]string{event.SmartContractAddress.String()},
		0, // Init the NFT to 0 before validation
	)
	if err := csr.Validate(); err != nil {
		return err
	}

	// Set the NFTID in the store if it has not been registered yet
	nftID := event.Id.Uint64()
	_, found := k.GetCSR(ctx, nftID)
	if found {
		return sdkerrors.Wrapf(ErrDuplicateNFTID, "EventHandler::RegisterEvent: this NFT id has already been registered")
	}
	csr.Id = nftID

	// Set the CSR in the store
	k.SetCSR(ctx, csr)

	return nil
}

// Update a CSR existing in the store given data from the evm transaction
func (k Keeper) UpdateEvent(ctx sdk.Context, data []byte) error {
	var event types.UpdateCSREvent
	// Unpack the data
	err := TurnstileContract.UnpackIntoInterface(&event, types.TurnstileEventUpdate, data)
	if err != nil {
		return err
	}
	// Validate that the contract entered can be registered
	err = k.ValidateContract(ctx, event.SmartContractAddress)
	if err != nil {
		return err
	}

	// Check if the NFT that is updated exists
	nftID := event.Id.Uint64()
	csr, found := k.GetCSR(ctx, nftID)
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

// ValidateContract checks if the smart contract can be registered to a CSR. It checks
// if the address is a smart contract address and whether it has been registered to an
// existing csr
func (k Keeper) ValidateContract(ctx sdk.Context, contract common.Address) error {
	// Check if the smart contract is already registered -> prevent double registration
	nftID, found := k.GetNFTByContract(ctx, contract.String())
	if found {
		return sdkerrors.Wrapf(ErrPrevRegisteredSmartContract,
			"EventHandler::ValidateContract this smart contract is already registered to an existing NFT: %d", nftID)
	}

	// Check if the user is attempting to register a non-smart contract address
	account := k.evmKeeper.GetAccount(ctx, contract)
	if account == nil || !account.IsContract() {
		return sdkerrors.Wrapf(ErrRegisterEOA,
			"EventHandler::ValidateContract user is attempting to register a nil or non-smart contract address")
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
