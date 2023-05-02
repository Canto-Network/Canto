package keeper

import (
	"github.com/Canto-Network/Canto/v6/x/liquidstaking/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

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

	bondDenom := k.stakingKeeper.BondDenom(ctx)
	minimumRequirement, _ := k.GetMinimumRequirements(ctx)
	// amount must be greater than or equal to one chunk size
	if !amount.IsGTE(minimumRequirement) {
		err = sdkerrors.Wrapf(types.ErrInvalidAmount, "amount must be greater than or equal to %s", minimumRequirement.String())
		return
	}

	// Check if there are any pairing insurances
	var pairingInsurances []types.Insurance
	validatorMap := make(map[string]stakingtypes.Validator)
	err = k.IteratePairingInsurances(ctx, func(insurance types.Insurance) (bool, error) {
		if _, ok := validatorMap[insurance.ValidatorAddress]; !ok {
			// If validator is not in map, get validator from staking keeper
			valAddr, err := sdk.ValAddressFromBech32(insurance.ValidatorAddress)
			if err != nil {
				return false, err
			}
			validator, found := k.stakingKeeper.GetValidator(ctx, valAddr)
			valid, err := k.isValidValidator(ctx, validator, found)
			if err != nil {
				return false, nil
			}
			if valid {
				validatorMap[insurance.ValidatorAddress] = validator
			} else {
				return false, nil
			}
		}
		pairingInsurances = append(pairingInsurances, insurance)
		return false, nil
	})
	if err != nil {
		return
	}
	if len(pairingInsurances) == 0 {
		err = types.ErrNoPairingInsurance
		return
	}

	// Liquid stakers can send amount of tokens corresponding multiple of chunk size and create multiple chunks
	// Check the liquid staker's balance
	n := amount.Amount.Quo(minimumRequirement.Amount).Int64()
	amount = sdk.NewCoin(bondDenom, types.ChunkSize.Mul(sdk.NewInt(n)))
	if !k.bankKeeper.HasBalance(ctx, delAddr, amount) {
		err = sdkerrors.Wrapf(sdkerrors.ErrInsufficientFunds, "insufficient funds to liquid stake")
		return
	}

	if n > int64(availableChunks) {
		n = int64(availableChunks)
		amount = sdk.NewCoin(bondDenom, types.ChunkSize.Mul(sdk.NewInt(n)))
	}

	// TODO: Must be proved that this is safe, we must call this before sending
	nas := k.GetNetAmountState(ctx)
	types.SortInsurances(validatorMap, pairingInsurances)
	totalNewShares := sdk.ZeroDec()
	totalLsTokenMintAmount := sdk.ZeroInt()
	for i := int64(0); i < n; i++ {
		// We can create paired chunk only with available pairing insurances
		if len(pairingInsurances) == 0 {
			break
		}
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
		chunk.InsuranceId = cheapestInsurance.Id
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

		// TODO: bond denom must be set at Genesis
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
