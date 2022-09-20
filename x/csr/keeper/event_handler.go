package keeper

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/ethereum/go-ethereum/common"

	"github.com/Canto-Network/Canto/v2/x/csr/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Register events occur in the Turnstile Contract when a user is attempting to create a new
// NFT with a smart contract that was just deployed. This event handler will unpack the
// event data, validate that the smart contract address, check that the receiver address is not null,
// and validate that this NFT is new. Only register can create new NFTs. Returns an error if the
// register event fails.
func (k Keeper) RegisterEvent(ctx sdk.Context, data []byte) error {
	var event types.RegisterCSREvent
	// Unpack the data
	err := TurnstileContract.UnpackIntoInterface(&event, types.TurnstileEventRegister, data)
	if err != nil {
		return err
	}

	k.Logger(ctx).Info("In Register Event this is the event ", event)

	// Validate that the contract entered can be registered
	err = k.ValidateContract(ctx, event.SmartContractAddress)
	if err != nil {
		return err
	}

	k.Logger(ctx).Info("Contract has been validated ", event)

	// Check that the receiver account  exists in the evm store
	if acct := k.evmKeeper.GetAccount(ctx, event.Receiver); acct == nil {
		return sdkerrors.Wrapf(ErrNonexistentAcct, "EventHandler::RegisterEvent: account does not exist: %s", event.Receiver)
	}

	k.Logger(ctx).Info("Account has been validated ", event)

	// Set the NFTID in the store if it has not been registered yet
	nftID := event.Id.Uint64()
	_, found := k.GetCSR(ctx, nftID)
	if found {
		return sdkerrors.Wrapf(ErrDuplicateNFTID, "EventHandler::RegisterEvent: this NFT id has already been registered")
	}

	// Create CSR object and perform stateless validation
	csr := types.NewCSR(
		[]string{event.SmartContractAddress.String()},
		nftID,
	)
	if err := csr.Validate(); err != nil {
		return err
	}

	k.Logger(ctx).Info("CSR has been validated ", csr)

	// Set the CSR in the store
	k.SetCSR(ctx, csr)

	k.Logger(ctx).Info("CSR has been set in state ", csr)

	return nil
}

// Update events occur in the Turnstile contract when a user is attempting to assign their newly
// deployed smart contract to an existing NFT. This event handler will unpack the data, validate
// that the smart contract to be assigned is valid, check that NFT id exists, and append the smart contract
// to the NFT id entered. Update is permissionless in the sense that you do not have to be the owner
// of the NFT to be able to add new smart contracts to it.
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

	// Check if the NFT that is being updated exists in the CSR store
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
// if the address is a smart contract address, whether the smart contract has code, and
// whether the contract is already assigned to some other NFT.
func (k Keeper) ValidateContract(ctx sdk.Context, contract common.Address) error {
	// Check if the smart contract is already registered -> prevent double registration
	nftID, found := k.GetNFTByContract(ctx, contract.String())
	if found {
		return sdkerrors.Wrapf(ErrPrevRegisteredSmartContract,
			"EventHandler::ValidateContract this smart contract is already registered to an existing NFT: %d", nftID)
	}

	// Check if the user is attempting to register a non-smart contract address (i.e. an EOA or non-existent address)
	account := k.evmKeeper.GetAccount(ctx, contract)
	if account == nil || !account.IsContract() {
		return sdkerrors.Wrapf(ErrRegisterEOA,
			"EventHandler::ValidateContract user is attempting to register a nil or non-smart contract address")
	}
	return nil
}
