package keeper_test

import (
	"bytes"
	"fmt"
	"math/rand"
	"time"

	liquidstakingkeeper "github.com/Canto-Network/Canto/v6/x/liquidstaking"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/simapp"
	"github.com/cosmos/cosmos-sdk/x/staking"
	"github.com/cosmos/cosmos-sdk/x/staking/teststaking"

	"github.com/Canto-Network/Canto/v6/x/liquidstaking/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/simulation"
	evidencetypes "github.com/cosmos/cosmos-sdk/x/evidence/types"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	ethermint "github.com/evmos/ethermint/types"
)

var onePower int64 = 1
var tenPercentFeeRate = sdk.NewDecWithPrec(10, 2)

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
	ctx sdk.Context,
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
		insurance, err := suite.app.LiquidStakingKeeper.DoProvideInsurance(ctx, msg)
		suite.NoError(err)
		providedInsurances = append(providedInsurances, insurance)
	}
	suite.mustPassInvariants()
	return providedInsurances
}

func (suite *KeeperTestSuite) liquidStakes(ctx sdk.Context, delegators []sdk.AccAddress, amounts []sdk.Coin) []types.Chunk {
	var chunks []types.Chunk
	for i, delegator := range delegators {
		msg := types.NewMsgLiquidStake(delegator.String(), amounts[i])
		createdChunks, _, _, err := suite.app.LiquidStakingKeeper.DoLiquidStake(ctx, msg)
		suite.NoError(err)
		for _, chunk := range createdChunks {
			chunks = append(chunks, chunk)
		}
	}
	suite.mustPassInvariants()
	return chunks
}

func (suite *KeeperTestSuite) TestProvideInsurance() {
	suite.resetEpochs()
	valAddrs, _ := suite.CreateValidators(
		[]int64{onePower, onePower, onePower},
		tenPercentFeeRate,
		nil,
	)
	_, oneInsurance := suite.app.LiquidStakingKeeper.GetMinimumRequirements(suite.ctx)
	providers, _ := suite.AddTestAddrs(10, oneInsurance.Amount)

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
				Amount:           oneInsurance,
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
				Amount:           oneInsurance.SubAmount(sdk.NewInt(1)),
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
	suite.mustPassInvariants()
}

func (suite *KeeperTestSuite) TestLiquidStakeSuccess() {
	suite.resetEpochs()
	valAddrs, _ := suite.CreateValidators(
		[]int64{onePower, onePower, onePower},
		tenPercentFeeRate,
		nil,
	)
	oneChunk, oneInsurance := suite.app.LiquidStakingKeeper.GetMinimumRequirements(suite.ctx)
	providers, balances := suite.AddTestAddrs(10, oneInsurance.Amount)
	suite.provideInsurances(suite.ctx, providers, valAddrs, balances, sdk.ZeroDec(), nil)

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
	{
		suite.NoError(suite.app.LiquidStakingKeeper.IterateAllChunks(suite.ctx, func(chunk types.Chunk) (bool, error) {
			suite.True(chunk.Equal(createdChunks[idx]))
			idx++
			return false, nil
		}))
		suite.Equal(len(createdChunks), idx, "number of created chunks should be equal to number of chunks in db")
	}

	lsTokenAfter := suite.app.BankKeeper.GetBalance(suite.ctx, del1, liquidBondDenom)
	{
		suite.NoError(err)
		suite.True(amt1.Amount.Equal(newShares.TruncateInt()), "delegation shares should be equal to amount")
		suite.True(amt1.Amount.Equal(lsTokenMintAmount), "at first try mint rate is 1, so mint amount should be equal to amount")
		suite.True(lsTokenAfter.Sub(lsTokenBefore).Amount.Equal(lsTokenMintAmount), "liquid staker must have minted ls tokens in account balance")
	}

	// NetAmountState should be updated correctly
	afterNas := suite.app.LiquidStakingKeeper.GetNetAmountState(suite.ctx)
	{
		suite.True(afterNas.LsTokensTotalSupply.Equal(lsTokenMintAmount), "total ls token supply should be equal to minted amount")
		suite.True(nas.TotalLiquidTokens.Add(amt1.Amount).Equal(afterNas.TotalLiquidTokens))
		suite.True(nas.NetAmount.Add(amt1.Amount.ToDec()).Equal(afterNas.NetAmount))
		suite.True(afterNas.MintRate.Equal(sdk.OneDec()), "no rewards yet, so mint rate should be 1")
	}
	suite.mustPassInvariants()
}

func (suite *KeeperTestSuite) TestLiquidStakeFail() {
	suite.resetEpochs()
	valAddrs, _ := suite.CreateValidators(
		[]int64{onePower, onePower, onePower},
		tenPercentFeeRate,
		nil,
	)
	oneChunk, oneInsurance := suite.app.LiquidStakingKeeper.GetMinimumRequirements(suite.ctx)

	addrs, balances := suite.AddTestAddrs(types.MaxPairedChunks-1, oneChunk.Amount)

	// TC: There are no pairing insurances yet. Insurances must be provided to liquid stake
	acc1 := addrs[0]
	msg := types.NewMsgLiquidStake(acc1.String(), oneChunk)
	_, _, _, err := suite.app.LiquidStakingKeeper.DoLiquidStake(suite.ctx, msg)
	suite.ErrorContains(err, types.ErrNoPairingInsurance.Error())

	providers, providerBalances := suite.AddTestAddrs(10, oneInsurance.Amount)
	suite.provideInsurances(suite.ctx, providers, valAddrs, providerBalances, sdk.ZeroDec(), nil)

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
	_ = suite.liquidStakes(suite.ctx, addrs, balances)

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
	suite.liquidStakes(suite.ctx, []sdk.AccAddress{acc1}, []sdk.Coin{oneChunk})

	// TC: MaxPairedChunks is reached, no more chunks can be paired
	newAddrs, newBalances := suite.AddTestAddrs(1, oneChunk.Amount)
	msg = types.NewMsgLiquidStake(newAddrs[0].String(), newBalances[0])
	_, _, _, err = suite.app.LiquidStakingKeeper.DoLiquidStake(suite.ctx, msg)
	suite.ErrorIs(err, types.ErrMaxPairedChunkSizeExceeded)

	suite.mustPassInvariants()
}

func (suite *KeeperTestSuite) TestLiquidStakeWithAdvanceBlocks() {
	fixedInsuranceFeeRate := tenPercentFeeRate
	env := suite.setupLiquidStakeTestingEnv(testingEnvOptions{
		desc:                  "TestLiquidStakeWithAdvanceBlocks",
		numVals:               3,
		fixedValFeeRate:       tenPercentFeeRate,
		valFeeRates:           nil,
		fixedPower:            onePower,
		powers:                nil,
		numInsurances:         10,
		fixedInsuranceFeeRate: fixedInsuranceFeeRate,
		insuranceFeeRates:     nil,
		numPairedChunks:       3,
	})

	_, oneInsurance := suite.app.LiquidStakingKeeper.GetMinimumRequirements(suite.ctx)
	unitDelegationRewardPerRewardEpoch, _ := sdk.NewIntFromString("29999994000000000000")
	unitInsuranceCommissionPerRewardEpoch, pureUnitRewardPerRewardEpoch := suite.getUnitDistribution(unitDelegationRewardPerRewardEpoch, fixedInsuranceFeeRate)

	nas := suite.app.LiquidStakingKeeper.GetNetAmountState(suite.ctx)
	fmt.Println(nas)
	pairedChunksInt := sdk.NewInt(int64(len(env.pairedChunks)))
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
			TotalRemainingInsuranceCommissions: sdk.ZeroDec(),
			TotalChunksBalance:                 sdk.ZeroInt(),
			TotalLiquidTokens:                  currentLiquidatedTokens,
			TotalInsuranceTokens:               oneInsurance.Amount.Mul(sdk.NewInt(int64(len(env.insurances)))),
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
			pureUnitRewardPerRewardEpoch.Mul(pairedChunksInt).String(),
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
			pureUnitRewardPerRewardEpoch.Mul(pairedChunksInt).Mul(sdk.NewInt(suite.rewardEpochCount)).String(),
			nas.RewardModuleAccBalance.String(),
		)
		suite.Equal(
			unitInsuranceCommissionPerRewardEpoch.Mul(pairedChunksInt).Mul(sdk.NewInt(suite.rewardEpochCount)).String(),
			nas.TotalPairedInsuranceCommissions.String(),
		)
		suite.Equal("0.999989200118798693", nas.MintRate.String())
		suite.True(nas.MintRate.LT(beforeNas.MintRate))
	}
}

func (suite *KeeperTestSuite) TestLiquidUnstakeWithAdvanceBlocks() {
	fixedInsuranceFeeRate := tenPercentFeeRate
	env := suite.setupLiquidStakeTestingEnv(testingEnvOptions{
		desc:                  "TestLiquidUnstakeWithAdvanceBlocks",
		numVals:               3,
		fixedValFeeRate:       tenPercentFeeRate,
		valFeeRates:           nil,
		fixedPower:            onePower,
		powers:                nil,
		numInsurances:         10,
		fixedInsuranceFeeRate: fixedInsuranceFeeRate,
		insuranceFeeRates:     nil,
		numPairedChunks:       3,
	})
	oneChunk, oneInsurance := suite.app.LiquidStakingKeeper.GetMinimumRequirements(suite.ctx)
	pairedChunksInt := sdk.NewInt(int64(len(env.pairedChunks)))
	mostExpensivePairedChunk := suite.getMostExpensivePairedChunk(env.pairedChunks)
	nas := suite.app.LiquidStakingKeeper.GetNetAmountState(suite.ctx)
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
			TotalRemainingInsuranceCommissions: sdk.ZeroDec(),
			TotalChunksBalance:                 sdk.ZeroInt(),
			TotalLiquidTokens:                  currentLiquidatedTokens,
			TotalInsuranceTokens:               oneInsurance.Amount.Mul(sdk.NewInt(int64(len(env.insurances)))),
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

	unitDelegationRewardPerRewardEpoch, _ := sdk.NewIntFromString("29999994000000000000")
	unitInsuranceCommissionPerRewardEpoch, pureUnitRewardPerRewardEpoch := suite.getUnitDistribution(unitDelegationRewardPerRewardEpoch, fixedInsuranceFeeRate)

	// each delegation reward per epoch(=1 block in test) * number of paired chunks
	// = 29999994000000000000 * 3
	notClaimedRewards := pureUnitRewardPerRewardEpoch.Mul(pairedChunksInt)
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

	undelegator := env.delegators[0]
	// Queue liquid unstake 1 chunk
	fmt.Println("Queue liquid unstake 1 chunk")
	beforeBondDenomBalance := suite.app.BankKeeper.GetBalance(suite.ctx, undelegator, env.bondDenom)
	beforeLiquidBondDenomBalance := suite.app.BankKeeper.GetBalance(suite.ctx, undelegator, env.liquidBondDenom)
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
		bondDenomBalance := suite.app.BankKeeper.GetBalance(suite.ctx, undelegator, env.bondDenom)
		liquidBondDenomBalance := suite.app.BankKeeper.GetBalance(suite.ctx, undelegator, env.liquidBondDenom)
		suite.Equal(sdk.ZeroInt(), bondDenomBalance.Sub(beforeBondDenomBalance).Amount, "unbonding period is just started so no tokens are backed yet")
		suite.Equal(
			lsTokensToEscrow,
			beforeLiquidBondDenomBalance.Sub(liquidBondDenomBalance).Amount,
			"ls tokens are escrowed by module",
		)
		suite.Equal(
			lsTokensToEscrow,
			suite.app.BankKeeper.GetBalance(suite.ctx, types.LsTokenEscrowAcc, env.liquidBondDenom).Amount,
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
			pureUnitRewardPerRewardEpoch.Mul(pairedChunksInt).Mul(sdk.NewInt(suite.rewardEpochCount)).String(),
			nas.RewardModuleAccBalance.String(),
			fmt.Sprintf("before unstaking triggered there are collecting reward process so reward module got %d chunk's rewards", pairedChunksInt.Int64()),
		)
		suite.Equal(
			unitInsuranceCommissionPerRewardEpoch.Mul(sdk.NewInt(suite.rewardEpochCount)).String(),
			nas.TotalUnpairingInsuranceCommissions.String(),
		)
		suite.Equal(
			unitInsuranceCommissionPerRewardEpoch.Mul(sdk.NewInt(suite.rewardEpochCount).Mul(sdk.NewInt(2))).String(),
			nas.TotalPairedInsuranceCommissions.Sub(beforeNas.TotalPairedInsuranceCommissions).String(),
		)
		suite.Equal(
			oneInsurance.Amount.String(),
			nas.TotalUnpairingInsuranceTokens.Sub(beforeNas.TotalUnpairingInsuranceTokens).String(),
			"",
		)
		suite.Equal(
			unitInsuranceCommissionPerRewardEpoch.Mul(sdk.NewInt(suite.rewardEpochCount)).String(),
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
	suite.Equal(len(env.pairedChunks)-len(UnpairingForUnstakingChunks), len(pairedChunksAfterUnstake))
	pairedChunksInt = sdk.NewInt(int64(len(pairedChunksAfterUnstake)))

	suite.advanceEpoch()
	suite.advanceHeight(1, "The insurance commission and reward are claimed\nThe unstaking is completed")

	// Now number of paired chunk is decreased and still reward is fixed,
	// so the unit reward per epoch is increased.
	unitDelegationRewardPerRewardEpoch, _ = sdk.NewIntFromString("44999986500000000000")
	unitInsuranceCommissionPerRewardEpoch, pureUnitRewardPerRewardEpoch = suite.getUnitDistribution(unitDelegationRewardPerRewardEpoch, fixedInsuranceFeeRate)

	beforeNas = nas
	nas = suite.app.LiquidStakingKeeper.GetNetAmountState(suite.ctx)
	fmt.Println(nas)
	afterBondDenomBalance := suite.app.BankKeeper.GetBalance(suite.ctx, undelegator, env.bondDenom).Amount
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
			pureUnitRewardPerRewardEpoch.Mul(pairedChunksInt).String(),
			nas.RewardModuleAccBalance.Sub(beforeNas.RewardModuleAccBalance).String(),
			"reward module account balance must be increased by pure reward per epoch * reward epoch count",
		)
		suite.Equal(
			unitInsuranceCommissionPerRewardEpoch.Mul(pairedChunksInt).String(),
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
	suite.resetEpochs()
	valAddrs, _ := suite.CreateValidators(
		[]int64{onePower, onePower, onePower},
		tenPercentFeeRate,
		nil,
	)
	oneChunk, oneInsurance := suite.app.LiquidStakingKeeper.GetMinimumRequirements(suite.ctx)
	providers, providerBalances := suite.AddTestAddrs(10, oneInsurance.Amount)
	suite.provideInsurances(suite.ctx, providers, valAddrs, providerBalances, sdk.ZeroDec(), nil)
	delegators, delegatorBalances := suite.AddTestAddrs(3, oneChunk.Amount)
	undelegator := delegators[0]

	for _, tc := range []struct {
		name        string
		msg         *types.MsgLiquidUnstake
		setupFunc   func(sdk.Context)
		expectedErr string
	}{
		{
			"no paired chunk to unstake",
			&types.MsgLiquidUnstake{
				DelegatorAddress: undelegator.String(),
				Amount:           oneChunk,
			},
			nil,
			types.ErrNoPairedChunk.Error(),
		},
		{
			"must be multiple of chunk size",
			&types.MsgLiquidUnstake{
				DelegatorAddress: undelegator.String(),
				Amount:           oneChunk.AddAmount(sdk.NewInt(1)),
			},
			func(ctx sdk.Context) {
				_ = suite.liquidStakes(ctx, []sdk.AccAddress{delegators[0]}, []sdk.Coin{delegatorBalances[0]})
			},
			types.ErrInvalidAmount.Error(),
		},
		{
			"must be bond denom",
			&types.MsgLiquidUnstake{
				DelegatorAddress: undelegator.String(),
				Amount:           sdk.NewCoin("invalidDenom", oneChunk.Amount),
			},
			func(ctx sdk.Context) {
				_ = suite.liquidStakes(ctx, []sdk.AccAddress{delegators[0]}, []sdk.Coin{delegatorBalances[0]})
			},
			types.ErrInvalidBondDenom.Error(),
		},
		{
			"try to unstake 2 chunks but there is only one paired chunk",
			&types.MsgLiquidUnstake{
				DelegatorAddress: undelegator.String(),
				Amount:           oneChunk.AddAmount(oneChunk.Amount),
			},
			func(ctx sdk.Context) {
				_ = suite.liquidStakes(ctx, []sdk.AccAddress{delegators[0]}, []sdk.Coin{delegatorBalances[0]})
			},
			types.ErrExceedAvailableChunks.Error(),
		},
		{
			"",
			&types.MsgLiquidUnstake{
				DelegatorAddress: undelegator.String(),
				Amount:           oneChunk.Add(oneChunk),
			},
			func(ctx sdk.Context) {
				_ = suite.liquidStakes(ctx, delegators, delegatorBalances)
			},
			sdkerrors.ErrInsufficientFunds.Error(),
		},
	} {
		suite.Run(tc.name, func() {
			s.Require().NoError(tc.msg.ValidateBasic())
			cachedCtx, _ := s.ctx.CacheContext()
			if tc.setupFunc != nil {
				tc.setupFunc(cachedCtx)
			}
			_, _, err := suite.app.LiquidStakingKeeper.QueueLiquidUnstake(cachedCtx, tc.msg)
			suite.ErrorContains(err, tc.expectedErr)
		})
	}
}

func (suite *KeeperTestSuite) TestCancelProvideInsuranceSuccess() {
	suite.resetEpochs()
	valAddrs, _ := suite.CreateValidators(
		[]int64{onePower, onePower, onePower},
		tenPercentFeeRate,
		nil,
	)
	_, minimumCoverage := suite.app.LiquidStakingKeeper.GetMinimumRequirements(suite.ctx)
	providers, balances := suite.AddTestAddrs(10, minimumCoverage.Amount)
	insurances := suite.provideInsurances(suite.ctx, providers, valAddrs, balances, sdk.ZeroDec(), nil)

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
	suite.mustPassInvariants()
}

func (suite *KeeperTestSuite) TestDoCancelProvideInsuranceFail() {
	env := suite.setupLiquidStakeTestingEnv(testingEnvOptions{
		desc:                  "TestDoCancelProvideInsuranceFail",
		numVals:               3,
		fixedValFeeRate:       tenPercentFeeRate,
		valFeeRates:           nil,
		fixedPower:            onePower,
		powers:                nil,
		numInsurances:         3,
		fixedInsuranceFeeRate: tenPercentFeeRate,
		insuranceFeeRates:     nil,
		numPairedChunks:       1,
	})
	onlyPairedInsurance := env.insurances[0]

	tcs := []struct {
		name        string
		msg         *types.MsgCancelProvideInsurance
		expectedErr error
	}{
		{
			name: "invalid provider",
			msg: types.NewMsgCancelProvideInsurance(
				env.providers[1].String(),
				env.insurances[2].Id,
			),
			expectedErr: types.ErrNotProviderOfInsurance,
		},
		{
			name: "invalid insurance id",
			msg: types.NewMsgCancelProvideInsurance(
				env.providers[1].String(),
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
	suite.mustPassInvariants()
}

func (suite *KeeperTestSuite) TestDoWithdrawInsurance() {
	env := suite.setupLiquidStakeTestingEnv(testingEnvOptions{
		desc:                  "TestDoWithdrawInsurance",
		numVals:               3,
		fixedValFeeRate:       tenPercentFeeRate,
		valFeeRates:           nil,
		fixedPower:            onePower,
		powers:                nil,
		numInsurances:         3,
		fixedInsuranceFeeRate: tenPercentFeeRate,
		insuranceFeeRates:     nil,
		numPairedChunks:       3,
	})

	toBeWithdrawnInsurance, _ := suite.app.LiquidStakingKeeper.GetInsurance(suite.ctx, env.insurances[0].Id)
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

	unpairedInsurance, _ := suite.app.LiquidStakingKeeper.GetInsurance(suite.ctx, env.insurances[0].Id)
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
	suite.mustPassInvariants()
}

func (suite *KeeperTestSuite) TestDoWithdrawInsuranceFail() {
	suite.resetEpochs()
	valAddrs, _ := suite.CreateValidators(
		[]int64{onePower, onePower, onePower},
		tenPercentFeeRate,
		nil,
	)
	_, oneInsurance := suite.app.LiquidStakingKeeper.GetMinimumRequirements(suite.ctx)
	providers, providerBalances := suite.AddTestAddrs(3, oneInsurance.Amount.Add(sdk.NewInt(100)))
	insurances := suite.provideInsurances(suite.ctx, providers, valAddrs, providerBalances, sdk.NewDecWithPrec(10, 2), nil)

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
	suite.mustPassInvariants()
}

func (suite *KeeperTestSuite) TestDoWithdrawInsuranceCommission() {
	fixedInsuranceFeeRate := tenPercentFeeRate
	env := suite.setupLiquidStakeTestingEnv(testingEnvOptions{
		desc:                  "TestDoWithdrawInsuranceCommission",
		numVals:               3,
		fixedValFeeRate:       tenPercentFeeRate,
		valFeeRates:           nil,
		fixedPower:            onePower,
		powers:                nil,
		numInsurances:         3,
		fixedInsuranceFeeRate: fixedInsuranceFeeRate,
		insuranceFeeRates:     nil,
		numPairedChunks:       3,
	})

	unitDelegationRewardPerRewardEpoch, _ := sdk.NewIntFromString("29999994000000000000")
	unitInsuranceCommissionPerRewardEpoch, _ := suite.getUnitDistribution(unitDelegationRewardPerRewardEpoch, fixedInsuranceFeeRate)

	provider := env.providers[0]
	targetInsurance := env.insurances[0]
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
		unitInsuranceCommissionPerRewardEpoch.Mul(sdk.NewInt(suite.rewardEpochCount)).String(),
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
	suite.mustPassInvariants()
}

func (suite *KeeperTestSuite) TestDoWithdrawInsuranceCommissionFail() {
	valAddrs, _ := suite.CreateValidators(
		[]int64{onePower, onePower, onePower},
		tenPercentFeeRate,
		nil,
	)
	_, oneInsurance := suite.app.LiquidStakingKeeper.GetMinimumRequirements(suite.ctx)
	providers, providerBalances := suite.AddTestAddrs(3, oneInsurance.Amount.Add(sdk.NewInt(100)))
	insurances := suite.provideInsurances(
		suite.ctx,
		providers,
		valAddrs,
		providerBalances,
		tenPercentFeeRate,
		nil,
	)
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
	suite.mustPassInvariants()
}

func (suite *KeeperTestSuite) TestDoDepositInsurance() {
	validators, _ := suite.CreateValidators(
		[]int64{1, 1, 1},
		tenPercentFeeRate,
		nil,
	)
	_, oneInsurance := suite.app.LiquidStakingKeeper.GetMinimumRequirements(suite.ctx)
	providers, _ := suite.AddTestAddrs(3, oneInsurance.Amount.Add(sdk.NewInt(100)))
	insurances := suite.provideInsurances(
		suite.ctx,
		providers,
		validators,
		[]sdk.Coin{oneInsurance, oneInsurance, oneInsurance},
		tenPercentFeeRate,
		nil,
	) // all providers still have 100 acanto after provide insurance

	msgDepositInsurance := types.NewMsgDepositInsurance(
		providers[0].String(),
		insurances[0].Id,
		sdk.NewCoin(oneInsurance.Denom, sdk.NewInt(100)),
	)

	err := suite.app.LiquidStakingKeeper.DoDepositInsurance(suite.ctx, msgDepositInsurance)
	suite.NoError(err)
	suite.mustPassInvariants()
}

func (suite *KeeperTestSuite) TestDoDepositInsuranceFail() {
	valAddrs, _ := suite.CreateValidators(
		[]int64{onePower, onePower, onePower},
		tenPercentFeeRate,
		nil,
	)
	_, oneInsurance := suite.app.LiquidStakingKeeper.GetMinimumRequirements(suite.ctx)
	providers, providerBalances := suite.AddTestAddrs(3, oneInsurance.Amount.Add(sdk.NewInt(100)))
	insurances := suite.provideInsurances(
		suite.ctx,
		providers,
		valAddrs,
		providerBalances,
		tenPercentFeeRate,
		nil,
	)
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
	suite.mustPassInvariants()
}

func (suite *KeeperTestSuite) TestRankInsurances() {
	env := suite.setupLiquidStakeTestingEnv(testingEnvOptions{
		desc:                  "TestRankInsurances",
		numVals:               3,
		fixedValFeeRate:       tenPercentFeeRate,
		valFeeRates:           nil,
		fixedPower:            onePower,
		powers:                nil,
		numInsurances:         3,
		fixedInsuranceFeeRate: tenPercentFeeRate,
		insuranceFeeRates:     nil,
		numPairedChunks:       3,
	})
	_, oneInsurance := suite.app.LiquidStakingKeeper.GetMinimumRequirements(suite.ctx)
	var idsOfPairedInsurances []uint64
	for _, insurance := range env.insurances {
		idsOfPairedInsurances = append(idsOfPairedInsurances, insurance.Id)
	}

	// INITIAL STATE: all paired chunks are working fine and there are no additional insurances yet
	newlyRankedInInsurances, rankOutInsurances, err := suite.app.LiquidStakingKeeper.RankInsurances(suite.ctx)
	suite.NoError(err)
	suite.Len(newlyRankedInInsurances, 0)
	suite.Len(rankOutInsurances, 0)

	suite.advanceHeight(1, "")

	// Cheap insurances which are competitive than current paired insurances are provided
	otherProviders, otherProviderBalances := suite.AddTestAddrs(3, oneInsurance.Amount)
	newInsurances := suite.provideInsurances(
		suite.ctx,
		otherProviders,
		env.valAddrs,
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
	suite.mustPassInvariants()
}

func (suite *KeeperTestSuite) TestEndBlocker() {
	env := suite.setupLiquidStakeTestingEnv(testingEnvOptions{
		desc:                  "TestEndBlocker",
		numVals:               3,
		fixedValFeeRate:       tenPercentFeeRate,
		valFeeRates:           nil,
		fixedPower:            onePower,
		powers:                nil,
		numInsurances:         3,
		fixedInsuranceFeeRate: tenPercentFeeRate,
		insuranceFeeRates:     nil,
		numPairedChunks:       3,
	})

	// Queue withdraw insurance request
	toBeWithdrawnInsurance, _ := suite.app.LiquidStakingKeeper.GetInsurance(suite.ctx, env.insurances[0].Id)
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

	_, oneInsurance := suite.app.LiquidStakingKeeper.GetMinimumRequirements(suite.ctx)
	newValAddrs, _ := suite.CreateValidators(
		[]int64{onePower, onePower, onePower},
		tenPercentFeeRate,
		nil,
	)
	newProviders, newProviderBalances := suite.AddTestAddrs(3, oneInsurance.Amount)
	newInsurances := suite.provideInsurances(
		suite.ctx,
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
		for _, insurance := range env.insurances {
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
		suite.ctx,
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

// TODO: Re-delegating validator has down-time slashing history, then shares are not equal to chunk size?
// But it should have same value with chunk size when converted to tokens. This part should be verified.
func (suite *KeeperTestSuite) TestPairedChunkTombstonedAndRedelegated() {
	env := suite.setupLiquidStakeTestingEnv(
		testingEnvOptions{
			"TestPairedChunkTombstonedAndRedelegated",
			3,
			tenPercentFeeRate,
			nil,
			onePower,
			nil,
			10,
			tenPercentFeeRate,
			nil,
			3,
		},
	)
	unitDelegationRewardPerRewardEpoch, _ := sdk.NewIntFromString("29999994000000000000")
	unitInsuranceCommissionPerRewardEpoch, _ := suite.getUnitDistribution(unitDelegationRewardPerRewardEpoch, tenPercentFeeRate)

	suite.advanceHeight(1, "liquid staking started")
	fmt.Println(suite.app.LiquidStakingKeeper.GetNetAmountState(suite.ctx))

	toBeTombstonedValidator := env.valAddrs[0]
	toBeTombstonedValidatorPubKey := env.pubKeys[0]
	toBeTombstonedChunk := env.pairedChunks[0]
	selfDelegationToken := suite.app.StakingKeeper.TokensFromConsensusPower(suite.ctx, onePower)
	// handle a signature to set signing info
	suite.app.SlashingKeeper.HandleValidatorSignature(
		suite.ctx,
		toBeTombstonedValidatorPubKey.Address(),
		selfDelegationToken.Int64(),
		true,
	)

	val := suite.app.StakingKeeper.Validator(suite.ctx, toBeTombstonedValidator)
	power := val.GetConsensusPower(suite.app.StakingKeeper.PowerReduction(suite.ctx))
	// TODO: We can control block height so we can check unpairing insurance covers the slashing or not in other TC.
	// infraction height should be before re-delegation to see it.
	// TODO: can refactor by using tombstone function?
	evidence := &evidencetypes.Equivocation{
		Height:           0,
		Time:             time.Unix(0, 0),
		Power:            power,
		ConsensusAddress: sdk.ConsAddress(toBeTombstonedValidatorPubKey.Address()).String(),
	}

	del, _ := suite.app.StakingKeeper.GetDelegation(
		suite.ctx,
		toBeTombstonedChunk.DerivedAddress(),
		toBeTombstonedValidator,
	)
	valTokensBeforeTombstoned := val.GetTokens()
	delTokens := val.TokensFromShares(del.GetShares())

	fmt.Println("DOUBLE SIGN SLASHING FOR VALIDATOR: " + toBeTombstonedValidator.String())
	suite.app.EvidenceKeeper.HandleEquivocationEvidence(suite.ctx, evidence)

	{
		valTombstoned := suite.app.StakingKeeper.Validator(suite.ctx, toBeTombstonedValidator)
		valTokensAfterTombstoned := valTombstoned.GetTokens()
		delTokensAfterTombstoned := valTombstoned.TokensFromShares(del.GetShares())
		valTokensDiff := valTokensBeforeTombstoned.Sub(valTokensAfterTombstoned)

		suite.Equal("250000050000000000000000", valTokensDiff.String())
		suite.Equal(
			valTokensBeforeTombstoned.ToDec().Mul(
				slashingtypes.DefaultSlashFractionDoubleSign,
			).TruncateInt(),
			valTokensDiff,
		)
		suite.Equal(
			types.ChunkSize.ToDec().Mul(slashingtypes.DefaultSlashFractionDoubleSign),
			delTokens.Sub(delTokensAfterTombstoned),
		)
		suite.True(
			suite.app.StakingKeeper.Validator(suite.ctx, toBeTombstonedValidator).IsJailed(),
			"validator must be jailed because it is tombstoned",
		)
		suite.True(
			suite.app.SlashingKeeper.IsTombstoned(
				suite.ctx, sdk.ConsAddress(toBeTombstonedValidatorPubKey.Address()),
			),
			"validator must be tombstoned",
		)
		suite.True(
			valTokensAfterTombstoned.LT(valTokensBeforeTombstoned),
			"double signing penalty must be applied",
		)
	}

	suite.advanceEpoch()
	suite.advanceHeight(1, "epoch reached after validator is tombstoned")
	fmt.Println(suite.app.LiquidStakingKeeper.GetNetAmountState(suite.ctx))
	passedRewardsEpochAfterTombstoned := int64(1)

	// check chunk is started to be re-paired with new insurances
	// and chunk delegation token value is recovered or not
	tombstonedChunk, _ := suite.app.LiquidStakingKeeper.GetChunk(suite.ctx, toBeTombstonedChunk.Id)
	{
		suite.Equal(
			env.insurances[4].Id,
			tombstonedChunk.PairedInsuranceId,
			"insurances[3] cannot be ranked in because it points to the tombstoned validator",
		)
		suite.Equal(types.CHUNK_STATUS_PAIRED, tombstonedChunk.Status)
		suite.Equal(toBeTombstonedChunk.PairedInsuranceId, tombstonedChunk.UnpairingInsuranceId)
		unpairingInsurance, _ := suite.app.LiquidStakingKeeper.GetInsurance(suite.ctx, tombstonedChunk.UnpairingInsuranceId)
		suite.Equal(
			unitInsuranceCommissionPerRewardEpoch.String(),
			suite.app.BankKeeper.GetBalance(
				suite.ctx,
				unpairingInsurance.FeePoolAddress(),
				env.bondDenom,
			).Amount.String(),
			fmt.Sprintf(
				"tombstoned insurance got commission for %d reward epochs",
				suite.rewardEpochCount-passedRewardsEpochAfterTombstoned,
			),
		)
		// Tombstoned validator got only 1 reward epoch commission because it is tombstoned before epoch is passed.
		// So other validator's delegation rewards are increased by the amount of tombstoned validator's delegation reward.
		numValidDels := int64(len(env.pairedChunks) - 1)
		additionalCommission := unitInsuranceCommissionPerRewardEpoch.Quo(sdk.NewInt(numValidDels))
		suite.Equal(
			unitInsuranceCommissionPerRewardEpoch.MulRaw(suite.rewardEpochCount).Add(additionalCommission).String(),
			suite.app.BankKeeper.GetBalance(
				suite.ctx,
				env.insurances[1].FeePoolAddress(),
				env.bondDenom,
			).Amount.String(),
			fmt.Sprintf(
				"normal insurance got (commission for %d reward epochs + "+
					"tombstoned delegation reward / number of valid delegations) "+
					"which means unit delegation reward is increased temporarily.\n"+
					"this is temporary because in this liquidstaking epoch, re-delegation happened so "+
					"every delegation reward will be same from now.",
				suite.rewardEpochCount,
			),
		)
	}
	newInsurance, _ := suite.app.LiquidStakingKeeper.GetInsurance(suite.ctx, tombstonedChunk.PairedInsuranceId)
	reDelegatedVal := suite.app.StakingKeeper.Validator(suite.ctx, newInsurance.GetValidator())
	// re-delegation obj must exist
	reDelegation, found := suite.app.StakingKeeper.GetRedelegation(
		suite.ctx,
		tombstonedChunk.DerivedAddress(),
		toBeTombstonedValidator,
		newInsurance.GetValidator(),
	)
	{
		suite.True(found, "re-delegation obj must exist")
		suite.Equal(types.ChunkSize.String(), reDelegation.Entries[0].InitialBalance.String())
		suite.Equal(types.ChunkSize.ToDec().String(), reDelegation.Entries[0].SharesDst.String())
		del, _ = suite.app.StakingKeeper.GetDelegation(
			suite.ctx,
			tombstonedChunk.DerivedAddress(),
			newInsurance.GetValidator(),
		)
		afterCovered := reDelegatedVal.TokensFromShares(del.GetShares())
		suite.Equal(types.ChunkSize.ToDec().String(), afterCovered.String())
	}
	suite.advanceHeight(1, "delegation rewards are accumulated")
	fmt.Println(suite.app.LiquidStakingKeeper.GetNetAmountState(suite.ctx))

	suite.advanceEpoch()
	suite.advanceHeight(1, "unpairing insurance because of tombstoned is unpaired now")
	fmt.Println(suite.app.LiquidStakingKeeper.GetNetAmountState(suite.ctx))

	{
		unpairedInsurance, _ := suite.app.LiquidStakingKeeper.GetInsurance(suite.ctx, tombstonedChunk.UnpairingInsuranceId)
		unpairedInsuranceVal, found := suite.app.StakingKeeper.GetValidator(suite.ctx, unpairedInsurance.GetValidator())
		suite.Equal(types.INSURANCE_STATUS_UNPAIRED, unpairedInsurance.Status)
		suite.Error(
			suite.app.LiquidStakingKeeper.IsValidValidator(suite.ctx, unpairedInsuranceVal, found),
			"validator of unpaired insurance is tombstoned",
		)
	}
}

func (suite *KeeperTestSuite) TestPairedChunkTombstonedAndUnpaired() {
	env := suite.setupLiquidStakeTestingEnv(
		testingEnvOptions{
			"TestPairedChunkTombstonedAndUnpaired",
			3,
			sdk.NewDecWithPrec(10, 2),
			nil,
			onePower,
			nil,
			4,
			sdk.NewDecWithPrec(10, 2),
			nil,
			3,
		},
	)
	unitDelegationRewardPerRewardEpoch, _ := sdk.NewIntFromString("29999994000000000000")
	unitInsuranceCommissionPerRewardEpoch, _ := suite.getUnitDistribution(unitDelegationRewardPerRewardEpoch, tenPercentFeeRate)

	suite.advanceHeight(1, "liquid staking started")
	fmt.Println(suite.app.LiquidStakingKeeper.GetNetAmountState(suite.ctx))

	toBeTombstonedValidator := env.valAddrs[0]
	toBeTombstonedValidatorPubKey := env.pubKeys[0]
	toBeTombstonedChunk := env.pairedChunks[0]
	pairedInsurance, _ := suite.app.LiquidStakingKeeper.GetInsurance(suite.ctx, toBeTombstonedChunk.PairedInsuranceId)
	{
		suite.Equal(
			toBeTombstonedValidator,
			env.insurances[0].GetValidator(),
			"insurance 0 will be unpaired",
		)
		suite.Equal(
			env.insurances[0].GetValidator(),
			env.insurances[3].GetValidator(),
			"in re-pairing process insurance 3 will never be ranked in because it also points to tombstoned validator",
		)
	}

	selfDelegationToken := suite.app.StakingKeeper.TokensFromConsensusPower(suite.ctx, onePower)
	// handle a signature to set signing info
	suite.app.SlashingKeeper.HandleValidatorSignature(
		suite.ctx,
		toBeTombstonedValidatorPubKey.Address(),
		selfDelegationToken.Int64(),
		true,
	)

	val := suite.app.StakingKeeper.Validator(suite.ctx, toBeTombstonedValidator)
	power := val.GetConsensusPower(suite.app.StakingKeeper.PowerReduction(suite.ctx))
	evidence := &evidencetypes.Equivocation{
		Height:           0,
		Time:             time.Unix(0, 0),
		Power:            power,
		ConsensusAddress: sdk.ConsAddress(toBeTombstonedValidatorPubKey.Address()).String(),
	}

	pairedInsuranceBalance := suite.app.BankKeeper.GetBalance(suite.ctx, pairedInsurance.DerivedAddress(), env.bondDenom)
	suite.app.EvidenceKeeper.HandleEquivocationEvidence(suite.ctx, evidence)

	suite.advanceHeight(1, "one block passed afetr validator is tombstoned because of double signing")
	fmt.Println(suite.app.LiquidStakingKeeper.GetNetAmountState(suite.ctx))

	suite.advanceEpoch()
	suite.advanceHeight(1, "chunk started to be unpairing")
	passedRewardsEpochAfterTombstoned := int64(2)

	pairedInsuranceBalanceAfterCoveringSlash := suite.app.BankKeeper.GetBalance(suite.ctx, pairedInsurance.DerivedAddress(), env.bondDenom)
	fmt.Println(suite.app.LiquidStakingKeeper.GetNetAmountState(suite.ctx))
	tombstonedChunk, _ := suite.app.LiquidStakingKeeper.GetChunk(suite.ctx, toBeTombstonedChunk.Id)
	pairedInsuranceBeforeSlashed, _ := suite.app.LiquidStakingKeeper.GetInsurance(suite.ctx, env.insurances[0].Id)
	candidateInsurance, _ := suite.app.LiquidStakingKeeper.GetInsurance(suite.ctx, env.insurances[3].Id)
	{
		suite.Equal(
			types.CHUNK_STATUS_UNPAIRING,
			tombstonedChunk.Status,
			"even though there was a one candidate insurance but + "+
				"that insurance also pointed to tombstoned validator",
		)
		suite.Equal(
			types.INSURANCE_STATUS_UNPAIRING,
			pairedInsuranceBeforeSlashed.Status,
			"insurance 0 is unpairing because it points to tombstoned validator",
		)
		suite.True(pairedInsuranceBalanceAfterCoveringSlash.IsLT(pairedInsuranceBalance))
		suite.Equal(
			types.INSURANCE_STATUS_PAIRING,
			candidateInsurance.Status,
			"insurance 3 is still in pairing status because it points to tombstoned validator, "+
				"so it couldn't join as a new paired insurance",
		)
		tombstonedInsurance, _ := suite.app.LiquidStakingKeeper.GetInsurance(suite.ctx, pairedInsuranceBeforeSlashed.Id)
		suite.Equal(
			unitInsuranceCommissionPerRewardEpoch.MulRaw(suite.rewardEpochCount-passedRewardsEpochAfterTombstoned).String(),
			suite.app.BankKeeper.GetBalance(suite.ctx, tombstonedInsurance.FeePoolAddress(), env.bondDenom).Amount.String(),
			fmt.Sprintf(
				"after tombstoned, tombstoned insurance couldn't get commissions corresponding %d * unit commission",
				passedRewardsEpochAfterTombstoned,
			),
		)
		validInsurancesAfterTombstoned := int64(2) // 3 -> 2 because insurance 0 got tombstoned validator
		additionalCommissionPerRewardEpoch := unitInsuranceCommissionPerRewardEpoch.QuoRaw(validInsurancesAfterTombstoned)
		suite.Equal(
			unitInsuranceCommissionPerRewardEpoch.MulRaw(suite.rewardEpochCount).Add(
				additionalCommissionPerRewardEpoch.MulRaw(passedRewardsEpochAfterTombstoned),
			).String(),
			suite.app.BankKeeper.GetBalance(suite.ctx, env.insurances[1].FeePoolAddress(), env.bondDenom).Amount.String(),
			"after tombstoned, valid insurance got additional commission because "+
				"validator set becomes small but rewards are fixed as 100 canto in testing environment",
		)

		unbondingDelegation, _ := suite.app.StakingKeeper.GetUnbondingDelegation(
			suite.ctx,
			tombstonedChunk.DerivedAddress(),
			val.GetOperator(),
		)
		suite.Len(
			unbondingDelegation.Entries,
			1,
			"there were no candidate insurance to pair, so unbonding of chunk started",
		)
		suite.Equal(
			unbondingDelegation.Entries[0].InitialBalance.String(),
			types.ChunkSize.String(),
			"there were no candidate insurance to pair, so unbonding of chunk started",
		)
	}

	suite.advanceHeight(1, "")
	fmt.Println(suite.app.LiquidStakingKeeper.GetNetAmountState(suite.ctx))

	suite.advanceEpoch()
	suite.advanceHeight(1, "unpairing of chunk is finished")

	{
		tombstonedChunkAfterUnpairing, _ := suite.app.LiquidStakingKeeper.GetChunk(suite.ctx, toBeTombstonedChunk.Id)
		suite.Equal(types.CHUNK_STATUS_PAIRING, tombstonedChunkAfterUnpairing.Status)
		suite.Equal(
			suite.app.BankKeeper.GetBalance(suite.ctx, tombstonedChunk.DerivedAddress(), env.bondDenom).Amount.String(),
			types.ChunkSize.String(),
			"chunk's balance must be equal to chunk size",
		)
		suite.Equal(
			types.Empty,
			tombstonedChunkAfterUnpairing.UnpairingInsuranceId,
			"unpairing insurance already finished its duty before chunk is unpaired",
		)
		unpairedInsurance, _ := suite.app.LiquidStakingKeeper.GetInsurance(suite.ctx, pairedInsurance.Id)
		suite.Equal(types.INSURANCE_STATUS_UNPAIRED, unpairedInsurance.Status)
	}
}

func (suite *KeeperTestSuite) TestMultiplePairedChunksTombstonedAndRedelegated() {
	env := suite.setupLiquidStakeTestingEnv(
		testingEnvOptions{
			"TestMultiplePairedChunksTombstonedAndRedelegated",
			3,
			sdk.NewDecWithPrec(10, 2),
			nil,
			onePower,
			nil,
			// insurance 0,3,6, will be invalid insurances
			// and insurance 10, 11, 13 will be replaced.
			14,
			sdk.NewDecWithPrec(10, 2),
			nil,
			9,
		},
	)
	_, oneInsurnace := suite.app.LiquidStakingKeeper.GetMinimumRequirements(suite.ctx)

	suite.advanceHeight(1, "liquid staking started")
	fmt.Println(suite.app.LiquidStakingKeeper.GetNetAmountState(suite.ctx))

	toBeTombstonedValidator := env.valAddrs[0]
	toBeTombstonedValidatorPubKey := env.pubKeys[0]
	toBeTombstonedChunks := []types.Chunk{env.pairedChunks[0], env.pairedChunks[3], env.pairedChunks[6]}
	pairedInsurances := []types.Insurance{env.insurances[0], env.insurances[3], env.insurances[6]}
	toBeNewlyRankedInsurances := []types.Insurance{env.insurances[10], env.insurances[11], env.insurances[13]}
	{
		// 0, 3, 6 are paired currently but will be unpaired because it points to toBeTombstonedValidator
		for i := 0; i < len(pairedInsurances); i++ {
			suite.Equal(pairedInsurances[i].Id, toBeTombstonedChunks[i].PairedInsuranceId)
			suite.Equal(toBeTombstonedValidator, pairedInsurances[i].GetValidator())
		}
		// 10, 11, 13 are not paired currently but will be paired because it points to valid validator
		for i := 0; i < len(toBeNewlyRankedInsurances); i++ {
			suite.NotEqual(toBeTombstonedValidator, toBeNewlyRankedInsurances[i].GetValidator())
		}
	}

	selfDelegationToken := suite.app.StakingKeeper.TokensFromConsensusPower(suite.ctx, onePower)
	// handle a signature to set signing info
	suite.app.SlashingKeeper.HandleValidatorSignature(
		suite.ctx,
		toBeTombstonedValidatorPubKey.Address(),
		selfDelegationToken.Int64(),
		true,
	)
	val := suite.app.StakingKeeper.Validator(suite.ctx, toBeTombstonedValidator)
	power := val.GetConsensusPower(suite.app.StakingKeeper.PowerReduction(suite.ctx))
	evidence := &evidencetypes.Equivocation{
		Height:           0,
		Time:             time.Unix(0, 0),
		Power:            power,
		ConsensusAddress: sdk.ConsAddress(toBeTombstonedValidatorPubKey.Address()).String(),
	}
	suite.app.EvidenceKeeper.HandleEquivocationEvidence(suite.ctx, evidence)
	suite.advanceHeight(1, "one block passed afetr validator is tombstoned because of double signing")
	fmt.Println(suite.app.LiquidStakingKeeper.GetNetAmountState(suite.ctx))

	suite.advanceEpoch()
	suite.advanceHeight(1, "re-pairing of chunks is finished")

	{
		for i, pairedInsuranceBeforeTombstoned := range pairedInsurances {
			tombstonedInsurance, _ := suite.app.LiquidStakingKeeper.GetInsurance(suite.ctx, pairedInsuranceBeforeTombstoned.Id)
			suite.Equal(types.INSURANCE_STATUS_UNPAIRING, tombstonedInsurance.Status)
			chunk, _ := suite.app.LiquidStakingKeeper.GetChunk(suite.ctx, tombstonedInsurance.ChunkId)
			suite.Equal(types.CHUNK_STATUS_PAIRED, chunk.Status)
			suite.Equal(tombstonedInsurance.Id, chunk.UnpairingInsuranceId)
			newInsurance, _ := suite.app.LiquidStakingKeeper.GetInsurance(suite.ctx, toBeNewlyRankedInsurances[i].Id)
			suite.Equal(types.INSURANCE_STATUS_PAIRED, newInsurance.Status)
			suite.Equal(newInsurance.Id, chunk.PairedInsuranceId)

			// check re-delegation happened or not
			reDelegation, found := suite.app.StakingKeeper.GetRedelegation(
				suite.ctx,
				chunk.DerivedAddress(),
				tombstonedInsurance.GetValidator(),
				newInsurance.GetValidator(),
			)
			suite.True(found)
			suite.Equal(types.ChunkSize.String(), reDelegation.Entries[0].InitialBalance.String())
			suite.Equal(types.ChunkSize.ToDec().String(), reDelegation.Entries[0].SharesDst.String())
			del, _ := suite.app.StakingKeeper.GetDelegation(
				suite.ctx,
				chunk.DerivedAddress(),
				newInsurance.GetValidator(),
			)
			suite.Equal(types.ChunkSize.ToDec().String(), del.GetShares().String())
			reDelegatedVal := suite.app.StakingKeeper.Validator(suite.ctx, newInsurance.GetValidator())
			suite.Equal(
				types.ChunkSize.ToDec().String(),
				reDelegatedVal.TokensFromShares(del.GetShares()).String(),
			)
		}
	}

	suite.advanceHeight(1, "")

	suite.advanceEpoch()
	suite.advanceHeight(1, "un-pairing insurances are unpaired")
	{
		for _, pairedInsuranceBeforeTombstoned := range pairedInsurances {
			// insurance finished duty
			unpairedInsurance, _ := suite.app.LiquidStakingKeeper.GetInsurance(suite.ctx, pairedInsuranceBeforeTombstoned.Id)
			suite.Equal(types.INSURANCE_STATUS_UNPAIRED, unpairedInsurance.Status)
			suite.True(
				suite.app.BankKeeper.GetBalance(
					suite.ctx,
					unpairedInsurance.DerivedAddress(),
					env.bondDenom,
				).IsLT(oneInsurnace),
				"it covered penalty at epoch",
			)
			chunk, _ := suite.app.LiquidStakingKeeper.GetChunk(suite.ctx, unpairedInsurance.ChunkId)
			suite.Equal(types.Empty, chunk.UnpairingInsuranceId)
		}
	}
}

func (suite *KeeperTestSuite) TestMultiplePairedChunksTombstonedAndUnpaired() {
	env := suite.setupLiquidStakeTestingEnv(
		testingEnvOptions{
			"TestMultiplePairedChunksTombstonedAndUnpaired",
			3,
			sdk.NewDecWithPrec(10, 2),
			nil,
			onePower,
			nil,
			9,
			sdk.NewDecWithPrec(10, 2),
			nil,
			9,
		},
	)
	suite.advanceHeight(1, "liquid staking started")
	fmt.Println(suite.app.LiquidStakingKeeper.GetNetAmountState(suite.ctx))
	toBeTombstonedValidator := env.valAddrs[0]
	toBeTombstonedValidatorPubKey := env.pubKeys[0]
	toBeTombstonedChunks := []types.Chunk{env.pairedChunks[0], env.pairedChunks[3], env.pairedChunks[6]}
	pairedInsurances := []types.Insurance{env.insurances[0], env.insurances[3], env.insurances[6]}
	{
		for i := 0; i < len(pairedInsurances); i++ {
			suite.Equal(pairedInsurances[i].Id, toBeTombstonedChunks[i].PairedInsuranceId)
			suite.Equal(toBeTombstonedValidator, pairedInsurances[i].GetValidator())
		}
	}

	selfDelegationToken := suite.app.StakingKeeper.TokensFromConsensusPower(suite.ctx, onePower)
	// handle a signature to set signing info
	suite.app.SlashingKeeper.HandleValidatorSignature(
		suite.ctx,
		toBeTombstonedValidatorPubKey.Address(),
		selfDelegationToken.Int64(),
		true,
	)
	val := suite.app.StakingKeeper.Validator(suite.ctx, toBeTombstonedValidator)
	power := val.GetConsensusPower(suite.app.StakingKeeper.PowerReduction(suite.ctx))
	evidence := &evidencetypes.Equivocation{
		Height:           0,
		Time:             time.Unix(0, 0),
		Power:            power,
		ConsensusAddress: sdk.ConsAddress(toBeTombstonedValidatorPubKey.Address()).String(),
	}
	var pairedInsuranceBalances []sdk.Coin
	for _, pairedInsurance := range pairedInsurances {
		pairedInsuranceBalances = append(
			pairedInsuranceBalances,
			suite.app.BankKeeper.GetBalance(suite.ctx, pairedInsurance.DerivedAddress(), env.bondDenom),
		)
	}

	suite.app.EvidenceKeeper.HandleEquivocationEvidence(suite.ctx, evidence)
	suite.advanceHeight(1, "one block passed afetr validator is tombstoned because of double signing")
	fmt.Println(suite.app.LiquidStakingKeeper.GetNetAmountState(suite.ctx))

	suite.advanceEpoch()
	suite.advanceHeight(1, "chunks started to be unpairing")

	var tombstonedChunks []types.Chunk
	for _, toBeTombstonedChunk := range toBeTombstonedChunks {
		tombstonedChunk, _ := suite.app.LiquidStakingKeeper.GetChunk(suite.ctx, toBeTombstonedChunk.Id)
		tombstonedChunks = append(tombstonedChunks, tombstonedChunk)
	}

	var pairedInsuranceBalancesAfterCoveringSlash []sdk.Coin
	for _, pairedInsurance := range pairedInsurances {
		pairedInsuranceBalancesAfterCoveringSlash = append(
			pairedInsuranceBalancesAfterCoveringSlash,
			suite.app.BankKeeper.GetBalance(suite.ctx, pairedInsurance.DerivedAddress(), env.bondDenom),
		)
	}

	for i, tombstonedChunk := range tombstonedChunks {
		pairedInsuranceBeforeSlashed, _ := suite.app.LiquidStakingKeeper.GetInsurance(suite.ctx, toBeTombstonedChunks[i].PairedInsuranceId)
		{
			suite.Equal(
				types.CHUNK_STATUS_UNPAIRING, tombstonedChunk.Status,
				"there are no candidate insurances so it started unpairing",
			)
			suite.Equal(
				types.INSURANCE_STATUS_UNPAIRING, pairedInsuranceBeforeSlashed.Status,
				fmt.Sprintf("insurance %d is unpairing because it points to tombstoned validator", i),
			)
			suite.True(pairedInsuranceBalancesAfterCoveringSlash[i].IsLT(pairedInsuranceBalances[i]))
			// get undelegation obj
			unbondingDelegation, _ := suite.app.StakingKeeper.GetUnbondingDelegation(
				suite.ctx,
				tombstonedChunk.DerivedAddress(),
				val.GetOperator(),
			)
			suite.Len(
				unbondingDelegation.Entries,
				1,
			)
			suite.Equal(
				unbondingDelegation.Entries[0].InitialBalance.String(),
				types.ChunkSize.String(),
			)
		}
	}

	suite.advanceHeight(1, "")
	fmt.Println(suite.app.LiquidStakingKeeper.GetNetAmountState(suite.ctx))

	suite.advanceEpoch()
	suite.advanceHeight(1, "unpairing of chunk is finished")

	{
		for i, toBeTombstonedChunk := range toBeTombstonedChunks {
			tombstonedChunkAfterUnpairing, _ := suite.app.LiquidStakingKeeper.GetChunk(suite.ctx, toBeTombstonedChunk.Id)
			suite.Equal(types.CHUNK_STATUS_PAIRING, tombstonedChunkAfterUnpairing.Status)
			suite.Equal(
				suite.app.BankKeeper.GetBalance(suite.ctx, tombstonedChunks[i].DerivedAddress(), env.bondDenom).Amount.String(),
				types.ChunkSize.String(),
				"chunk's balance must be equal to chunk size",
			)
			suite.Equal(
				types.Empty,
				tombstonedChunkAfterUnpairing.UnpairingInsuranceId,
				"unpairing insurance already finished its duty before chunk is unpaired",
			)
			unpairedInsurance, _ := suite.app.LiquidStakingKeeper.GetInsurance(suite.ctx, pairedInsurances[i].Id)
			suite.Equal(types.INSURANCE_STATUS_UNPAIRED, unpairedInsurance.Status)
		}
	}

}

// TODO: 2. TestMultiplePairedChunksTombstonedAndRepaired
// Some chunks can be re-paired but others can't which means there are some standards and we need to test it
func (suite *KeeperTestSuite) TestUnpairingForUnstakingChunkTombstoned() {
	env := suite.setupLiquidStakeTestingEnv(
		testingEnvOptions{
			"TestUnpairingForUnstakingChunkTombstoned",
			3,
			tenPercentFeeRate,
			nil,
			onePower,
			nil,
			3,
			tenPercentFeeRate,
			nil,
			3,
		},
	)
	numPassedRewardEpochsBeforeUnstaked := int64(0)
	suite.advanceHeight(1, "liquid staking started")
	numPassedRewardEpochsBeforeUnstaked++
	fmt.Println(suite.app.LiquidStakingKeeper.GetNetAmountState(suite.ctx))

	// Remember target chunk to be unstaked is a chunk which have most expensive insurance.
	// All insurance fees are fixed so last insurance is the target insurnace so
	// last chunk will be started to be unpairing for unstkaing.
	toBeUnstakedChunk := env.pairedChunks[2]
	toBeTombstonedValidator := env.valAddrs[2]
	toBeTombstonedValidatorPubKey := env.pubKeys[2]

	oneChunk, _ := suite.app.LiquidStakingKeeper.GetMinimumRequirements(suite.ctx)
	undelegator := env.delegators[0]
	undelegatorInitialBalance := suite.app.BankKeeper.GetBalance(suite.ctx, undelegator, env.bondDenom)
	msg := types.NewMsgLiquidUnstake(undelegator.String(), oneChunk)
	_, _, err := suite.app.LiquidStakingKeeper.QueueLiquidUnstake(suite.ctx, msg)
	suite.NoError(err)
	afterEscrowLsTokens := suite.app.BankKeeper.GetBalance(suite.ctx, undelegator, env.liquidBondDenom)

	suite.advanceEpoch()
	suite.advanceHeight(1, "unstaking started")
	numPassedRewardEpochsBeforeUnstaked++
	fmt.Println(suite.app.LiquidStakingKeeper.GetNetAmountState(suite.ctx))

	// 29999994 + 14999997(1 / num paired chunks)
	unitDelegationRewardPerRewardEpoch, _ := sdk.NewIntFromString("44999991000000000000")
	_, pureUnitRewardPerRewardEpoch := suite.getUnitDistribution(unitDelegationRewardPerRewardEpoch, tenPercentFeeRate)

	var pairedInsuranceBalanceAfterUnstakingStarted sdk.Coin
	var pairedInsuranceCommissionAfterUnstakingStarted sdk.Coin
	var escrowedLsTokens sdk.Coin
	{
		// check whether liquid unstaking started or not
		chunk, _ := suite.app.LiquidStakingKeeper.GetChunk(suite.ctx, toBeUnstakedChunk.Id)
		suite.Equal(types.CHUNK_STATUS_UNPAIRING_FOR_UNSTAKING, chunk.Status)
		info, _ := suite.app.LiquidStakingKeeper.GetUnpairingForUnstakingChunkInfo(suite.ctx, chunk.Id)
		suite.Equal(chunk.Id, info.ChunkId)
		escrowedLsTokens = info.EscrowedLstokens
		insurance, _ := suite.app.LiquidStakingKeeper.GetInsurance(suite.ctx, toBeUnstakedChunk.PairedInsuranceId)
		suite.Equal(types.INSURANCE_STATUS_UNPAIRING, insurance.Status)
		pairedInsuranceBalanceAfterUnstakingStarted = suite.app.BankKeeper.GetBalance(
			suite.ctx,
			insurance.DerivedAddress(),
			env.bondDenom,
		)
		pairedInsuranceCommissionAfterUnstakingStarted = suite.app.BankKeeper.GetBalance(
			suite.ctx,
			insurance.FeePoolAddress(),
			env.bondDenom,
		)

		unbondingDelegation, _ := suite.app.StakingKeeper.GetUnbondingDelegation(
			suite.ctx,
			chunk.DerivedAddress(),
			toBeTombstonedValidator,
		)
		suite.Len(unbondingDelegation.Entries, 1)
		suite.Equal(unbondingDelegation.Entries[0].InitialBalance.String(), types.ChunkSize.String())
	}

	suite.advanceHeight(1, "")
	fmt.Println(suite.app.LiquidStakingKeeper.GetNetAmountState(suite.ctx))

	selfDelegationToken := suite.app.StakingKeeper.TokensFromConsensusPower(suite.ctx, onePower)
	// handle a signature to set signing info
	suite.app.SlashingKeeper.HandleValidatorSignature(
		suite.ctx,
		toBeTombstonedValidatorPubKey.Address(),
		selfDelegationToken.Int64(),
		true,
	)
	val := suite.app.StakingKeeper.Validator(suite.ctx, toBeTombstonedValidator)
	power := val.GetConsensusPower(suite.app.StakingKeeper.PowerReduction(suite.ctx))
	evidence := &evidencetypes.Equivocation{
		Height:           0,
		Time:             time.Unix(0, 0),
		Power:            power,
		ConsensusAddress: sdk.ConsAddress(toBeTombstonedValidatorPubKey.Address()).String(),
	}

	fmt.Println("DOUBLE SIGN SLASHING FOR VALIDATOR: " + toBeTombstonedValidator.String())
	suite.app.EvidenceKeeper.HandleEquivocationEvidence(suite.ctx, evidence)
	suite.advanceHeight(1, "one block passed afetr validator is tombstoned because of double signing")
	fmt.Println(suite.app.LiquidStakingKeeper.GetNetAmountState(suite.ctx))

	var penalty sdk.Int
	{
		chunk, _ := suite.app.LiquidStakingKeeper.GetChunk(suite.ctx, toBeUnstakedChunk.Id)
		unbondingDelegation, _ := suite.app.StakingKeeper.GetUnbondingDelegation(
			suite.ctx,
			chunk.DerivedAddress(),
			toBeTombstonedValidator,
		)
		penalty = unbondingDelegation.Entries[0].InitialBalance.Sub(unbondingDelegation.Entries[0].Balance)
		suite.True(
			penalty.GT(sdk.ZeroInt()),
			"penalty applied to unbonding delegation "+
				"but insurance not yet covered because epoch not yet reached",
		)
		insurance, _ := suite.app.LiquidStakingKeeper.GetInsurance(suite.ctx, chunk.UnpairingInsuranceId)
		suite.Equal(
			pairedInsuranceBalanceAfterUnstakingStarted,
			suite.app.BankKeeper.GetBalance(suite.ctx, insurance.DerivedAddress(), env.bondDenom),
			"insurance not yet covered penalty because epoch not yet reached",
		)
	}

	rewardPoolBalanceBefore := suite.app.BankKeeper.GetBalance(suite.ctx, types.RewardPool, env.bondDenom)
	suite.advanceEpoch()
	suite.advanceHeight(1, "epoch reached after validator is tombstoned because of double signing")
	fmt.Println(suite.app.LiquidStakingKeeper.GetNetAmountState(suite.ctx))
	rewardPoolBalanceAfter := suite.app.BankKeeper.GetBalance(suite.ctx, types.RewardPool, env.bondDenom)

	{
		_, found := suite.app.LiquidStakingKeeper.GetChunk(suite.ctx, toBeUnstakedChunk.Id)
		suite.False(found, "liquid unstaking of chunk is finished")
		undelegatorBalance := suite.app.BankKeeper.GetBalance(suite.ctx, undelegator, env.bondDenom)
		suite.Equal(
			types.ChunkSize.Sub(penalty).String(),
			undelegatorBalance.Sub(undelegatorInitialBalance).Amount.String(),
			"undelegator got (chunk size - penalty) tokens after unstaking",
		)
		rewardAfter := rewardPoolBalanceAfter.Sub(rewardPoolBalanceBefore).Amount
		expectedRewardAfter := penalty.Add(
			pureUnitRewardPerRewardEpoch.MulRaw(2).MulRaw(suite.rewardEpochCount - numPassedRewardEpochsBeforeUnstaked),
		)
		// TODO: remove this margin error
		suite.Equal(
			"8099991000000",
			expectedRewardAfter.Sub(rewardAfter).String(),
			"penalty is sent to reward pool also, by the way there are very small margin error because "+
				"during the test, there were a moment when validator power is 1 because of unbonding",
		)
		insurance, _ := suite.app.LiquidStakingKeeper.GetInsurance(suite.ctx, toBeUnstakedChunk.PairedInsuranceId)
		suite.Equal(types.INSURANCE_STATUS_UNPAIRED, insurance.Status)
		balance := suite.app.BankKeeper.GetBalance(suite.ctx, insurance.DerivedAddress(), env.bondDenom)
		suite.Equal(
			penalty.String(),
			pairedInsuranceBalanceAfterUnstakingStarted.Sub(balance).Amount.String(),
			"insurance covered penalty after epoch reached",
		)
		penaltyRatio := penalty.ToDec().Quo(types.ChunkSize.ToDec())
		discounted := penaltyRatio.Mul(escrowedLsTokens.Amount.ToDec())
		afterFinishUnbonding := suite.app.BankKeeper.GetBalance(suite.ctx, undelegator, env.liquidBondDenom)
		suite.Equal(
			discounted.TruncateInt().String(),
			afterFinishUnbonding.Sub(afterEscrowLsTokens).Amount.String(),
			"discounted liquid staking tokens are sent to undelegator",
		)

		commission := suite.app.BankKeeper.GetBalance(suite.ctx, insurance.FeePoolAddress(), env.bondDenom)
		suite.Equal(
			commission.String(),
			pairedInsuranceCommissionAfterUnstakingStarted.String(),
			"insurance commission is not affected by slashing",
		)
	}
}

// TestCumulativePenaltyByMultipleDownTimeSlashingAndTombstone tests how much penalty is applied to chunk
// when there were maximum possible downtime slashing (+ tombstone).
// To estimate the appropriate amount of insurance collateral, we should test worst-case scenarios.
func (suite *KeeperTestSuite) TestCumulativePenaltyByMultipleDownTimeSlashingAndTombstone() {
	tcs := []struct {
		name string
		// blockTime is the time passed between two blocks
		blockTime              time.Duration
		includeTombstone       bool
		expectedPenaltyPercent int
	}{
		{
			name:                   "block time is 1 second",
			blockTime:              time.Second,
			includeTombstone:       false,
			expectedPenaltyPercent: 61,
		},
		{
			name:                   "block time is 1 second including tombstone",
			blockTime:              time.Second,
			includeTombstone:       true,
			expectedPenaltyPercent: 64,
		},
		{
			name:                   "block time is 5 second",
			blockTime:              5 * time.Second,
			includeTombstone:       false,
			expectedPenaltyPercent: 18,
		},

		{
			name:                   "block time is 5 second including tombstone",
			blockTime:              5 * time.Second,
			includeTombstone:       true,
			expectedPenaltyPercent: 22,
		},
	}
	for _, tc := range tcs {
		initialHeight := int64(1)
		suite.ctx = suite.ctx.WithBlockHeight(initialHeight) // start with clean height
		valNum := 2
		delAddrs, _ := suite.AddTestAddrs(valNum, suite.app.StakingKeeper.TokensFromConsensusPower(suite.ctx, 200))
		valAddrs := simapp.ConvertAddrsToValAddrs(delAddrs)
		pubKeys := suite.createTestPubKeys(valNum)
		tstaking := teststaking.NewHelper(suite.T(), suite.ctx, suite.app.StakingKeeper)
		tstaking.Denom = suite.app.StakingKeeper.BondDenom(suite.ctx)
		power := int64(100)
		selfDelegations := make([]sdk.Int, valNum)
		// create validators which have the same power
		for i, valAddr := range valAddrs {
			selfDelegations[i] = tstaking.CreateValidatorWithValPower(valAddr, pubKeys[i], power, true)
		}
		staking.EndBlocker(suite.ctx, suite.app.StakingKeeper)

		// Let's create 2 chunk and 2 insurance
		oneChunk, oneInsurance := suite.app.LiquidStakingKeeper.GetMinimumRequirements(suite.ctx)
		providers, providerBalances := suite.AddTestAddrs(2, oneInsurance.Amount)
		suite.provideInsurances(suite.ctx, providers, valAddrs, providerBalances, tenPercentFeeRate, nil)
		delegators, delegatorBalances := suite.AddTestAddrs(2, oneChunk.Amount)
		pairedChunks := suite.liquidStakes(suite.ctx, delegators, delegatorBalances)
		suite.ctx = suite.ctx.WithBlockHeight(suite.ctx.BlockHeight() + 1)
		staking.EndBlocker(suite.ctx, suite.app.StakingKeeper)

		downValAddr := valAddrs[0]
		downValPubKey := pubKeys[0]
		toBeUnpairedChunk := pairedChunks[0]
		toBeDrainedInsuranceId := pairedChunks[0].PairedInsuranceId
		fmt.Printf("balance of insurance to be drained: %s\n", oneInsurance.Amount.String())

		epoch := suite.app.LiquidStakingKeeper.GetEpoch(suite.ctx)
		epochTime := suite.ctx.BlockTime().Add(epoch.Duration)
		called := 0
		for {
			validator, _ := suite.app.StakingKeeper.GetValidatorByConsAddr(suite.ctx, sdk.GetConsAddress(downValPubKey))
			suite.downTimeSlashing(
				suite.ctx,
				downValPubKey,
				validator.GetConsensusPower(suite.app.StakingKeeper.PowerReduction(suite.ctx)),
				called,
				tc.blockTime,
			)
			suite.unjail(suite.ctx, downValAddr, downValPubKey, tc.blockTime)
			called++

			if suite.ctx.BlockTime().After(epochTime) {
				break
			}
		}
		if tc.includeTombstone {
			suite.tombstone(suite.ctx, downValAddr, downValPubKey)
		}

		validatorAfterSlashed, _ := suite.app.StakingKeeper.GetValidatorByConsAddr(suite.ctx, sdk.GetConsAddress(downValPubKey))
		cumulativePenalty := types.ChunkSize.ToDec().Sub(validatorAfterSlashed.TokensFromShares(types.ChunkSize.ToDec()))
		fmt.Printf("%d downtime slashing occurred during epoch(%0.f days)\n", called, epoch.Duration.Hours()/24)
		damagedPercent := cumulativePenalty.Quo(types.ChunkSize.ToDec()).MulInt64(100).TruncateInt64()
		suite.Equal(tc.expectedPenaltyPercent, int(damagedPercent))
		fmt.Printf(
			"accumulated penalty: %s | %d percent of ChunkSize tokens\n",
			cumulativePenalty.String(), damagedPercent,
		)
		suite.advanceEpoch()
		staking.EndBlocker(suite.ctx, suite.app.StakingKeeper)
		liquidstakingkeeper.EndBlocker(suite.ctx, suite.app.LiquidStakingKeeper)
		fmt.Println("chunk unbonding is started")
		{
			unPairingChunk, _ := suite.app.LiquidStakingKeeper.GetChunk(suite.ctx, toBeUnpairedChunk.Id)
			unpairingInsurance, _ := suite.app.LiquidStakingKeeper.GetInsurance(suite.ctx, toBeDrainedInsuranceId)
			suite.Equal(
				types.CHUNK_STATUS_UNPAIRING,
				unPairingChunk.Status,
				"chunk unbonding is started",
			)
			ubd, _ := suite.app.StakingKeeper.GetUnbondingDelegation(
				suite.ctx,
				unPairingChunk.DerivedAddress(),
				unpairingInsurance.GetValidator(),
			)
			suite.Len(ubd.Entries, 1)
			suite.Equal(
				types.ChunkSize.Sub(cumulativePenalty.Ceil().TruncateInt()).String(),
				ubd.Entries[0].InitialBalance.String(),
				"it is slashed so when unbonding, initial balance is less than chunk size tokens",
			)
		}

		rewardModuleAccBalance := suite.app.BankKeeper.GetBalance(suite.ctx, types.RewardPool, suite.denom)
		suite.advanceEpoch()
		staking.EndBlocker(suite.ctx, suite.app.StakingKeeper)
		liquidstakingkeeper.EndBlocker(suite.ctx, suite.app.LiquidStakingKeeper)
		fmt.Println("chunk unbonding is finished")
		rewardModuleAccBalanceAfter := suite.app.BankKeeper.GetBalance(suite.ctx, types.RewardPool, suite.denom)
		suite.True(
			rewardModuleAccBalanceAfter.Amount.GT(rewardModuleAccBalance.Amount),
		)
		diff := rewardModuleAccBalanceAfter.Amount.Sub(rewardModuleAccBalance.Amount)
		fmt.Printf("reward module account balance increased by %s\n", diff.String())
		unpairingInsurance, _ := suite.app.LiquidStakingKeeper.GetInsurance(suite.ctx, toBeDrainedInsuranceId)
		unpairingInsuranceBalance := suite.app.BankKeeper.GetBalance(suite.ctx, unpairingInsurance.DerivedAddress(), suite.denom)
		suite.True(unpairingInsuranceBalance.IsZero(),
			"unpairing insurance is used all of its balance to cover penalty by"+
				"sending it to reward pool",
		)
	}
}

func (suite *KeeperTestSuite) downTimeSlashing(
	ctx sdk.Context, downValPubKey cryptotypes.PubKey, power int64, called int, blockTime time.Duration,
) (penalty sdk.Int) {
	validator, _ := suite.app.StakingKeeper.GetValidatorByConsAddr(suite.ctx, sdk.GetConsAddress(downValPubKey))
	valTokens := validator.GetTokens()
	expectedPenalty := suite.expectedPenalty(
		suite.ctx,
		power,
		suite.app.SlashingKeeper.SlashFractionDowntime(suite.ctx),
	)

	height := suite.ctx.BlockHeader().Height
	window := suite.app.SlashingKeeper.SignedBlocksWindow(suite.ctx)
	i := height
	for ; i <= height+window; i++ {
		suite.ctx = suite.ctx.WithBlockHeight(i).WithBlockTime(suite.ctx.BlockTime().Add(blockTime))
		suite.app.SlashingKeeper.HandleValidatorSignature(suite.ctx, downValPubKey.Address(), power, true)
	}
	min := suite.app.SlashingKeeper.MinSignedPerWindow(ctx)
	height = suite.ctx.BlockHeight()
	for ; i <= height+min+1; i++ {
		suite.ctx = suite.ctx.WithBlockHeight(i).WithBlockTime(suite.ctx.BlockTime().Add(blockTime))
		suite.app.SlashingKeeper.HandleValidatorSignature(suite.ctx, downValPubKey.Address(), power, false)
	}

	updates := staking.EndBlocker(suite.ctx, suite.app.StakingKeeper)
	jailedOrNot := false
	for _, update := range updates {
		if bytes.Equal(update.PubKey.GetEd25519(), downValPubKey.Bytes()) && update.Power == 0 {
			jailedOrNot = true
			break
		}
	}

	suite.Equal(true, jailedOrNot, fmt.Sprintf("called-%d validator should have been jailed", called))
	// validator should have been jailed and slashed
	validator, _ = suite.app.StakingKeeper.GetValidatorByConsAddr(suite.ctx, sdk.GetConsAddress(downValPubKey))
	valTokensAfter := validator.GetTokens()
	suite.Equal(stakingtypes.Unbonding, validator.GetStatus())
	penalty = valTokens.Sub(valTokensAfter)
	suite.Equal(expectedPenalty.String(), penalty.String(), fmt.Sprintf("called: %d", called))
	return
}

func (suite *KeeperTestSuite) tombstone(
	ctx sdk.Context, valAddr sdk.ValAddress, valPubKey cryptotypes.PubKey,
) {
	validator := suite.app.StakingKeeper.Validator(ctx, valAddr)
	power := validator.GetConsensusPower(suite.app.StakingKeeper.PowerReduction(ctx))
	evidence := &evidencetypes.Equivocation{
		Height:           0,
		Time:             time.Unix(0, 0),
		Power:            power,
		ConsensusAddress: sdk.ConsAddress(valPubKey.Address()).String(),
	}
	fmt.Println("DOUBLE SIGN SLASHING FOR VALIDATOR: " + valAddr.String())
	suite.app.EvidenceKeeper.HandleEquivocationEvidence(ctx, evidence)

	suite.True(
		suite.app.StakingKeeper.Validator(ctx, valAddr).IsJailed(),
		"validator must be jailed because it is tombstoned",
	)
	suite.True(
		suite.app.SlashingKeeper.IsTombstoned(ctx, sdk.ConsAddress(valPubKey.Address())),
		"validator must be tombstoned",
	)

}

func (suite *KeeperTestSuite) unjail(ctx sdk.Context, valAddr sdk.ValAddress, pubKey cryptotypes.PubKey, blockTime time.Duration) {
	jailDuration := suite.app.SlashingKeeper.GetParams(suite.ctx).DowntimeJailDuration
	blockNum := int64(jailDuration / blockTime)
	suite.ctx = suite.ctx.WithBlockHeight(
		suite.ctx.BlockHeight() + blockNum,
	).WithBlockTime(
		suite.ctx.BlockTime().Add(jailDuration),
	)
	suite.NoError(suite.app.SlashingKeeper.Unjail(suite.ctx, valAddr))
	updates := staking.EndBlocker(suite.ctx, suite.app.StakingKeeper)
	suite.Len(updates, 1, "validator should have been bonded again")
	suite.Equal(pubKey.Bytes(), updates[0].PubKey.GetEd25519(), "validator is bonded again!")
}

func (suite *KeeperTestSuite) expectedPenalty(ctx sdk.Context, power int64, slashFactor sdk.Dec) sdk.Int {
	amount := suite.app.StakingKeeper.TokensFromConsensusPower(ctx, power)
	slashAmountDec := amount.ToDec().Mul(slashFactor)
	return slashAmountDec.TruncateInt()
}

func (suite *KeeperTestSuite) getUnitDistribution(
	unitDelegationRewardPerRewardEpoch sdk.Int,
	fixedInsuranceFeeRate sdk.Dec,
) (sdk.Int, sdk.Int) {
	unitInsuranceCommissionPerRewardEpoch := unitDelegationRewardPerRewardEpoch.ToDec().Mul(fixedInsuranceFeeRate).TruncateInt()
	pureUnitRewardPerRewardEpoch := unitDelegationRewardPerRewardEpoch.Sub(unitInsuranceCommissionPerRewardEpoch)
	fmt.Println("unitDelegationRewardPerRewardEpoch: ", unitDelegationRewardPerRewardEpoch.String())
	fmt.Println("unitInsuranceCommissionPerRewardEpoch: ", unitInsuranceCommissionPerRewardEpoch.String())
	fmt.Println("pureUnitRewardPerRewardEpoch: ", pureUnitRewardPerRewardEpoch.String())
	return unitInsuranceCommissionPerRewardEpoch, pureUnitRewardPerRewardEpoch
}
