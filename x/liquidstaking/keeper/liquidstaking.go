package keeper

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/cosmos/cosmos-sdk/x/evidence/exported"
	evidencetypes "github.com/cosmos/cosmos-sdk/x/evidence/types"

	"github.com/Canto-Network/Canto/v7/x/liquidstaking/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

// CoverRedelegationPenalty covers the penalty of re-delegation from unpairing insurance.
func (k Keeper) CoverRedelegationPenalty(ctx sdk.Context) {
	bondDenom := k.stakingKeeper.BondDenom(ctx)
	// For all paired chunks, if chunk have an unpairing insurance, then
	// this chunk is re-delegation on-goning.
	k.IterateAllRedelegationInfos(ctx, func(info types.RedelegationInfo) bool {
		chunk, srcIns, dstIns, entry := k.mustValidateRedelegationInfo(ctx, info)
		dstDel := k.stakingKeeper.Delegation(ctx, chunk.DerivedAddress(), dstIns.GetValidator())
		diff := entry.SharesDst.Sub(dstDel.GetShares())
		if diff.IsPositive() {
			dstVal, found := k.stakingKeeper.GetValidator(ctx, dstIns.GetValidator())
			if !found {
				panic(fmt.Sprintf("validator: %s not found", dstIns.GetValidator()))
			}
			penaltyAmt := dstVal.TokensFromShares(diff).Ceil().TruncateInt()
			if penaltyAmt.IsPositive() {
				penaltyAmt = k.CalcCeiledPenalty(dstVal, penaltyAmt)
				srcInsBal := k.bankKeeper.GetBalance(ctx, srcIns.DerivedAddress(), bondDenom)
				if srcInsBal.Amount.LT(penaltyAmt) {
					panic(fmt.Sprintf(
						"unpairing insurance: %s cannot cover penalty during re-delegation: %s",
						srcIns.DerivedAddress(), penaltyAmt.String()),
					)
				}
				// happy case: unpairing insurance can cover penalty, so cover it.
				if err := k.bankKeeper.SendCoins(
					ctx, srcIns.DerivedAddress(), chunk.DerivedAddress(), sdk.NewCoins(sdk.NewCoin(bondDenom, penaltyAmt)),
				); err != nil {
					panic(err)
				}
				k.mustDelegatePenaltyAmt(ctx, chunk, penaltyAmt, dstVal, srcIns.Id, types.AttributeValueReasonUnpairingInsCoverPenalty)
			}
		}
		return false
	})
}

// CollectRewardAndFee collects reward of chunk and
// distributes it to insurance, dynamic fee and reward module account.
// 1. Send commission to insurance based on chunk reward.
// 2. Deduct dynamic fee from remaining and burn it.
// 3. Send rest of rewards to reward module account.
func (k Keeper) CollectRewardAndFee(
	ctx sdk.Context, dynamicFeeRate sdk.Dec, chunk types.Chunk, ins types.Insurance,
) {
	// At upper callstack(=DistributeReward), we withdrawed delegation reward of chunk.
	// So balance of chunk is delegation reward.
	delRewards := k.bankKeeper.SpendableCoins(ctx, chunk.DerivedAddress())
	var insCommissions sdk.Coins
	var dynamicFees sdk.Coins
	var remainingRewards sdk.Coins

	for _, delRewardCoin := range delRewards {
		insuranceCommissionAmt := delRewardCoin.Amount.ToDec().Mul(ins.FeeRate).TruncateInt()
		if insuranceCommissionAmt.IsPositive() {
			insCommissions = insCommissions.Add(sdk.NewCoin(delRewardCoin.Denom, insuranceCommissionAmt))
		}

		pureRewardAmt := delRewardCoin.Amount.Sub(insuranceCommissionAmt)
		dynamicFeeAmt := pureRewardAmt.ToDec().Mul(dynamicFeeRate).Ceil().TruncateInt()
		remainingRewardAmt := pureRewardAmt.Sub(dynamicFeeAmt)

		if dynamicFeeAmt.IsPositive() {
			dynamicFees = dynamicFees.Add(sdk.NewCoin(delRewardCoin.Denom, dynamicFeeAmt))
		}
		if remainingRewardAmt.IsPositive() {
			remainingRewards = remainingRewards.Add(sdk.NewCoin(delRewardCoin.Denom, remainingRewardAmt))
		}
	}

	var inputs []banktypes.Input
	var outputs []banktypes.Output
	switch delRewards.Len() {
	case 0:
		return
	default:
		// Dynamic Fee can be zero if the utilization rate is low.
		if dynamicFees.IsValid() && dynamicFees.IsAllPositive() {
			// Collect dynamic fee and burn it first.
			if err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, chunk.DerivedAddress(), types.ModuleName, dynamicFees); err != nil {
				panic(err)
			}
			if err := k.bankKeeper.BurnCoins(ctx, types.ModuleName, dynamicFees); err != nil {
				panic(err)
			}
		}

		// If insurance fee rate was zero, insurance commissions are not positive.
		if insCommissions.IsValid() && insCommissions.IsAllPositive() {
			inputs = append(inputs, banktypes.NewInput(chunk.DerivedAddress(), insCommissions))
			outputs = append(outputs, banktypes.NewOutput(ins.FeePoolAddress(), insCommissions))
		}
		if remainingRewards.IsValid() && remainingRewards.IsAllPositive() {
			inputs = append(inputs, banktypes.NewInput(chunk.DerivedAddress(), remainingRewards))
			outputs = append(outputs, banktypes.NewOutput(types.RewardPool, remainingRewards))
		}
	}
	if err := k.bankKeeper.InputOutputCoins(ctx, inputs, outputs); err != nil {
		panic(err)
	}
}

// DistributeReward withdraws delegation rewards from all paired chunks
// Keeper.CollectRewardAndFee will be called during withdrawing process.
func (k Keeper) DistributeReward(ctx sdk.Context) {
	nase, _, _, _ := k.GetNetAmountStateEssentials(ctx)
	k.IterateAllChunks(ctx, func(chunk types.Chunk) bool {
		if chunk.Status != types.CHUNK_STATUS_PAIRED {
			return false
		}
		pairedIns, validator, _ := k.mustValidatePairedChunk(ctx, chunk)
		_, err := k.distributionKeeper.WithdrawDelegationRewards(ctx, chunk.DerivedAddress(), validator.GetOperator())
		if err != nil {
			panic(err)
		}

		k.CollectRewardAndFee(ctx, nase.FeeRate, chunk, pairedIns)
		return false
	})
}

// CoverSlashingAndHandleMatureUnbondings covers slashing and handles mature unbondings.
func (k Keeper) CoverSlashingAndHandleMatureUnbondings(ctx sdk.Context) {
	k.IterateAllChunks(ctx, func(chunk types.Chunk) bool {
		switch chunk.Status {
		// Finish mature unbondings triggered in previous epoch
		case types.CHUNK_STATUS_UNPAIRING_FOR_UNSTAKING:
			k.completeLiquidUnstake(ctx, chunk)

		case types.CHUNK_STATUS_UNPAIRING:
			k.handleUnpairingChunk(ctx, chunk)

		case types.CHUNK_STATUS_PAIRED:
			k.handlePairedChunk(ctx, chunk)
		}
		return false
	})
}

// RemoveDeletableRedelegationInfos remove infos which are matured and deletable.
func (k Keeper) RemoveDeletableRedelegationInfos(ctx sdk.Context) {
	k.IterateAllRedelegationInfos(ctx, func(info types.RedelegationInfo) bool {
		if info.Matured(ctx.BlockTime()) {
			k.DeleteRedelegationInfo(ctx, info.ChunkId)
		}
		return false
	})
	return
}

// HandleQueuedLiquidUnstakes processes unstaking requests that were queued before the epoch.
func (k Keeper) HandleQueuedLiquidUnstakes(ctx sdk.Context) []types.Chunk {
	var unstakedChunks []types.Chunk
	var completionTime time.Time
	var chunkIds []string
	var err error
	k.IterateAllUnpairingForUnstakingChunkInfos(ctx, func(info types.UnpairingForUnstakingChunkInfo) bool {
		// Get chunk
		chunk := k.mustGetChunk(ctx, info.ChunkId)
		if chunk.Status != types.CHUNK_STATUS_PAIRED {
			// When it is queued with chunk, it must be paired but not now.
			// (e.g. validator got huge slash after unstake request is queued, so the chunk is not valid now)
			return false
		}
		ins, _, del := k.mustValidatePairedChunk(ctx, chunk)
		completionTime, err = k.stakingKeeper.Undelegate(
			ctx, chunk.DerivedAddress(), ins.GetValidator(), del.GetShares(),
		)
		if err != nil {
			panic(err)
		}
		_, chunk = k.startUnpairingForLiquidUnstake(ctx, ins, chunk)
		unstakedChunks = append(unstakedChunks, chunk)
		chunkIds = append(chunkIds, strconv.FormatUint(chunk.Id, 10))
		return false
	})
	if len(chunkIds) > 0 {
		ctx.EventManager().EmitEvents(sdk.Events{
			sdk.NewEvent(
				types.EventTypeBeginLiquidUnstake,
				sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
				sdk.NewAttribute(types.AttributeKeyChunkIds, strings.Join(chunkIds, ", ")),
				sdk.NewAttribute(types.AttributeKeyCompletionTime, completionTime.Format(time.RFC3339)),
			),
		})
	}
	return unstakedChunks
}

// HandleUnprocessedQueuedLiquidUnstakes checks if there are any unprocessed queued liquid unstakes.
// And if there are any, refund the escrowed ls tokens to requester and delete the info.
func (k Keeper) HandleUnprocessedQueuedLiquidUnstakes(ctx sdk.Context) {
	k.IterateAllUnpairingForUnstakingChunkInfos(ctx, func(info types.UnpairingForUnstakingChunkInfo) bool {
		chunk := k.mustGetChunk(ctx, info.ChunkId)
		if chunk.Status != types.CHUNK_STATUS_UNPAIRING_FOR_UNSTAKING {
			// Unstaking is not processed. Let's refund the chunk and delete info.
			if err := k.bankKeeper.SendCoins(ctx, types.LsTokenEscrowAcc, info.GetDelegator(), sdk.NewCoins(info.EscrowedLstokens)); err != nil {
				panic(err)
			}
			k.DeleteUnpairingForUnstakingChunkInfo(ctx, info.ChunkId)
			ctx.EventManager().EmitEvents(sdk.Events{
				sdk.NewEvent(
					types.EventTypeDeleteQueuedLiquidUnstake,
					sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
					sdk.NewAttribute(types.AttributeKeyDelegator, info.DelegatorAddress),
				),
			})
		}
		return false
	})
}

// HandleQueuedWithdrawInsuranceRequests processes withdraw insurance requests that were queued before the epoch.
// Unpairing insurances will be unpaired in the next epoch.is unpaired.
func (k Keeper) HandleQueuedWithdrawInsuranceRequests(ctx sdk.Context) []types.Insurance {
	var withdrawnInsurances []types.Insurance
	var withdrawnInsIds []string
	k.IterateAllWithdrawInsuranceRequests(ctx, func(req types.WithdrawInsuranceRequest) bool {
		ins := k.mustGetInsurance(ctx, req.InsuranceId)
		if ins.Status != types.INSURANCE_STATUS_PAIRED && ins.Status != types.INSURANCE_STATUS_UNPAIRING {
			panic(fmt.Sprintf("ins %d is not paired or unpairing", ins.Id))
		}

		// get chunk from ins
		chunk := k.mustGetChunk(ctx, ins.ChunkId)
		if chunk.Status == types.CHUNK_STATUS_PAIRED {
			// If not paired, state change already happened in CoverSlashingAndHandleMatureUnbondings
			chunk.SetStatus(types.CHUNK_STATUS_UNPAIRING)
			if ins.Id == chunk.PairedInsuranceId {
				chunk.UnpairingInsuranceId = chunk.PairedInsuranceId
				chunk.EmptyPairedInsurance()
			}
			k.SetChunk(ctx, chunk)
		}
		ins.SetStatus(types.INSURANCE_STATUS_UNPAIRING_FOR_WITHDRAWAL)
		k.SetInsurance(ctx, ins)
		k.DeleteWithdrawInsuranceRequest(ctx, ins.Id)
		withdrawnInsurances = append(withdrawnInsurances, ins)
		withdrawnInsIds = append(withdrawnInsIds, strconv.FormatUint(ins.Id, 10))
		return false
	})
	if len(withdrawnInsIds) > 0 {
		ctx.EventManager().EmitEvents(sdk.Events{
			sdk.NewEvent(
				types.EventTypeBeginWithdrawInsurance,
				sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
				sdk.NewAttribute(types.AttributeKeyInsuranceIds, strings.Join(withdrawnInsIds, ", ")),
			),
		})
	}
	return withdrawnInsurances
}

// GetAllRePairableChunksAndOutInsurances returns all re-pairable chunks and out insurances.
// Re-pairable chunks contains chunks with the following statuses
// - Pairing
// - Paired
// - Unpairing but not in un-bonding
// Out insurances contains insurances with the following statuses
// - Serving unpairing chunk(not damaged) which have no unbonding delegation
// - Paired but the validator is not valid anymore
func (k Keeper) GetAllRePairableChunksAndOutInsurances(ctx sdk.Context) (
	rePairableChunks []types.Chunk, outInsurances []types.Insurance,
	validPairedInsuranceMap map[uint64]struct{},
) {
	validPairedInsuranceMap = make(map[uint64]struct{})
	k.IterateAllChunks(ctx, func(chunk types.Chunk) bool {
		switch chunk.Status {
		case types.CHUNK_STATUS_UNPAIRING:
			err := k.validateUnpairingChunk(ctx, chunk)
			if errors.Is(err, types.ErrMustHaveNoUnbondingDelegation) {
				// unbonding of chunk is triggered because insurance cannot cover the penalty of chunk.
				// In next epoch, insurance send all of it's balance to chunk
				// and all balances of chunk will go to reward pool.
				// After that, insurance will be unpaired also.
				return false
			}
			if err != nil {
				panic(err)
			}
			unpairingIns := k.mustGetInsurance(ctx, chunk.UnpairingInsuranceId)
			outInsurances = append(outInsurances, unpairingIns)
			rePairableChunks = append(rePairableChunks, chunk)
		case types.CHUNK_STATUS_PAIRING:
			rePairableChunks = append(rePairableChunks, chunk)
		case types.CHUNK_STATUS_PAIRED:
			pairedIns, validator, _ := k.mustValidatePairedChunk(ctx, chunk)
			if err := k.ValidateValidator(ctx, validator); err != nil {
				k.startUnpairing(ctx, pairedIns, chunk)
				chunk = k.mustGetChunk(ctx, chunk.Id)
				unpairingIns := k.mustGetInsurance(ctx, chunk.UnpairingInsuranceId)
				outInsurances = append(outInsurances, unpairingIns)
			} else {
				validPairedInsuranceMap[pairedIns.Id] = struct{}{}
			}
			rePairableChunks = append(rePairableChunks, chunk)
		default:
			return false
		}
		return false
	})
	return
}

// RankInsurances ranks insurances and returns following:
// 1. newly ranked insurances
// - rank in insurance which is not paired currently
// - NOTE: no change is needed for already ranked in and paired insurances
// 2. Ranked out insurances
// - current unpairing insurances + paired insurances which is failed to rank in
func (k Keeper) RankInsurances(ctx sdk.Context) (
	newlyRankedInInsurances, rankOutInsurances []types.Insurance,
) {
	candidatesValidatorMap := make(map[string]stakingtypes.Validator)
	rePairableChunks, currentOutInsurances, validPairedInsuranceMap := k.GetAllRePairableChunksAndOutInsurances(ctx)

	// candidateInsurances will be ranked
	var candidateInsurances []types.Insurance
	k.IterateAllInsurances(ctx, func(ins types.Insurance) (stop bool) {
		// Only pairing or paired insurances are candidates to be ranked
		if ins.Status != types.INSURANCE_STATUS_PAIRED &&
			ins.Status != types.INSURANCE_STATUS_PAIRING {
			return false
		}
		if _, ok := candidatesValidatorMap[ins.ValidatorAddress]; !ok {
			// Only insurance which directs valid validator can be ranked in
			validator, found := k.stakingKeeper.GetValidator(ctx, ins.GetValidator())
			if !found {
				return false
			}
			if err := k.ValidateValidator(ctx, validator); err != nil {
				return false
			}
			candidatesValidatorMap[ins.ValidatorAddress] = validator
		}
		candidateInsurances = append(candidateInsurances, ins)
		return false
	})

	types.SortInsurances(candidatesValidatorMap, candidateInsurances, false)
	var rankInInsurances []types.Insurance
	var rankOutCandidates []types.Insurance
	if len(rePairableChunks) > len(candidateInsurances) {
		// All candidates can be ranked in because there are enough chunks
		rankInInsurances = candidateInsurances
	} else {
		// There are more candidates than chunks so we need to decide which candidates are ranked in or out
		rankInInsurances = candidateInsurances[:len(rePairableChunks)]
		rankOutCandidates = candidateInsurances[len(rePairableChunks):]
	}

	for _, ins := range rankOutCandidates {
		if ins.Status == types.INSURANCE_STATUS_PAIRED {
			rankOutInsurances = append(rankOutInsurances, ins)
		}
	}
	rankOutInsurances = append(rankOutInsurances, currentOutInsurances...)

	for _, ins := range rankInInsurances {
		// If insurance is already paired, we just skip it
		// because it is already ranked in and paired so there are no changes.
		if _, ok := validPairedInsuranceMap[ins.Id]; !ok {
			newlyRankedInInsurances = append(newlyRankedInInsurances, ins)
		}
	}
	return
}

// RePairRankedInsurances re-pairs ranked insurances.
func (k Keeper) RePairRankedInsurances(
	ctx sdk.Context, newlyRankedInInsurances, rankOutInsurances []types.Insurance,
) {
	// create rankOutInsChunkMap to fast access chunk by rank out insurance id
	var rankOutInsChunkMap = make(map[uint64]types.Chunk)
	for _, outIns := range rankOutInsurances {
		chunk := k.mustGetChunk(ctx, outIns.ChunkId)
		rankOutInsChunkMap[outIns.Id] = chunk
	}

	// newInsurancesWithDifferentValidators will replace out insurance by re-delegation
	// because there are no rank out insurances which have same validator
	var newInsurancesWithDifferentValidators []types.Insurance

	// Create handledOutInsurances map to track which out insurances are handled
	handledOutInsurances := make(map[uint64]struct{})
	// Short circuit
	// Try to replace outInsurance with newRankInInsurance which has same validator.
	for _, newRankInIns := range newlyRankedInInsurances {
		hasSameValidator := false
		for _, outIns := range rankOutInsurances {
			if _, ok := handledOutInsurances[outIns.Id]; ok {
				continue
			}
			// Happy case. Same validator so we can skip re-delegation
			if newRankInIns.ValidatorAddress == outIns.ValidatorAddress {
				// get chunk by outIns.ChunkId
				chunk := k.mustGetChunk(ctx, outIns.ChunkId)
				k.rePairChunkAndInsurance(ctx, chunk, newRankInIns, outIns)
				hasSameValidator = true
				// mark outIns as handled, so we will not handle it again
				handledOutInsurances[outIns.Id] = struct{}{}
				break
			}
		}
		if !hasSameValidator {
			newInsurancesWithDifferentValidators = append(newInsurancesWithDifferentValidators, newRankInIns)
		}
	}

	// Which ranked-out insurances are not handled yet?
	remainedOutInsurances := make([]types.Insurance, 0)
	for _, outIns := range rankOutInsurances {
		if _, ok := handledOutInsurances[outIns.Id]; !ok {
			remainedOutInsurances = append(remainedOutInsurances, outIns)
		}
	}

	// pairing chunks are immediately pairable, so just delegate it.
	var pairingChunks []types.Chunk
	pairingChunks = k.GetAllPairingChunks(ctx)
	bondDenom := k.stakingKeeper.BondDenom(ctx)
	for len(pairingChunks) > 0 && len(newInsurancesWithDifferentValidators) > 0 {
		// pop first chunk
		chunk := pairingChunks[0]
		pairingChunks = pairingChunks[1:]

		// pop cheapest insurance
		newIns := newInsurancesWithDifferentValidators[0]
		newInsurancesWithDifferentValidators = newInsurancesWithDifferentValidators[1:]

		validator, found := k.stakingKeeper.GetValidator(ctx, newIns.GetValidator())
		if !found {
			panic(fmt.Sprintf("validator not found(validator: %s, newInsuranceId: %d)", newIns.GetValidator(), newIns.Id))
		}

		// pairing chunk is immediately pairable so just delegate it
		chunkBal := k.bankKeeper.GetBalance(ctx, chunk.DerivedAddress(), bondDenom)
		if chunkBal.Amount.LT(types.ChunkSize) {
			panic(fmt.Sprintf("pairing chunk balance is below chunk size(bal: %s, chunkId: %d)", chunkBal, chunk.Id))
		}
		_, _, newShares, err := k.mustPairChunkAndDelegate(ctx, chunk, newIns, newIns.GetValidator(), chunkBal.Amount)
		if err != nil {
			panic(err)
		}
		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypeDelegate,
				sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
				sdk.NewAttribute(types.AttributeKeyChunkId, fmt.Sprintf("%d", chunk.Id)),
				sdk.NewAttribute(types.AttributeKeyInsuranceId, fmt.Sprintf("%d", newIns.Id)),
				sdk.NewAttribute(types.AttributeKeyDelegator, chunk.DerivedAddress().String()),
				sdk.NewAttribute(types.AttributeKeyValidator, validator.GetOperator().String()),
				sdk.NewAttribute(sdk.AttributeKeyAmount, types.ChunkSize.String()),
				sdk.NewAttribute(types.AttributeKeyNewShares, newShares.String()),
				sdk.NewAttribute(types.AttributeKeyReason, types.AttributeValueReasonPairingChunkPaired),
			),
		)
	}

	// reset handledOutInsurances to track which out insurances are handled
	handledOutInsurances = make(map[uint64]struct{})
	// rest of rankOutInsurances are replaced with newInsurancesWithDifferentValidators by re-delegation
	// if there are remaining newInsurancesWithDifferentValidators
	for _, outIns := range remainedOutInsurances {
		if len(newInsurancesWithDifferentValidators) == 0 {
			// We don't have any new insurance to replace
			break
		}
		srcVal := outIns.GetValidator()
		// We don't allow chunks to re-delegate from Unbonding validator.
		// Because we cannot expect when this re-delegation will be completed. (It depends on unbonding time of validator).
		// Current version of this module exepects that re-delegation will be completed at endblocker of staking module in next epoch.
		// But if validator is unbonding, it will be completed before the epoch so we cannot track it.
		if k.stakingKeeper.Validator(ctx, srcVal).IsUnbonding() {
			continue
		}

		// Pop cheapest insurance
		newIns := newInsurancesWithDifferentValidators[0]
		newInsurancesWithDifferentValidators = newInsurancesWithDifferentValidators[1:]
		chunk := rankOutInsChunkMap[outIns.Id]

		// get delegation shares of srcValidator
		delegation, found := k.stakingKeeper.GetDelegation(ctx, chunk.DerivedAddress(), outIns.GetValidator())
		if !found {
			panic(fmt.Sprintf("delegation not found(delegator: %s, validator: %s)", chunk.DerivedAddress(), outIns.GetValidator()))
		}
		completionTime, err := k.stakingKeeper.BeginRedelegation(
			ctx, chunk.DerivedAddress(), outIns.GetValidator(), newIns.GetValidator(), delegation.GetShares(),
		)
		if err != nil {
			panic(err)
		}

		if !k.stakingKeeper.Validator(ctx, srcVal).IsUnbonded() {
			// Start to track new redelegation which will be completed at next epoch.
			// We track it because some additional slashing can happened during re-delegation period.
			// If src validator is already unbonded then we don't track it because it immediately re-delegated.
			k.SetRedelegationInfo(ctx, types.NewRedelegationInfo(chunk.Id, completionTime))
		}
		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypeBeginRedelegate,
				sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
				sdk.NewAttribute(types.AttributeKeyChunkId, fmt.Sprintf("%d", chunk.Id)),
				sdk.NewAttribute(types.AttributeKeySrcValidator, outIns.ValidatorAddress),
				sdk.NewAttribute(types.AttributeKeyDstValidator, newIns.ValidatorAddress),
				sdk.NewAttribute(types.AttributeKeyCompletionTime, completionTime.Format(time.RFC3339)),
			),
		)
		k.rePairChunkAndInsurance(ctx, chunk, newIns, outIns)
		handledOutInsurances[outIns.Id] = struct{}{}
	}

	// What ranked-out insurances are not handled yet?
	restOutInsurances := make([]types.Insurance, 0)
	for _, outIns := range remainedOutInsurances {
		if _, ok := handledOutInsurances[outIns.Id]; !ok {
			restOutInsurances = append(restOutInsurances, outIns)
		}
	}

	// No more candidate insurances to replace, so just start unbonding.
	for _, outIns := range restOutInsurances {
		chunk := k.mustGetChunk(ctx, outIns.ChunkId)
		chunk.SetStatus(types.CHUNK_STATUS_UNPAIRING)
		chunk.EmptyPairedInsurance()
		chunk.UnpairingInsuranceId = outIns.Id
		k.SetChunk(ctx, chunk)
		if outIns.Status != types.INSURANCE_STATUS_UNPAIRING_FOR_WITHDRAWAL &&
			outIns.Status != types.INSURANCE_STATUS_UNPAIRING {
			outIns.Status = types.INSURANCE_STATUS_UNPAIRING
			k.SetInsurance(ctx, outIns)
		}
		// get delegation shares of out insurance
		delegation, found := k.stakingKeeper.GetDelegation(ctx, chunk.DerivedAddress(), outIns.GetValidator())
		if !found {
			panic(fmt.Sprintf("delegation not found(chunkId: %d, validator: %s)", chunk.Id, outIns.GetValidator()))
		}
		// validate unbond amount
		completionTime, err := k.stakingKeeper.Undelegate(ctx, chunk.DerivedAddress(), outIns.GetValidator(), delegation.GetShares())
		if err != nil {
			panic(err)
		}
		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypeBeginUndelegate,
				sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
				sdk.NewAttribute(types.AttributeKeyChunkId, fmt.Sprintf("%d", chunk.Id)),
				sdk.NewAttribute(types.AttributeKeyValidator, outIns.ValidatorAddress),
				sdk.NewAttribute(types.AttributeKeyCompletionTime, completionTime.Format(time.RFC3339)),
				sdk.NewAttribute(types.AttributeKeyReason, types.AttributeValueReasonNoCandidateIns),
			),
		)
		continue
	}
}

func (k Keeper) DoLiquidStake(ctx sdk.Context, msg *types.MsgLiquidStake) (
	chunks []types.Chunk, totalNewShares sdk.Dec, totalLsTokenMintAmount sdk.Int, err error,
) {
	delAddr := msg.GetDelegator()
	amount := msg.Amount

	if err = k.ShouldBeBondDenom(ctx, amount.Denom); err != nil {
		return
	}
	// Liquid stakers can send amount of tokens corresponding multiple of chunk size and create multiple chunks
	if err = k.ShouldBeMultipleOfChunkSize(amount.Amount); err != nil {
		return
	}
	chunksToCreate := amount.Amount.Quo(types.ChunkSize)

	nase, _, _, _ := k.GetNetAmountStateEssentials(ctx)

	if nase.RemainingChunkSlots.LT(chunksToCreate) {
		err = sdkerrors.Wrapf(
			types.ErrExceedAvailableChunks,
			"requested chunks to create: %d, available chunks: %s",
			chunksToCreate,
			nase.RemainingChunkSlots.String(),
		)
		return
	}

	pairingInsurances, validatorMap := k.GetPairingInsurances(ctx)

	numPairingInsurances := sdk.NewIntFromUint64(uint64(len(pairingInsurances)))
	if chunksToCreate.GT(numPairingInsurances) {
		err = types.ErrNoPairingInsurance
		return
	}

	types.SortInsurances(validatorMap, pairingInsurances, false)
	totalNewShares, newShares := sdk.ZeroDec(), sdk.ZeroDec()
	totalLsTokenMintAmount, lsTokenMintAmount := sdk.ZeroInt(), sdk.ZeroInt()
	liquidBondDenom := k.GetLiquidBondDenom(ctx)
	chunkSizeCoins := sdk.NewCoins(sdk.NewCoin(amount.Denom, types.ChunkSize))
	for {
		if !chunksToCreate.IsPositive() {
			break
		}
		cheapestIns := pairingInsurances[0]
		pairingInsurances = pairingInsurances[1:]

		// Now we have the cheapest pairing insurance and valid msg liquid stake! Let's create a chunk
		// Create a chunk
		chunkId := k.getNextChunkIdWithUpdate(ctx)
		chunk := types.NewChunk(chunkId)

		// Escrow liquid staker's coins
		if err = k.bankKeeper.SendCoins(ctx, delAddr, chunk.DerivedAddress(), chunkSizeCoins); err != nil {
			return
		}

		// Delegate to the validator
		// Delegator: DerivedAddress(chunk.Id)
		// Validator: insurance.ValidatorAddress
		// Amount: msg.Amount
		chunk, cheapestIns, newShares, err = k.mustPairChunkAndDelegate(
			ctx, chunk, cheapestIns, cheapestIns.GetValidator(), types.ChunkSize,
		)
		if err != nil {
			return
		}
		totalNewShares = totalNewShares.Add(newShares)

		// Mint the liquid staking token
		lsTokenMintAmount = types.ChunkSize
		if nase.LsTokensTotalSupply.IsPositive() {
			lsTokenMintAmount = nase.MintRate.MulTruncate(types.ChunkSize.ToDec()).TruncateInt()
		}
		if !lsTokenMintAmount.IsPositive() {
			err = sdkerrors.Wrapf(types.ErrInvalidAmount, "amount must be greater than or equal to %s", amount.String())
			return
		}
		mintedCoin := sdk.NewCoin(liquidBondDenom, lsTokenMintAmount)
		if err = k.bankKeeper.MintCoins(ctx, types.ModuleName, sdk.NewCoins(mintedCoin)); err != nil {
			return
		}
		totalLsTokenMintAmount = totalLsTokenMintAmount.Add(lsTokenMintAmount)
		if err = k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, delAddr, sdk.NewCoins(mintedCoin)); err != nil {
			return
		}
		chunks = append(chunks, chunk)
		chunksToCreate = chunksToCreate.Sub(sdk.OneInt())
	}
	return
}

// QueueLiquidUnstake queues MsgLiquidUnstake.
// Actual unstaking will be done in the next epoch.
func (k Keeper) QueueLiquidUnstake(ctx sdk.Context, msg *types.MsgLiquidUnstake) (
	toBeUnstakedChunks []types.Chunk,
	infos []types.UnpairingForUnstakingChunkInfo,
	err error,
) {
	delAddr := msg.GetDelegator()
	amount := msg.Amount

	if err = k.ShouldBeBondDenom(ctx, amount.Denom); err != nil {
		return
	}
	if err = k.ShouldBeMultipleOfChunkSize(amount.Amount); err != nil {
		return
	}

	chunksToLiquidUnstake := amount.Amount.Quo(types.ChunkSize)

	nase, pairedChunksWithInsuranceId, pairedInsurances, validatorMap := k.GetNetAmountStateEssentials(ctx)

	// purelyPairedInsurances contains paired insurances which serve chunk which is not in queue for unstaking.
	var purelyPairedInsurances []types.Insurance
	for _, pairedIns := range pairedInsurances {
		chunk := pairedChunksWithInsuranceId[pairedIns.Id]
		// check whether the chunk is already have unstaking requests in queue.
		_, found := k.GetUnpairingForUnstakingChunkInfo(ctx, chunk.Id)
		if found {
			delete(pairedChunksWithInsuranceId, pairedIns.Id)
			continue
		}
		purelyPairedInsurances = append(purelyPairedInsurances, pairedIns)
	}

	pairedChunks := sdk.NewIntFromUint64(uint64(len(pairedChunksWithInsuranceId)))
	if pairedChunks.IsZero() {
		err = types.ErrNoPairedChunk
		return
	}
	if chunksToLiquidUnstake.GT(pairedChunks) {
		err = sdkerrors.Wrapf(
			types.ErrExceedAvailableChunks,
			"requested chunks to liquid unstake: %d, paired chunks: %d",
			chunksToLiquidUnstake,
			pairedChunks,
		)
		return
	}
	// Sort insurances by descend order
	types.SortInsurances(validatorMap, purelyPairedInsurances, true)

	// How much ls tokens must be burned

	liquidBondDenom := k.GetLiquidBondDenom(ctx)
	for {
		if !chunksToLiquidUnstake.IsPositive() {
			break
		}
		// Escrow ls tokens from the delegator
		lsTokenBurnAmount := types.ChunkSize
		if nase.LsTokensTotalSupply.IsPositive() {
			lsTokenBurnAmount = lsTokenBurnAmount.ToDec().Mul(nase.MintRate).TruncateInt()
		}
		lsTokensToBurn := sdk.NewCoin(liquidBondDenom, lsTokenBurnAmount)
		if err = k.bankKeeper.SendCoins(
			ctx, delAddr, types.LsTokenEscrowAcc, sdk.NewCoins(lsTokensToBurn),
		); err != nil {
			return
		}

		mostExpensiveInsurance := purelyPairedInsurances[0]
		purelyPairedInsurances = purelyPairedInsurances[1:]
		chunkToBeUndelegated := pairedChunksWithInsuranceId[mostExpensiveInsurance.Id]
		_, found := k.GetUnpairingForUnstakingChunkInfo(ctx, chunkToBeUndelegated.Id)
		if found {
			err = sdkerrors.Wrapf(
				types.ErrAlreadyInQueue,
				"chunk id: %d, delegator address: %s",
				chunkToBeUndelegated.Id,
				msg.DelegatorAddress,
			)
			return
		}

		info := types.NewUnpairingForUnstakingChunkInfo(
			chunkToBeUndelegated.Id,
			msg.DelegatorAddress,
			lsTokensToBurn,
		)
		toBeUnstakedChunks = append(toBeUnstakedChunks, pairedChunksWithInsuranceId[mostExpensiveInsurance.Id])
		infos = append(infos, info)
		k.SetUnpairingForUnstakingChunkInfo(ctx, info)
		chunksToLiquidUnstake = chunksToLiquidUnstake.Sub(sdk.OneInt())
	}
	return
}

func (k Keeper) DoProvideInsurance(ctx sdk.Context, msg *types.MsgProvideInsurance) (ins types.Insurance, err error) {
	providerAddr := msg.GetProvider()
	valAddr := msg.GetValidator()
	feeRate := msg.FeeRate
	amount := msg.Amount

	if err = k.ShouldBeBondDenom(ctx, amount.Denom); err != nil {
		return
	}
	// Check if the amount is greater than minimum coverage
	_, minimumCollateral := k.GetMinimumRequirements(ctx)
	if amount.IsLT(minimumCollateral) {
		err = sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "amount must be greater than minimum collateral: %s", minimumCollateral)
		return
	}

	// Check if the validator is valid
	validator, found := k.stakingKeeper.GetValidator(ctx, valAddr)
	if !found {
		err = sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "validator does not exist: %s", valAddr.String())
		return
	}
	if err = k.ValidateValidator(ctx, validator); err != nil {
		return
	}

	if feeRate.Add(validator.GetCommission()).GTE(types.MaximumInsValFeeRate) {
		err = sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "fee rate(validator fee rate + insurance fee rate) must be less than %s", types.MaximumInsValFeeRate.String())
		return
	}

	// Create an insurnace
	insId := k.getNextInsuranceIdWithUpdate(ctx)
	ins = types.NewInsurance(insId, providerAddr.String(), valAddr.String(), feeRate)

	// Escrow provider's balance
	if err = k.bankKeeper.SendCoins(ctx, providerAddr, ins.DerivedAddress(), sdk.NewCoins(amount)); err != nil {
		return
	}
	k.SetInsurance(ctx, ins)

	return
}

func (k Keeper) DoCancelProvideInsurance(ctx sdk.Context, msg *types.MsgCancelProvideInsurance) (ins types.Insurance, err error) {
	providerAddr := msg.GetProvider()
	insId := msg.Id

	if ins, err = k.validateInsurance(ctx, insId, providerAddr, types.INSURANCE_STATUS_PAIRING); err != nil {
		return
	}

	// Unescrow provider's balance
	escrowed := k.bankKeeper.SpendableCoins(ctx, ins.DerivedAddress())
	fees := k.bankKeeper.SpendableCoins(ctx, ins.FeePoolAddress())

	var inputs []banktypes.Input
	var outputs []banktypes.Output
	if escrowed.IsValid() && escrowed.IsAllPositive() {
		inputs = append(inputs, banktypes.NewInput(ins.DerivedAddress(), escrowed))
		outputs = append(outputs, banktypes.NewOutput(providerAddr, escrowed))
	}
	if fees.IsValid() && fees.IsAllPositive() {
		inputs = append(inputs, banktypes.NewInput(ins.FeePoolAddress(), fees))
		outputs = append(outputs, banktypes.NewOutput(providerAddr, fees))
	}
	if err := k.bankKeeper.InputOutputCoins(ctx, inputs, outputs); err != nil {
		return ins, err
	}
	k.DeleteInsurance(ctx, insId)
	return
}

// DoWithdrawInsurance withdraws insurance immediately if it is unpaired.
// If it is paired then it will be queued and unpaired at the epoch.
func (k Keeper) DoWithdrawInsurance(ctx sdk.Context, msg *types.MsgWithdrawInsurance) (
	ins types.Insurance, req types.WithdrawInsuranceRequest, err error,
) {
	if ins, err = k.validateInsurance(ctx, msg.Id, msg.GetProvider(), types.INSURANCE_STATUS_UNSPECIFIED); err != nil {
		return
	}

	// If insurnace is paired or unpairing, then queue request
	// If insurnace is unpaired then immediately withdraw ins
	switch ins.Status {
	case types.INSURANCE_STATUS_PAIRED:
		req = types.NewWithdrawInsuranceRequest(msg.Id)
		k.SetWithdrawInsuranceRequest(ctx, req)
	case types.INSURANCE_STATUS_UNPAIRED:
		// Withdraw immediately
		err = k.withdrawInsurance(ctx, ins)
	default:
		err = sdkerrors.Wrapf(types.ErrNotInWithdrawableStatus, "ins status: %s", ins.Status)
	}
	return
}

// DoWithdrawInsuranceCommission withdraws insurance commission immediately.
func (k Keeper) DoWithdrawInsuranceCommission(
	ctx sdk.Context,
	msg *types.MsgWithdrawInsuranceCommission,
) (feePoolBals sdk.Coins, err error) {
	providerAddr := msg.GetProvider()
	insId := msg.Id

	ins, err := k.validateInsurance(ctx, insId, providerAddr, types.INSURANCE_STATUS_UNSPECIFIED)
	if err != nil {
		return
	}

	feePoolBals = k.bankKeeper.SpendableCoins(ctx, ins.FeePoolAddress())
	if !feePoolBals.IsValid() || !feePoolBals.IsAllPositive() {
		err = sdkerrors.Wrapf(types.ErrInsCommissionsNotWithdrawable, "feePoolBals: %s(insurnaceId: %d)", feePoolBals, ins.Id)
		return
	}
	if err = k.bankKeeper.SendCoins(ctx, ins.FeePoolAddress(), providerAddr, feePoolBals); err != nil {
		return
	}
	insBals := k.bankKeeper.SpendableCoins(ctx, ins.DerivedAddress())
	if insBals.IsZero() && feePoolBals.IsZero() {
		k.DeleteInsurance(ctx, insId)
	}
	return
}

// DoDepositInsurance deposits more coin to insurance.
func (k Keeper) DoDepositInsurance(ctx sdk.Context, msg *types.MsgDepositInsurance) (err error) {
	providerAddr := msg.GetProvider()
	insuranceId := msg.Id
	amount := msg.Amount

	ins, err := k.validateInsurance(ctx, insuranceId, providerAddr, types.INSURANCE_STATUS_UNSPECIFIED)
	if err != nil {
		return
	}

	if err = k.ShouldBeBondDenom(ctx, amount.Denom); err != nil {
		return
	}
	if err = k.bankKeeper.SendCoins(ctx, providerAddr, ins.DerivedAddress(), sdk.NewCoins(amount)); err != nil {
		return
	}

	return
}

// DoClaimDiscountedReward claims discounted reward by paying lstoken.
func (k Keeper) DoClaimDiscountedReward(ctx sdk.Context, msg *types.MsgClaimDiscountedReward) (
	claim sdk.Coins,
	discountedMintRate sdk.Dec,
	err error,
) {
	if err = k.ShouldBeLiquidBondDenom(ctx, msg.Amount.Denom); err != nil {
		return
	}

	nase, _, _, _ := k.GetNetAmountStateEssentials(ctx)
	// discount rate >= minimum discount rate
	// if discount rate(e.g. 10%) is lower than minimum discount rate(e.g. 20%), then it is not profitable to claim reward.
	if nase.DiscountRate.LT(msg.MinimumDiscountRate) {
		err = sdkerrors.Wrapf(types.ErrDiscountRateTooLow, "current discount rate: %s", nase.DiscountRate)
		return
	}
	discountedMintRate = nase.MintRate.Mul(sdk.OneDec().Sub(nase.DiscountRate))

	var claimableCoin sdk.Coin
	var burnAmt sdk.Coin

	rewardPoolCoins := k.bankKeeper.SpendableCoins(ctx, types.RewardPool)
	bondDenom := k.stakingKeeper.BondDenom(ctx)
	claimableCoin = sdk.NewCoin(bondDenom, rewardPoolCoins.AmountOf(bondDenom))
	burnAmt = msg.Amount

	// sanity check to avoid division by zero
	if discountedMintRate.IsZero() {
		err = sdkerrors.Wrapf(types.ErrInvalidAmount, "discounted mint rate must be greater than zero")
		return
	}
	// claim amount = (ls token amount / discounted mint rate)
	claimAmt := burnAmt.Amount.ToDec().QuoTruncate(discountedMintRate).TruncateInt()
	// Requester can claim only up to claimable amount
	if claimAmt.GT(claimableCoin.Amount) {
		// requester cannot claim more than claimable amount
		claimAmt = claimableCoin.Amount
		// burn amount = (claim amount * discounted mint rate)
		burnAmt.Amount = claimAmt.ToDec().Mul(discountedMintRate).Ceil().TruncateInt()
	}

	claimCoins := sdk.NewCoins(sdk.NewCoin(k.stakingKeeper.BondDenom(ctx), claimAmt))
	if err = k.burnLsTokens(ctx, msg.GetRequestser(), burnAmt); err != nil {
		return
	}
	// send claimCoins to requester
	if err = k.bankKeeper.SendCoins(ctx, types.RewardPool, msg.GetRequestser(), claimCoins); err != nil {
		return
	}

	return
}

func (k Keeper) SetLiquidBondDenom(ctx sdk.Context, denom string) {
	store := ctx.KVStore(k.storeKey)
	store.Set(types.KeyPrefixLiquidBondDenom, []byte(denom))
}

func (k Keeper) GetLiquidBondDenom(ctx sdk.Context) string {
	store := ctx.KVStore(k.storeKey)
	return string(store.Get(types.KeyPrefixLiquidBondDenom))
}

func (k Keeper) ValidateValidator(ctx sdk.Context, validator stakingtypes.Validator) error {
	pubKey, err := validator.ConsPubKey()
	if err != nil {
		return err
	}
	if k.slashingKeeper.IsTombstoned(ctx, sdk.ConsAddress(pubKey.Address())) {
		return types.ErrTombstonedValidator
	}

	if validator.GetStatus() == stakingtypes.Unspecified ||
		validator.GetTokens().IsNil() ||
		validator.GetDelegatorShares().IsNil() ||
		validator.InvalidExRate() {
		return types.ErrInvalidValidatorStatus
	}
	return nil
}

// Get minimum requirements for liquid staking
// Liquid staker must provide at least one chunk amount
// Insurance provider must provide at least slashing coverage
func (k Keeper) GetMinimumRequirements(ctx sdk.Context) (oneChunkAmount, slashingCoverage sdk.Coin) {
	bondDenom := k.stakingKeeper.BondDenom(ctx)
	oneChunkAmount = sdk.NewCoin(bondDenom, types.ChunkSize)
	fraction := sdk.MustNewDecFromStr(types.MinimumCollateral)
	slashingCoverage = sdk.NewCoin(bondDenom, oneChunkAmount.Amount.ToDec().Mul(fraction).TruncateInt())
	return
}

// ShouldBeMultipleOfChunkSize returns error if amount is not a multiple of chunk size
func (k Keeper) ShouldBeMultipleOfChunkSize(amount sdk.Int) error {
	if !amount.IsPositive() || !amount.Mod(types.ChunkSize).IsZero() {
		return sdkerrors.Wrapf(types.ErrInvalidAmount, "got: %s", amount.String())
	}
	return nil
}

// ShouldBeBondDenom returns error if denom is not the same as the bond denom
func (k Keeper) ShouldBeBondDenom(ctx sdk.Context, denom string) error {
	if denom == k.stakingKeeper.BondDenom(ctx) {
		return nil
	}
	return sdkerrors.Wrapf(types.ErrInvalidBondDenom, "expected: %s, got: %s", k.stakingKeeper.BondDenom(ctx), denom)
}

func (k Keeper) ShouldBeLiquidBondDenom(ctx sdk.Context, denom string) error {
	if denom == k.GetLiquidBondDenom(ctx) {
		return nil
	}
	return sdkerrors.Wrapf(types.ErrInvalidLiquidBondDenom, "expected: %s, got: %s", k.GetLiquidBondDenom(ctx), denom)
}

func (k Keeper) burnEscrowedLsTokens(ctx sdk.Context, lsTokensToBurn sdk.Coin) error {
	if err := k.ShouldBeLiquidBondDenom(ctx, lsTokensToBurn.Denom); err != nil {
		return err
	}

	if err := k.bankKeeper.SendCoinsFromAccountToModule(
		ctx,
		types.LsTokenEscrowAcc,
		types.ModuleName,
		sdk.NewCoins(lsTokensToBurn),
	); err != nil {
		return err
	}
	if err := k.bankKeeper.BurnCoins(ctx, types.ModuleName, sdk.NewCoins(lsTokensToBurn)); err != nil {
		return err
	}
	return nil
}

func (k Keeper) burnLsTokens(ctx sdk.Context, from sdk.AccAddress, lsTokensToBurn sdk.Coin) error {
	if err := k.ShouldBeLiquidBondDenom(ctx, lsTokensToBurn.Denom); err != nil {
		return err
	}

	if err := k.bankKeeper.SendCoinsFromAccountToModule(
		ctx,
		from,
		types.ModuleName,
		sdk.NewCoins(lsTokensToBurn),
	); err != nil {
		return err
	}
	if err := k.bankKeeper.BurnCoins(ctx, types.ModuleName, sdk.NewCoins(lsTokensToBurn)); err != nil {
		return err
	}
	return nil
}

// completeInsuranceDuty completes insurance duty.
// the status of chunk is not changed here. it should be changed in the caller side.
func (k Keeper) completeInsuranceDuty(ctx sdk.Context, ins types.Insurance) types.Insurance {
	// insurance duty is over
	ins.EmptyChunk()
	validator, found := k.stakingKeeper.GetValidator(ctx, ins.GetValidator())
	if found &&
		k.ValidateValidator(ctx, validator) == nil &&
		k.IsSufficientInsurance(ctx, ins) &&
		ins.Status != types.INSURANCE_STATUS_UNPAIRING_FOR_WITHDRAWAL {
		// This insurance is still valid, so set status to pairing.
		ins.SetStatus(types.INSURANCE_STATUS_PAIRING)
	} else {
		ins.SetStatus(types.INSURANCE_STATUS_UNPAIRED)
	}
	k.SetInsurance(ctx, ins)
	return ins
}

// completeLiquidStake completes liquid stake.
func (k Keeper) completeLiquidUnstake(ctx sdk.Context, chunk types.Chunk) {
	if chunk.Status != types.CHUNK_STATUS_UNPAIRING_FOR_UNSTAKING {
		// We don't have to return error or panic here.
		// This function is called during iteration, so just return without any processing.
		ctx.Logger().Error("chunk status is not unpairing for unstake", "chunkId", chunk.Id, "status", chunk.Status)
		return
	}
	var err error
	if err = k.validateUnpairingChunk(ctx, chunk); err != nil {
		panic(err)
	}

	bondDenom := k.stakingKeeper.BondDenom(ctx)
	liquidBondDenom := k.GetLiquidBondDenom(ctx)
	unpairingIns, _ := k.GetInsurance(ctx, chunk.UnpairingInsuranceId)
	// handle mature unbondings
	info := k.mustGetUnpairingForUnstakingChunkInfo(ctx, chunk.Id)
	lsTokensToBurn := info.EscrowedLstokens
	penaltyAmt := types.ChunkSize.Sub(k.bankKeeper.GetBalance(ctx, chunk.DerivedAddress(), bondDenom).Amount)
	if penaltyAmt.IsPositive() {
		sendAmt := penaltyAmt
		var dstAddr sdk.AccAddress
		// If this value is true, it means that the unpairing insurance cannot cover the penalty.
		var exceedInsBal bool
		unpairingInsBal := k.bankKeeper.GetBalance(ctx, unpairingIns.DerivedAddress(), bondDenom)
		if sendAmt.LTE(unpairingInsBal.Amount) {
			// insurance can cover the penalty
			dstAddr = chunk.DerivedAddress()
		} else {
			// EDGE CASE: unpairing insurance cannot cover penalties incurred during the unbonding period.
			// send all its bondDenom balance to reward pool
			sendAmt = unpairingInsBal.Amount
			dstAddr = types.RewardPool
			exceedInsBal = true
		}
		sendCoins := sdk.NewCoins(sdk.NewCoin(bondDenom, sendAmt))
		if err = k.bankKeeper.SendCoins(ctx, unpairingIns.DerivedAddress(), dstAddr, sendCoins); err != nil {
			panic(err)
		}
		if exceedInsBal {
			// The chunk has already been damaged because unpairing insurance was unable to cover the penalty.
			// Let's refund some lsTokens to unstaker since escrowed lsTokens were for the entire chunk (250K tokens), not the damaged chunk.
			penaltyRatio := penaltyAmt.ToDec().QuoTruncate(types.ChunkSize.ToDec())
			discountAmt := penaltyRatio.Mul(lsTokensToBurn.Amount.ToDec()).TruncateInt()
			refundCoin := sdk.NewCoin(liquidBondDenom, discountAmt)

			// refund
			if refundCoin.IsValid() && refundCoin.IsPositive() {
				// send discount lstokens to info.Delegator
				if err = k.bankKeeper.SendCoins(ctx, types.LsTokenEscrowAcc, info.GetDelegator(), sdk.NewCoins(refundCoin)); err != nil {
					panic(err)
				}
				lsTokensToBurn = lsTokensToBurn.Sub(refundCoin)
			}
		}
	}
	k.completeInsuranceDuty(ctx, unpairingIns)
	if lsTokensToBurn.IsValid() && lsTokensToBurn.IsPositive() {
		if err = k.burnEscrowedLsTokens(ctx, lsTokensToBurn); err != nil {
			panic(err)
		}
	}
	chunkBals := k.bankKeeper.SpendableCoins(ctx, chunk.DerivedAddress())
	// TODO: un-comment below lines while fuzzing tests to check when below condition is true
	// if !types.ChunkSize.Sub(penaltyAmt).Equal(chunkBals.AmountOf(bondDenom)) {
	// 	panic("investigating it")
	// }
	if chunkBals.IsValid() && chunkBals.IsAllPositive() {
		// We already got and burnt escrowed lsTokens, so give chunk back to unstaker.
		if err = k.bankKeeper.SendCoins(ctx, chunk.DerivedAddress(), info.GetDelegator(), chunkBals); err != nil {
			panic(err)
		}
	}
	k.DeleteUnpairingForUnstakingChunkInfo(ctx, chunk.Id)
	k.DeleteChunk(ctx, chunk.Id)
	return
}

// handleUnpairingChunk handles unpairing chunk which created previous epoch.
// Those chunks completed their unbonding already.
func (k Keeper) handleUnpairingChunk(ctx sdk.Context, chunk types.Chunk) {
	if chunk.Status != types.CHUNK_STATUS_UNPAIRING {
		// We don't have to return error or panic here.
		// This function is called during iteration, so just return without any processing.
		ctx.Logger().Error("chunk status is not unpairing", "chunkId", chunk.Id, "status", chunk.Status)
		return
	}
	var err error
	bondDenom := k.stakingKeeper.BondDenom(ctx)

	if err = k.validateUnpairingChunk(ctx, chunk); err != nil {
		panic(err)
	}
	unpairingIns, _ := k.GetInsurance(ctx, chunk.UnpairingInsuranceId)
	chunkBal := k.bankKeeper.GetBalance(ctx, chunk.DerivedAddress(), bondDenom).Amount
	penaltyAmt := types.ChunkSize.Sub(chunkBal)

	if penaltyAmt.IsPositive() {
		unpairingInsBal := k.bankKeeper.GetBalance(ctx, unpairingIns.DerivedAddress(), bondDenom).Amount
		var sendCoin sdk.Coin
		var dstAddr sdk.AccAddress
		if penaltyAmt.GT(unpairingInsBal) {
			// unpairing insurance's balance is in-sufficient to pay penaltyAmt
			// send whole insurance balance to reward pool
			// send damaged chunk to reward pool
			sendCoin = sdk.NewCoin(bondDenom, unpairingInsBal)
			dstAddr = types.RewardPool
		} else {
			// insurance balance is sufficient to pay penaltyAmt
			// chunk receive penaltyAmt from insurance
			sendCoin = sdk.NewCoin(bondDenom, penaltyAmt)
			dstAddr = chunk.DerivedAddress()
		}

		// insurance pay penaltyAmt
		if sendCoin.IsValid() && sendCoin.IsPositive() {
			if err = k.bankKeeper.SendCoins(ctx, unpairingIns.DerivedAddress(), dstAddr, sdk.NewCoins(sendCoin)); err != nil {
				panic(err)
			}
			// update chunk balance
			chunkBal = k.bankKeeper.GetBalance(ctx, chunk.DerivedAddress(), bondDenom).Amount
		}
	}
	unpairingIns = k.completeInsuranceDuty(ctx, unpairingIns)

	// If chunk got damaged, all of its coins will be sent to reward module account and chunk will be deleted
	if chunkBal.LT(types.ChunkSize) {
		chunkBals := k.bankKeeper.SpendableCoins(ctx, chunk.DerivedAddress())
		var sendCoins sdk.Coins
		if chunkBals.IsValid() && chunkBals.IsAllPositive() {
			sendCoins = chunkBals
		}
		if sendCoins.IsValid() && sendCoins.IsAllPositive() {
			if err = k.bankKeeper.SendCoins(ctx, chunk.DerivedAddress(), types.RewardPool, sendCoins); err != nil {
				panic(err)
			}
		}
		k.DeleteChunk(ctx, chunk.Id)
		// Insurance already sent all of its balance to chunk, but we cannot delete it yet
		// because it can have remaining commissions.
		if k.bankKeeper.SpendableCoins(ctx, unpairingIns.FeePoolAddress()).IsZero() {
			// if insurance has no commissions, we can delete it
			k.DeleteInsurance(ctx, unpairingIns.Id)
		}
		return
	}
	chunk.SetStatus(types.CHUNK_STATUS_PAIRING)
	chunk.EmptyPairedInsurance()
	chunk.EmptyUnpairingInsurance()
	k.SetChunk(ctx, chunk)
	return
}

func (k Keeper) handlePairedChunk(ctx sdk.Context, chunk types.Chunk) {
	if chunk.Status != types.CHUNK_STATUS_PAIRED {
		k.Logger(ctx).Error("chunk status is not paired", "chunkId", chunk.Id, "status", chunk.Status)
		return
	}

	var err error
	bondDenom := k.stakingKeeper.BondDenom(ctx)
	pairedIns, validator, del := k.mustValidatePairedChunk(ctx, chunk)

	// Check whether delegation value is decreased by slashing
	// The check process should use TokensFromShares to get the current delegation value
	tokens := validator.TokensFromShares(del.GetShares()).Ceil().TruncateInt()
	var penaltyAmt sdk.Int
	if tokens.GTE(types.ChunkSize) {
		// There is no penalty
		penaltyAmt = sdk.ZeroInt()
	} else {
		penaltyAmt = k.CalcCeiledPenalty(validator, types.ChunkSize.Sub(tokens))
	}
	var undelegated bool
	if penaltyAmt.IsPositive() {
		if k.isRepairingChunk(ctx, chunk) {
			// If chunk is repairing and validator is tombstoned then check evidence and
			// decide which insurance should pay penalty.
			err = k.ValidateValidator(ctx, validator)
			switch err {
			case nil:
				// validator is not tombstoned
				// no need to handle this case
			case types.ErrTombstonedValidator:
				latestEvidence, err := k.findLatestEvidence(ctx, validator)
				if err != nil {
					panic(err)
				}

				if latestEvidence == nil {
					panic("tombstoned validator but have no evidence, impossible")
				}
				epoch := k.GetEpoch(ctx)
				if epoch.GetStartHeight() >= latestEvidence.GetHeight() {
					coveredAmt := k.mustCoverDoubleSignPenaltyFromUnpairingInsurance(ctx, chunk)
					penaltyAmt = penaltyAmt.Sub(coveredAmt)
					penaltyAmt = sdk.MaxInt(penaltyAmt, sdk.ZeroInt())
					// update variables after cover double sign penalty
					_, validator, del = k.mustValidatePairedChunk(ctx, chunk)
				}
				// If epoch.StartHeight < lastEvidence.Height, then it means
				// there was double sign slashing after re-pairing, so in this case
				// unpairing insurance doesn't have to pay for penalty
			case types.ErrInvalidValidatorStatus:
				// Ths case must not happen.
				panic(err)
			}
		}
		if penaltyAmt.IsPositive() {
			pairedInsBal := k.bankKeeper.GetBalance(ctx, pairedIns.DerivedAddress(), bondDenom)
			// EDGE CASE: paired insurance cannot cover penalty
			if penaltyAmt.GT(pairedInsBal.Amount) {
				// At this time, insurance does not cover the penalty because it has already been determined that the chunk was damaged.
				// Just un-delegate(=unpair) the chunk, so it can be naturally handled by the unpairing logic in the next epoch.
				// Insurance will send penalty to the reward pool at next epoch and chunk's token will go to reward pool.
				// Check the logic of handleUnpairingChunk for detail.
				k.startUnpairing(ctx, pairedIns, chunk)
				k.mustUndelegate(ctx, chunk, validator, del, types.AttributeValueReasonNotEnoughPairedInsCoverage)
				undelegated = true
			} else {
				// happy case: paired insurance can cover penalty and there is no un-covered penalty from unpairing insurance.
				// 1. Send penalty to chunk
				// 2. chunk delegate additional tokens to validator
				penaltyCoin := sdk.NewCoin(bondDenom, penaltyAmt)
				// send penalty to chunk
				if err = k.bankKeeper.SendCoins(ctx, pairedIns.DerivedAddress(), chunk.DerivedAddress(), sdk.NewCoins(penaltyCoin)); err != nil {
					panic(err)
				}
				// delegate additional tokens to validator as chunk.DerivedAddress()
				k.mustDelegatePenaltyAmt(ctx, chunk, penaltyCoin.Amount, validator, pairedIns.Id, types.AttributeValueReasonPairedInsCoverPenalty)
				// update variables after delegate
				_, validator, del = k.mustValidatePairedChunk(ctx, chunk)
			}
		}
	}

	// After cover penalty, check whether paired insurance is sufficient or not.
	// If not sufficient, start unpairing and un-delegate.
	if !undelegated && !k.IsEnoughToCoverSlash(ctx, pairedIns) {
		// To remove complexity, if insurance is not enough to cover double sign slashing then
		// un-pair and un-delegate.
		// By doing this, we can avoid the case that current paired insurance cannot cover penalty while
		// re-delegation or re-pairing period.
		k.startUnpairing(ctx, pairedIns, chunk)
		k.mustUndelegate(ctx, chunk, validator, del, types.AttributeValueReasonPairedInsBalUnderDoubleSignSlashing)
	}

	// If validator of paired insurance is not valid, start unpairing.
	if err := k.ValidateValidator(ctx, validator); err != nil {
		k.startUnpairing(ctx, pairedIns, chunk)
	}

	if chunk.HasUnpairingInsurance() {
		// Unpairing insurance created at previous epoch finished its duty.
		unpairingIns := k.mustGetInsurance(ctx, chunk.UnpairingInsuranceId)
		k.completeInsuranceDuty(ctx, unpairingIns)
	}

	// If unpairing insurance of updated chunk is Unpaired or Pairing
	// which means it already completed its duty during unpairing period,
	// we can safely remove unpairing insurance id from the chunk.
	k.mustClearUnpairedInsurance(ctx, chunk.Id)
	return
}

// IsSufficientInsurance checks whether insurance has sufficient balance to cover slashing or not.
func (k Keeper) IsSufficientInsurance(ctx sdk.Context, insurance types.Insurance) bool {
	insBal := k.bankKeeper.GetBalance(ctx, insurance.DerivedAddress(), k.stakingKeeper.BondDenom(ctx))
	_, minimumCollateral := k.GetMinimumRequirements(ctx)
	return insBal.Amount.GTE(minimumCollateral.Amount)
}

// IsEnoughToCoverSlash checks whether insurance has sufficient balance to cover slashing or not.
func (k Keeper) IsEnoughToCoverSlash(ctx sdk.Context, insurance types.Insurance) bool {
	params := k.slashingKeeper.GetParams(ctx)
	downTimePenaltyAmt := types.ChunkSize.ToDec().Mul(params.SlashFractionDowntime).Ceil().TruncateInt()
	insBal := k.bankKeeper.GetBalance(ctx, insurance.DerivedAddress(), k.stakingKeeper.BondDenom(ctx))
	doubleSignPenaltyAmt := types.ChunkSize.ToDec().Mul(params.SlashFractionDoubleSign).Ceil().TruncateInt()
	return insBal.Amount.GTE(downTimePenaltyAmt.Add(doubleSignPenaltyAmt))
}

// startUnpairing changes status of insurance and chunk to unpairing.
// Actual unpairing process including un-delegate chunk will be done after ranking in EndBlocker.
func (k Keeper) startUnpairing(ctx sdk.Context, ins types.Insurance, chunk types.Chunk) {
	ins.SetStatus(types.INSURANCE_STATUS_UNPAIRING)
	chunk.UnpairingInsuranceId = chunk.PairedInsuranceId
	chunk.EmptyPairedInsurance()
	chunk.SetStatus(types.CHUNK_STATUS_UNPAIRING)
	k.SetChunk(ctx, chunk)
	k.SetInsurance(ctx, ins)
}

// startUnpairingForLiquidUnstake changes status of insurance to unpairing and
// chunk to UnpairingForUnstaking.
func (k Keeper) startUnpairingForLiquidUnstake(ctx sdk.Context, ins types.Insurance, chunk types.Chunk) (types.Insurance, types.Chunk) {
	chunk.SetStatus(types.CHUNK_STATUS_UNPAIRING_FOR_UNSTAKING)
	chunk.UnpairingInsuranceId = chunk.PairedInsuranceId
	chunk.EmptyPairedInsurance()
	ins.SetStatus(types.INSURANCE_STATUS_UNPAIRING)
	k.SetChunk(ctx, chunk)
	k.SetInsurance(ctx, ins)
	return ins, chunk
}

// mustUndelegate undelegates chunk from validator.
func (k Keeper) mustUndelegate(
	ctx sdk.Context, chunk types.Chunk, validator stakingtypes.Validator, del stakingtypes.Delegation, reason string,
) {
	completionTime, err := k.stakingKeeper.Undelegate(ctx, chunk.DerivedAddress(), validator.GetOperator(), del.GetShares())
	if err != nil {
		panic(err)
	}
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeBeginUndelegate,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
			sdk.NewAttribute(types.AttributeKeyChunkId, fmt.Sprintf("%d", chunk.Id)),
			sdk.NewAttribute(types.AttributeKeyValidator, validator.GetOperator().String()),
			sdk.NewAttribute(types.AttributeKeyCompletionTime, completionTime.Format(time.RFC3339)),
			sdk.NewAttribute(types.AttributeKeyReason, reason),
		),
	)
}

// mustDelegatePenaltyAmt delegates amt to validator as chunk.
func (k Keeper) mustDelegatePenaltyAmt(
	ctx sdk.Context, chunk types.Chunk, amt sdk.Int, validator stakingtypes.Validator,
	insId uint64, reason string,
) {
	// delegate additional tokens to validator as chunk.DerivedAddress()
	newShares, err := k.stakingKeeper.Delegate(ctx, chunk.DerivedAddress(), amt, stakingtypes.Unbonded, validator, true)
	if err != nil {
		panic(err)
	}
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			stakingtypes.EventTypeDelegate,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
			sdk.NewAttribute(types.AttributeKeyChunkId, fmt.Sprintf("%d", chunk.Id)),
			sdk.NewAttribute(types.AttributeKeyInsuranceId, fmt.Sprintf("%d", insId)),
			sdk.NewAttribute(types.AttributeKeyDelegator, chunk.DerivedAddress().String()),
			sdk.NewAttribute(types.AttributeKeyValidator, validator.GetOperator().String()),
			sdk.NewAttribute(sdk.AttributeKeyAmount, amt.String()),
			sdk.NewAttribute(types.AttributeKeyNewShares, newShares.String()),
			sdk.NewAttribute(types.AttributeKeyReason, reason),
		),
	)
}

// withdrawInsurance withdraws insurance and commissions from insurance account immediately.
func (k Keeper) withdrawInsurance(ctx sdk.Context, insurance types.Insurance) error {
	var inputs []banktypes.Input
	var outputs []banktypes.Output

	insBals := k.bankKeeper.SpendableCoins(ctx, insurance.DerivedAddress())
	if insBals.IsValid() && insBals.IsAllPositive() {
		inputs = append(inputs, banktypes.NewInput(insurance.DerivedAddress(), insBals))
		outputs = append(outputs, banktypes.NewOutput(insurance.GetProvider(), insBals))
	}
	commissions := k.bankKeeper.SpendableCoins(ctx, insurance.FeePoolAddress())
	if commissions.IsValid() && commissions.IsAllPositive() {
		inputs = append(inputs, banktypes.NewInput(insurance.FeePoolAddress(), commissions))
		outputs = append(outputs, banktypes.NewOutput(insurance.GetProvider(), commissions))
	}
	if err := k.bankKeeper.InputOutputCoins(ctx, inputs, outputs); err != nil {
		return err
	}

	insBals = k.bankKeeper.SpendableCoins(ctx, insurance.DerivedAddress())
	commissions = k.bankKeeper.SpendableCoins(ctx, insurance.FeePoolAddress())
	if insBals.IsZero() && commissions.IsZero() {
		k.DeleteInsurance(ctx, insurance.Id)
	}
	return nil
}

// mustPairChunkAndDelegate pairs chunk and delegate it to validator pointed by insurance.
func (k Keeper) mustPairChunkAndDelegate(
	ctx sdk.Context, chunk types.Chunk, ins types.Insurance, valAddr sdk.ValAddress, amt sdk.Int,
) (types.Chunk, types.Insurance, sdk.Dec, error) {
	validator, found := k.stakingKeeper.GetValidator(ctx, valAddr)
	if !found {
		return types.Chunk{}, types.Insurance{}, sdk.Dec{}, fmt.Errorf("validator %s not found", valAddr)
	}
	newShares, err := k.stakingKeeper.Delegate(
		ctx, chunk.DerivedAddress(), amt, stakingtypes.Unbonded, validator, true,
	)
	if err != nil {
		return types.Chunk{}, types.Insurance{}, sdk.Dec{}, err
	}
	chunk.PairedInsuranceId = ins.Id
	ins.ChunkId = chunk.Id
	chunk.SetStatus(types.CHUNK_STATUS_PAIRED)
	ins.SetStatus(types.INSURANCE_STATUS_PAIRED)
	k.SetChunk(ctx, chunk)
	k.SetInsurance(ctx, ins)
	return chunk, ins, newShares, nil
}

func (k Keeper) rePairChunkAndInsurance(ctx sdk.Context, chunk types.Chunk, newIns, outIns types.Insurance) {
	chunk.UnpairingInsuranceId = outIns.Id
	if outIns.Status != types.INSURANCE_STATUS_UNPAIRING_FOR_WITHDRAWAL {
		outIns.SetStatus(types.INSURANCE_STATUS_UNPAIRING)
	}
	chunk.PairedInsuranceId = newIns.Id
	newIns.ChunkId = chunk.Id
	newIns.SetStatus(types.INSURANCE_STATUS_PAIRED)
	chunk.SetStatus(types.CHUNK_STATUS_PAIRED)
	k.SetInsurance(ctx, outIns)
	k.SetInsurance(ctx, newIns)
	k.SetChunk(ctx, chunk)
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeRePairedWithNewInsurance,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
			sdk.NewAttribute(types.AttributeKeyChunkId, fmt.Sprintf("%d", chunk.Id)),
			sdk.NewAttribute(types.AttributeKeyNewInsuranceId, fmt.Sprintf("%d", newIns.Id)),
			sdk.NewAttribute(types.AttributeKeyOutInsuranceId, fmt.Sprintf("%d", outIns.Id)),
		),
	)
}

// validateUnpairingChunk validates unpairing or unpairing for unstaking chunk.
func (k Keeper) validateUnpairingChunk(ctx sdk.Context, chunk types.Chunk) error {
	// get paired insurance from chunk
	unpairingIns, found := k.GetInsurance(ctx, chunk.UnpairingInsuranceId)
	if !found {
		return sdkerrors.Wrapf(types.ErrNotFoundUnpairingInsurance, "insuranceId: %d(chunkId: %d)", chunk.UnpairingInsuranceId, chunk.Id)
	}
	if chunk.HasPairedInsurance() {
		return sdkerrors.Wrapf(types.ErrMustHaveNoPairedInsurance, "chunkId: %d", chunk.Id)
	}
	if _, found = k.stakingKeeper.GetUnbondingDelegation(ctx, chunk.DerivedAddress(), unpairingIns.GetValidator()); found {
		// UnbondingDelegation already must be removed by staking keeper EndBlocker
		// because Endblocker of liquidstaking module is called after staking module.
		return sdkerrors.Wrapf(types.ErrMustHaveNoUnbondingDelegation, "chunkId: %d", chunk.Id)
	}
	return nil
}

func (k Keeper) validateInsurance(
	ctx sdk.Context, insId uint64, providerAddr sdk.AccAddress, expectedStatus types.InsuranceStatus,
) (types.Insurance, error) {
	// Check if the ins exists
	ins, found := k.GetInsurance(ctx, insId)
	if !found {
		return ins, sdkerrors.Wrapf(types.ErrNotFoundInsurance, "ins id: %d", insId)
	}

	// Check if the provider is the same
	if ins.ProviderAddress != providerAddr.String() {
		return ins, sdkerrors.Wrapf(types.ErrNotProviderOfInsurance, "ins id: %d", insId)
	}

	if expectedStatus != types.INSURANCE_STATUS_UNSPECIFIED {
		if ins.Status != expectedStatus {
			return ins, sdkerrors.Wrapf(types.ErrInvalidInsuranceStatus, "expected: %s, actual: %s(insuranceId: %d)", expectedStatus, ins.Status, insId)
		}
	}
	return ins, nil
}

// mustValidaRedelegationInfo validates re-delegation info and returns chunk, srcInsurance, dstInsurance, entry.
// If it is invalid, it panics.
func (k Keeper) mustValidateRedelegationInfo(ctx sdk.Context, info types.RedelegationInfo) (
	chunk types.Chunk,
	srcIns types.Insurance,
	dstIns types.Insurance,
	entry stakingtypes.RedelegationEntry,
) {
	chunk = k.mustGetChunk(ctx, info.ChunkId)
	if chunk.Status != types.CHUNK_STATUS_PAIRED {
		panic(fmt.Sprintf("chunk id: %d is not paired", info.ChunkId))
	}
	// In re-delegation situation, chunk must have an unpairing insurance.
	if !chunk.HasUnpairingInsurance() || !chunk.HasPairedInsurance() {
		panic(fmt.Sprintf("both paired and unpairing insurance must exists while module is tracking re-delegation(chunkId: %d)", info.ChunkId))
	}
	srcIns = k.mustGetInsurance(ctx, chunk.UnpairingInsuranceId)
	dstIns = k.mustGetInsurance(ctx, chunk.PairedInsuranceId)
	reDels := k.stakingKeeper.GetAllRedelegations(
		ctx,
		chunk.DerivedAddress(),
		srcIns.GetValidator(),
		dstIns.GetValidator(),
	)
	if len(reDels) != 1 {
		panic(fmt.Sprintf("chunk id: %d must have one re-delegation, got: %d", chunk.Id, len(reDels)))
	}
	red := reDels[0]
	if len(red.Entries) != 1 {
		panic(fmt.Sprintf("chunk id: %d must have one re-delegation entry", chunk.Id))
	}
	entry = red.Entries[0]
	return
}

// mustValidatePairedChunk validates paired chunk and return paired insurance and its validator.
// If it is invalid, then it panics.
func (k Keeper) mustValidatePairedChunk(ctx sdk.Context, chunk types.Chunk) (
	types.Insurance, stakingtypes.Validator, stakingtypes.Delegation,
) {
	ins := k.mustGetInsurance(ctx, chunk.PairedInsuranceId)
	validator, found := k.stakingKeeper.GetValidator(ctx, ins.GetValidator())
	if !found {
		panic(fmt.Sprintf("validator of paired ins %s not found(insuranceId: %d)", ins.GetValidator(), ins.Id))
	}
	// Get delegation of chunk
	delegation, found := k.stakingKeeper.GetDelegation(ctx, chunk.DerivedAddress(), validator.GetOperator())
	if !found {
		panic(fmt.Sprintf("delegation not found: %s(chunkId: %d)", chunk.DerivedAddress(), chunk.Id))
	}
	return ins, validator, delegation
}

// mustClearUnpairedInsurance clears unpaired insurance of chunk.
func (k Keeper) mustClearUnpairedInsurance(ctx sdk.Context, id uint64) {
	chunk := k.mustGetChunk(ctx, id)
	if chunk.HasUnpairingInsurance() {
		unpairingIns := k.mustGetInsurance(ctx, chunk.UnpairingInsuranceId)
		if unpairingIns.IsUnpaired() {
			chunk.EmptyUnpairingInsurance()
			k.SetChunk(ctx, chunk)
		}
	}
	return
}

// isRepairingChunk returns true if the chunk is repairing without re-delegation obj.
func (k Keeper) isRepairingChunk(ctx sdk.Context, chunk types.Chunk) bool {
	if chunk.HasPairedInsurance() && chunk.HasUnpairingInsurance() {
		pairedIns := k.mustGetInsurance(ctx, chunk.PairedInsuranceId)
		unpairingIns := k.mustGetInsurance(ctx, chunk.UnpairingInsuranceId)
		if pairedIns.ValidatorAddress == unpairingIns.ValidatorAddress {
			return true
		}
	}
	return false
}

func (k Keeper) findLatestEvidence(ctx sdk.Context, validator stakingtypes.Validator) (latest *evidencetypes.Equivocation, err error) {
	k.evidenceKeeper.IterateEvidence(ctx, func(evidence exported.Evidence) (stop bool) {
		if v, ok := evidence.(*evidencetypes.Equivocation); ok {
			consAddr, err := validator.GetConsAddr()
			if err != nil {
				return true
			}
			if v.GetConsensusAddress().Equals(consAddr) {
				if latest == nil {
					latest = v
					return false
				}
				if v.GetHeight() > latest.GetHeight() {
					latest = v
				}
			}
		}
		return false
	})
	return
}

// mustCoverDoubleSignPenaltyFromUnpairingInsurance covers dobule sign slashing penalty from unpairing insurance.
func (k Keeper) mustCoverDoubleSignPenaltyFromUnpairingInsurance(ctx sdk.Context, chunk types.Chunk) sdk.Int {
	unpairingIns := k.mustGetInsurance(ctx, chunk.UnpairingInsuranceId)
	bondDenom := k.stakingKeeper.BondDenom(ctx)

	validator, found := k.stakingKeeper.GetValidator(ctx, unpairingIns.GetValidator())
	if !found {
		panic(fmt.Sprintf("validator not found: %s", unpairingIns.GetValidator()))
	}

	params := k.slashingKeeper.GetParams(ctx)
	coverAmt := types.ChunkSize.ToDec().Mul(params.SlashFractionDoubleSign).Ceil().TruncateInt()
	coverAmt = k.CalcCeiledPenalty(validator, coverAmt)
	dstAddr := chunk.DerivedAddress()
	unpairingInsBal := k.bankKeeper.GetBalance(ctx, unpairingIns.DerivedAddress(), bondDenom)
	if coverAmt.GT(unpairingInsBal.Amount) {
		panic(fmt.Sprintf("unpairing insurance balance is not enough to cover double sign slashing penalty: %s", unpairingIns.DerivedAddress()))
	}
	coveredCoins := sdk.NewCoins(sdk.NewCoin(bondDenom, coverAmt))
	if coveredCoins.IsValid() && coveredCoins.IsAllPositive() {
		if err := k.bankKeeper.SendCoins(ctx, unpairingIns.DerivedAddress(), dstAddr, coveredCoins); err != nil {
			panic(err)
		}
		k.mustDelegatePenaltyAmt(ctx, chunk, coverAmt, validator, unpairingIns.Id, types.AttributeValueReasonUnpairingInsCoverPenalty)
	}
	return coverAmt
}

func (k Keeper) CalcCeiledPenalty(validator stakingtypes.Validator, penaltyAmt sdk.Int) sdk.Int {
	penaltyShares, err := validator.SharesFromTokens(penaltyAmt)
	if err != nil {
		panic(err)
	}
	// If penaltyShares is not integer, we need to ceil it.
	// If not, then after we cover penalty and check tokens value, it will be less than chunkSize.
	if !penaltyShares.IsInteger() {
		penaltyShares = penaltyShares.Ceil()
		return validator.TokensFromShares(penaltyShares).Ceil().TruncateInt()
	}
	return penaltyAmt
}
