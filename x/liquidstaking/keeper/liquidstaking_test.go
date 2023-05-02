package keeper_test

import (
	"fmt"
	"math/rand"

	"github.com/Canto-Network/Canto/v6/x/liquidstaking/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/simulation"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	ethermint "github.com/evmos/ethermint/types"
)

func (suite *KeeperTestSuite) getPairedChunks() []types.Chunk {
	var pairedChunks []types.Chunk
	suite.app.LiquidStakingKeeper.IterateAllChunks(suite.ctx, func(chunk types.Chunk) (bool, error) {
		if chunk.Status == types.CHUNK_STATUS_PAIRED {
			pairedChunks = append(pairedChunks, chunk)
		}
		return false, nil
	})
	return pairedChunks
}

func (suite *KeeperTestSuite) getUnpairingForUnstakingChunks() []types.Chunk {
	var UnpairingForUnstakingChunks []types.Chunk
	suite.app.LiquidStakingKeeper.IterateAllChunks(suite.ctx, func(chunk types.Chunk) (bool, error) {
		if chunk.Status == types.CHUNK_STATUS_UNPAIRING_FOR_UNSTAKING {
			UnpairingForUnstakingChunks = append(UnpairingForUnstakingChunks, chunk)
		}
		return false, nil
	})
	return UnpairingForUnstakingChunks
}

// getMostExpensivePairedChunk returns the paired chunk which have most expensive insurance
func (suite *KeeperTestSuite) getMostExpensivePairedChunk(pairedChunks []types.Chunk) types.Chunk {
	chunksWithInsuranceId := make(map[uint64]types.Chunk)
	var insurances []types.Insurance
	validatorMap := make(map[string]stakingtypes.Validator)
	for _, chunk := range pairedChunks {
		insurance, _ := suite.app.LiquidStakingKeeper.GetInsurance(suite.ctx, chunk.PairedInsuranceId)
		if _, ok := validatorMap[insurance.ValidatorAddress]; !ok {
			validator, _ := suite.app.StakingKeeper.GetValidator(suite.ctx, insurance.GetValidator())
			validatorMap[insurance.ValidatorAddress] = validator
		}
		insurances = append(insurances, insurance)
		chunksWithInsuranceId[insurance.Id] = chunk
	}
	types.SortInsurances(validatorMap, insurances, true)
	return chunksWithInsuranceId[insurances[0].Id]
}

// Provide insurance with random fee (1 ~ 10%),
// if fixed fee is given, then use 10% as fee.
func (suite *KeeperTestSuite) provideInsurances(
	providers []sdk.AccAddress,
	valAddrs []sdk.ValAddress,
	amounts []sdk.Coin,
	fixedFeeRate sdk.Dec,
	feeRates []sdk.Dec,
) []types.Insurance {
	s := rand.NewSource(0)
	r := rand.New(s)

	valNum := len(valAddrs)
	var providedInsurances []types.Insurance
	for i, provider := range providers {
		msg := types.NewMsgProvideInsurance(provider.String(), amounts[i])
		msg.ValidatorAddress = valAddrs[i%valNum].String()
		if fixedFeeRate.IsPositive() {
			msg.FeeRate = fixedFeeRate
		} else if feeRates != nil && len(feeRates) > 0 {
			msg.FeeRate = feeRates[i]
		} else {
			// 1 ~ 10% insurance fee
			msg.FeeRate = sdk.NewDecWithPrec(int64(simulation.RandIntBetween(r, 1, 10)), 2)
		}
		msg.Amount = amounts[i]
		insurance, err := suite.app.LiquidStakingKeeper.DoProvideInsurance(suite.ctx, msg)
		suite.NoError(err)
		providedInsurances = append(providedInsurances, insurance)
	}
	return providedInsurances
}

func (suite *KeeperTestSuite) liquidStakes(delegators []sdk.AccAddress, amounts []sdk.Coin) []types.Chunk {
	var chunks []types.Chunk
	for i, delegator := range delegators {
		msg := types.NewMsgLiquidStake(delegator.String(), amounts[i])
		createdChunks, _, _, err := suite.app.LiquidStakingKeeper.DoLiquidStake(suite.ctx, msg)
		suite.NoError(err)
		for _, chunk := range createdChunks {
			chunks = append(chunks, chunk)
		}
	}
	return chunks
}

func (suite *KeeperTestSuite) TestProvideInsurance() {
	valAddrs := suite.CreateValidators([]int64{10, 10, 10})
	_, minimumCoverage := suite.app.LiquidStakingKeeper.GetMinimumRequirements(suite.ctx)
	providers, _ := suite.AddTestAddrs(10, minimumCoverage.Amount)

	for _, tc := range []struct {
		name        string
		msg         *types.MsgProvideInsurance
		validate    func(ctx sdk.Context, insurance types.Insurance)
		expectedErr string
	}{
		{
			"success",
			&types.MsgProvideInsurance{
				ProviderAddress:  providers[0].String(),
				ValidatorAddress: valAddrs[0].String(),
				Amount:           minimumCoverage,
				FeeRate:          sdk.ZeroDec(),
			},
			func(ctx sdk.Context, createdInsurance types.Insurance) {
				insurance, found := suite.app.LiquidStakingKeeper.GetInsurance(ctx, createdInsurance.Id)
				suite.True(found)
				suite.True(insurance.Equal(createdInsurance))
			},
			"",
		},
		{
			"insurance is smaller than minimum coverage",
			&types.MsgProvideInsurance{
				ProviderAddress:  providers[0].String(),
				ValidatorAddress: valAddrs[0].String(),
				Amount:           minimumCoverage.SubAmount(sdk.NewInt(1)),
				FeeRate:          sdk.Dec{},
			},
			nil,
			"amount must be greater than minimum coverage",
		},
	} {
		suite.Run(tc.name, func() {
			s.Require().NoError(tc.msg.ValidateBasic())
			cachedCtx, _ := s.ctx.CacheContext()
			insurance, err := suite.app.LiquidStakingKeeper.DoProvideInsurance(cachedCtx, tc.msg)
			if tc.expectedErr != "" {
				suite.ErrorContains(err, tc.expectedErr)
			} else {
				suite.NoError(err)
				tc.validate(cachedCtx, insurance)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestLiquidStakeSuccess() {
	suite.resetEpochs()
	valAddrs := suite.CreateValidators([]int64{10, 10, 10})
	oneChunk, oneInsurance := suite.app.LiquidStakingKeeper.GetMinimumRequirements(suite.ctx)
	providers, balances := suite.AddTestAddrs(10, oneInsurance.Amount)
	suite.provideInsurances(providers, valAddrs, balances, sdk.ZeroDec(), nil)

	delegators, balances := suite.AddTestAddrs(10, oneChunk.Amount)
	nas := suite.app.LiquidStakingKeeper.GetNetAmountState(suite.ctx)

	liquidBondDenom := suite.app.LiquidStakingKeeper.GetLiquidBondDenom(suite.ctx)
	// First try
	del1 := delegators[0]
	amt1 := balances[0]
	msg := types.NewMsgLiquidStake(del1.String(), amt1)
	lsTokenBefore := suite.app.BankKeeper.GetBalance(suite.ctx, del1, liquidBondDenom)
	createdChunks, newShares, lsTokenMintAmount, err := suite.app.LiquidStakingKeeper.DoLiquidStake(suite.ctx, msg)
	// Check created chunks are stored in db correctly
	idx := 0
	suite.NoError(suite.app.LiquidStakingKeeper.IterateAllChunks(suite.ctx, func(chunk types.Chunk) (bool, error) {
		suite.True(chunk.Equal(createdChunks[idx]))
		idx++
		return false, nil
	}))
	suite.Equal(len(createdChunks), idx, "number of created chunks should be equal to number of chunks in db")

	lsTokenAfter := suite.app.BankKeeper.GetBalance(suite.ctx, del1, liquidBondDenom)
	suite.NoError(err)
	suite.True(amt1.Amount.Equal(newShares.TruncateInt()), "delegation shares should be equal to amount")
	suite.True(amt1.Amount.Equal(lsTokenMintAmount), "at first try mint rate is 1, so mint amount should be equal to amount")
	suite.True(lsTokenAfter.Sub(lsTokenBefore).Amount.Equal(lsTokenMintAmount), "liquid staker must have minted ls tokens in account balance")

	// NetAmountState should be updated correctly
	afterNas := suite.app.LiquidStakingKeeper.GetNetAmountState(suite.ctx)
	suite.True(afterNas.LsTokensTotalSupply.Equal(lsTokenMintAmount), "total ls token supply should be equal to minted amount")
	suite.True(nas.TotalLiquidTokens.Add(amt1.Amount).Equal(afterNas.TotalLiquidTokens))
	suite.True(nas.NetAmount.Add(amt1.Amount.ToDec()).Equal(afterNas.NetAmount))
	suite.True(afterNas.MintRate.Equal(sdk.OneDec()), "no rewards yet, so mint rate should be 1")
}

func (suite *KeeperTestSuite) TestLiquidStakeFail() {
	suite.resetEpochs()
	valAddrs := suite.CreateValidators([]int64{10, 10, 10})
	oneChunk, oneInsurance := suite.app.LiquidStakingKeeper.GetMinimumRequirements(suite.ctx)

	addrs, balances := suite.AddTestAddrs(types.MaxPairedChunks-1, oneChunk.Amount)

	// TC: There are no pairing insurances yet. Insurances must be provided to liquid stake
	acc1 := addrs[0]
	msg := types.NewMsgLiquidStake(acc1.String(), oneChunk)
	_, _, _, err := suite.app.LiquidStakingKeeper.DoLiquidStake(suite.ctx, msg)
	suite.ErrorContains(err, types.ErrNoPairingInsurance.Error())

	providers, providerBalances := suite.AddTestAddrs(10, oneInsurance.Amount)
	suite.provideInsurances(providers, valAddrs, providerBalances, sdk.ZeroDec(), nil)

	// TC: Not enough amount to liquid stake
	// acc1 tries to liquid stake 2 * ChunkSize tokens, but he has only ChunkSize tokens
	msg = types.NewMsgLiquidStake(acc1.String(), oneChunk.AddAmount(types.ChunkSize))
	cachedCtx, writeCache := suite.ctx.CacheContext()
	_, _, _, err = suite.app.LiquidStakingKeeper.DoLiquidStake(cachedCtx, msg)
	if err == nil {
		writeCache()
	}
	suite.ErrorContains(err, sdkerrors.ErrInsufficientFunds.Error())

	msg.Amount.Denom = "unknown"
	_, _, _, err = suite.app.LiquidStakingKeeper.DoLiquidStake(suite.ctx, msg)
	suite.ErrorContains(err, types.ErrInvalidBondDenom.Error())
	msg.Amount.Denom = suite.denom

	// Pairs (MaxPairedChunks - 1) chunks, 1 chunk left now
	_ = suite.liquidStakes(addrs, balances)

	// Fund coins to acc1
	suite.fundAccount(acc1, types.ChunkSize.Mul(sdk.NewInt(2)))
	// Now acc1 have 2 * ChunkSize tokens as balance and try to liquid stake 2 * ChunkSize tokens
	acc1Balance := suite.app.BankKeeper.GetBalance(suite.ctx, acc1, suite.denom)
	suite.True(acc1Balance.Amount.Equal(types.ChunkSize.Mul(sdk.NewInt(2))))
	// TC: Enough to liquid stake 2 chunks, but current available chunk size is 1
	_, _, _, err = suite.app.LiquidStakingKeeper.DoLiquidStake(suite.ctx, msg)
	suite.ErrorContains(err, types.ErrExceedAvailableChunks.Error())

	// TC: amount must be multiple of chunk size
	oneTokenAmount := sdk.TokensFromConsensusPower(1, ethermint.PowerReduction)
	msg.Amount = msg.Amount.SubAmount(oneTokenAmount)
	_, _, _, err = suite.app.LiquidStakingKeeper.DoLiquidStake(suite.ctx, msg)
	suite.ErrorContains(err, types.ErrInvalidAmount.Error())
	msg.Amount = msg.Amount.AddAmount(oneTokenAmount)

	// liquid stake ChunkSize tokens so maximum chunk size is reached
	suite.liquidStakes([]sdk.AccAddress{acc1}, []sdk.Coin{oneChunk})

	// TC: MaxPairedChunks is reached, no more chunks can be paired
	newAddrs, newBalances := suite.AddTestAddrs(1, oneChunk.Amount)
	msg = types.NewMsgLiquidStake(newAddrs[0].String(), newBalances[0])
	_, _, _, err = suite.app.LiquidStakingKeeper.DoLiquidStake(suite.ctx, msg)
	suite.ErrorIs(err, types.ErrMaxPairedChunkSizeExceeded)
}

func (suite *KeeperTestSuite) TestLiquidStakeWithAdvanceBlocks() {
	// SETUP TEST ---------------------------------------------------
	suite.resetEpochs()
	// 3 validators we have
	valAddrs := suite.CreateValidators([]int64{1, 1, 1})
	oneChunk, oneInsurance := suite.app.LiquidStakingKeeper.GetMinimumRequirements(suite.ctx)
	providers, providerBalances := suite.AddTestAddrs(10, oneInsurance.Amount)
	fixedInsuranceFeeRate := sdk.NewDecWithPrec(10, 2)
	suite.provideInsurances(providers, valAddrs, providerBalances, fixedInsuranceFeeRate, nil)

	// 3 delegators
	delegators, delegatorBalances := suite.AddTestAddrs(3, oneChunk.Amount)
	nas := suite.app.LiquidStakingKeeper.GetNetAmountState(suite.ctx)
	suite.True(nas.IsZeroState(), "nothing happened yet so it must be zero state")

	// liquid stake 3 chunks (each delegator liquid stakes 1 chunk)
	pairedChunks := suite.liquidStakes(delegators, delegatorBalances)
	pairedChunksInt := sdk.NewInt(int64(len(pairedChunks)))

	bondDenom := suite.app.StakingKeeper.BondDenom(suite.ctx)
	liquidBondDenom := suite.app.LiquidStakingKeeper.GetLiquidBondDenom(suite.ctx)
	fmt.Printf(`
===============================================================================
Initial state of TC
- num of validators: %d
- fixed validator fee rate: %s
- num of delegators: %d
- num of insurances: %d
- fixed insurance fee rate: %s
- bonded denom: %s
- liquid bond denom: %s
===============================================================================
`,
		len(valAddrs),
		"10%",
		len(delegators),
		len(providers),
		fixedInsuranceFeeRate,
		bondDenom,
		liquidBondDenom,
	)
	// ---------------------------------------------------

	unitDelegationRewardPerEpoch, _ := sdk.NewIntFromString("29999994000000000000")
	unitInsuranceCommissionPerEpoch, pureUnitRewardPerEpoch := suite.getUnitDistribution(unitDelegationRewardPerEpoch, fixedInsuranceFeeRate)

	nas = suite.app.LiquidStakingKeeper.GetNetAmountState(suite.ctx)
	fmt.Println(nas)
	// 1 chunk size * number of paired chunks (=3) tokens are liquidated
	currentLiquidatedTokens := types.ChunkSize.Mul(pairedChunksInt)
	currentInsuranceTokens := oneInsurance.Amount.Mul(pairedChunksInt)
	{
		suite.True(nas.Equal(types.NetAmountState{
			MintRate:                           sdk.OneDec(),
			LsTokensTotalSupply:                currentLiquidatedTokens,
			NetAmount:                          currentLiquidatedTokens.ToDec(),
			TotalDelShares:                     currentLiquidatedTokens.ToDec(),
			TotalRemainingRewards:              sdk.ZeroDec(),
			TotalChunksBalance:                 sdk.ZeroInt(),
			TotalLiquidTokens:                  currentLiquidatedTokens,
			TotalInsuranceTokens:               oneInsurance.Amount.Mul(sdk.NewInt(int64(len(providers)))),
			TotalInsuranceCommissions:          sdk.ZeroInt(),
			TotalPairedInsuranceTokens:         currentInsuranceTokens,
			TotalPairedInsuranceCommissions:    sdk.ZeroInt(),
			TotalUnpairingInsuranceTokens:      sdk.ZeroInt(),
			TotalUnpairingInsuranceCommissions: sdk.ZeroInt(),
			TotalUnpairedInsuranceTokens:       sdk.ZeroInt(),
			TotalUnpairedInsuranceCommissions:  sdk.ZeroInt(),
			TotalUnbondingBalance:              sdk.ZeroInt(),
			RewardModuleAccBalance:             sdk.ZeroInt(),
		}), "no epoch(=1 block in test) processed yet, so there are no mint rate change and remaining rewards yet")
	}

	suite.advanceHeight(1, "")
	beforeNas := nas
	nas = suite.app.LiquidStakingKeeper.GetNetAmountState(suite.ctx)
	fmt.Println(nas)
	{
		suite.Equal(
			pureUnitRewardPerEpoch.Mul(pairedChunksInt).String(),
			nas.TotalRemainingRewards.Sub(beforeNas.TotalRemainingRewards).TruncateInt().String(),
		)
		suite.Equal("0.999994600030239830", nas.MintRate.String())
	}

	suite.advanceEpoch()
	suite.advanceHeight(1, "delegation reward are distributed to insurance and reward module")
	beforeNas = nas
	nas = suite.app.LiquidStakingKeeper.GetNetAmountState(suite.ctx)
	fmt.Println(nas)
	{
		suite.True(nas.TotalRemainingRewards.IsZero(), "remaining rewards are distributed")
		suite.Equal(
			pureUnitRewardPerEpoch.Mul(pairedChunksInt).Mul(sdk.NewInt(suite.rewardEpochCount)).String(),
			nas.RewardModuleAccBalance.String(),
		)
		suite.Equal(
			unitInsuranceCommissionPerEpoch.Mul(pairedChunksInt).Mul(sdk.NewInt(suite.rewardEpochCount)).String(),
			nas.TotalPairedInsuranceCommissions.String(),
		)
		suite.Equal("0.999989200118798693", nas.MintRate.String())
		suite.True(nas.MintRate.LT(beforeNas.MintRate))
	}
}

func (suite *KeeperTestSuite) TestLiquidUnstakeWithAdvanceBlocks() {
	// SETUP TEST ---------------------------------------------------
	suite.resetEpochs()
	// 3 validators we have
	valAddrs := suite.CreateValidators([]int64{1, 1, 1})
	oneChunk, oneInsurance := suite.app.LiquidStakingKeeper.GetMinimumRequirements(suite.ctx)
	providers, providerBalances := suite.AddTestAddrs(10, oneInsurance.Amount)
	fixedInsuranceFeeRate := sdk.NewDecWithPrec(10, 2)
	suite.provideInsurances(providers, valAddrs, providerBalances, fixedInsuranceFeeRate, nil)

	// 3 delegators
	delegators, delegatorBalances := suite.AddTestAddrs(3, oneChunk.Amount)
	nas := suite.app.LiquidStakingKeeper.GetNetAmountState(suite.ctx)
	suite.True(nas.IsZeroState(), "nothing happened yet so it must be zero state")
	undelegator := delegators[0]

	// liquid stake 3 chunks (each delegator liquid stakes 1 chunk)
	pairedChunks := suite.liquidStakes(delegators, delegatorBalances)
	pairedChunksInt := sdk.NewInt(int64(len(pairedChunks)))
	mostExpensivePairedChunk := suite.getMostExpensivePairedChunk(pairedChunks)

	bondDenom := suite.app.StakingKeeper.BondDenom(suite.ctx)
	liquidBondDenom := suite.app.LiquidStakingKeeper.GetLiquidBondDenom(suite.ctx)
	fmt.Printf(`
===============================================================================
Initial state of TC
- num of validators: %d
- fixed validator fee rate: %s
- num of delegators: %d
- num of insurances: %d
- fixed insurance fee rate: %s
- bonded denom: %s
- liquid bond denom: %s
===============================================================================
`,
		len(valAddrs),
		"10%",
		len(delegators),
		len(providers),
		fixedInsuranceFeeRate,
		bondDenom,
		liquidBondDenom,
	)
	// ---------------------------------------------------

	nas = suite.app.LiquidStakingKeeper.GetNetAmountState(suite.ctx)
	fmt.Println(nas)
	// 1 chunk size * number of paired chunks (=3) tokens are liquidated
	currentLiquidatedTokens := types.ChunkSize.Mul(pairedChunksInt)
	currentInsuranceTokens := oneInsurance.Amount.Mul(pairedChunksInt)
	{
		suite.True(nas.Equal(types.NetAmountState{
			MintRate:                           sdk.OneDec(),
			LsTokensTotalSupply:                currentLiquidatedTokens,
			NetAmount:                          currentLiquidatedTokens.ToDec(),
			TotalDelShares:                     currentLiquidatedTokens.ToDec(),
			TotalRemainingRewards:              sdk.ZeroDec(),
			TotalChunksBalance:                 sdk.ZeroInt(),
			TotalLiquidTokens:                  currentLiquidatedTokens,
			TotalInsuranceTokens:               oneInsurance.Amount.Mul(sdk.NewInt(int64(len(providers)))),
			TotalInsuranceCommissions:          sdk.ZeroInt(),
			TotalPairedInsuranceTokens:         currentInsuranceTokens,
			TotalPairedInsuranceCommissions:    sdk.ZeroInt(),
			TotalUnpairingInsuranceTokens:      sdk.ZeroInt(),
			TotalUnpairingInsuranceCommissions: sdk.ZeroInt(),
			TotalUnpairedInsuranceTokens:       sdk.ZeroInt(),
			TotalUnpairedInsuranceCommissions:  sdk.ZeroInt(),
			TotalUnbondingBalance:              sdk.ZeroInt(),
			RewardModuleAccBalance:             sdk.ZeroInt(),
		}), "no epoch(=1 block in test) processed yet, so there are no mint rate change and remaining rewards yet")
	}
	// advance 1 block(= epoch period in test environment) so reward is accumulated which means mint rate is changed
	suite.advanceHeight(1, "")

	unitDelegationRewardPerEpoch, _ := sdk.NewIntFromString("29999994000000000000")
	unitInsuranceCommissionPerEpoch, pureUnitRewardPerEpoch := suite.getUnitDistribution(unitDelegationRewardPerEpoch, fixedInsuranceFeeRate)

	// each delegation reward per epoch(=1 block in test) * number of paired chunks
	// = 29999994000000000000 * 3
	notClaimedRewards := pureUnitRewardPerEpoch.Mul(pairedChunksInt)
	beforeNas := nas
	nas = suite.app.LiquidStakingKeeper.GetNetAmountState(suite.ctx)
	fmt.Println(nas)
	{
		suite.Equal(
			notClaimedRewards.ToDec(),
			nas.TotalRemainingRewards.Sub(beforeNas.TotalRemainingRewards),
			"one epoch(=1 block in test) passed, so remaining rewards must be increased",
		)
		suite.Equal(notClaimedRewards.ToDec(), nas.NetAmount.Sub(beforeNas.NetAmount), "net amount must be increased by not claimed rewards")
		suite.Equal("0.999994600030239830", nas.MintRate.String(), "mint rate increased because of reward accumulation")
	}

	// Queue liquid unstake 1 chunk
	fmt.Println("Queue liquid unstake 1 chunk")
	beforeBondDenomBalance := suite.app.BankKeeper.GetBalance(suite.ctx, undelegator, bondDenom)
	beforeLiquidBondDenomBalance := suite.app.BankKeeper.GetBalance(suite.ctx, undelegator, liquidBondDenom)
	msg := types.NewMsgLiquidUnstake(undelegator.String(), oneChunk)
	lsTokensToEscrow := nas.MintRate.Mul(oneChunk.Amount.ToDec()).TruncateInt()
	toBeUnstakedChunks, pendingLiquidUnstakes, err := suite.app.LiquidStakingKeeper.QueueLiquidUnstake(suite.ctx, msg)
	{
		suite.NoError(err)
		suite.Equal(1, len(toBeUnstakedChunks), "we just queued liuquid unstaking for 1 chunk")
		suite.Equal(1, len(pendingLiquidUnstakes), "we just queued liuquid unstaking for 1 chunk")
		suite.Equal(toBeUnstakedChunks[0].Id, pendingLiquidUnstakes[0].ChunkId)
		suite.Equal(undelegator.String(), pendingLiquidUnstakes[0].DelegatorAddress)
		suite.Equal(
			mostExpensivePairedChunk.PairedInsuranceId,
			toBeUnstakedChunks[0].PairedInsuranceId,
			"queued chunk must have the most expensive insurance paired with the previously paired chunk",
		)
		// Check if the liquid unstaker escrowed ls tokens
		bondDenomBalance := suite.app.BankKeeper.GetBalance(suite.ctx, undelegator, bondDenom)
		liquidBondDenomBalance := suite.app.BankKeeper.GetBalance(suite.ctx, undelegator, liquidBondDenom)
		suite.Equal(sdk.ZeroInt(), bondDenomBalance.Sub(beforeBondDenomBalance).Amount, "unbonding period is just started so no tokens are backed yet")
		suite.Equal(
			lsTokensToEscrow,
			beforeLiquidBondDenomBalance.Sub(liquidBondDenomBalance).Amount,
			"ls tokens are escrowed by module",
		)
		suite.Equal(
			lsTokensToEscrow,
			suite.app.BankKeeper.GetBalance(suite.ctx, types.LsTokenEscrowAcc, liquidBondDenom).Amount,
			"module got ls tokens from liquid unstaker",
		)
	}

	// The actual unstaking started in this epoch
	suite.advanceEpoch()
	suite.advanceHeight(1, "The actual unstaking started\nThe insurance commission and reward are claimed")
	beforeNas = nas
	nas = suite.app.LiquidStakingKeeper.GetNetAmountState(suite.ctx)
	fmt.Println(nas)

	// Check NetAmounState changed right
	{
		suite.Equal(
			beforeNas.TotalDelShares.Sub(nas.TotalDelShares).TruncateInt().String(),
			oneChunk.Amount.String(),
			"unstaking 1 chunk is started which means undelegate is already triggered so total del shares must be decreased by 1 chunk amount",
		)
		suite.Equal(
			nas.LsTokensTotalSupply.String(),
			beforeNas.LsTokensTotalSupply.String(),
			"unstaking is not finished so ls tokens total supply must not be changed",
		)
		suite.Equal(
			nas.TotalUnbondingBalance.String(),
			oneChunk.Amount.String(),
			"unstaking 1 chunk is started which means undelegate is already triggered",
		)
		suite.True(nas.TotalRemainingRewards.IsZero(), "all rewards are claimed")
		suite.Equal(
			pureUnitRewardPerEpoch.Mul(pairedChunksInt).Mul(sdk.NewInt(suite.rewardEpochCount)).String(),
			nas.RewardModuleAccBalance.String(),
			fmt.Sprintf("before unstaking triggered there are collecting reward process so reward module got %d chunk's rewards", pairedChunksInt.Int64()),
		)
		suite.Equal(
			unitInsuranceCommissionPerEpoch.Mul(sdk.NewInt(suite.rewardEpochCount)).String(),
			nas.TotalUnpairingInsuranceCommissions.String(),
		)
		suite.Equal(
			unitInsuranceCommissionPerEpoch.Mul(sdk.NewInt(suite.rewardEpochCount).Mul(sdk.NewInt(2))).String(),
			nas.TotalPairedInsuranceCommissions.Sub(beforeNas.TotalPairedInsuranceCommissions).String(),
		)
		suite.Equal(
			oneInsurance.Amount.String(),
			nas.TotalUnpairingInsuranceTokens.Sub(beforeNas.TotalUnpairingInsuranceTokens).String(),
			"",
		)
		suite.Equal(
			unitInsuranceCommissionPerEpoch.Mul(sdk.NewInt(suite.rewardEpochCount)).String(),
			nas.TotalUnpairingInsuranceCommissions.Sub(beforeNas.TotalUnpairingInsuranceCommissions).String(),
			"TotalUnpairingInsuranceTokens must be increased by insurance commission per epoch",
		)
		suite.True(nas.MintRate.LT(beforeNas.MintRate), "mint rate decreased because of reward is accumulated")
	}

	// After epoch reached, toBeUnstakedChunks should be unstaked
	unstakedChunk, found := suite.app.LiquidStakingKeeper.GetChunk(suite.ctx, toBeUnstakedChunks[0].Id)
	// Check status of chunks
	{
		suite.True(found)
		suite.Equal(unstakedChunk.Status, types.CHUNK_STATUS_UNPAIRING_FOR_UNSTAKING)
		suite.Equal(unstakedChunk.UnpairingInsuranceId, toBeUnstakedChunks[0].PairedInsuranceId)
	}
	// check states after liquid unstake
	pairedChunksAfterUnstake := suite.getPairedChunks()
	// check UnpairingForUnstaking chunks
	UnpairingForUnstakingChunks := suite.getUnpairingForUnstakingChunks()
	// paired chunk count should be decreased by number of unstaked chunks
	suite.Equal(len(pairedChunks)-len(UnpairingForUnstakingChunks), len(pairedChunksAfterUnstake))
	pairedChunksInt = sdk.NewInt(int64(len(pairedChunksAfterUnstake)))

	suite.advanceEpoch()
	suite.advanceHeight(1, "The insurance commission and reward are claimed\nThe unstaking is completed")

	// Now number of paired chunk is decreased and still reward is fixed,
	// so the unit reward per epoch is increased.
	unitDelegationRewardPerEpoch, _ = sdk.NewIntFromString("44999986500000000000")
	unitInsuranceCommissionPerEpoch, pureUnitRewardPerEpoch = suite.getUnitDistribution(unitDelegationRewardPerEpoch, fixedInsuranceFeeRate)

	beforeNas = nas
	nas = suite.app.LiquidStakingKeeper.GetNetAmountState(suite.ctx)
	fmt.Println(nas)
	afterBondDenomBalance := suite.app.BankKeeper.GetBalance(suite.ctx, undelegator, bondDenom).Amount
	// Get bondDeno balance of undelegator
	{
		suite.Equal(
			oneInsurance.Amount.String(),
			nas.TotalUnpairedInsuranceTokens.Sub(beforeNas.TotalUnpairedInsuranceTokens).String(),
			"unstkaing 1 chunk is finished so the insurance is released",
		)
		suite.Equal(beforeNas.TotalDelShares.String(), nas.TotalDelShares.String())
		suite.Equal(beforeNas.TotalLiquidTokens.String(), nas.TotalLiquidTokens.String())
		suite.Equal(
			beforeNas.TotalUnbondingBalance.Sub(oneChunk.Amount).String(),
			nas.TotalUnbondingBalance.String(),
			"unstaking(=unbonding) is finished",
		)
		suite.True(nas.LsTokensTotalSupply.LT(beforeNas.LsTokensTotalSupply), "ls tokens are burned")
		suite.True(nas.TotalRemainingRewards.IsZero(), "all rewards are claimed")
		suite.Equal(
			pureUnitRewardPerEpoch.Mul(pairedChunksInt).String(),
			nas.RewardModuleAccBalance.Sub(beforeNas.RewardModuleAccBalance).String(),
			"reward module account balance must be increased by pure reward per epoch * reward epoch count",
		)
		suite.Equal(
			unitInsuranceCommissionPerEpoch.Mul(pairedChunksInt).String(),
			nas.TotalPairedInsuranceCommissions.Sub(beforeNas.TotalPairedInsuranceCommissions).String(),
		)
		suite.Equal(
			afterBondDenomBalance.String(),
			oneChunk.Amount.String(),
			"got chunk tokens back after unstaking",
		)
	}
}

func (suite *KeeperTestSuite) TestQueueLiquidUnstakeFail() {
	valAddrs := suite.CreateValidators([]int64{10, 10, 10})
	minimumRequirement, minimumCoverage := suite.app.LiquidStakingKeeper.GetMinimumRequirements(suite.ctx)
	providers, providerBalances := suite.AddTestAddrs(10, minimumCoverage.Amount)
	suite.provideInsurances(providers, valAddrs, providerBalances, sdk.ZeroDec(), nil)

	// Now we have 1 paired chunks
	delegators, delegatorBalances := suite.AddTestAddrs(3, minimumRequirement.Amount)
	undelegator := delegators[0]

	msg := types.NewMsgLiquidUnstake(
		undelegator.String(),
		minimumRequirement,
	)
	_, _, err := suite.app.LiquidStakingKeeper.QueueLiquidUnstake(suite.ctx, msg)
	suite.ErrorContains(err, types.ErrNoPairedChunk.Error())

	// create one paired chunk
	_ = suite.liquidStakes([]sdk.AccAddress{delegators[0]}, []sdk.Coin{delegatorBalances[0]})

	msg.Amount.Amount = msg.Amount.Amount.Sub(sdk.NewInt(1))
	// TC: Must be multiple of chunk size
	_, _, err = suite.app.LiquidStakingKeeper.QueueLiquidUnstake(suite.ctx, msg)
	suite.ErrorContains(err, types.ErrInvalidAmount.Error())
	msg.Amount = msg.Amount.Add(sdk.NewCoin(suite.denom, sdk.OneInt())) // now amount is valid

	// TC: Must be bond denom
	msg.Amount.Denom = "invalid"
	_, _, err = suite.app.LiquidStakingKeeper.QueueLiquidUnstake(suite.ctx, msg)
	suite.ErrorContains(err, types.ErrInvalidBondDenom.Error())
	msg.Amount.Denom = suite.denom // now denom is valid

	// TC: Want to liquid unstake 2 chunks but current paired chunk is only one
	msg.Amount.Amount = minimumRequirement.Amount.Mul(sdk.NewInt(2))
	_, _, err = suite.app.LiquidStakingKeeper.QueueLiquidUnstake(suite.ctx, msg)
	suite.ErrorContains(err, types.ErrExceedAvailableChunks.Error())

	// Now we have 3 paired chunks
	_ = suite.liquidStakes(delegators[1:], delegatorBalances[1:])

	// TC: Want to liquid unstake 2 chunks but unstaker have lstokens corresponding to 1 chunk size
	_, _, err = suite.app.LiquidStakingKeeper.QueueLiquidUnstake(suite.ctx, msg)
	suite.ErrorContains(err, sdkerrors.ErrInsufficientFunds.Error())
}

func (suite *KeeperTestSuite) TestCancelProvideInsuranceSuccess() {
	valAddrs := suite.CreateValidators([]int64{10, 10, 10})
	_, minimumCoverage := suite.app.LiquidStakingKeeper.GetMinimumRequirements(suite.ctx)
	providers, balances := suite.AddTestAddrs(10, minimumCoverage.Amount)
	insurances := suite.provideInsurances(providers, valAddrs, balances, sdk.ZeroDec(), nil)

	provider := providers[0]
	insurance := insurances[0]
	escrowed := suite.app.BankKeeper.GetBalance(suite.ctx, insurance.DerivedAddress(), suite.denom)
	beforeProviderBalance := suite.app.BankKeeper.GetBalance(suite.ctx, provider, suite.denom)
	msg := types.NewMsgCancelProvideInsurance(provider.String(), insurance.Id)
	canceledInsurance, err := suite.app.LiquidStakingKeeper.DoCancelProvideInsurance(suite.ctx, msg)
	suite.NoError(err)
	suite.True(insurance.Equal(canceledInsurance))
	afterProviderBalance := suite.app.BankKeeper.GetBalance(suite.ctx, provider, suite.denom)
	suite.True(afterProviderBalance.Amount.Equal(beforeProviderBalance.Amount.Add(escrowed.Amount)), "provider should get back escrowed amount")
}

func (suite *KeeperTestSuite) TestDoCancelProvideInsuranceFail() {
	suite.resetEpochs()
	// create valAddrs
	valAddrs := suite.CreateValidators([]int64{1, 1, 1})
	oneChunk, oneInsurance := suite.app.LiquidStakingKeeper.GetMinimumRequirements(suite.ctx)
	// create providers
	providers, providerBalances := suite.AddTestAddrs(3, oneInsurance.Amount.Add(sdk.NewInt(100)))
	// provide insurances
	insurances := suite.provideInsurances(providers, valAddrs, providerBalances, sdk.NewDecWithPrec(10, 2), nil)
	delegators, delegatorBalances := suite.AddTestAddrs(1, oneChunk.Amount)
	// liquid stake
	suite.liquidStakes(delegators, delegatorBalances)
	onlyPairedInsurance := insurances[0]

	tcs := []struct {
		name        string
		msg         *types.MsgCancelProvideInsurance
		expectedErr error
	}{
		{
			name: "invalid provider",
			msg: types.NewMsgCancelProvideInsurance(
				providers[1].String(),
				insurances[2].Id,
			),
			expectedErr: types.ErrNotProviderOfInsurance,
		},
		{
			name: "invalid insurance id",
			msg: types.NewMsgCancelProvideInsurance(
				providers[1].String(),
				120,
			),
			expectedErr: types.ErrNotFoundInsurance,
		},
		{
			name: "this is no pairing insurance",
			msg: types.NewMsgCancelProvideInsurance(
				onlyPairedInsurance.ProviderAddress,
				onlyPairedInsurance.Id,
			),
			expectedErr: types.ErrInvalidInsuranceStatus,
		},
	}

	for _, tc := range tcs {
		_, err := suite.app.LiquidStakingKeeper.DoCancelProvideInsurance(suite.ctx, tc.msg)
		if tc.expectedErr == nil {
			suite.NoError(err)
		}
		suite.ErrorContains(err, tc.expectedErr.Error())
	}
}

func (suite *KeeperTestSuite) TestDoWithdrawInsurance() {
	// SETUP TEST ---------------------------------------------------
	suite.resetEpochs()
	// 3 validators we have
	valAddrs := suite.CreateValidators([]int64{10, 10, 10})
	oneChunk, oneInsurance := suite.app.LiquidStakingKeeper.GetMinimumRequirements(suite.ctx)
	providers, providerBalances := suite.AddTestAddrs(3, oneInsurance.Amount)
	// 3 insurances (insurance fee rates are all same as 10%)
	insurances := suite.provideInsurances(providers, valAddrs, providerBalances, sdk.NewDecWithPrec(10, 2), nil)
	var idsOfPairedInsurances []uint64
	for _, insurance := range insurances {
		idsOfPairedInsurances = append(idsOfPairedInsurances, insurance.Id)
	}
	// 3 delegators
	delegators, delegatorBalances := suite.AddTestAddrs(3, oneChunk.Amount)
	// liquid stakes 3 chunks
	suite.liquidStakes(delegators, delegatorBalances)
	// ---------------------------------------------------

	toBeWithdrawnInsurance, _ := suite.app.LiquidStakingKeeper.GetInsurance(suite.ctx, insurances[0].Id)
	_, err := suite.app.LiquidStakingKeeper.DoWithdrawInsurance(
		suite.ctx,
		types.NewMsgWithdrawInsurance(
			toBeWithdrawnInsurance.ProviderAddress,
			toBeWithdrawnInsurance.Id,
		),
	)
	suite.NoError(err)
	suite.advanceEpoch()
	suite.advanceHeight(1, "queued withdraw insurance request is handled and there are no additional insurances yet so unpairing triggered")

	suite.advanceHeight(1, "")

	suite.advanceEpoch()
	suite.advanceHeight(1, "unpairing is done")

	unpairedInsurance, _ := suite.app.LiquidStakingKeeper.GetInsurance(suite.ctx, insurances[0].Id)
	suite.Equal(types.INSURANCE_STATUS_UNPAIRED, unpairedInsurance.Status)

	beforeProviderBalance := suite.app.BankKeeper.GetBalance(suite.ctx, unpairedInsurance.GetProvider(), suite.denom)
	unpairedInsuranceBalance := suite.app.BankKeeper.GetBalance(suite.ctx, unpairedInsurance.DerivedAddress(), suite.denom)
	unpairedInsuranceCommission := suite.app.BankKeeper.GetBalance(suite.ctx, unpairedInsurance.FeePoolAddress(), suite.denom)
	_, err = suite.app.LiquidStakingKeeper.DoWithdrawInsurance(
		suite.ctx,
		types.NewMsgWithdrawInsurance(
			unpairedInsurance.ProviderAddress,
			unpairedInsurance.Id,
		),
	)
	suite.NoError(err)
	afterProviderBalance := suite.app.BankKeeper.GetBalance(suite.ctx, unpairedInsurance.GetProvider(), suite.denom)
	suite.Equal(
		beforeProviderBalance.Amount.Add(unpairedInsuranceBalance.Amount).Add(unpairedInsuranceCommission.Amount).String(),
		afterProviderBalance.Amount.String(),
	)
}

func (suite *KeeperTestSuite) TestDoWithdrawInsuranceFail() {
	suite.resetEpochs()
	// create valAddrs
	valAddrs := suite.CreateValidators([]int64{1, 1, 1})
	_, oneInsurance := suite.app.LiquidStakingKeeper.GetMinimumRequirements(suite.ctx)
	// create providers
	providers, providerBalances := suite.AddTestAddrs(3, oneInsurance.Amount.Add(sdk.NewInt(100)))
	// provide insurances
	insurances := suite.provideInsurances(providers, valAddrs, providerBalances, sdk.NewDecWithPrec(10, 2), nil)
	// withdraw insurances[0]
	suite.app.LiquidStakingKeeper.DoWithdrawInsurance(
		suite.ctx,
		types.NewMsgWithdrawInsurance(
			insurances[0].ProviderAddress,
			insurances[0].Id,
		),
	)
	suite.advanceEpoch()
	suite.advanceHeight(1, "insurance enters into UnpairingForWithdrawal status")

	tcs := []struct {
		name        string
		msg         *types.MsgWithdrawInsurance
		expectedErr error
	}{
		{
			name: "invalid provider",
			msg: types.NewMsgWithdrawInsurance(
				providers[1].String(),
				insurances[0].Id,
			),
			expectedErr: types.ErrNotProviderOfInsurance,
		},
		{
			name: "invalid insurance id",
			msg: types.NewMsgWithdrawInsurance(
				providers[0].String(),
				120,
			),
			expectedErr: types.ErrNotFoundInsurance,
		},
		{
			name: "invalid insurance status",
			msg: types.NewMsgWithdrawInsurance(
				providers[0].String(),
				insurances[0].Id,
			),
			expectedErr: types.ErrNotInWithdrawableStatus,
		},
	}

	for _, tc := range tcs {
		_, err := suite.app.LiquidStakingKeeper.DoWithdrawInsurance(suite.ctx, tc.msg)
		if tc.expectedErr == nil {
			suite.NoError(err)
		}
		suite.ErrorContains(err, tc.expectedErr.Error())
	}
}

func (suite *KeeperTestSuite) TestDoWithdrawInsuranceCommission() {
	suite.resetEpochs()
	valAddrs := suite.CreateValidators([]int64{1, 1, 1})
	oneChunk, oneInsurance := suite.app.LiquidStakingKeeper.GetMinimumRequirements(suite.ctx)
	providers, providerBalances := suite.AddTestAddrs(3, oneInsurance.Amount)
	// 3 insurances (insurance fee rates are all same as 10%)
	fixedInsuranceFeeRate := sdk.NewDecWithPrec(10, 2)
	insurances := suite.provideInsurances(providers, valAddrs, providerBalances, fixedInsuranceFeeRate, nil)
	// 3 delegators
	delegators, delegatorBalances := suite.AddTestAddrs(3, oneChunk.Amount)
	// liquid stakes 3 chunks
	suite.liquidStakes(delegators, delegatorBalances)

	unitDelegationRewardPerEpoch, _ := sdk.NewIntFromString("29999994000000000000")
	// unitInsuranceCommissionPerEpoch, _ := suite.getUnitDistribution(unitDelegationRewardPerEpoch, fixedInsuranceFeeRate)
	unitInsuranceCommissionPerEpoch, _ := suite.getUnitDistribution(unitDelegationRewardPerEpoch, fixedInsuranceFeeRate)

	provider := providers[0]
	targetInsurance := insurances[0]
	beforeInsuranceCommission := suite.app.BankKeeper.GetBalance(suite.ctx, targetInsurance.FeePoolAddress(), suite.denom)
	suite.advanceHeight(1, "")
	afterInsuranceCommission := suite.app.BankKeeper.GetBalance(suite.ctx, targetInsurance.FeePoolAddress(), suite.denom)
	suite.Equal(
		afterInsuranceCommission.String(),
		beforeInsuranceCommission.String(),
		"epoch is not reached yet so no insurance commission is distributed",
	)

	suite.advanceEpoch()
	suite.advanceHeight(1, "cumulated delegation reward is distributed to withdraw fee pool")
	afterInsuranceCommission = suite.app.BankKeeper.GetBalance(suite.ctx, targetInsurance.FeePoolAddress(), suite.denom)
	suite.Equal(
		unitInsuranceCommissionPerEpoch.Mul(sdk.NewInt(suite.rewardEpochCount)).String(),
		afterInsuranceCommission.Amount.String(),
		"cumulated delegation reward is distributed to withdraw fee pool",
	)

	beforeProviderBalance := suite.app.BankKeeper.GetBalance(suite.ctx, provider, suite.denom)
	// withdraw insurance commission
	err := suite.app.LiquidStakingKeeper.DoWithdrawInsuranceCommission(
		suite.ctx,
		types.NewMsgWithdrawInsuranceCommission(
			targetInsurance.ProviderAddress,
			targetInsurance.Id,
		),
	)
	suite.NoError(err)
	afterProviderBalance := suite.app.BankKeeper.GetBalance(suite.ctx, provider, suite.denom)
	suite.Equal(
		afterInsuranceCommission.String(),
		afterProviderBalance.Sub(beforeProviderBalance).String(),
		"provider did withdraw insurance commission",
	)
}

func (suite *KeeperTestSuite) TestDoWithdrawInsuranceCommissionFail() {
	// create valAddrs
	valAddrs := suite.CreateValidators([]int64{1, 1, 1})
	_, oneInsurance := suite.app.LiquidStakingKeeper.GetMinimumRequirements(suite.ctx)
	// create providers
	providers, providerBalances := suite.AddTestAddrs(3, oneInsurance.Amount.Add(sdk.NewInt(100)))
	// provide insurances
	insurances := suite.provideInsurances(providers, valAddrs, providerBalances, sdk.NewDecWithPrec(10, 2), nil)

	tcs := []struct {
		name        string
		msg         *types.MsgWithdrawInsuranceCommission
		expectedErr error
	}{
		{
			name: "invalid provider",
			msg: types.NewMsgWithdrawInsuranceCommission(
				providers[1].String(),
				insurances[0].Id,
			),
			expectedErr: types.ErrNotProviderOfInsurance,
		},
		{
			name: "invalid insurance id",
			msg: types.NewMsgWithdrawInsuranceCommission(
				providers[0].String(),
				120,
			),
			expectedErr: types.ErrNotFoundInsurance,
		},
	}

	for _, tc := range tcs {
		err := suite.app.LiquidStakingKeeper.DoWithdrawInsuranceCommission(suite.ctx, tc.msg)
		if tc.expectedErr == nil {
			suite.NoError(err)
		}
		suite.ErrorContains(err, tc.expectedErr.Error())
	}
}

func (suite *KeeperTestSuite) TestDoDepositInsurance() {
	// create validators
	validators := suite.CreateValidators([]int64{1, 1, 1})
	_, oneInsurance := suite.app.LiquidStakingKeeper.GetMinimumRequirements(suite.ctx)
	// create providers
	providers, _ := suite.AddTestAddrs(3, oneInsurance.Amount.Add(sdk.NewInt(100)))
	// provide insurances
	insurances := suite.provideInsurances(
		providers,
		validators,
		[]sdk.Coin{oneInsurance, oneInsurance, oneInsurance},
		sdk.NewDecWithPrec(10, 2),
		nil,
	)
	// all providers still have 100 acanto after provide insurance

	msgDepositInsurance := types.NewMsgDepositInsurance(
		providers[0].String(),
		insurances[0].Id,
		sdk.NewCoin(oneInsurance.Denom, sdk.NewInt(100)),
	)

	err := suite.app.LiquidStakingKeeper.DoDepositInsurance(suite.ctx, msgDepositInsurance)
	suite.NoError(err)
}

func (suite *KeeperTestSuite) TestDoDepositInsuranceFail() {
	// create valAddrs
	valAddrs := suite.CreateValidators([]int64{1, 1, 1})
	_, oneInsurance := suite.app.LiquidStakingKeeper.GetMinimumRequirements(suite.ctx)
	// create providers
	providers, providerBalances := suite.AddTestAddrs(3, oneInsurance.Amount.Add(sdk.NewInt(100)))
	// provide insurances
	insurances := suite.provideInsurances(providers, valAddrs, providerBalances, sdk.NewDecWithPrec(10, 2), nil)

	tcs := []struct {
		name        string
		msg         *types.MsgDepositInsurance
		expectedErr error
	}{
		{
			name: "invalid provider",
			msg: types.NewMsgDepositInsurance(
				providers[1].String(),
				insurances[0].Id,
				sdk.NewCoin(oneInsurance.Denom, sdk.NewInt(100)),
			),
			expectedErr: types.ErrNotProviderOfInsurance,
		},
		{
			name: "invalid insurance id",
			msg: types.NewMsgDepositInsurance(
				providers[0].String(),
				120,
				sdk.NewCoin(oneInsurance.Denom, sdk.NewInt(100)),
			),
			expectedErr: types.ErrNotFoundInsurance,
		},
		{
			name: "invalid insurance denom",
			msg: types.NewMsgDepositInsurance(
				providers[0].String(),
				insurances[0].Id,
				sdk.NewCoin("invalidDenom", sdk.NewInt(100)),
			),
			expectedErr: types.ErrInvalidBondDenom,
		},
	}

	for _, tc := range tcs {
		err := suite.app.LiquidStakingKeeper.DoDepositInsurance(suite.ctx, tc.msg)
		if tc.expectedErr == nil {
			suite.NoError(err)
		}
		suite.ErrorContains(err, tc.expectedErr.Error())
	}
}

func (suite *KeeperTestSuite) TestRankInsurances() {
	// SETUP TEST ---------------------------------------------------
	suite.resetEpochs()
	// 3 validators we have
	valAddrs := suite.CreateValidators([]int64{1, 1, 1})
	oneChunk, oneInsurance := suite.app.LiquidStakingKeeper.GetMinimumRequirements(suite.ctx)
	providers, providerBalances := suite.AddTestAddrs(3, oneInsurance.Amount)
	// 3 insurances (insurance fee rates are all same as 10%)
	insurances := suite.provideInsurances(providers, valAddrs, providerBalances, sdk.NewDecWithPrec(10, 2), nil)
	var idsOfPairedInsurances []uint64
	for _, insurance := range insurances {
		idsOfPairedInsurances = append(idsOfPairedInsurances, insurance.Id)
	}
	// 3 delegators
	delegators, delegatorBalances := suite.AddTestAddrs(3, oneChunk.Amount)
	// liquid stakes 3 chunks
	suite.liquidStakes(delegators, delegatorBalances)
	// ---------------------------------------------------

	// INITIAL STATE: all paired chunks are working fine and there are no additional insurances yet
	newlyRankedInInsurances, rankOutInsurances, err := suite.app.LiquidStakingKeeper.RankInsurances(suite.ctx)
	suite.NoError(err)
	suite.Len(newlyRankedInInsurances, 0)
	suite.Len(rankOutInsurances, 0)

	suite.advanceHeight(1, "")

	// Cheap insurances which are competitive than current paired insurances are provided
	otherProviders, otherProviderBalances := suite.AddTestAddrs(3, oneInsurance.Amount)
	newInsurances := suite.provideInsurances(
		otherProviders,
		valAddrs,
		otherProviderBalances,
		sdk.ZeroDec(),
		// fee rates(1~3%) of new insurances are all lower than current paired insurances (10%)
		[]sdk.Dec{sdk.NewDecWithPrec(1, 2), sdk.NewDecWithPrec(2, 2), sdk.NewDecWithPrec(3, 2)},
	)
	var idsOfNewInsurances []uint64
	for _, insurance := range newInsurances {
		idsOfNewInsurances = append(idsOfNewInsurances, insurance.Id)
	}

	newlyRankedInInsurances, rankOutInsurances, err = suite.app.LiquidStakingKeeper.RankInsurances(suite.ctx)
	suite.NoError(err)
	suite.Len(newlyRankedInInsurances, 3)
	suite.Len(rankOutInsurances, 3)
	// make sure idsOfNewInsurances are all in newlyRankedInInsurances
	for _, id := range idsOfNewInsurances {
		found := false
		for _, newlyRankedInInsurance := range newlyRankedInInsurances {
			if newlyRankedInInsurance.Id == id {
				found = true
				break
			}
		}
		suite.True(found)
	}
	// make sure idsOfPairedInsurances are all in rankOutInsurances
	for _, id := range idsOfPairedInsurances {
		found := false
		for _, rankOutInsurance := range rankOutInsurances {
			if rankOutInsurance.Id == id {
				found = true
				break
			}
		}
		suite.True(found)
	}
}

func (suite *KeeperTestSuite) TestEndBlocker() {
	// SETUP TEST ---------------------------------------------------
	suite.resetEpochs()
	// 3 validators we have
	valAddrs := suite.CreateValidators([]int64{1, 1, 1})
	oneChunk, oneInsurance := suite.app.LiquidStakingKeeper.GetMinimumRequirements(suite.ctx)
	providers, providerBalances := suite.AddTestAddrs(3, oneInsurance.Amount)
	// 3 insurances (insurance fee rates are all same as 10%)
	insurances := suite.provideInsurances(providers, valAddrs, providerBalances, sdk.NewDecWithPrec(10, 2), nil)
	// 3 delegators
	delegators, delegatorBalances := suite.AddTestAddrs(3, oneChunk.Amount)
	// liquid stakes 3 chunks
	suite.liquidStakes(delegators, delegatorBalances)
	// ---------------------------------------------------

	// Queue withdraw insurance request
	toBeWithdrawnInsurance, _ := suite.app.LiquidStakingKeeper.GetInsurance(suite.ctx, insurances[0].Id)
	chunkToBeUnpairing, _ := suite.app.LiquidStakingKeeper.GetChunk(suite.ctx, toBeWithdrawnInsurance.ChunkId)
	_, err := suite.app.LiquidStakingKeeper.DoWithdrawInsurance(
		suite.ctx,
		types.NewMsgWithdrawInsurance(
			toBeWithdrawnInsurance.ProviderAddress,
			toBeWithdrawnInsurance.Id,
		),
	)
	suite.NoError(err)
	suite.advanceEpoch()
	suite.advanceHeight(1, "queued withdraw insurance request is handled and there are no additional insurances yet so unpairing triggered")
	{
		// Check unbonding obj exists
		unbondingDelegation, found := suite.app.StakingKeeper.GetUnbondingDelegation(
			suite.ctx,
			chunkToBeUnpairing.DerivedAddress(),
			toBeWithdrawnInsurance.GetValidator(),
		)
		suite.True(found)
		suite.Equal(toBeWithdrawnInsurance.GetValidator().String(), unbondingDelegation.ValidatorAddress)
	}

	suite.advanceHeight(1, "")

	suite.advanceEpoch()
	suite.advanceHeight(1, "withdrawal and unbonding of chunkToBeUnpairing is finished")
	withdrawnInsurance, _ := suite.app.LiquidStakingKeeper.GetInsurance(suite.ctx, toBeWithdrawnInsurance.Id)
	pairingChunk, _ := suite.app.LiquidStakingKeeper.GetChunk(suite.ctx, chunkToBeUnpairing.Id)
	{
		suite.Equal(types.CHUNK_STATUS_PAIRING, pairingChunk.Status)
		suite.Equal(uint64(0), pairingChunk.UnpairingInsuranceId)
		suite.Equal(types.INSURANCE_STATUS_UNPAIRED, withdrawnInsurance.Status)
	}

	suite.advanceHeight(1, "")

	newValAddrs := suite.CreateValidators([]int64{1, 1, 1})
	newProviders, newProviderBalances := suite.AddTestAddrs(3, oneInsurance.Amount)
	newInsurances := suite.provideInsurances(
		newProviders,
		newValAddrs,
		newProviderBalances,
		sdk.NewDecWithPrec(1, 2), // much cheaper than current paired insurances
		nil,
	)

	suite.advanceEpoch()
	suite.advanceHeight(1, "pairing chunk is paired now")
	{
		// get newInsurances from module so it presents latest state of insurances
		var updatedNewInsurances []types.Insurance
		for _, newInsurance := range newInsurances {
			insurance, _ := suite.app.LiquidStakingKeeper.GetInsurance(suite.ctx, newInsurance.Id)
			updatedNewInsurances = append(updatedNewInsurances, insurance)
		}

		var updatedOldInsurances []types.Insurance
		for _, insurance := range insurances {
			insurance, _ := suite.app.LiquidStakingKeeper.GetInsurance(suite.ctx, insurance.Id)
			updatedOldInsurances = append(updatedOldInsurances, insurance)
		}

		pairedChunk, _ := suite.app.LiquidStakingKeeper.GetChunk(suite.ctx, pairingChunk.Id)
		suite.Equal(types.CHUNK_STATUS_PAIRED, pairedChunk.Status)
		suite.NoError(suite.app.LiquidStakingKeeper.IterateAllChunks(suite.ctx, func(chunk types.Chunk) (bool, error) {
			if chunk.Status == types.CHUNK_STATUS_PAIRED {
				found := false
				for _, newInsurance := range updatedNewInsurances {
					if chunk.PairedInsuranceId == newInsurance.Id &&
						newInsurance.ChunkId == chunk.Id &&
						newInsurance.Status == types.INSURANCE_STATUS_PAIRED {
						found = true
						break
					}
				}
				suite.True(found, "chunk must be paired with one of new insurances(ranked-in)")

				found = false
				// old insurances(= ranked-out) must not be paired with chunks
				for _, oldInsurance := range updatedOldInsurances {
					if chunk.PairedInsuranceId == oldInsurance.Id {
						found = true
						break
					}
					suite.True(oldInsurance.Status != types.INSURANCE_STATUS_PAIRED, "ranked-out oldInsurance must not be paired")
				}
				suite.False(found, "chunk must not be paired with one of old insurances(ranked-out)")
			}
			return false, nil
		}))
	}

	suite.advanceHeight(1, "")

	pairedInsurances := newInsurances
	newProviders, newProviderBalances = suite.AddTestAddrs(3, oneInsurance.Amount)
	newInsurances = suite.provideInsurances(
		newProviders,
		newValAddrs,
		newProviderBalances,
		sdk.NewDecWithPrec(1, 3), // much cheaper than current paired insurances
		nil,
	)

	suite.advanceEpoch()
	suite.advanceHeight(1, "all paired chunks are started to be re-paired with new insurances")
	{
		// get newInsurances from module so it presents latest state of insurances
		var updatedNewInsurances []types.Insurance
		for _, newInsurance := range newInsurances {
			insurance, _ := suite.app.LiquidStakingKeeper.GetInsurance(suite.ctx, newInsurance.Id)
			updatedNewInsurances = append(updatedNewInsurances, insurance)
		}

		var updatedOldInsurances []types.Insurance
		for _, insurance := range pairedInsurances {
			insurance, _ := suite.app.LiquidStakingKeeper.GetInsurance(suite.ctx, insurance.Id)
			updatedOldInsurances = append(updatedOldInsurances, insurance)
		}

		suite.NoError(suite.app.LiquidStakingKeeper.IterateAllChunks(suite.ctx, func(chunk types.Chunk) (bool, error) {
			if chunk.Status == types.CHUNK_STATUS_PAIRED {
				found := false
				for _, newInsurance := range updatedNewInsurances {
					if chunk.PairedInsuranceId == newInsurance.Id &&
						newInsurance.ChunkId == chunk.Id &&
						newInsurance.Status == types.INSURANCE_STATUS_PAIRED {
						found = true
						break
					}
				}
				suite.True(found, "chunk must be paired with one of new insurances(ranked-in)")

				found = false
				for _, oldInsurance := range updatedOldInsurances {
					if chunk.PairedInsuranceId == oldInsurance.Id {
						found = true
						break
					}
				}
				suite.False(found, "chunk must not be paired with one of old insurances(ranked-out)")
			}
			return false, nil
		}))
	}

}

func (suite *KeeperTestSuite) getUnitDistribution(
	unitDelegationRewardPerEpoch sdk.Int,
	fixedInsuranceFeeRate sdk.Dec,
) (sdk.Int, sdk.Int) {
	unitInsuranceCommissionPerEpoch := unitDelegationRewardPerEpoch.ToDec().Mul(fixedInsuranceFeeRate).TruncateInt()
	pureUnitRewardPerEpoch := unitDelegationRewardPerEpoch.Sub(unitInsuranceCommissionPerEpoch)
	fmt.Println("unitDelegationRewardPerEpoch: ", unitDelegationRewardPerEpoch.String())
	fmt.Println("unitInsuranceCommissionPerEpoch: ", unitInsuranceCommissionPerEpoch.String())
	fmt.Println("pureUnitRewardPerEpoch: ", pureUnitRewardPerEpoch.String())
	return unitInsuranceCommissionPerEpoch, pureUnitRewardPerEpoch
}
