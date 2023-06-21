package keeper

import (
	"fmt"

	"github.com/Canto-Network/Canto/v6/x/liquidstaking/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

// CollectRewardAndFee collects reward of chunk and
// distribute to a module(=fee), reward pool and insurance.
// 1. Send commission to insurance based on chunk reward.
// 2. Deduct dynamic fee from remaining and burn it.
// 3. Send rest of rewards to reward module account.
func (k Keeper) CollectRewardAndFee(
	ctx sdk.Context,
	dynamicFeeRate sdk.Dec,
	chunk types.Chunk,
	insurance types.Insurance,
) {
	delegationRewards := k.bankKeeper.GetAllBalances(ctx, chunk.DerivedAddress())
	insuranceCommissions := make(sdk.Coins, delegationRewards.Len())
	dynamicFees := make(sdk.Coins, delegationRewards.Len())
	remainingRewards := make(sdk.Coins, delegationRewards.Len())

	for i, delReward := range delegationRewards {
		if delReward.IsZero() {
			continue
		}
		insuranceCommission := delReward.Amount.ToDec().Mul(insurance.FeeRate).TruncateInt()
		insuranceCommissions[i] = sdk.NewCoin(
			delReward.Denom,
			insuranceCommission,
		)
		pureReward := delReward.Amount.Sub(insuranceCommission)
		dynamicFee := pureReward.ToDec().Mul(dynamicFeeRate).Ceil().TruncateInt()
		remainingReward := pureReward.Sub(dynamicFee)
		dynamicFees[i] = sdk.NewCoin(
			delReward.Denom,
			dynamicFee,
		)
		remainingRewards[i] = sdk.NewCoin(
			delReward.Denom,
			remainingReward,
		)
	}
	fmt.Printf("Collect Reward for validator: %s\n", insurance.GetValidator())
	fmt.Printf("Delegation Reward: %s\n", delegationRewards.String())
	fmt.Printf("Insurance Commission: %s\n", insuranceCommissions.String())
	fmt.Printf("Dynamic Fee: %s\n", dynamicFees.String())
	fmt.Printf("Reamining Reward: %s\n", remainingRewards.String())

	var inputs []banktypes.Input
	var outputs []banktypes.Output
	switch remainingRewards.Len() {
	case 0:
		return
	default:
		if !dynamicFees.IsZero() {
			// Collect dynamic fee and burn it first.
			if err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, chunk.DerivedAddress(), types.ModuleName, dynamicFees); err != nil {
				panic(err)
			}
			if err := k.bankKeeper.BurnCoins(ctx, types.ModuleName, dynamicFees); err != nil {
				panic(err)
			}
		}

		inputs = []banktypes.Input{
			banktypes.NewInput(chunk.DerivedAddress(), insuranceCommissions),
			banktypes.NewInput(chunk.DerivedAddress(), remainingRewards),
		}
		outputs = []banktypes.Output{
			banktypes.NewOutput(insurance.FeePoolAddress(), insuranceCommissions),
			banktypes.NewOutput(types.RewardPool, remainingRewards),
		}
	}
	if err := k.bankKeeper.InputOutputCoins(ctx, inputs, outputs); err != nil {
		panic(err)
	}
}

// DistributeReward withdraws delegation rewards from all paired chunks
// Keeper.CollectRewardAndFee will be called during withdrawing process.
func (k Keeper) DistributeReward(ctx sdk.Context) {
	feeRate := k.CalcDynamicFeeRate(ctx)
	err := k.IterateAllChunks(ctx, func(chunk types.Chunk) (bool, error) {
		var insurance types.Insurance
		var found bool
		switch chunk.Status {
		case types.CHUNK_STATUS_PAIRED:
			insurance, found = k.GetInsurance(ctx, chunk.PairedInsuranceId)
			if !found {
				panic(types.ErrNotFoundInsurance.Error())
			}
		default:
			return false, nil
		}
		validator, found := k.stakingKeeper.GetValidator(ctx, insurance.GetValidator())
		err := k.IsValidValidator(ctx, validator, found)
		if err == types.ErrNotFoundValidator {
			return true, err
		}
		// TODO: remove print when go to production
		fmt.Printf("Chunk %d Balance Before Withdraw Delegation Rewards\n", chunk.Id)
		// TODO: remove when go to production
		bal := k.bankKeeper.GetBalance(ctx, chunk.DerivedAddress(), "acanto")
		if bal.IsPositive() {
			panic("chunk %d balance is not zero")
		}
		fmt.Println(bal.String())
		_, err = k.distributionKeeper.WithdrawDelegationRewards(ctx, chunk.DerivedAddress(), validator.GetOperator())
		// chunk balance -> chunk reward address
		if err != nil {
			return true, err
		}
		fmt.Printf("Chunk %d Balance After Withdraw Delegation Rewards\n", chunk.Id)
		fmt.Println(k.bankKeeper.GetBalance(ctx, chunk.DerivedAddress(), "acanto").String())

		k.CollectRewardAndFee(ctx, feeRate, chunk, insurance)
		return false, nil
	})
	if err != nil {
		panic(err.Error())
	}
}

// CoverSlashingAndHandleMatureUnbondings covers slashing and handles mature unbondings.
func (k Keeper) CoverSlashingAndHandleMatureUnbondings(ctx sdk.Context) {
	var err error
	err = k.IterateAllChunks(ctx, func(chunk types.Chunk) (bool, error) {
		switch chunk.Status {
		// Finish mature unbondings triggered in previous epoch
		case types.CHUNK_STATUS_UNPAIRING_FOR_UNSTAKING:
			if err := k.completeLiquidUnstake(ctx, chunk); err != nil {
				panic(err)
			}

		case types.CHUNK_STATUS_UNPAIRING:
			if err := k.handleUnpairingChunk(ctx, chunk); err != nil {
				panic(err)
			}

		case types.CHUNK_STATUS_PAIRED:
			if err := k.handlePairedChunk(ctx, chunk); err != nil {
				panic(err)
			}
		}
		return false, nil
	})
	if err != nil {
		panic(err.Error())
	}
}

// HandleQueuedLiquidUnstakes processes unstaking requests that were queued before the epoch.
// 1. Get all pending liquid unstakes
// 2. For each pending liquid unstake, get chunk and insurance
// 3. Validate unbond amount
// 4. Un-delegate chunk
// 5. Update chunk status
// 6. Update insurance status
// 7. Set unpairing for unstake chunk info which will be used by CoverSlashingAndHandleMatureUnbondings
// 8. Delete pending liquid unstake
func (k Keeper) HandleQueuedLiquidUnstakes(ctx sdk.Context) ([]types.Chunk, error) {
	var unstakedChunks []types.Chunk
	// TODO: Should use Queue for processing in sequence? MintRate is ok?, insurance issue? etc...
	infos := k.GetAllUnpairingForUnstakingChunkInfos(ctx)
	for _, info := range infos {
		// Get chunk
		chunk, found := k.GetChunk(ctx, info.ChunkId)
		if !found {
			return nil, sdkerrors.Wrapf(types.ErrNotFoundChunk, "id: %d", info.ChunkId)
		}
		if chunk.Status != types.CHUNK_STATUS_PAIRED {
			// Chunk is already in unstaking process, so we skip it
			continue
		}
		// get insurance
		insurance, found := k.GetInsurance(ctx, chunk.PairedInsuranceId)
		if !found {
			return nil, sdkerrors.Wrapf(types.ErrNotFoundInsurance, "id: %d", chunk.PairedInsuranceId)
		}
		shares, err := k.stakingKeeper.ValidateUnbondAmount(ctx, chunk.DerivedAddress(), insurance.GetValidator(), types.ChunkSize)
		if err != nil {
			return nil, err
		}
		_, err = k.stakingKeeper.Undelegate(
			ctx,
			chunk.DerivedAddress(),
			insurance.GetValidator(),
			shares,
		)
		if err != nil {
			return nil, err
		}
		_, chunk = k.startUnpairingForLiquidUnstake(ctx, insurance, chunk)
		unstakedChunks = append(unstakedChunks, chunk)
	}
	return unstakedChunks, nil
}

// HandleQueuedWithdrawInsuranceRequests processes withdraw insurance requests that were queued before the epoch.
// Unpairing insurances will be unpaired in the next epoch.is unpaired.
// 1. Get all pending withdraw insurance requests
// 2. For each pending withdraw insurance request, get insurance
// 3. Validate insurance status
// 4. Get chunk from insurance
// 5. Validate chunk status
// 6. Unpair chunk and insurance
// 7. Update chunk status
// 8. Update insurance status
// 9. Delete pending withdraw insurance request
func (k Keeper) HandleQueuedWithdrawInsuranceRequests(ctx sdk.Context) ([]types.Insurance, error) {
	var withdrawnInsurances []types.Insurance
	reqs := k.GetAllWithdrawInsuranceRequests(ctx)
	for _, req := range reqs {
		// get insurance
		insurance, found := k.GetInsurance(ctx, req.InsuranceId)
		if !found {
			return nil, sdkerrors.Wrapf(types.ErrNotFoundInsurance, "id: %d", req.InsuranceId)
		}
		if insurance.Status != types.INSURANCE_STATUS_PAIRED {
			return nil, sdkerrors.Wrapf(types.ErrInvalidInsuranceStatus, "id: %d, status: %s", insurance.Id, insurance.Status)
		}

		// get chunk from insurance
		chunk, found := k.GetChunk(ctx, insurance.ChunkId)
		if !found {
			return nil, sdkerrors.Wrapf(types.ErrNotFoundChunk, "id: %d", insurance.ChunkId)
		}
		chunk.SetStatus(types.CHUNK_STATUS_UNPAIRING)
		chunk.UnpairingInsuranceId = chunk.PairedInsuranceId
		chunk.PairedInsuranceId = 0
		insurance.SetStatus(types.INSURANCE_STATUS_UNPAIRING_FOR_WITHDRAWAL)
		k.SetInsurance(ctx, insurance)
		k.SetChunk(ctx, chunk)
		k.DeleteWithdrawInsuranceRequest(ctx, insurance.Id)
		withdrawnInsurances = append(withdrawnInsurances, insurance)
	}
	return withdrawnInsurances, nil
}

// GetAllRePairableChunksAndOutInsurances returns all active chunks and related insurances.
// Active chunks are chunks which are paired or not unpairing.
// Not unpairing chunk have no un-bonding info.
func (k Keeper) GetAllRePairableChunksAndOutInsurances(ctx sdk.Context) (
	rePairableChunks []types.Chunk,
	outInsurances []types.Insurance,
	pairedInsuranceMap map[uint64]struct{},
	err error,
) {
	pairedInsuranceMap = make(map[uint64]struct{})
	if err = k.IterateAllChunks(ctx, func(chunk types.Chunk) (bool, error) {
		switch chunk.Status {
		case types.CHUNK_STATUS_UNPAIRING:
			insurance, found := k.GetInsurance(ctx, chunk.UnpairingInsuranceId)
			if !found {
				return false, sdkerrors.Wrapf(types.ErrNotFoundInsurance, "insurance id: %d", chunk.UnpairingInsuranceId)
			}
			_, found = k.stakingKeeper.GetUnbondingDelegation(ctx, chunk.DerivedAddress(), insurance.GetValidator())
			if found {
				// unbonding of chunk is triggered because insurance cannot cover the penalty of chunk.
				// In next epoch, insurance send all of it's balance to chunk
				// and all balances of chunk will go to reward pool.
				// After that, insurance will be unpaired also.
				return false, nil
			}
			outInsurances = append(outInsurances, insurance)
			rePairableChunks = append(rePairableChunks, chunk)
		case types.CHUNK_STATUS_PAIRING:
			rePairableChunks = append(rePairableChunks, chunk)
		case types.CHUNK_STATUS_PAIRED:
			insurance, found := k.GetInsurance(ctx, chunk.PairedInsuranceId)
			if !found {
				return false, sdkerrors.Wrapf(types.ErrNotFoundInsurance, "insurance id: %d", chunk.UnpairingInsuranceId)
			}
			pairedInsuranceMap[insurance.Id] = struct{}{}
			rePairableChunks = append(rePairableChunks, chunk)
		default:
			return false, nil
		}
		return false, nil
	}); err != nil {
		return
	}
	return
}

// RankInsurances ranks insurances and returns following:
// 1. newly ranked insurances
// - rank in insurance which is not paired currently
// - no change is needed for already ranked in and paired insurances
// 2. Ranked out insurances
// - current unpairing insurances + paired insurances which is failed to rank in
func (k Keeper) RankInsurances(ctx sdk.Context) (
	newlyRankedInInsurances []types.Insurance,
	rankOutInsurances []types.Insurance,
	err error,
) {
	candidatesValidatorMap := make(map[string]stakingtypes.Validator)
	rePairableChunks, currentOutInsurances, pairedInsuranceMap, err := k.GetAllRePairableChunksAndOutInsurances(ctx)

	// candidateInsurances will be ranked
	var candidateInsurances []types.Insurance
	if err = k.IterateAllInsurances(ctx, func(insurance types.Insurance) (stop bool, err error) {
		// Only pairing or paired insurances are candidates to be ranked
		if insurance.Status != types.INSURANCE_STATUS_PAIRED &&
			insurance.Status != types.INSURANCE_STATUS_PAIRING {
			return false, nil
		}

		if _, ok := candidatesValidatorMap[insurance.ValidatorAddress]; !ok {
			validator, found := k.stakingKeeper.GetValidator(ctx, insurance.GetValidator())
			err := k.IsValidValidator(ctx, validator, found)
			if err != nil {
				if insurance.Status == types.INSURANCE_STATUS_PAIRED {
					// CRITICAL & EDGE CASE:
					// paired insurance must have valid validator
					return false, err
				} else if insurance.Status == types.INSURANCE_STATUS_PAIRING {
					// EDGE CASE:
					// Skip pairing insurance which have invalid validator
					return false, nil
				}
			}
			candidatesValidatorMap[insurance.ValidatorAddress] = validator
		}
		candidateInsurances = append(candidateInsurances, insurance)
		return false, nil
	}); err != nil {
		return
	}

	types.SortInsurances(candidatesValidatorMap, candidateInsurances, false)
	var rankInInsurances []types.Insurance
	var rankOutCandidates []types.Insurance
	if len(rePairableChunks) > len(candidateInsurances) {
		rankInInsurances = candidateInsurances
	} else {
		rankInInsurances = candidateInsurances[:len(rePairableChunks)]
		rankOutCandidates = candidateInsurances[len(rePairableChunks):]
	}

	for _, insurance := range rankOutCandidates {
		if insurance.Status == types.INSURANCE_STATUS_PAIRED {
			rankOutInsurances = append(rankOutInsurances, insurance)
		}
	}
	rankOutInsurances = append(rankOutInsurances, currentOutInsurances...)

	for _, insurance := range rankInInsurances {
		if _, ok := pairedInsuranceMap[insurance.Id]; !ok {
			newlyRankedInInsurances = append(newlyRankedInInsurances, insurance)
		}
	}
	return
}

// RePairRankedInsurances re-pairs ranked insurances.
func (k Keeper) RePairRankedInsurances(
	ctx sdk.Context,
	newlyRankedInInsurances,
	rankOutInsurances []types.Insurance,
) (err error) {
	var rankOutInsuranceChunkMap = make(map[uint64]types.Chunk)
	for _, outInsurance := range rankOutInsurances {
		chunk, found := k.GetChunk(ctx, outInsurance.ChunkId)
		if !found {
			return sdkerrors.Wrapf(types.ErrNotFoundChunk, "chunk id: %d", outInsurance.ChunkId)
		}
		rankOutInsuranceChunkMap[outInsurance.Id] = chunk
	}

	// newInsurancesWithDifferentValidators will be replaced by re-delegate
	// because there are no rankout insurances which have same validator
	var newInsurancesWithDifferentValidators []types.Insurance
	// Short circuit
	// Try to replace outInsurance with inInsurance which has same validator.
	for _, newRankInInsurance := range newlyRankedInInsurances {
		hasSameValidator := false
		for oi, outInsurance := range rankOutInsurances {
			// Happy case. Same validator so we can skip re-delegation
			if newRankInInsurance.GetValidator().Equals(outInsurance.GetValidator()) {
				// get chunk by outInsurance.ChunkId
				chunk, found := k.GetChunk(ctx, outInsurance.ChunkId)
				if !found {
					panic("chunk not found")
				}
				// TODO: outInsurance is removed at next epoch? and also it covers penalty if slashing happened after?
				k.rePairChunkAndInsurance(ctx, chunk, newRankInInsurance, outInsurance)
				hasSameValidator = true
				// Remove already checked outInsurance
				rankOutInsurances = append(rankOutInsurances[:oi], rankOutInsurances[oi+1:]...)
				break
			}
		}
		if !hasSameValidator {
			newInsurancesWithDifferentValidators = append(newInsurancesWithDifferentValidators, newRankInInsurance)
		}
	}

	// pairing chunks are immediately pairable
	var pairingChunks []types.Chunk
	if pairingChunks, err = k.GetAllPairingChunks(ctx); err != nil {
		return
	}
	for len(pairingChunks) > 0 && len(newInsurancesWithDifferentValidators) > 0 {
		chunk := pairingChunks[0]
		pairingChunks = pairingChunks[1:]

		newInsurance := newInsurancesWithDifferentValidators[0]
		newInsurancesWithDifferentValidators = newInsurancesWithDifferentValidators[1:]

		validator, found := k.stakingKeeper.GetValidator(ctx, newInsurance.GetValidator())
		if !found {
			err = sdkerrors.Wrapf(types.ErrNotFoundValidator, "validator: %s", newInsurance.GetValidator())
			return
		}

		if _, _, _, err = k.pairChunkAndInsurance(ctx, chunk, newInsurance, validator); err != nil {
			return
		}
	}

	if len(newInsurancesWithDifferentValidators) == 0 {
		for _, outInsurance := range rankOutInsurances {
			chunk, found := k.GetChunk(ctx, outInsurance.ChunkId)
			if !found {
				err = sdkerrors.Wrapf(types.ErrNotFoundChunk, "chunkId: %d", outInsurance.ChunkId)
				return
			}
			if chunk.Status != types.CHUNK_STATUS_UNPAIRING {
				// CRITICAL: Must be unpairing status
				return sdkerrors.Wrapf(types.ErrInvalidChunkStatus, "chunkId: %d", outInsurance.ChunkId)
			}
			del, found := k.stakingKeeper.GetDelegation(ctx, chunk.DerivedAddress(), outInsurance.GetValidator())
			if !found {
				err = sdkerrors.Wrapf(types.ErrNotFoundDelegation, "delegator: %s, validator: %s", chunk.DerivedAddress(), outInsurance.GetValidator())
				return
			}
			if _, err = k.stakingKeeper.Undelegate(ctx, chunk.DerivedAddress(), outInsurance.GetValidator(), del.GetShares()); err != nil {
				return
			}
			continue
		}
		return
	}

	// rest of rankOutInsurances are replaced with newInsurancesWithDifferentValidators
	for _, outInsurance := range rankOutInsurances {
		// Pop cheapest insurance
		newInsurance := newInsurancesWithDifferentValidators[0]
		newInsurancesWithDifferentValidators = newInsurancesWithDifferentValidators[1:] // TODO: check out of index can be happen or not
		chunk := rankOutInsuranceChunkMap[outInsurance.Id]

		// get delegation shares of srcValidator
		delegation, found := k.stakingKeeper.GetDelegation(ctx, chunk.DerivedAddress(), outInsurance.GetValidator())
		if !found {
			return sdkerrors.Wrapf(types.ErrNotFoundDelegation, "delegator: %s, validator: %s", chunk.DerivedAddress(), outInsurance.GetValidator())
		}
		if _, err = k.stakingKeeper.BeginRedelegation(
			ctx,
			chunk.DerivedAddress(),
			outInsurance.GetValidator(),
			newInsurance.GetValidator(),
			delegation.GetShares(),
		); err != nil {
			return err
		}
		k.rePairChunkAndInsurance(ctx, chunk, newInsurance, outInsurance)
	}
	return
}

func (k Keeper) DoLiquidStake(ctx sdk.Context, msg *types.MsgLiquidStake) (chunks []types.Chunk, newShares sdk.Dec, lsTokenMintAmount sdk.Int, err error) {
	delAddr := msg.GetDelegator()
	amount := msg.Amount

	if err = k.ShouldBeBondDenom(ctx, amount.Denom); err != nil {
		return
	}
	// Liquid stakers can send amount of tokens corresponding multiple of chunk size and create multiple chunks
	if err = k.ShouldBeMultipleOfChunkSize(amount.Amount); err != nil {
		return
	}
	chunksToCreate := amount.Amount.Quo(types.ChunkSize).Int64()

	availableChunkSlots := k.GetAvailableChunkSlots(ctx).Int64()
	if (availableChunkSlots - chunksToCreate) < 0 {
		err = sdkerrors.Wrapf(
			types.ErrExceedAvailableChunks,
			"requested chunks to create: %d, available chunks: %d",
			chunksToCreate,
			availableChunkSlots,
		)
		return
	}

	pairingInsurances, validatorMap := k.getPairingInsurances(ctx)
	if chunksToCreate > int64(len(pairingInsurances)) {
		err = types.ErrNoPairingInsurance
		return
	}

	nas := k.GetNetAmountState(ctx)
	types.SortInsurances(validatorMap, pairingInsurances, false)
	totalNewShares := sdk.ZeroDec()
	totalLsTokenMintAmount := sdk.ZeroInt()
	for i := int64(0); i < chunksToCreate; i++ {
		cheapestInsurance := pairingInsurances[0]
		pairingInsurances = pairingInsurances[1:]

		// Now we have the cheapest pairing insurance and valid msg liquid stake! Let's create a chunk
		// Create a chunk
		chunkId := k.getNextChunkIdWithUpdate(ctx)
		chunk := types.NewChunk(chunkId)

		// Escrow liquid staker's coins
		if err = k.bankKeeper.SendCoins(
			ctx,
			delAddr,
			chunk.DerivedAddress(),
			sdk.NewCoins(sdk.NewCoin(amount.Denom, types.ChunkSize)),
		); err != nil {
			return
		}
		validator := validatorMap[cheapestInsurance.ValidatorAddress]

		// Delegate to the validator
		// Delegator: DerivedAddress(chunk.Id)
		// Validator: insurance.ValidatorAddress
		// Amount: msg.Amount
		chunk, cheapestInsurance, newShares, err = k.pairChunkAndInsurance(
			ctx,
			chunk,
			cheapestInsurance,
			validator,
		)
		if err != nil {
			return
		}
		totalNewShares = totalNewShares.Add(newShares)

		liquidBondDenom := k.GetLiquidBondDenom(ctx)
		// Mint the liquid staking token
		lsTokenMintAmount = amount.Amount
		if nas.LsTokensTotalSupply.IsPositive() {
			lsTokenMintAmount = types.NativeTokenToLiquidStakeToken(amount.Amount, nas.LsTokensTotalSupply, nas.NetAmount)
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

	chunksToLiquidUnstake := amount.Amount.Quo(types.ChunkSize).Int64()

	chunksWithInsuranceId := make(map[uint64]types.Chunk)
	var insurances []types.Insurance
	validatorMap := make(map[string]stakingtypes.Validator)
	err = k.IterateAllChunks(ctx, func(chunk types.Chunk) (stop bool, err error) {
		if chunk.Status != types.CHUNK_STATUS_PAIRED {
			return false, nil
		}
		pairedInsurance, found := k.GetInsurance(ctx, chunk.PairedInsuranceId)
		if found == false {
			return false, types.ErrNotFoundInsurance
		}

		if _, ok := validatorMap[pairedInsurance.ValidatorAddress]; !ok {
			// If validator is not in map, get validator from staking keeper
			validator, found := k.stakingKeeper.GetValidator(ctx, pairedInsurance.GetValidator())
			err := k.IsValidValidator(ctx, validator, found)
			if err != nil {
				return false, nil
			}
			validatorMap[pairedInsurance.ValidatorAddress] = validator
		}
		insurances = append(insurances, pairedInsurance)
		chunksWithInsuranceId[chunk.PairedInsuranceId] = chunk
		return false, nil
	})
	if err != nil {
		return
	}

	pairedChunks := int64(len(chunksWithInsuranceId))
	if pairedChunks == 0 {
		err = types.ErrNoPairedChunk
		return
	}
	if chunksToLiquidUnstake > pairedChunks {
		err = sdkerrors.Wrapf(
			types.ErrExceedAvailableChunks,
			"requested chunks to liquid unstake: %d, paired chunks: %d",
			chunksToLiquidUnstake,
			pairedChunks,
		)
		return
	}
	// Sort insurances by descend order
	types.SortInsurances(validatorMap, insurances, true)

	// How much ls tokens must be burned
	nas := k.GetNetAmountState(ctx)
	liquidBondDenom := k.GetLiquidBondDenom(ctx)
	for i := int64(0); i < chunksToLiquidUnstake; i++ {
		// Escrow ls tokens from the delegator
		lsTokenBurnAmount := types.ChunkSize
		if nas.LsTokensTotalSupply.IsPositive() {
			lsTokenBurnAmount = lsTokenBurnAmount.ToDec().Mul(nas.MintRate).TruncateInt()
		}
		lsTokensToBurn := sdk.NewCoin(liquidBondDenom, lsTokenBurnAmount)
		if err = k.bankKeeper.SendCoins(
			ctx, delAddr, types.LsTokenEscrowAcc, sdk.NewCoins(lsTokensToBurn),
		); err != nil {
			return
		}

		mostExpensiveInsurance := insurances[i]
		chunkToBeUndelegated := chunksWithInsuranceId[mostExpensiveInsurance.Id]
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
		toBeUnstakedChunks = append(toBeUnstakedChunks, chunksWithInsuranceId[insurances[i].Id])
		infos = append(infos, info)
		k.SetUnpairingForUnstakingChunkInfo(ctx, info)
	}
	return
}

func (k Keeper) DoProvideInsurance(ctx sdk.Context, msg *types.MsgProvideInsurance) (insurance types.Insurance, err error) {
	providerAddr := msg.GetProvider()
	valAddr := msg.GetValidator()
	feeRate := msg.FeeRate
	amount := msg.Amount

	if err = k.ShouldBeBondDenom(ctx, amount.Denom); err != nil {
		return
	}
	// Check if the amount is greater than minimum coverage
	_, minimumCoverage := k.GetMinimumRequirements(ctx)
	if amount.Amount.LT(minimumCoverage.Amount) {
		err = sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "amount must be greater than minimum coverage: %s", minimumCoverage.String())
		return
	}

	// Check if the validator is valid
	validator, found := k.stakingKeeper.GetValidator(ctx, valAddr)
	err = k.IsValidValidator(ctx, validator, found)
	if err != nil {
		return
	}

	// Create an insurance
	insuranceId := k.getNextInsuranceIdWithUpdate(ctx)
	insurance = types.NewInsurance(insuranceId, providerAddr.String(), valAddr.String(), feeRate)

	// Escrow provider's balance
	if err = k.bankKeeper.SendCoins(
		ctx,
		providerAddr,
		insurance.DerivedAddress(),
		sdk.NewCoins(amount),
	); err != nil {
		return
	}
	k.SetInsurance(ctx, insurance)

	return
}

func (k Keeper) DoCancelProvideInsurance(ctx sdk.Context, msg *types.MsgCancelProvideInsurance) (insurance types.Insurance, err error) {
	providerAddr := msg.GetProvider()
	insuranceId := msg.Id

	// Check if the insurance exists
	insurance, found := k.GetInsurance(ctx, insuranceId)
	if !found {
		err = sdkerrors.Wrapf(types.ErrNotFoundInsurance, "insurance id: %d", insuranceId)
		return
	}

	if insurance.Status != types.INSURANCE_STATUS_PAIRING {
		err = sdkerrors.Wrapf(types.ErrInvalidInsuranceStatus, "insurance id: %d", insuranceId)
		return
	}

	// Check if the provider is the same
	if insurance.ProviderAddress != providerAddr.String() {
		err = sdkerrors.Wrapf(types.ErrNotProviderOfInsurance, "insurance id: %d", insuranceId)
		return
	}

	// Unescrow provider's balance
	escrowed := k.bankKeeper.GetBalance(ctx, insurance.DerivedAddress(), k.stakingKeeper.BondDenom(ctx))
	if err = k.bankKeeper.SendCoins(
		ctx,
		insurance.DerivedAddress(),
		providerAddr,
		sdk.NewCoins(escrowed),
	); err != nil {
		return
	}
	k.DeleteInsurance(ctx, insuranceId)
	return
}

// DoWithdrawInsurance withdraws insurance immediately if it is unpaired.
// If it is paired then it will be queued and unpaired at the epoch.
func (k Keeper) DoWithdrawInsurance(ctx sdk.Context, msg *types.MsgWithdrawInsurance) (withdrawnInsurance types.Insurance, err error) {
	// Get insurance
	insurance, found := k.GetInsurance(ctx, msg.Id)
	if !found {
		err = sdkerrors.Wrapf(types.ErrNotFoundInsurance, "insurance id: %d", msg.Id)
		return
	}
	if msg.ProviderAddress != insurance.ProviderAddress {
		err = sdkerrors.Wrapf(types.ErrNotProviderOfInsurance, "insurance id: %d", msg.Id)
		return
	}

	// If insurance is paired then queue request
	// If insurnace is unpaired then immediately withdraw insurance
	switch insurance.Status {
	case types.INSURANCE_STATUS_PAIRED:
		k.SetWithdrawInsuranceRequest(ctx, types.NewWithdrawInsuranceRequest(msg.Id))
	case types.INSURANCE_STATUS_UNPAIRED:
		// Withdraw immediately
		err = k.withdrawInsurance(ctx, insurance)
	default:
		err = sdkerrors.Wrapf(types.ErrNotInWithdrawableStatus, "insurance status: %s", insurance.Status)
	}
	return
}

// DoWithdrawInsuranceCommission withdraws insurance commission immediately.
func (k Keeper) DoWithdrawInsuranceCommission(ctx sdk.Context, msg *types.MsgWithdrawInsuranceCommission) (err error) {
	providerAddr := msg.GetProvider()
	insuranceId := msg.Id

	// Check if the insurance exists
	insurance, found := k.GetInsurance(ctx, insuranceId)
	if !found {
		err = sdkerrors.Wrapf(types.ErrNotFoundInsurance, "insurance id: %d", insuranceId)
		return
	}

	// Check if the provider is the same
	if insurance.ProviderAddress != providerAddr.String() {
		err = sdkerrors.Wrapf(types.ErrNotProviderOfInsurance, "insurance id: %d", insuranceId)
		return
	}

	// Get all balances of the insurance
	balances := k.bankKeeper.GetAllBalances(ctx, insurance.FeePoolAddress())
	inputs := []banktypes.Input{
		banktypes.NewInput(insurance.FeePoolAddress(), balances),
	}
	outputs := []banktypes.Output{
		banktypes.NewOutput(providerAddr, balances),
	}
	if err = k.bankKeeper.InputOutputCoins(ctx, inputs, outputs); err != nil {
		return
	}
	return
}

// DoDepositInsurance deposits more coin to insurance.
func (k Keeper) DoDepositInsurance(ctx sdk.Context, msg *types.MsgDepositInsurance) (err error) {
	providerAddr := msg.GetProvider()
	insuranceId := msg.Id
	amount := msg.Amount

	insurance, found := k.GetInsurance(ctx, insuranceId)
	if !found {
		err = sdkerrors.Wrapf(types.ErrNotFoundInsurance, "insurance id: %d", insuranceId)
		return
	}

	if insurance.ProviderAddress != providerAddr.String() {
		err = sdkerrors.Wrapf(types.ErrNotProviderOfInsurance, "insurance id: %d", insuranceId)
		return
	}

	if err = k.ShouldBeBondDenom(ctx, amount.Denom); err != nil {
		return
	}

	if err = k.bankKeeper.SendCoins(
		ctx,
		providerAddr,
		insurance.DerivedAddress(),
		sdk.NewCoins(amount),
	); err != nil {
		return
	}
	return
}

// DoClaimDiscountedReward claims discounted reward by paying lstoken.
func (k Keeper) DoClaimDiscountedReward(ctx sdk.Context, msg *types.MsgClaimDiscountedReward) (err error) {
	if err = k.ShouldBeLiquidBondDenom(ctx, msg.Amount.Denom); err != nil {
		return
	}

	discountRate := k.CalcDiscountRate(ctx)
	// discount rate >= minimum discount rate
	// if discount rate(e.g. 10%) is lower than minimum discount rate(e.g. 20%), then it is not profitable to claim reward.
	if discountRate.LT(msg.MinimumDiscountRate) {
		err = sdkerrors.Wrapf(types.ErrDiscountRateTooLow, "current discount rate: %s", discountRate)
		return
	}
	nas := k.GetNetAmountState(ctx)
	discountedMintRate := nas.MintRate.Mul(sdk.OneDec().Sub(discountRate))

	var claimableAmt sdk.Coin
	var burnAmt sdk.Coin

	claimableAmt = k.bankKeeper.GetBalance(ctx, types.RewardPool, k.stakingKeeper.BondDenom(ctx))
	burnAmt = msg.Amount

	// claim amount = (ls token amount / discounted mint rate)
	claimAmt := burnAmt.Amount.ToDec().Quo(discountedMintRate).TruncateInt()
	// Requester can claim only up to claimable amount
	if claimAmt.GT(claimableAmt.Amount) {
		// requester cannot claim more than claimable amount
		claimAmt = claimableAmt.Amount
		// burn amount = (claim amount * discounted mint rate)
		burnAmt.Amount = claimAmt.ToDec().Mul(discountedMintRate).Ceil().TruncateInt()
	}

	if err = k.burnLsTokens(ctx, msg.GetRequestser(), burnAmt); err != nil {
		return
	}
	// send claimAmt to requester (error)
	if err = k.bankKeeper.SendCoins(
		ctx,
		types.RewardPool,
		msg.GetRequestser(),
		sdk.NewCoins(claimableAmt),
	); err != nil {
		return
	}
	return
}

// CalcDiscountRate calculates the current discount rate.
// reward module account's balance / (num paired chunks * chunk size)
func (k Keeper) CalcDiscountRate(ctx sdk.Context) sdk.Dec {
	accumulated := k.bankKeeper.GetBalance(ctx, types.RewardPool, k.stakingKeeper.BondDenom(ctx))
	numPairedChunks := k.getNumPairedChunks(ctx)
	if accumulated.IsZero() || numPairedChunks == 0 {
		return sdk.ZeroDec()
	}
	discountRate := accumulated.Amount.ToDec().Quo(
		sdk.NewInt(numPairedChunks).Mul(types.ChunkSize).ToDec(),
	)
	return sdk.MinDec(discountRate, types.MaximumDiscountRate)
}

func (k Keeper) SetLiquidBondDenom(ctx sdk.Context, denom string) {
	store := ctx.KVStore(k.storeKey)
	store.Set(types.KeyLiquidBondDenom, []byte(denom))
}

func (k Keeper) GetLiquidBondDenom(ctx sdk.Context) string {
	store := ctx.KVStore(k.storeKey)
	return string(store.Get(types.KeyLiquidBondDenom))
}

func (k Keeper) IsValidValidator(ctx sdk.Context, validator stakingtypes.Validator, found bool) error {
	if !found {
		return types.ErrNotFoundValidator
	}
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
func (k Keeper) completeInsuranceDuty(ctx sdk.Context, insurance types.Insurance) (types.Chunk, types.Insurance, error) {
	// get chunk
	chunk, found := k.GetChunk(ctx, insurance.ChunkId)
	if !found {
		return chunk, insurance, sdkerrors.Wrapf(types.ErrNotFoundChunk, "chunk id: %d", insurance.ChunkId)
	}

	// insurance duty is over
	insurance.ChunkId = types.Empty
	insurance.SetStatus(types.INSURANCE_STATUS_UNPAIRED)

	switch chunk.Status {
	case types.CHUNK_STATUS_UNPAIRING_FOR_UNSTAKING, types.CHUNK_STATUS_UNPAIRING, types.CHUNK_STATUS_PAIRED:
		chunk.UnpairingInsuranceId = types.Empty
	}

	k.SetInsurance(ctx, insurance)
	k.SetChunk(ctx, chunk)
	return chunk, insurance, nil
}

// completeLiquidStake completes liquid stake.
// TODO: write TC for penalty situation
func (k Keeper) completeLiquidUnstake(ctx sdk.Context, chunk types.Chunk) error {
	if chunk.Status != types.CHUNK_STATUS_UNPAIRING_FOR_UNSTAKING {
		return sdkerrors.Wrapf(types.ErrInvalidChunkStatus, "chunk status: %s", chunk.Status)
	}
	var err error

	bondDenom := k.stakingKeeper.BondDenom(ctx)
	liquidBondDenom := k.GetLiquidBondDenom(ctx)

	// get paired insurance from chunk
	unpairingInsurance, found := k.GetInsurance(ctx, chunk.UnpairingInsuranceId)
	if !found {
		return sdkerrors.Wrapf(types.ErrNotFoundInsurance, "insurance id: %d", chunk.UnpairingInsuranceId)
	}
	if chunk.PairedInsuranceId != 0 {
		return sdkerrors.Wrapf(types.ErrUnpairingChunkHavePairedChunk, "paired insurance id: %d", chunk.PairedInsuranceId)
	}

	// unpairing for unstake chunk only have unpairing insurance
	_, found = k.stakingKeeper.GetUnbondingDelegation(ctx, chunk.DerivedAddress(), unpairingInsurance.GetValidator())
	if found {
		// UnbondingDelegation must be removed by staking keeper EndBlocker
		// because Endblocker of liquidstaking module is called after staking module.
		return sdkerrors.Wrapf(types.ErrUnbondingDelegationNotRemoved, "chunk id: %d", chunk.Id)
	}
	// handle mature unbondings
	info, found := k.GetUnpairingForUnstakingChunkInfo(ctx, chunk.Id)
	if !found {
		return sdkerrors.Wrapf(types.ErrNotFoundUnpairingForUnstakingChunkInfo, "chunk id: %d", chunk.Id)
	}
	lsTokensToBurn := info.EscrowedLstokens
	unstakedTokens := sdk.NewCoin(bondDenom, types.ChunkSize)
	penalty := types.ChunkSize.Sub(k.bankKeeper.GetBalance(ctx, chunk.DerivedAddress(), bondDenom).Amount)
	if penalty.IsPositive() {
		// send penalty to reward pool
		if err = k.bankKeeper.SendCoins(
			ctx,
			unpairingInsurance.DerivedAddress(),
			types.RewardPool,
			sdk.NewCoins(sdk.NewCoin(bondDenom, penalty)),
		); err != nil {
			return err
		}
		penaltyRatio := penalty.ToDec().Quo(types.ChunkSize.ToDec())
		discount := penaltyRatio.Mul(lsTokensToBurn.Amount.ToDec())
		refund := sdk.NewCoin(liquidBondDenom, discount.TruncateInt())

		// send discount lstokens to info.Delegator
		if err = k.bankKeeper.SendCoins(
			ctx,
			types.LsTokenEscrowAcc,
			info.GetDelegator(),
			sdk.NewCoins(refund),
		); err != nil {
			return err
		}
		lsTokensToBurn = lsTokensToBurn.Sub(refund)
		unstakedTokens.Amount = unstakedTokens.Amount.Sub(penalty)
	}
	// insurance duty is over
	if chunk, unpairingInsurance, err = k.completeInsuranceDuty(ctx, unpairingInsurance); err != nil {
		return err
	}
	if err = k.burnEscrowedLsTokens(ctx, lsTokensToBurn); err != nil {
		return err
	}
	// TODO: remove panic after fuzzing tests, it will be better to send chunk balance instead of unstakedTokens
	chunkBalance := k.bankKeeper.GetBalance(ctx, chunk.DerivedAddress(), bondDenom)
	if !types.ChunkSize.Sub(penalty).Equal(chunkBalance.Amount) {
		panic("investigating it")
	}
	if err = k.bankKeeper.SendCoins(
		ctx,
		chunk.DerivedAddress(),
		info.GetDelegator(),
		sdk.NewCoins(unstakedTokens),
	); err != nil {
		return err
	}
	k.DeleteUnpairingForUnstakingChunkInfo(ctx, chunk.Id)
	k.DeleteChunk(ctx, chunk.Id)
	return nil
}

// handleUnpairingChunk handles unpairing chunk which created previous epoch.
// Those chunks completed their unbonding already.
func (k Keeper) handleUnpairingChunk(ctx sdk.Context, chunk types.Chunk) error {
	if chunk.Status != types.CHUNK_STATUS_UNPAIRING {
		return sdkerrors.Wrapf(types.ErrInvalidChunkStatus, "chunk id: %d, status: %s", chunk.Id, chunk.Status)
	}
	var err error
	bondDenom := k.stakingKeeper.BondDenom(ctx)

	// get paired insurance from chunk
	unpairingInsurance, found := k.GetInsurance(ctx, chunk.UnpairingInsuranceId)
	if !found {
		return sdkerrors.Wrapf(types.ErrNotFoundInsurance, "insurance id: %d", chunk.UnpairingInsuranceId)
	}
	if chunk.HasPairedInsurance() {
		return sdkerrors.Wrapf(types.ErrUnpairingChunkHavePairedChunk, "paired insurance id: %d", chunk.PairedInsuranceId)
	}
	if _, found = k.stakingKeeper.GetUnbondingDelegation(ctx, chunk.DerivedAddress(), unpairingInsurance.GetValidator()); found {
		// UnbondingDelegation must be removed by staking keeper EndBlocker
		// because Endblocker of liquidstaking module is called after staking module.
		return sdkerrors.Wrapf(types.ErrUnbondingDelegationNotRemoved, "chunk id: %d", chunk.Id)
	}

	chunkBalance := k.bankKeeper.GetBalance(ctx, chunk.DerivedAddress(), bondDenom).Amount
	insuranceBalance := k.bankKeeper.GetBalance(ctx, unpairingInsurance.DerivedAddress(), bondDenom).Amount
	penalty := types.ChunkSize.Sub(chunkBalance)
	if penalty.IsPositive() {
		var sendAmount sdk.Coin
		if penalty.GT(insuranceBalance) {
			sendAmount = sdk.NewCoin(bondDenom, insuranceBalance)
		} else {
			sendAmount = sdk.NewCoin(bondDenom, penalty)
		}

		// Send penalty to chunk
		// unpairing chunk must be not damaged to become pairing chunk
		if err = k.bankKeeper.SendCoins(
			ctx,
			unpairingInsurance.DerivedAddress(),
			chunk.DerivedAddress(),
			sdk.NewCoins(sendAmount),
		); err != nil {
			return err
		}
		chunkBalance = k.bankKeeper.GetBalance(ctx, chunk.DerivedAddress(), bondDenom).Amount
	}
	if chunk, unpairingInsurance, err = k.completeInsuranceDuty(ctx, unpairingInsurance); err != nil {
		return err
	}

	// If chunk got damaged, all of its coins will be sent to reward module account and chunk will be deleted
	if chunkBalance.LT(types.ChunkSize) {
		allBalances := k.bankKeeper.GetAllBalances(ctx, chunk.DerivedAddress())
		var inputs []banktypes.Input
		var outputs []banktypes.Output
		inputs = append(inputs, banktypes.NewInput(chunk.DerivedAddress(), allBalances))
		outputs = append(outputs, banktypes.NewOutput(types.RewardPool, allBalances))

		if err = k.bankKeeper.InputOutputCoins(ctx, inputs, outputs); err != nil {
			return err
		}
		k.DeleteChunk(ctx, chunk.Id)
		// Insurance already sent all of its balance to chunk, but we cannot delete it yet
		// because it can have remaining commissions.
		if k.bankKeeper.GetAllBalances(ctx, unpairingInsurance.FeePoolAddress()).IsZero() {
			// if insurance has no commissions, we can delete it
			k.DeleteInsurance(ctx, unpairingInsurance.Id)
		}
		return nil
	}
	chunk.SetStatus(types.CHUNK_STATUS_PAIRING)
	k.SetChunk(ctx, chunk)

	return nil
}

// TODO: Unpairing insurance should cover infraction height before replacing.
func (k Keeper) handlePairedChunk(ctx sdk.Context, chunk types.Chunk) error {
	if chunk.Status != types.CHUNK_STATUS_PAIRED {
		return sdkerrors.Wrapf(types.ErrInvalidChunkStatus, "chunk id: %d, status: %s", chunk.Id, chunk.Status)
	}

	var err error
	bondDenom := k.stakingKeeper.BondDenom(ctx)
	// Get insurance from chunk
	pairedInsurance, found := k.GetInsurance(ctx, chunk.PairedInsuranceId)
	if !found {
		return sdkerrors.Wrapf(types.ErrNotFoundInsurance, "insurance id: %d", chunk.PairedInsuranceId)
	}

	validator, found := k.stakingKeeper.GetValidator(ctx, pairedInsurance.GetValidator())
	err = k.IsValidValidator(ctx, validator, found)
	// TODO: Should we un-pair insurances which have invalid validator?
	if err == types.ErrNotFoundValidator {
		return sdkerrors.Wrapf(err, "validator: %s", pairedInsurance.GetValidator())
	}

	// Get delegation of chunk
	delegation, found := k.stakingKeeper.GetDelegation(ctx, chunk.DerivedAddress(), validator.GetOperator())
	if !found {
		return sdkerrors.Wrapf(types.ErrNotFoundDelegation, "delegator: %s, validator: %s", chunk.DerivedAddress(), validator.GetOperator())
	}
	// TODO: Consider ReDelegation

	insuranceOutOfBalance := false
	// Check whether delegation value is decreased by slashing
	// The check process should use TokensFromShares to get the current delegation value
	tokens := validator.TokensFromShares(delegation.GetShares())
	penalty := types.ChunkSize.ToDec().Sub(tokens)
	if penalty.IsPositive() {
		// TODO: Check when slashing happened and decide which insurances (unpairing or paired) should cover penalty.
		// check penalty is bigger than insurance balance
		insuranceBalance := k.bankKeeper.GetBalance(
			ctx,
			pairedInsurance.DerivedAddress(),
			bondDenom,
		)
		// EDGE CASE: Insurance cannot cover penalty
		if penalty.GT(insuranceBalance.Amount.ToDec()) {
			insuranceOutOfBalance = true
			k.startUnpairing(ctx, pairedInsurance, chunk)
			// start unbonding of chunk because it is damaged
			if _, err = k.stakingKeeper.Undelegate(
				ctx, chunk.DerivedAddress(),
				validator.GetOperator(),
				delegation.GetShares(),
			); err != nil {
				return err
			}
		} else {
			// Insurance can cover penalty
			// 1. Send penalty to chunk
			// 2. chunk delegate additional tokens to validator

			var penaltyCoin sdk.Coin
			if penalty.GT(penalty.TruncateDec()) {
				penaltyCoin = sdk.NewCoin(bondDenom, penalty.Ceil().TruncateInt())
			} else {
				penaltyCoin = sdk.NewCoin(bondDenom, penalty.TruncateInt())
			}
			// send penalty to chunk
			if err = k.bankKeeper.SendCoins(
				ctx,
				pairedInsurance.DerivedAddress(),
				chunk.DerivedAddress(),
				sdk.NewCoins(penaltyCoin),
			); err != nil {
				return err
			}
			// delegate additional tokens to validator as chunk.DerivedAddress()
			if _, err = k.stakingKeeper.Delegate(
				ctx,
				chunk.DerivedAddress(),
				penaltyCoin.Amount,
				stakingtypes.Unbonded,
				validator,
				true,
			); err != nil {
				return err
			}
		}
	}

	if !insuranceOutOfBalance && !k.IsSufficientInsurance(ctx, pairedInsurance) {
		k.startUnpairing(ctx, pairedInsurance, chunk)
	}

	// TODO: use IsValidValidator but should it be handled above?
	if err := k.IsValidValidator(ctx, validator, found); err != nil {
		// Find all insurances which have same validator with this
		var invalidInsurances []types.Insurance
		if err = k.IterateAllInsurances(ctx, func(insurance types.Insurance) (bool, error) {
			if insurance.Status != types.INSURANCE_STATUS_PAIRED {
				return false, nil
			}
			if insurance.GetValidator().Equals(pairedInsurance.GetValidator()) {
				invalidInsurances = append(invalidInsurances, insurance)
			}
			return false, nil
		}); err != nil {
			return err
		}
		for _, insurance := range invalidInsurances {
			chunk, found := k.GetChunk(ctx, insurance.ChunkId)
			if !found {
				return sdkerrors.Wrapf(types.ErrNotFoundChunk, "chunk id: %d", insurance.ChunkId)
			}
			k.startUnpairing(ctx, insurance, chunk)
		}
	}

	unpairingInsurance, found := k.GetInsurance(ctx, chunk.UnpairingInsuranceId)
	if found {
		if _, _, err = k.completeInsuranceDuty(ctx, unpairingInsurance); err != nil {
			return err
		}
	}
	return nil
}

// IsSufficientInsurance checks whether insurance has sufficient balance to cover slashing or not.
func (k Keeper) IsSufficientInsurance(ctx sdk.Context, insurance types.Insurance) bool {
	insuranceBalance := k.bankKeeper.GetBalance(ctx, insurance.DerivedAddress(), k.stakingKeeper.BondDenom(ctx))
	_, slashingCoverage := k.GetMinimumRequirements(ctx)
	if insuranceBalance.Amount.LT(slashingCoverage.Amount) {
		return false
	}
	return true
}

// GetAvailableChunkSlots returns a number of chunk which can be paired.
func (k Keeper) GetAvailableChunkSlots(ctx sdk.Context) sdk.Int {
	return k.MaxPairedChunks(ctx).Sub(sdk.NewInt(k.getNumPairedChunks(ctx)))
}

// startUnpairing changes status of insurance and chunk to unpairing.
// Actual unpairing process including un-delegate chunk will be done after ranking in EndBlocker.
func (k Keeper) startUnpairing(
	ctx sdk.Context,
	insurance types.Insurance,
	chunk types.Chunk,
) {
	insurance.SetStatus(types.INSURANCE_STATUS_UNPAIRING)
	chunk.UnpairingInsuranceId = chunk.PairedInsuranceId
	chunk.PairedInsuranceId = 0
	chunk.SetStatus(types.CHUNK_STATUS_UNPAIRING)
	k.SetChunk(ctx, chunk)
	k.SetInsurance(ctx, insurance)
}

// startUnpairingForLiquidUnstake changes status of insurance to unpairing and
// chunk to UnpairingForUnstaking.
func (k Keeper) startUnpairingForLiquidUnstake(
	ctx sdk.Context,
	insurance types.Insurance,
	chunk types.Chunk,
) (types.Insurance, types.Chunk) {
	chunk.SetStatus(types.CHUNK_STATUS_UNPAIRING_FOR_UNSTAKING)
	chunk.UnpairingInsuranceId = chunk.PairedInsuranceId
	chunk.PairedInsuranceId = types.Empty
	insurance.SetStatus(types.INSURANCE_STATUS_UNPAIRING)
	k.SetChunk(ctx, chunk)
	k.SetInsurance(ctx, insurance)
	return insurance, chunk
}

// withdrawInsurance withdraws insurance and commissions from insurance account immediately.
func (k Keeper) withdrawInsurance(ctx sdk.Context, insurance types.Insurance) error {
	insuranceTokens := k.bankKeeper.GetAllBalances(ctx, insurance.DerivedAddress())
	commissions := k.bankKeeper.GetAllBalances(ctx, insurance.FeePoolAddress())
	inputs := []banktypes.Input{
		banktypes.NewInput(insurance.DerivedAddress(), insuranceTokens),
		banktypes.NewInput(insurance.FeePoolAddress(), commissions),
	}
	outpus := []banktypes.Output{
		banktypes.NewOutput(insurance.GetProvider(), insuranceTokens),
		banktypes.NewOutput(insurance.GetProvider(), commissions),
	}
	if err := k.bankKeeper.InputOutputCoins(ctx, inputs, outpus); err != nil {
		return err
	}
	k.DeleteInsurance(ctx, insurance.Id)
	return nil
}

// pairChunkAndInsurance pairs chunk and insurance.
func (k Keeper) pairChunkAndInsurance(
	ctx sdk.Context,
	chunk types.Chunk,
	insurance types.Insurance,
	validator stakingtypes.Validator,
) (types.Chunk, types.Insurance, sdk.Dec, error) {
	newShares, err := k.stakingKeeper.Delegate(
		ctx,
		chunk.DerivedAddress(),
		types.ChunkSize,
		stakingtypes.Unbonded,
		validator,
		true,
	)
	if err != nil {
		return types.Chunk{}, types.Insurance{}, sdk.Dec{}, err
	}
	chunk.PairedInsuranceId = insurance.Id
	insurance.ChunkId = chunk.Id
	chunk.SetStatus(types.CHUNK_STATUS_PAIRED)
	insurance.SetStatus(types.INSURANCE_STATUS_PAIRED)
	k.SetChunk(ctx, chunk)
	k.SetInsurance(ctx, insurance)
	return chunk, insurance, newShares, nil
}

func (k Keeper) rePairChunkAndInsurance(ctx sdk.Context, chunk types.Chunk, newInsurance, outInsurance types.Insurance) {
	chunk.UnpairingInsuranceId = outInsurance.Id
	outInsurance.SetStatus(types.INSURANCE_STATUS_UNPAIRING)
	chunk.PairedInsuranceId = newInsurance.Id
	newInsurance.ChunkId = chunk.Id
	newInsurance.SetStatus(types.INSURANCE_STATUS_PAIRED)
	chunk.SetStatus(types.CHUNK_STATUS_PAIRED)
	k.SetInsurance(ctx, outInsurance)
	k.SetInsurance(ctx, newInsurance)
	k.SetChunk(ctx, chunk)
}

func (k Keeper) getNumPairedChunks(ctx sdk.Context) (numPairedChunks int64) {
	k.IterateAllChunks(ctx, func(chunk types.Chunk) (bool, error) {
		if chunk.Status != types.CHUNK_STATUS_PAIRED {
			return false, nil
		}
		numPairedChunks++
		return false, nil
	})
	return
}
