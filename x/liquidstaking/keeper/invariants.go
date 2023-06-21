package keeper

import (
	"fmt"

	"github.com/Canto-Network/Canto/v6/x/liquidstaking/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// TODO: Add invariants_test
func RegisterInvariants(ir sdk.InvariantRegistry, k Keeper) {
	ir.RegisterRoute(types.ModuleName, "net-account",
		NetAmountInvariant(k))
	ir.RegisterRoute(types.ModuleName, "chunks",
		ChunksInvariant(k))
	ir.RegisterRoute(types.ModuleName, "insurances",
		InsurancesInvariant(k))
	ir.RegisterRoute(types.ModuleName, "unpairing-for-unstaking-chunk-infos",
		UnpairingForUnstakingChunkInfosInvariant(k))
	ir.RegisterRoute(types.ModuleName, "withdraw-insurance-requests",
		WithdrawInsuranceRequestsInvariant(k))
}

func AllInvariants(k Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		for _, inv := range []func(Keeper) sdk.Invariant{
			NetAmountInvariant,
			ChunksInvariant,
			InsurancesInvariant,
			UnpairingForUnstakingChunkInfosInvariant,
			WithdrawInsuranceRequestsInvariant,
		} {
			res, stop := inv(k)(ctx)
			if stop {
				return res, stop
			}
		}
		return "", false
	}
}

func NetAmountInvariant(k Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		nas := k.GetNetAmountState(ctx)
		// if net amount is positive, it means that there are paired chunks.
		if nas.LsTokensTotalSupply.IsPositive() && !nas.NetAmount.IsPositive() {
			return "found positive lsToken supply with non-positive net amount", true
		}
		// if ls tokens supply is not positive, it means that all chunks are un-paired.
		// any unbonding balance or liquid tokens must not exist.
		if !nas.LsTokensTotalSupply.IsPositive() &&
			(nas.TotalUnbondingBalance.IsPositive() || nas.TotalLiquidTokens.IsPositive()) {
			return "found non-positive lsToken supply with positive unbonding balance or liquid tokens", true
		}

		return "", false
	}
}

func ChunksInvariant(k Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		msg := ""
		brokenCount := 0
		k.IterateAllChunks(ctx, func(chunk types.Chunk) (bool, error) {
			switch chunk.Status {
			case types.CHUNK_STATUS_PAIRING:
				// must have empty paired insurance
				if chunk.PairedInsuranceId != types.Empty {
					msg += fmt.Sprintf("pairing chunk(id: %d) have non-empty paired insurance\n", chunk.Id)
					brokenCount++
					return false, nil
				}

				// must have balance more than ChunkSize tokens
				balance := k.bankKeeper.GetBalance(ctx, chunk.DerivedAddress(), k.stakingKeeper.BondDenom(ctx))
				if balance.Amount.LT(types.ChunkSize) {
					msg += fmt.Sprintf("pairing chunk(id: %d) have balance less than ChunkSize\n", chunk.Id)
					brokenCount++
					return false, nil
				}
			case types.CHUNK_STATUS_PAIRED:
				// must have paired insurance
				if chunk.PairedInsuranceId == types.Empty {
					msg += fmt.Sprintf("paired chunk(id: %d) have empty paired insurance\n", chunk.Id)
					return false, nil
				}
				insurance, found := k.GetInsurance(ctx, chunk.PairedInsuranceId)
				if !found {
					msg += fmt.Sprintf("not found paired insurance for paired chunk(id: %d)\n", chunk.Id)
					brokenCount++
					return false, nil
				}
				if insurance.Status != types.INSURANCE_STATUS_PAIRED {
					msg += fmt.Sprintf("paired chunk(id: %d) have paired insurance with invalid status: %s\n", chunk.Id, insurance.Status)
					brokenCount++
					return false, nil
				}
				// must have valid Delegation object
				delegation, found := k.stakingKeeper.GetDelegation(ctx, chunk.DerivedAddress(), insurance.GetValidator())
				if !found {
					msg += fmt.Sprintf("not found delegation for paired chunk(id: %d)\n", chunk.Id)
					brokenCount++
					return false, nil
				}
				delShares := delegation.GetShares()
				if delShares.LT(types.ChunkSize.ToDec()) {
					msg += fmt.Sprintf("delegation tokens of paired chunk(id: %d) is less than chunk size tokens: %s\n", chunk.Id, delShares.String())
					brokenCount++
					return false, nil
				}
			case types.CHUNK_STATUS_UNPAIRING, types.CHUNK_STATUS_UNPAIRING_FOR_UNSTAKING:
				// must have unpairing insurance
				if chunk.UnpairingInsuranceId == types.Empty {
					msg += fmt.Sprintf("unpairing chunk(id: %d) have empty unpairing insurance\n", chunk.Id)
					brokenCount++
					return false, nil
				}
				insurance, found := k.GetInsurance(ctx, chunk.UnpairingInsuranceId)
				if !found {
					msg += fmt.Sprintf("not found unpairing insurance for unpairing chunk(id: %d)\n", chunk.Id)
					brokenCount++
					return false, nil
				}
				if k.IsEpochReached(ctx) {
					// skip in this case to check unbonding delegation entry it because
					// mature unbonding delegation is deleted in the begin blocker of staking module
					// and invariant checks will begin at the end blocker of crisis module.
					return false, nil
				}
				// must have unbonding delegation
				ubd, found := k.stakingKeeper.GetUnbondingDelegation(ctx, chunk.DerivedAddress(), insurance.GetValidator())
				if !found {
					msg += fmt.Sprintf("not found unbonding delegation for unpairing chunk(id: %d)\n", chunk.Id)
					brokenCount++
					return false, nil
				}
				// must have valid Delegation object
				if len(ubd.Entries) != 1 {
					msg += fmt.Sprintf("unbonding delegation for unpairing chunk(id: %d) have more than 1 entries\n", chunk.Id)
					brokenCount++
					return false, nil
				}
				if ubd.Entries[0].InitialBalance.LT(types.ChunkSize) {
					msg += fmt.Sprintf("unbonding delegation tokens of unpairing chunk(id: %d) is less than chunk size tokens: %s\n", chunk.Id, ubd.Entries[0].InitialBalance.String())
					brokenCount++
					return false, nil
				}
			default:
				msg += fmt.Sprintf("chunk(id: %d) have invalid status: %s\n", chunk.Id, chunk.Status)
				brokenCount++
				return false, nil
			}
			return false, nil
		})
		if brokenCount > 0 {
			return sdk.FormatInvariant(types.ModuleName, "chunks", fmt.Sprintf(
				"found %d broken chunks:\n%s", brokenCount, msg)), true
		} else {
			return "", false
		}
	}
}

func InsurancesInvariant(k Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		msg := ""
		brokenCount := 0
		k.IterateAllInsurances(ctx, func(insurance types.Insurance) (bool, error) {
			switch insurance.Status {
			case types.INSURANCE_STATUS_PAIRING:
				// must have empty chunk
				if insurance.ChunkId != types.Empty {
					msg += fmt.Sprintf("pairing insurance(id: %d) have non-empty paired chunk\n", insurance.Id)
					brokenCount++
					return false, nil
				}
			case types.INSURANCE_STATUS_PAIRED:
				// must have paired chunk
				if insurance.ChunkId == types.Empty {
					msg += fmt.Sprintf("paired insurance(id: %d) have empty paired chunk\n", insurance.Id)
					brokenCount++
					return false, nil
				}
				chunk, found := k.GetChunk(ctx, insurance.ChunkId)
				if !found {
					msg += fmt.Sprintf("not found paired chunk for paired insurance(id: %d)\n", insurance.Id)
					brokenCount++
					return false, nil
				}
				if chunk.Status != types.CHUNK_STATUS_PAIRED {
					msg += fmt.Sprintf("paired insurance(id: %d) have invalid paired chunk status: %s\n", insurance.Id, chunk.Status)
					brokenCount++
					return false, nil
				}
			case types.INSURANCE_STATUS_UNPAIRING:
				// must have chunk to protect
				if insurance.ChunkId == types.Empty {
					msg += fmt.Sprintf("unpairing insurance(id: %d) have empty chunk to protect\n", insurance.Id)
					brokenCount++
					return false, nil
				}
				_, found := k.GetChunk(ctx, insurance.ChunkId)
				if !found {
					msg += fmt.Sprintf("not found chunk to protect for unpairing insurance(id: %d)\n", insurance.Id)
					brokenCount++
					return false, nil
				}

			case types.INSURANCE_STATUS_UNPAIRED:
				// must have empty chunk
				if insurance.ChunkId != types.Empty {
					msg += fmt.Sprintf("unpaired insurance(id: %d) have non-empty paired chunk\n", insurance.Id)
					brokenCount++
					return false, nil
				}
			case types.INSURANCE_STATUS_UNPAIRING_FOR_WITHDRAWAL:
				// must have chunk to protect
				if insurance.ChunkId == types.Empty {
					msg += fmt.Sprintf("unpairing for withdrawal insurance(id: %d) have empty chunk to protect\n", insurance.Id)
					brokenCount++
					return false, nil
				}
				chunk, found := k.GetChunk(ctx, insurance.ChunkId)
				if !found {
					msg += fmt.Sprintf("not found chunk to protect for unpairing for withdrawal insurance(id: %d)\n", insurance.Id)
					brokenCount++
					return false, nil
				}
				if chunk.Status == types.CHUNK_STATUS_PAIRING {
					msg += fmt.Sprintf("unpairing for withdrawal insurance(id: %d) have invalid chunk status: %s\n", insurance.Id, chunk.Status)
					brokenCount++
					return false, nil
				}
			default:
				msg += fmt.Sprintf("insurance(id: %d) have invalid status: %s\n", insurance.Id, insurance.Status)
				brokenCount++
				return false, nil
			}
			return false, nil
		})
		if brokenCount > 0 {
			return sdk.FormatInvariant(types.ModuleName, "insurances", fmt.Sprintf(
				"found %d broken insurances:\n%s", brokenCount, msg)), true
		} else {
			return "", false
		}
	}
}

func UnpairingForUnstakingChunkInfosInvariant(k Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		msg := ""
		brokenCount := 0

		infos := k.GetAllUnpairingForUnstakingChunkInfos(ctx)
		for _, info := range infos {
			// get chunk from chunk id
			chunk, found := k.GetChunk(ctx, info.ChunkId)
			if !found {
				msg += fmt.Sprintf("not found chunk(id: %d) for unpairing for unstaking chunk info\n", info.ChunkId)
				brokenCount++
				continue
			}
			if chunk.Status != types.CHUNK_STATUS_PAIRED &&
				chunk.Status != types.CHUNK_STATUS_UNPAIRING_FOR_UNSTAKING {
				msg += fmt.Sprintf("chunk(id: %d) have invalid status: %s\n", chunk.Id, chunk.Status)
				brokenCount++
				continue
			}
		}
		if brokenCount > 0 {
			return sdk.FormatInvariant(types.ModuleName, "unpairing for unstaking chunk infos", fmt.Sprintf(
				"found %d broken unpairing for unstaking chunk infos:\n%s", brokenCount, msg)), true
		} else {
			return "", false
		}
	}
}

func WithdrawInsuranceRequestsInvariant(k Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		msg := ""
		brokenCount := 0

		reqs := k.GetAllWithdrawInsuranceRequests(ctx)
		for _, req := range reqs {
			// get insurance from insurance id
			insurance, found := k.GetInsurance(ctx, req.InsuranceId)
			if !found {
				msg += fmt.Sprintf("not found insurance(id: %d) for withdraw insurance request\n", req.InsuranceId)
				brokenCount++
				continue
			}
			if insurance.Status != types.INSURANCE_STATUS_PAIRED {
				msg += fmt.Sprintf("insurance(id: %d) have invalid status: %s\n", insurance.Id, insurance.Status)
				brokenCount++
				continue
			}
		}
		if brokenCount > 0 {
			return sdk.FormatInvariant(types.ModuleName, "withdraw insurance requests", fmt.Sprintf(
				"found %d broken withdraw insurance requests:\n%s", brokenCount, msg)), true
		} else {
			return "", false
		}
	}
}
