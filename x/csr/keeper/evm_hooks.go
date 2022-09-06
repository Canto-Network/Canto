package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/Canto-Network/Canto/v2/contracts"
	"github.com/Canto-Network/Canto/v2/x/csr/types"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/core"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	evmtypes "github.com/evmos/ethermint/x/evm/types"
)

// Hooks wrapper struct for fees keeper
type Hooks struct {
	k Keeper
}

var (
	_                 evmtypes.EvmHooks = Hooks{}
	TurnstileContract abi.ABI           = contracts.TurnstileContract.ABI
	csrNftContract    abi.ABI           = contracts.CSRNFTContract.ABI
)

// Hooks return the wrapper hooks struct for the Keeper
func (k Keeper) Hooks() Hooks {
	return Hooks{k}
}

// PostTxProcessing implements EvmHooks.PostTxProcessing. After each successful
// interaction with a registered contract, the contract deployer receives
// a share from the transaction fees paid by the user.
func (h Hooks) PostTxProcessing(ctx sdk.Context, msg core.Message, receipt *ethtypes.Receipt) error {
	// Check if the csr module has been enabled
	params := h.k.GetParams(ctx)
	if !params.EnableCsr {
		return nil
	}

	// Ensure that transactions have a valid to address
	contract := msg.To()
	if contract == nil {
		return nil
	}

	// Check and process turnstile events if applicable
	// If a tx has turnstile events, then no fees will get distributed
	errEvents := h.processEvents(ctx, receipt)
	if errEvents != nil {
		return errEvents
	}

	// Grab the nft the smart contract corresponds to, if it has no nft -> return
	id, foundNFT := h.k.GetNFTByContract(ctx, contract.String())
	if !foundNFT {
		return nil
	}

	// Grab the account which will be receiving the tx fees
	csr, _ := h.k.GetCSR(ctx, id)
	beneficiary := csr.Account

	// Calculate fees to be distributed
	fee := sdk.NewIntFromUint64(receipt.GasUsed).Mul(sdk.NewIntFromBigInt(msg.GasPrice()))
	developerFee := sdk.NewDecFromInt(fee).Mul(params.CsrShares)
	evmDenom := h.k.evmKeeper.GetParams(ctx).EvmDenom
	csrFees := sdk.Coins{{Denom: evmDenom, Amount: developerFee.TruncateInt()}}

	err := h.k.bankKeeper.SendCoinsFromModuleToAccount(ctx, h.k.feeCollectorName, sdk.AccAddress(beneficiary), csrFees)

	if err != nil {
		return sdkerrors.Wrapf(ErrFeeCollectorDistribution, "EVMHook::PostTxProcessing failed to distribute fees from module account to nft beneficiary")
	}

	return nil
}

// returns true if there were any events emitted from the turnstile
func (h Hooks) processEvents(ctx sdk.Context, receipt *ethtypes.Receipt) error {
	// Get the turnstile and NFT address which are used to check events
	turnstileAddress, found := h.k.GetTurnstile(ctx)
	if !found {
		return sdkerrors.Wrapf(ErrContractDeployments, "Keeper::ProcessEvents the turnstile contract has not been found.")
	}
	nftAddress, found := h.k.GetCSRNFT(ctx)
	if !found {
		return sdkerrors.Wrapf(ErrCSRNFTNotDeployed, "Keeper::ProcessEvents the nft contract has not been found.")
	}

	for _, log := range receipt.Logs {
		// Check if the address matches the NFT or turnstile contracts
		eventID := log.Topics[0]
		switch log.Address {
		case turnstileAddress:
			event, err := TurnstileContract.EventByID(eventID)
			if err != nil {
				return err
			}

			// switch and process based on the turnstile event type
			switch event.Name {
			case types.TurnstileEventRegisterCSR:
				err = h.k.RegisterCSREvent(ctx, log.Data)
			case types.TurnstileEventUpdateCSR:
				err = h.k.UpdateCSREvent(ctx, log.Data)
			case types.TurnstileEventRetroactiveRegister:
				continue
			}
			if err != nil {
				return err
			}
		case nftAddress:
			// retrieve the event emitted from CSR NFT
			event, err := csrNftContract.EventByID(eventID)
			if err != nil {
				return err
			}

			// only the Withdrawal Event can be emitted from CSRNFT,
			switch event.Name {
			case types.WithdrawalEvent:
				// handle withdrawal
				err = h.k.WithdrawalEvent(ctx, log.Data)
			}
			if err != nil {
				return err
			}
		}
	}
	return nil
}
