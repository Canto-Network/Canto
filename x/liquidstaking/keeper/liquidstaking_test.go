package keeper_test

import (
	"fmt"
	"math/rand"

	"github.com/Canto-Network/Canto/v6/x/liquidstaking/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/simulation"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
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

// Provide insurance with random fee (1 ~ 10%)
func (suite *KeeperTestSuite) provideInsurances(providers []sdk.AccAddress, valAddrs []sdk.ValAddress, amounts []sdk.Coin) []types.Insurance {
	s := rand.NewSource(0)
	r := rand.New(s)

	valNum := len(valAddrs)
	var providedInsurances []types.Insurance
	for i, provider := range providers {
		msg := types.NewMsgInsuranceProvide(provider.String(), amounts[i])
		msg.ValidatorAddress = valAddrs[i%valNum].String()
		// 1 ~ 10% insurance fee
		msg.FeeRate = sdk.NewDecWithPrec(int64(simulation.RandIntBetween(r, 1, 10)), 2)
		msg.Amount = amounts[i]
		insurance, err := suite.app.LiquidStakingKeeper.DoInsuranceProvide(suite.ctx, msg)
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

func (suite *KeeperTestSuite) TestInsuranceProvide() {
	valAddrs := suite.CreateValidators([]int64{10, 10, 10})
	_, minimumCoverage := suite.app.LiquidStakingKeeper.GetMinimumRequirements(suite.ctx)
	providers, _ := suite.AddTestAddrs(10, minimumCoverage.Amount)

	for _, tc := range []struct {
		name        string
		msg         *types.MsgInsuranceProvide
		validate    func(ctx sdk.Context, insurance types.Insurance)
		expectedErr string
	}{
		{
			"success",
			&types.MsgInsuranceProvide{
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
			&types.MsgInsuranceProvide{
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
			insurance, err := suite.app.LiquidStakingKeeper.DoInsuranceProvide(cachedCtx, tc.msg)
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
	valAddrs := suite.CreateValidators([]int64{10, 10, 10})
	minimumRequirement, minimumCoverage := suite.app.LiquidStakingKeeper.GetMinimumRequirements(suite.ctx)
	providers, balances := suite.AddTestAddrs(10, minimumCoverage.Amount)
	suite.provideInsurances(providers, valAddrs, balances)

	delegators, balances := suite.AddTestAddrs(10, minimumRequirement.Amount)
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

	// TODO: Should test multiple liquidstake and advance blocks and chekc mintrate is right or not
}

func (suite *KeeperTestSuite) TestLiquidStakeFail() {
	valAddrs := suite.CreateValidators([]int64{10, 10, 10})
	minimumRequirement, minimumCoverage := suite.app.LiquidStakingKeeper.GetMinimumRequirements(suite.ctx)

	addrs, balances := suite.AddTestAddrs(types.MaxPairedChunks-1, minimumRequirement.Amount)

	// TC: There are no pairing insurances yet. Insurances must be provided to liquid stake
	acc1 := addrs[0]
	msg := types.NewMsgLiquidStake(acc1.String(), minimumRequirement)
	_, _, _, err := suite.app.LiquidStakingKeeper.DoLiquidStake(suite.ctx, msg)
	suite.ErrorContains(err, types.ErrNoPairingInsurance.Error())

	providers, providerBalances := suite.AddTestAddrs(10, minimumCoverage.Amount)
	suite.provideInsurances(providers, valAddrs, providerBalances)

	// TC: Not enough amount to liquid stake
	// acc1 tries to liquid stake 2 * ChunkSize tokens, but he has only ChunkSize tokens
	msg = types.NewMsgLiquidStake(acc1.String(), minimumRequirement.AddAmount(types.ChunkSize))
	_, _, _, err = suite.app.LiquidStakingKeeper.DoLiquidStake(suite.ctx, msg)
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
	suite.liquidStakes([]sdk.AccAddress{acc1}, []sdk.Coin{minimumRequirement})

	// TC: MaxPairedChunks is reached, no more chunks can be paired
	newAddrs, newBalances := suite.AddTestAddrs(1, minimumRequirement.Amount)
	msg = types.NewMsgLiquidStake(newAddrs[0].String(), newBalances[0])
	_, _, _, err = suite.app.LiquidStakingKeeper.DoLiquidStake(suite.ctx, msg)
	suite.ErrorIs(err, types.ErrMaxPairedChunkSizeExceeded)
}

// TODO: Must implement scenario test for liquid staking
func (suite *KeeperTestSuite) TestLiquidStakeWithAdvanceBlocks() {
	valAddrs := suite.CreateValidators([]int64{1, 1, 1})
	minimumRequirement, minimumCoverage := suite.app.LiquidStakingKeeper.GetMinimumRequirements(suite.ctx)
	providers, balances := suite.AddTestAddrs(10, minimumCoverage.Amount)
	suite.provideInsurances(providers, valAddrs, balances)

	nas1 := suite.app.LiquidStakingKeeper.GetNetAmountState(suite.ctx)
	delegators, delegatorBalances := suite.AddTestAddrs(3, minimumRequirement.Amount)
	_ = suite.liquidStakes(delegators, delegatorBalances)
	nas2 := suite.app.LiquidStakingKeeper.GetNetAmountState(suite.ctx)

	suite.advanceHeight(1)
	nas3 := suite.app.LiquidStakingKeeper.GetNetAmountState(suite.ctx)

	// Check NetAmountState while incrementing blocks by one
	suite.advanceHeight(1)
	nas4 := suite.app.LiquidStakingKeeper.GetNetAmountState(suite.ctx)

	fmt.Println(nas1)
	fmt.Println(nas2)
	fmt.Println(nas3)
	fmt.Println(nas4)

	suite.app.StakingKeeper.IterateAllDelegations(suite.ctx, func(delegation stakingtypes.Delegation) bool {
		fmt.Println(delegation)
		return false
	})

	suite.app.StakingKeeper.IterateBondedValidatorsByPower(suite.ctx, func(index int64, validator stakingtypes.ValidatorI) bool {
		fmt.Println("validator.address: ", validator.GetOperator())
		fmt.Println("validator.minSelfDelegation: ", validator.GetMinSelfDelegation())
		fmt.Println("validator.tokens: ", validator.GetTokens().ToDec())
		fmt.Println("validator.delegatorShares: ", validator.GetDelegatorShares())
		return false
	})
	suite.app.StakingKeeper.IterateHistoricalInfo(suite.ctx, func(info stakingtypes.HistoricalInfo) bool {
		fmt.Println(info)
		return false
	})
	suite.app.DistrKeeper.IterateValidatorAccumulatedCommissions(suite.ctx, func(validatorAddr sdk.ValAddress, commission distrtypes.ValidatorAccumulatedCommission) bool {
		fmt.Println(validatorAddr, commission)
		return false
	})
}

// TODO: Update liquid unstake scenario test with updated code
// func (suite *KeeperTestSuite) TestLiquidUnstakeWithAdvanceBlocks() {
// 	// 3 validators we have
// 	valAddrs := suite.CreateValidators([]int64{1, 1, 1})
// 	oneChunk, oneInsurance := suite.app.LiquidStakingKeeper.GetMinimumRequirements(suite.ctx)
// 	providers, providerBalances := suite.AddTestAddrs(10, oneInsurance.Amount)
// 	suite.provideInsurances(providers, valAddrs, providerBalances)
//
// 	// 3 delegators
// 	delegators, delegatorBalances := suite.AddTestAddrs(3, oneChunk.Amount)
// 	nas := suite.app.LiquidStakingKeeper.GetNetAmountState(suite.ctx)
// 	suite.True(nas.IsZeroState(), "nothing happened yet so it must be zero state")
//
// 	// liquid stake 3 chunks (each delegator liquid stakes 1 chunk)
// 	pairedChunks := suite.liquidStakes(delegators, delegatorBalances)
// 	mostExpensivePairedChunk := suite.getMostExpensivePairedChunk(pairedChunks)
//
// 	nas = suite.app.LiquidStakingKeeper.GetNetAmountState(suite.ctx)
// 	// 1 chunk size * number of paired chunks (=3) tokens are liquidated
// 	currentLiquidtedTokens := types.ChunkSize.Mul(sdk.NewInt(int64(len(pairedChunks))))
// 	currentInsuranceTokens := oneInsurance.Amount.Mul(sdk.NewInt(int64(len(pairedChunks))))
// 	suite.True(nas.Equal(types.NetAmountState{
// 		MintRate:               sdk.OneDec(),
// 		LsTokensTotalSupply:    currentLiquidtedTokens,
// 		NetAmount:              currentLiquidtedTokens.ToDec(),
// 		TotalDelShares:         currentLiquidtedTokens.ToDec(),
// 		TotalRemainingRewards:  sdk.ZeroDec(),
// 		TotalChunksBalance:     sdk.ZeroInt(),
// 		TotalLiquidTokens:      currentLiquidtedTokens,
// 		TotalInsuranceTokens:   currentInsuranceTokens,
// 		TotalUnbondingBalance:  sdk.ZeroInt(),
// 		RewardModuleAccBalance: sdk.ZeroInt(),
// 	}), "no block processed yet, so there are no mint rate change and remaining rewards yet")
//
// 	// advance 1 block(= epoch period in test environment) so reward is accumulated which means mint rate is changed
// 	suite.advanceHeight(1)
// 	eachDelegationRewardPerEpoch, _ := sdk.NewIntFromString("29999994000000000000")
// 	// each delegation reward per epoch(=1 block in test) * number of paired chunks
// 	// = 29999994000000000000 * 3
// 	notClaimedRewards := eachDelegationRewardPerEpoch.Mul(sdk.NewInt(int64(len(pairedChunks))))
// 	beforeNas := nas
// 	nas = suite.app.LiquidStakingKeeper.GetNetAmountState(suite.ctx)
// 	suite.Equal(
// 		nas.TotalRemainingRewards.Sub(beforeNas.TotalRemainingRewards),
// 		notClaimedRewards.ToDec(),
// 		"one epoch(=1 block in test) passed, so remaining rewards must be increased",
// 	)
// 	suite.Equal(nas.NetAmount.Sub(beforeNas.NetAmount), notClaimedRewards.ToDec(), "net amount must be increased by not claimed rewards")
// 	suite.Equal(nas.MintRate, sdk.MustNewDecFromStr("0.999994000037199769"), "mint rate increased because of reward accumulation")
//
// 	bondDenom := suite.app.StakingKeeper.BondDenom(suite.ctx)
// 	liquidBondDenom := suite.app.LiquidStakingKeeper.GetLiquidBondDenom(suite.ctx)
//
// 	// liquid unstake 1 chunk
// 	undelegator := delegators[0]
// 	beforeBondDenomBalance := suite.app.BankKeeper.GetBalance(suite.ctx, undelegator, bondDenom)
// 	beforeLiquidBondDenomBalance := suite.app.BankKeeper.GetBalance(suite.ctx, undelegator, liquidBondDenom)
//
// 	msg := types.NewMsgLiquidUnstake(undelegator.String(), oneChunk)
// 	lsTokensToEscrow := nas.MintRate.Mul(oneChunk.Amount.ToDec()).TruncateInt()
// 	unstakedChunks, unstakeUnobndingDelegationInfos, err := suite.app.LiquidStakingKeeper.QueueLiquidUnstake(suite.ctx, msg)
// 	suite.NoError(err)
// 	beforeNas = nas
// 	nas = suite.app.LiquidStakingKeeper.GetNetAmountState(suite.ctx)
//
// 	bondDenomBalance := suite.app.BankKeeper.GetBalance(suite.ctx, undelegator, bondDenom)
// 	liquidBondDenomBalance := suite.app.BankKeeper.GetBalance(suite.ctx, undelegator, liquidBondDenom)
// 	suite.Equal(bondDenomBalance.Sub(beforeBondDenomBalance).Amount, sdk.ZeroInt(), "unbonding period is just started so no tokens are backed yet")
// 	suite.Equal(
// 		beforeLiquidBondDenomBalance.Sub(liquidBondDenomBalance).Amount,
// 		lsTokensToEscrow,
// 		"ls tokens are escrowed by module",
// 	)
// 	suite.Equal(
// 		suite.app.BankKeeper.GetBalance(suite.ctx, types.LsTokenEscrowAcc, liquidBondDenom).Amount,
// 		lsTokensToEscrow,
// 		"module got ls tokens from liquid unstaker",
// 	)
//
// 	// check states after liquid unstake
// 	pairedChunksAfterUnstake := suite.getPairedChunks()
// 	suite.Len(unstakedChunks, 1)
// 	suite.Len(unstakeUnobndingDelegationInfos, 1)
// 	// unstakedChunk should be the most expensive insurance paired with the previously paired chunk.
// 	suite.Equal(unstakedChunks[0].Id, mostExpensivePairedChunk.Id)
// 	suite.Equal(unstakedChunks[0].InsuranceId, mostExpensivePairedChunk.InsuranceId)
// 	// paired chunk count should be decreased by number of unstaked chunks
// 	suite.Len(pairedChunksAfterUnstake, len(pairedChunks)-len(unstakedChunks))
//
// 	// check chagned net amount state after liquid unstake
// 	suite.Equal(
// 		beforeNas.TotalRemainingRewards.Sub(nas.TotalRemainingRewards),
// 		eachDelegationRewardPerEpoch.ToDec(),
// 		"delegation reward for one chunk is claimed by reward module account",
// 	)
// 	suite.Equal(
// 		beforeNas.TotalInsuranceTokens.Sub(nas.TotalInsuranceTokens),
// 		oneInsurance.Amount,
// 		"insurance tokens must be increased by one chunk insurance",
// 	)
// 	unstakedInsurance, _ := suite.app.LiquidStakingKeeper.GetInsurance(suite.ctx, unstakedChunks[0].InsuranceId)
// 	insuranceCommission := unstakedInsurance.FeeRate.Mul(eachDelegationRewardPerEpoch.ToDec()).TruncateInt()
// 	// check reward module balance
// 	suite.Equal(
// 		suite.app.BankKeeper.GetBalance(suite.ctx, types.RewardPool, suite.app.StakingKeeper.BondDenom(suite.ctx)).Amount,
// 		eachDelegationRewardPerEpoch.Sub(insuranceCommission),
// 		"reward module account collect chunk's delegation reward minus insurance commission",
// 	)
// 	suite.Equal(nas.RewardModuleAccBalance, eachDelegationRewardPerEpoch.Sub(insuranceCommission))
// 	// check insurance got fee
// 	suite.Equal(
// 		suite.app.BankKeeper.GetBalance(suite.ctx, unstakedInsurance.FeePoolAddress(), suite.app.StakingKeeper.BondDenom(suite.ctx)).Amount,
// 		insuranceCommission,
// 		"insurance got commission from chunk's delegation reward",
// 	)
// }
//
// func (suite *KeeperTestSuite) TestLiquidUnstakeFail() {
// 	valAddrs := suite.CreateValidators([]int64{10, 10, 10})
// 	minimumRequirement, minimumCoverage := suite.app.LiquidStakingKeeper.GetMinimumRequirements(suite.ctx)
// 	providers, rpvodierBalances := suite.AddTestAddrs(10, minimumCoverage.Amount)
// 	suite.provideInsurances(providers, valAddrs, rpvodierBalances)
//
// 	// Now we have 1 paired chunks
// 	delegators, delegatorBalances := suite.AddTestAddrs(3, minimumRequirement.Amount)
// 	undelegator := delegators[0]
//
// 	msg := types.NewMsgLiquidUnstake(
// 		undelegator.String(),
// 		minimumRequirement,
// 	)
// 	_, _, err := suite.app.LiquidStakingKeeper.QueueLiquidUnstake(suite.ctx, msg)
// 	suite.ErrorContains(err, types.ErrNoPairedChunk.Error())
//
// 	// create one paired chunk
// 	_ = suite.liquidStakes([]sdk.AccAddress{delegators[0]}, []sdk.Coin{delegatorBalances[0]})
//
// 	msg.Amount.Amount = msg.Amount.Amount.Sub(sdk.NewInt(1))
// 	// TC: Must be multiple of chunk size
// 	_, _, err = suite.app.LiquidStakingKeeper.QueueLiquidUnstake(suite.ctx, msg)
// 	suite.ErrorContains(err, types.ErrInvalidAmount.Error())
// 	msg.Amount = msg.Amount.Add(sdk.NewCoin(suite.denom, sdk.OneInt())) // now amount is valid
//
// 	// TC: Must be bond denom
// 	msg.Amount.Denom = "invalid"
// 	_, _, err = suite.app.LiquidStakingKeeper.QueueLiquidUnstake(suite.ctx, msg)
// 	suite.ErrorContains(err, types.ErrInvalidBondDenom.Error())
// 	msg.Amount.Denom = suite.denom // now denom is valid
//
// 	// TC: Want to liquid unstake 2 chunks but current paired chunk is only one
// 	msg.Amount.Amount = minimumRequirement.Amount.Mul(sdk.NewInt(2))
// 	_, _, err = suite.app.LiquidStakingKeeper.QueueLiquidUnstake(suite.ctx, msg)
// 	suite.ErrorContains(err, types.ErrExceedAvailableChunks.Error())
//
// 	// Now we have 3 paired chunks
// 	_ = suite.liquidStakes(delegators[1:], delegatorBalances[1:])
//
// 	// TC: Want to liquid unstake 2 chunks but unstaker have lstokens corresponding to 1 chunk size
// 	_, _, err = suite.app.LiquidStakingKeeper.QueueLiquidUnstake(suite.ctx, msg)
// 	suite.ErrorContains(err, sdkerrors.ErrInsufficientFunds.Error())
// }

func (suite *KeeperTestSuite) TestCancelInsuranceProvideSuccess() {
	valAddrs := suite.CreateValidators([]int64{10, 10, 10})
	_, minimumCoverage := suite.app.LiquidStakingKeeper.GetMinimumRequirements(suite.ctx)
	providers, balances := suite.AddTestAddrs(10, minimumCoverage.Amount)
	insurances := suite.provideInsurances(providers, valAddrs, balances)

	provider := providers[0]
	insurance := insurances[0]
	escrowed := suite.app.BankKeeper.GetBalance(suite.ctx, insurance.DerivedAddress(), suite.denom)
	beforeProviderBalance := suite.app.BankKeeper.GetBalance(suite.ctx, provider, suite.denom)
	msg := types.NewMsgCancelInsuranceProvide(provider.String(), insurance.Id)
	canceledInsurance, err := suite.app.LiquidStakingKeeper.DoCancelInsuranceProvide(suite.ctx, msg)
	suite.NoError(err)
	suite.True(insurance.Equal(canceledInsurance))
	afterProviderBalance := suite.app.BankKeeper.GetBalance(suite.ctx, provider, suite.denom)
	suite.True(afterProviderBalance.Amount.Equal(beforeProviderBalance.Amount.Add(escrowed.Amount)), "provider should get back escrowed amount")
}

func (suite *KeeperTestSuite) TestCancelInsuranceProvideFail() {
	valAddrs := suite.CreateValidators([]int64{10, 10, 10})
	minimumRequirement, minimumCoverage := suite.app.LiquidStakingKeeper.GetMinimumRequirements(suite.ctx)
	providers, balances := suite.AddTestAddrs(10, minimumCoverage.Amount)
	suite.provideInsurances(providers, valAddrs, balances)

	// TC: No insurance to cancel
	var notExistingInsuranceId uint64 = 9999
	provider := providers[0]

	_, err := suite.app.LiquidStakingKeeper.DoCancelInsuranceProvide(
		suite.ctx,
		types.NewMsgCancelInsuranceProvide(provider.String(), notExistingInsuranceId),
	)
	suite.ErrorIs(err, types.ErrPairingInsuranceNotFound, "only pairing insurances can be canceled")

	// TC: Paired insurance cannot be canceled
	delegators, delegatorBalances := suite.AddTestAddrs(10, minimumRequirement.Amount)
	del1 := delegators[0]
	amt1 := delegatorBalances[0]
	createdChunks, _, _, err := suite.app.LiquidStakingKeeper.DoLiquidStake(suite.ctx, types.NewMsgLiquidStake(del1.String(), amt1))
	chunk := createdChunks[0]
	insurance, found := suite.app.LiquidStakingKeeper.GetInsurance(suite.ctx, chunk.PairedInsuranceId)
	suite.True(found)

	_, err = suite.app.LiquidStakingKeeper.DoCancelInsuranceProvide(
		suite.ctx,
		types.NewMsgCancelInsuranceProvide(insurance.ProviderAddress, insurance.Id),
	)
	suite.ErrorIs(err, types.ErrPairingInsuranceNotFound, "only pairing insurances can be canceled")
}
