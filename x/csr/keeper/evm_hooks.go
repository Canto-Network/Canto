package keeper

import (
	"math/big"

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
)

// Hooks return the wrapper hooks struct for the Keeper
func (k Keeper) Hooks() Hooks {
	return Hooks{k}
}

// PostTxProcessing implements EvmHooks.PostTxProcessing. The EVM hook allows users to utilize the Turnstile
// Smart Contract to register and assign smart contracts that are deployed to a Contract Secured Revenue NFT and
// distribute transaction fees for contracts that are already registered to some NFT. After each successful EVM
// transaction, the PostTxProcessing hook will process all of the events from the log receipt of the EVM tx.
// If any of the events originate from the Turnstile address, there will be an event handler that processes the two
// Turnstile events – register and assign – accordingly. At the very end of the hook, the To address from the EVM
// tx to will be read in. From there, the hook will check if that contract address belongs to any NFT currently in
// the store. If so, the fees will be split and distributed to the Turnstile Address.
func (h Hooks) PostTxProcessing(ctx sdk.Context, msg core.Message, receipt *ethtypes.Receipt) error {
	// Check if the csr module has been enabled
	params := h.k.GetParams(ctx)
	if !params.EnableCsr {
		return nil
	}

	// Check and process turnstile events if applicable
	// If a tx has turnstile events, then no fees will get distributed
	err := h.processEvents(ctx, receipt)
	if err != nil {
		h.k.Logger(ctx).Error(err.Error())
	}

	// Grab the nft the smart contract corresponds to, if it has no nft -> return nil
	contract := msg.To()
	if contract == nil {
		return nil
	}

	nftID, foundNFT := h.k.GetNFTByContract(ctx, contract.String())
	if !foundNFT {
		return nil
	}

	csr, found := h.k.GetCSR(ctx, nftID)
	if !found {
		return sdkerrors.Wrapf(ErrNonexistentCSR, "EVMHook::PostTxProcessing the NFT ID was found but the CSR was not.")
	}

	// Calculate fees to be distributed = intFloor(GasUsed * GasPrice * csrShares)
	fee := sdk.NewIntFromUint64(receipt.GasUsed).Mul(sdk.NewIntFromBigInt(msg.GasPrice()))
	csrFee := sdk.NewDecFromInt(fee).Mul(params.CsrShares).TruncateInt()
	evmDenom := h.k.evmKeeper.GetParams(ctx).EvmDenom
	csrFees := sdk.Coins{{Denom: evmDenom, Amount: csrFee}}

	// Send fees from fee collector to module account before distribution
	err = h.k.bankKeeper.SendCoinsFromModuleToModule(ctx, h.k.FeeCollectorName, types.ModuleName, csrFees)
	if err != nil {
		return sdkerrors.Wrapf(ErrFeeDistribution, "EVMHook::PostTxProcessing failed to distribute fees from fee collector to module, %d", err)
	}

	// Get the turnstile which will receive funds for tx fees
	turnstileAddress, found := h.k.GetTurnstile(ctx)
	if !found {
		return sdkerrors.Wrapf(ErrContractDeployments, "Keeper::ProcessEvents the turnstile contract has not been found.")
	}

	// Distribute fees to turnstile contract by NFT ID distributeFees(amount, nftID)
	amount := csrFee.BigInt()
	_, err = h.k.CallMethod(ctx, "distributeFees", contracts.TurnstileContract, types.ModuleAddress, &turnstileAddress, amount, new(big.Int).SetUint64(nftID))
	if err != nil {
		return sdkerrors.Wrapf(ErrFeeDistribution, "EVMHook::PostTxProcessing failed to distribute fees from module account to turnstile, %d", err)
	}

	// Update TX count for this NFT
	csr.Txs += 1
	// Update the cumulative revenue accumulated by this NFT
	revenue := new(big.Int).SetBytes(csr.Revenue)
	revenue.Add(revenue, amount)
	csr.Revenue = revenue.Bytes()

	// Store updated CSR
	h.k.SetCSR(ctx, *csr)

	return nil
}

// returns an error if there was an issue processing any of the events (register, assign) from the turnstile address
func (h Hooks) processEvents(ctx sdk.Context, receipt *ethtypes.Receipt) error {
	// Get the turnstile which is used to check events
	turnstileAddress, found := h.k.GetTurnstile(ctx)
	if !found {
		return sdkerrors.Wrapf(ErrContractDeployments, "Keeper::ProcessEvents the turnstile contract has not been found.")
	}

	for _, log := range receipt.Logs {
		if len(log.Topics) == 0 {
			continue
		}

		// Check if the address matches the turnstile contracts
		eventID := log.Topics[0]
		if log.Address == turnstileAddress {
			event, err := TurnstileContract.EventByID(eventID)
			if err != nil {
				return err
			}

			// switch and process based on the turnstile event type
			switch event.Name {
			case types.TurnstileEventRegister:
				err = h.k.RegisterEvent(ctx, log.Data)
			case types.TurnstileEventUpdate:
				err = h.k.UpdateEvent(ctx, log.Data)
			}
			if err != nil {
				return err
			}
		}
	}
	return nil
}
