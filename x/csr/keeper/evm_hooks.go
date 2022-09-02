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

var _ evmtypes.EvmHooks = Hooks{}

var turnstileContract abi.ABI = contracts.TurnstileContract.ABI

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
	hasTurnstileEvent := h.processEvents(receipt)
	if hasTurnstileEvent {
		return nil
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
		return sdkerrors.Wrapf(ErrFeeCollectorDistribution, "EVMHOOK::PostTxProcessing failed to distribute fees from module account to nft beneficiary")
	}

	return nil
}

// returns true if there were any events emitted from the turnstile
func (h Hooks) processEvents(receipt *ethtypes.Receipt) bool {
	// Iterate through all logs in the receipt and check for events emitted by
	// the turnstile contract
	seenTurnstileEvent := false
	for _, log := range receipt.Logs {
		// Check if the address where the event was omitted is the turnstile contract
		// if log.Address != contracts.TurnstileContractAddress {
		// 	continue
		// }

		eventID := log.Topics[0]
		event, err := turnstileContract.EventByID(eventID)
		if err != nil {
			continue
		}

		// switch and process based on the event type
		switch event.Name {
		case types.TurnstileEventRegisterCSR:
			h.k.RegisterCSREvent(log.Data)
		case types.TurnstileEventUpdateCSR:
			h.k.UpdateCSREvent(log.Data)
		case types.TurnstileEventRetroactiveRegister:
			continue
		}
		seenTurnstileEvent = true
		// }
	}
	return seenTurnstileEvent
}
