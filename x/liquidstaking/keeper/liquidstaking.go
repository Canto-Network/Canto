package keeper

import (
	"github.com/Canto-Network/Canto/v6/x/liquidstaking/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

// DistributeReward withdraws delegation rewards from all paired chunks
// Keeper.CollectReward will be called during withdrawing process.
func (k Keeper) DistributeReward(ctx sdk.Context) {
	err := k.IterateAllChunks(ctx, func(chunk types.Chunk) (bool, error) {
		if chunk.Status != types.CHUNK_STATUS_PAIRED {
			return false, nil
		}
		// get an insurance from chunk
		insurance, found := k.GetInsurance(ctx, chunk.Id)
		if !found {
			panic(types.ErrNotFoundInsurance.Error())
		}
		validator, found := k.stakingKeeper.GetValidator(ctx, insurance.GetValidator())
		if !found {
			// Tombstoned validator can be existed in staking keeper
			panic(types.ErrValidatorNotFound.Error())
		}
		// withdraw delegation reward of chunk
		// AfterModifiedHook will call CollectReward
		_, err := k.distributionKeeper.WithdrawDelegationRewards(ctx, chunk.DerivedAddress(), validator.GetOperator())
		if err != nil {
			panic(err.Error())
		}
		return false, nil
	})
	if err != nil {
		panic(err.Error())
	}
}

func (k Keeper) CoverSlashingAndHandleMatureUnbondings(ctx sdk.Context) {
	var err error
	err = k.IterateAllChunks(ctx, func(chunk types.Chunk) (bool, error) {
		switch chunk.Status {
		// Finish mature unbondings triggered in previous epoch
		case types.CHUNK_STATUS_UNPAIRING_FOR_UNSTAKE:
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
func (k Keeper) HandleQueuedLiquidUnstakes(ctx sdk.Context) ([]types.Chunk, error) {
	var unstakedChunks []types.Chunk
	pendingLiquidunstakes := k.GetAllPendingLiquidUnstake(ctx)
	for _, plu := range pendingLiquidunstakes {
		// Get chunk
		chunk, found := k.GetChunk(ctx, plu.ChunkId)
		if !found {
			return nil, sdkerrors.Wrapf(types.ErrNotFoundChunk, "id: %d", plu.ChunkId)
		}
		if chunk.Status != types.CHUNK_STATUS_PAIRED {
			return nil, sdkerrors.Wrapf(types.ErrInvalidChunkStatus, "id: %d, status: %s", chunk.Id, chunk.Status)
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
		chunk.SetStatus(types.CHUNK_STATUS_UNPAIRING_FOR_UNSTAKE)
		chunk.UnpairingInsuranceId = chunk.PairedInsuranceId
		chunk.PairedInsuranceId = 0
		insurance.SetStatus(types.INSURANCE_STATUS_UNPAIRING)
		k.SetChunk(ctx, chunk)
		k.SetInsurance(ctx, insurance)
		unstakedChunks = append(unstakedChunks, chunk)
		// Set tracking obj
		k.SetUnpairingForUnstakeChunkInfo(
			ctx,
			types.NewUnpairingForUnstakeChunkInfo(chunk.Id, plu.DelegatorAddress, plu.EscrowedLstokens),
		)
		k.DeletePendingLiquidUnstake(ctx, plu)
	}
	return unstakedChunks, nil
}

// HandleQueuedWithdrawInsuranceRequests processes withdraw insurance requests that were queued before the epoch.
// It will unpair the chunk and insurance.
// Unpairing insurances will be unpaired in the next epoch.
// After insurance is unpaired, it can be withdrawn by MsgWithdrawInsurance immediately.
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
		insurance.SetStatus(types.INSURANCE_STATUS_UNPAIRING_FOR_WITHDRAW)
		chunk.UnpairingInsuranceId = chunk.PairedInsuranceId
		chunk.PairedInsuranceId = 0
		k.SetInsurance(ctx, insurance)
		k.SetChunk(ctx, chunk)
		withdrawnInsurances = append(withdrawnInsurances, insurance)
	}
	return nil, nil
}

func (k Keeper) DoLiquidStake(ctx sdk.Context, msg *types.MsgLiquidStake) (chunks []types.Chunk, newShares sdk.Dec, lsTokenMintAmount sdk.Int, err error) {
	delAddr := msg.GetDelegator()
	amount := msg.Amount

	// Check if max paired chunk size is exceeded or not
	currenPairedChunks := 0
	err = k.IterateAllChunks(ctx, func(chunk types.Chunk) (bool, error) {
		if chunk.Status == types.CHUNK_STATUS_PAIRED {
			currenPairedChunks++
		}
		return false, nil
	})
	if err != nil {
		return
	}
	availableChunks := types.MaxPairedChunks - currenPairedChunks
	if availableChunks <= 0 {
		err = sdkerrors.Wrapf(types.ErrMaxPairedChunkSizeExceeded, "current paired chunk size: %d", currenPairedChunks)
		return
	}

	if err = k.ShouldBeBondDenom(ctx, amount.Denom); err != nil {
		return
	}
	// Liquid stakers can send amount of tokens corresponding multiple of chunk size and create multiple chunks
	if err = k.ShouldBeMultipleOfChunkSize(amount.Amount); err != nil {
		return
	}
	chunksToCreate := amount.Amount.Quo(types.ChunkSize).Int64()
	if chunksToCreate > int64(availableChunks) {
		err = sdkerrors.Wrapf(
			types.ErrExceedAvailableChunks,
			"requested chunks to create: %d, available chunks: %d",
			chunksToCreate,
			availableChunks,
		)
		return
	}

	pairingInsurances, validatorMap := k.getPairingInsurances(ctx)
	if chunksToCreate > int64(len(pairingInsurances)) {
		err = types.ErrNoPairingInsurance
		return
	}

	// TODO: Must be proved that this is safe, we must call this before sending
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
			sdk.NewCoins(amount),
		); err != nil {
			return
		}
		chunk.PairedInsuranceId = cheapestInsurance.Id
		validator := validatorMap[cheapestInsurance.ValidatorAddress]

		// Delegate to the validator
		// Delegator: DerivedAddress(chunk.Id)
		// Validator: insurance.ValidatorAddress
		// Amount: msg.Amount
		newShares, err = k.stakingKeeper.Delegate(ctx, chunk.DerivedAddress(), amount.Amount, stakingtypes.Unbonded, validator, true)
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
		chunk.SetStatus(types.CHUNK_STATUS_PAIRED)
		cheapestInsurance.SetStatus(types.INSURANCE_STATUS_PAIRED)
		k.SetChunk(ctx, chunk)
		k.SetInsurance(ctx, cheapestInsurance)
		k.DeletePairingInsuranceIndex(ctx, cheapestInsurance)
		chunks = append(chunks, chunk)
	}
	return
}

// QueueLiquidUnstake queues MsgLiquidUnstake.
// Actual unstaking will be done in the next epoch.
func (k Keeper) QueueLiquidUnstake(ctx sdk.Context, msg *types.MsgLiquidUnstake) (
	unstakedChunks []types.Chunk,
	err error,
) {
	delAddr := msg.GetDelegator()
	amount := msg.Amount

	if err = k.ShouldBeMultipleOfChunkSize(amount.Amount); err != nil {
		return
	}
	if err = k.ShouldBeBondDenom(ctx, amount.Denom); err != nil {
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
			valid, err := k.isValidValidator(ctx, validator, found)
			if err != nil {
				return false, nil
			}
			if valid {
				validatorMap[pairedInsurance.ValidatorAddress] = validator
			} else {
				return false, nil
			}
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
		k.SetPendingLiquidUnstake(
			ctx,
			types.NewPendingLiquidUnstake(
				chunkToBeUndelegated.Id,
				chunkToBeUndelegated.DerivedAddress().String(), lsTokensToBurn,
			),
		)
	}
	return
}

func (k Keeper) DoInsuranceProvide(ctx sdk.Context, msg *types.MsgInsuranceProvide) (insurance types.Insurance, err error) {
	providerAddr := msg.GetProvider()
	valAddr := msg.GetValidator()
	feeRate := msg.FeeRate
	amount := msg.Amount

	// Check if the amount is greater than minimum coverage
	_, minimumCoverage := k.GetMinimumRequirements(ctx)
	if amount.Amount.LT(minimumCoverage.Amount) {
		err = sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "amount must be greater than minimum coverage: %s", minimumCoverage.String())
		return
	}

	// Check if the validator is valid
	validator, found := k.stakingKeeper.GetValidator(ctx, valAddr)
	_, err = k.isValidValidator(ctx, validator, found)
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
	k.SetPairingInsuranceIndex(ctx, insurance)
	k.SetInsurancesByProviderIndex(ctx, insurance)

	return
}

func (k Keeper) DoCancelInsuranceProvide(ctx sdk.Context, msg *types.MsgCancelInsuranceProvide) (insurance types.Insurance, err error) {
	providerAddr := msg.GetProvider()
	insuranceId := msg.Id

	// Check if the insurance exists
	insurance, found := k.GetPairingInsurance(ctx, insuranceId)
	if !found {
		err = sdkerrors.Wrapf(types.ErrPairingInsuranceNotFound, "insurance id: %d", insuranceId)
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
	k.DeleteInsurancesByProviderIndex(ctx, insurance)
	k.DeletePairingInsuranceIndex(ctx, insurance)
	return
}

// DoWithdrawInsurance withdraws insurance immediately if it is unpaired.
// If it is paired then it will be queued and unpaired at the epoch.
func (k Keeper) DoWithdrawInsurance(ctx sdk.Context, msg *types.MsgWithdrawInsurance) (withdrawnInsurance types.Insurance, err error) {
	// Get insurance
	insurance, found := k.GetPairingInsurance(ctx, msg.Id)
	if !found {
		err = sdkerrors.Wrapf(types.ErrPairingInsuranceNotFound, "insurance id: %d", msg.Id)
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

func (k Keeper) SetLiquidBondDenom(ctx sdk.Context, denom string) {
	store := ctx.KVStore(k.storeKey)
	store.Set(types.KeyLiquidBondDenom, []byte(denom))
}

func (k Keeper) GetLiquidBondDenom(ctx sdk.Context) string {
	store := ctx.KVStore(k.storeKey)
	return string(store.Get(types.KeyLiquidBondDenom))
}

func (k Keeper) isValidValidator(ctx sdk.Context, validator stakingtypes.Validator, found bool) (bool, error) {
	if !found {
		return false, types.ErrValidatorNotFound
	}
	pubKey, err := validator.ConsPubKey()
	if err != nil {
		return false, err
	}
	if k.slashingKeeper.IsTombstoned(ctx, sdk.ConsAddress(pubKey.Address())) {
		return false, types.ErrTombstonedValidator
	}
	return true, nil
}

// Get minimum requirements for liquid staking
// Liquid staker must provide at least one chunk amount
// Insurance provider must provide at least slashing coverage
func (k Keeper) GetMinimumRequirements(ctx sdk.Context) (oneChunkAmount, slashingCoverage sdk.Coin) {
	bondDenom := k.stakingKeeper.BondDenom(ctx)
	oneChunkAmount = sdk.NewCoin(bondDenom, types.ChunkSize)
	fraction := sdk.MustNewDecFromStr(types.SlashFraction)
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

func (k Keeper) completeInsuranceDuty(ctx sdk.Context, insurance types.Insurance) error {
	// get chunk
	chunk, found := k.GetChunk(ctx, insurance.ChunkId)
	if !found {
		return sdkerrors.Wrapf(types.ErrNotFoundChunk, "chunk id: %d", insurance.ChunkId)
	}

	// insurance duty is over
	insurance.ChunkId = 0
	chunk.UnpairingInsuranceId = chunk.PairedInsuranceId
	chunk.PairedInsuranceId = 0

	insurance.SetStatus(types.INSURANCE_STATUS_UNPAIRED)
	chunk.SetStatus(types.CHUNK_STATUS_UNSPECIFIED)
	k.SetInsurance(ctx, insurance)
	k.SetChunk(ctx, chunk)
	return nil
}

func (k Keeper) completeLiquidUnstake(ctx sdk.Context, chunk types.Chunk) error {
	var err error

	bondDenom := k.stakingKeeper.BondDenom(ctx)
	liquidBondDenom := k.GetLiquidBondDenom(ctx)

	// get paired insurance from chunk
	unpairingInsurnace, found := k.GetInsurance(ctx, chunk.UnpairingInsuranceId)
	if !found {
		return sdkerrors.Wrapf(types.ErrNotFoundInsurance, "insurance id: %d", chunk.UnpairingInsuranceId)
	}

	if chunk.PairedInsuranceId != 0 {
		return sdkerrors.Wrapf(types.ErrUnpairingChunkHavePairedChunk, "paired insurance id: %d", chunk.PairedInsuranceId)
	}

	// unpairing for unstake chunk only have unpairing insurance
	_, found = k.stakingKeeper.GetUnbondingDelegation(ctx, chunk.DerivedAddress(), unpairingInsurnace.GetValidator())
	if found {
		// UnbondingDelegation must be removed by staking keeper EndBlocker
		// because Endblocker of liquidstaking module is called after staking module.
		return sdkerrors.Wrapf(types.ErrUnbondingDelegationNotRemoved, "chunk id: %d", chunk.Id)
	}
	// handle mature unbondings
	info, found := k.GetUnpairingForUnstakeChunkInfo(ctx, chunk.Id)
	if !found {
		return sdkerrors.Wrapf(types.ErrNotFoundUnpairingForUnstakeChunkInfo, "chunk id: %d", chunk.Id)
	}
	lsTokensToBurn := info.EscrowedLstokens
	penalty := types.ChunkSize.Sub(k.bankKeeper.GetBalance(ctx, chunk.DerivedAddress(), bondDenom).Amount)
	if penalty.IsPositive() {
		// send penalty to reward pool
		if err = k.bankKeeper.SendCoins(
			ctx,
			unpairingInsurnace.DerivedAddress(),
			types.RewardPool,
			sdk.NewCoins(sdk.NewCoin(bondDenom, penalty)),
		); err != nil {
			return err
		}
		penaltyRatio := penalty.ToDec().Quo(types.ChunkSize.ToDec())
		discount := penaltyRatio.Mul(types.ChunkSize.ToDec())
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
	}
	// insurance duty is over
	if err = k.completeInsuranceDuty(ctx, unpairingInsurnace); err != nil {
		return err
	}
	if err = k.burnEscrowedLsTokens(ctx, lsTokensToBurn); err != nil {
		return err
	}
	k.DeleteUnpairingForUnstakeChunkInfo(ctx, chunk.Id)
	k.DeleteChunk(ctx, chunk.Id)
	return nil
}

func (k Keeper) handleUnpairingChunk(ctx sdk.Context, chunk types.Chunk) error {
	// TODO: Implement
	return nil
}

func (k Keeper) handlePairedChunk(ctx sdk.Context, chunk types.Chunk) error {
	// TODO: Implement
	return nil
}

func (k Keeper) withdrawInsurance(ctx sdk.Context, insurance types.Insurance) error {
	insuranceTokens := k.bankKeeper.GetAllBalances(ctx, insurance.DerivedAddress())
	if err := k.bankKeeper.SendCoins(ctx, insurance.DerivedAddress(), insurance.GetProvider(), insuranceTokens); err != nil {
		return err
	}
	commissions := k.bankKeeper.GetAllBalances(ctx, insurance.FeePoolAddress())
	if err := k.bankKeeper.SendCoins(ctx, insurance.DerivedAddress(), insurance.GetProvider(), commissions); err != nil {
		return err
	}
	k.DeleteInsurance(ctx, insurance.Id)
	return nil
}
