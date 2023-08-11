package keeper_test

import (
	"bytes"
	"fmt"
	"math/rand"
	"time"

	liquidstakingkeeper "github.com/Canto-Network/Canto/v7/x/liquidstaking"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/simapp"
	"github.com/cosmos/cosmos-sdk/x/staking"
	"github.com/cosmos/cosmos-sdk/x/staking/teststaking"

	"github.com/Canto-Network/Canto/v7/x/liquidstaking/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/simulation"
	evidencetypes "github.com/cosmos/cosmos-sdk/x/evidence/types"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	ethermint "github.com/evmos/ethermint/types"
)

var onePower int64 = 1
var TenPercentFeeRate = sdk.NewDecWithPrec(10, 2)
var FivePercentFeeRate = sdk.NewDecWithPrec(5, 2)
var OnePercentFeeRate = sdk.NewDecWithPrec(1, 2)

// fundingAccount is a rich account.
// Any accounts created during tests except validators must get funding from this account.
var fundingAccount = sdk.AccAddress(ed25519.GenPrivKey().PubKey().Address())

type insuranceState struct {
	// The sum of all insurances' amount (= DerivedAddress(Insurance.Id).Balance)
	TotalInsuranceTokens sdk.Int
	// The sum of all paired insurances' amount (=
	// DerivedAddress(Insurance.Id).Balance)
	TotalPairedInsuranceTokens sdk.Int
	// The sum of all unpairing insurances' amount (=
	// DerivedAddress(Insurance.Id).Balance)
	TotalUnpairingInsuranceTokens sdk.Int
	// The cumulative commissions of all insurances
	TotalRemainingInsuranceCommissions sdk.Dec
}

func (suite *KeeperTestSuite) getInsuranceState(ctx sdk.Context) insuranceState {
	// fill insurance state
	bondDenom := suite.app.StakingKeeper.BondDenom(ctx)
	totalPairedInsuranceTokens, totalUnpairingInsuranceTokens, totalInsuranceTokens := sdk.ZeroInt(), sdk.ZeroInt(), sdk.ZeroInt()
	totalRemainingInsuranceCommissions := sdk.ZeroDec()
	suite.app.LiquidStakingKeeper.IterateAllInsurances(ctx, func(insurance types.Insurance) (stop bool) {
		insuranceBalance := suite.app.BankKeeper.GetBalance(ctx, insurance.DerivedAddress(), bondDenom)
		commission := suite.app.BankKeeper.GetBalance(ctx, insurance.FeePoolAddress(), bondDenom)
		switch insurance.Status {
		case types.INSURANCE_STATUS_PAIRED:
			totalPairedInsuranceTokens = totalPairedInsuranceTokens.Add(insuranceBalance.Amount)
		case types.INSURANCE_STATUS_UNPAIRING:
			totalUnpairingInsuranceTokens = totalUnpairingInsuranceTokens.Add(insuranceBalance.Amount)
		}
		totalInsuranceTokens = totalInsuranceTokens.Add(insuranceBalance.Amount)
		totalRemainingInsuranceCommissions = totalRemainingInsuranceCommissions.Add(commission.Amount.ToDec())
		return false
	})
	return insuranceState{
		totalInsuranceTokens,
		totalPairedInsuranceTokens,
		totalUnpairingInsuranceTokens,
		totalRemainingInsuranceCommissions,
	}
}

func (suite *KeeperTestSuite) getPairedChunks() []types.Chunk {
	var pairedChunks []types.Chunk
	suite.app.LiquidStakingKeeper.IterateAllChunks(suite.ctx, func(chunk types.Chunk) bool {
		if chunk.Status == types.CHUNK_STATUS_PAIRED {
			pairedChunks = append(pairedChunks, chunk)
		}
		return false
	})
	return pairedChunks
}

func (suite *KeeperTestSuite) getUnpairingForUnstakingChunks() []types.Chunk {
	var UnpairingForUnstakingChunks []types.Chunk
	suite.app.LiquidStakingKeeper.IterateAllChunks(suite.ctx, func(chunk types.Chunk) bool {
		if chunk.Status == types.CHUNK_STATUS_UNPAIRING_FOR_UNSTAKING {
			UnpairingForUnstakingChunks = append(UnpairingForUnstakingChunks, chunk)
		}
		return false
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
		msg := types.NewMsgProvideInsurance(provider.String(), valAddrs[i%valNum].String(), amounts[i], sdk.ZeroDec())
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
		TenPercentFeeRate,
		nil,
	)
	suite.fundAccount(suite.ctx, fundingAccount, types.ChunkSize.MulRaw(500))
	_, oneInsurance := suite.app.LiquidStakingKeeper.GetMinimumRequirements(suite.ctx)
	providers, _ := suite.AddTestAddrsWithFunding(fundingAccount, 10, oneInsurance.Amount)

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
				FeeRate:          TenPercentFeeRate,
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
				FeeRate:          TenPercentFeeRate,
			},
			nil,
			"amount must be greater than minimum collateral",
		},
		{
			"fee rate >= maximum fee rate",
			&types.MsgProvideInsurance{
				ProviderAddress:  providers[0].String(),
				ValidatorAddress: valAddrs[0].String(),
				Amount:           oneInsurance,
				FeeRate:          TenPercentFeeRate.MulInt(sdk.NewInt(4)), // vFee 10% + 40% = 50%
			},
			nil,
			"fee rate(validator fee rate + insurance fee rate) must be less than 0.500000000000000000",
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
		TenPercentFeeRate,
		nil,
	)
	suite.fundAccount(suite.ctx, fundingAccount, types.ChunkSize.MulRaw(500))
	oneChunk, oneInsurance := suite.app.LiquidStakingKeeper.GetMinimumRequirements(suite.ctx)
	providers, balances := suite.AddTestAddrsWithFunding(fundingAccount, 10, oneInsurance.Amount)
	suite.provideInsurances(suite.ctx, providers, valAddrs, balances, sdk.ZeroDec(), nil)

	delegators, balances := suite.AddTestAddrsWithFunding(fundingAccount, 10, oneChunk.Amount)
	nase := suite.app.LiquidStakingKeeper.GetNetAmountStateEssentials(suite.ctx)

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
		suite.app.LiquidStakingKeeper.IterateAllChunks(suite.ctx, func(chunk types.Chunk) bool {
			suite.True(chunk.Equal(createdChunks[idx]))
			idx++
			return false
		})
		suite.Equal(len(createdChunks), idx, "number of created chunks should be equal to number of chunks in db")
	}

	lsTokenAfter := suite.app.BankKeeper.GetBalance(suite.ctx, del1, liquidBondDenom)
	{
		suite.NoError(err)
		suite.True(amt1.Amount.Equal(newShares.TruncateInt()), "delegation shares should be equal to amount")
		suite.True(amt1.Amount.Equal(lsTokenMintAmount), "at first try mint rate is 1, so mint amount should be equal to amount")
		suite.True(lsTokenAfter.Sub(lsTokenBefore).Amount.Equal(lsTokenMintAmount), "liquid staker must have minted ls tokens in account balance")
	}

	// NetAmountStateEssentials should be updated correctly
	afterNas := suite.app.LiquidStakingKeeper.GetNetAmountStateEssentials(suite.ctx)
	{
		suite.True(afterNas.LsTokensTotalSupply.Equal(lsTokenMintAmount), "total ls token supply should be equal to minted amount")
		suite.True(nase.TotalLiquidTokens.Add(amt1.Amount).Equal(afterNas.TotalLiquidTokens))
		suite.True(nase.NetAmount.Add(amt1.Amount.ToDec()).Equal(afterNas.NetAmount))
		suite.True(afterNas.MintRate.Equal(sdk.OneDec()), "no rewards yet, so mint rate should be 1")
	}
	suite.mustPassInvariants()
}

func (suite *KeeperTestSuite) TestLiquidStakeFail() {
	suite.resetEpochs()
	valAddrs, _ := suite.CreateValidators(
		[]int64{onePower, onePower, onePower},
		TenPercentFeeRate,
		nil,
	)
	oneChunk, oneInsurance := suite.app.LiquidStakingKeeper.GetMinimumRequirements(suite.ctx)
	suite.fundAccount(suite.ctx, fundingAccount, oneChunk.Amount.MulRaw(100).Add(oneInsurance.Amount.MulRaw(10)))
	nase := suite.app.LiquidStakingKeeper.GetNetAmountStateEssentials(suite.ctx)
	remainingChunkSlots := nase.RemainingChunkSlots
	suite.Equal(
		remainingChunkSlots, sdk.NewInt(10),
		"set total supply by creating funding account to fix max paired chunks",
	)
	addrs, balances := suite.AddTestAddrsWithFunding(fundingAccount, int(remainingChunkSlots.SubRaw(1).Int64()), oneChunk.Amount)

	// TC: There are no pairing insurances yet. Insurances must be provided to liquid stake
	acc1 := addrs[0]
	msg := types.NewMsgLiquidStake(acc1.String(), oneChunk)
	_, _, _, err := suite.app.LiquidStakingKeeper.DoLiquidStake(suite.ctx, msg)
	suite.ErrorContains(err, types.ErrNoPairingInsurance.Error())

	providers, providerBalances := suite.AddTestAddrsWithFunding(fundingAccount, int(remainingChunkSlots.Int64()), oneInsurance.Amount)
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
	suite.app.BankKeeper.SendCoins(
		suite.ctx,
		fundingAccount,
		acc1,
		sdk.NewCoins(
			sdk.NewCoin(suite.denom, types.ChunkSize.Mul(sdk.NewInt(2))),
		),
	)
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
	newAddrs, newBalances := suite.AddTestAddrsWithFunding(fundingAccount, 1, oneChunk.Amount)
	msg = types.NewMsgLiquidStake(newAddrs[0].String(), newBalances[0])
	_, _, _, err = suite.app.LiquidStakingKeeper.DoLiquidStake(suite.ctx, msg)
	suite.ErrorIs(err, types.ErrExceedAvailableChunks)

	suite.mustPassInvariants()
}

func (suite *KeeperTestSuite) TestLiquidStakeWithAdvanceBlocks() {
	fixedInsuranceFeeRate := TenPercentFeeRate
	env := suite.setupLiquidStakeTestingEnv(
		testingEnvOptions{
			"TestLiquidStakeWithAdvanceBlocks",
			3,
			TenPercentFeeRate,
			nil,
			onePower,
			nil,
			10,
			fixedInsuranceFeeRate,
			nil,
			3,
			types.ChunkSize.MulRaw(500),
		},
	)

	_, oneInsurance := suite.app.LiquidStakingKeeper.GetMinimumRequirements(suite.ctx)
	nase := suite.app.LiquidStakingKeeper.GetNetAmountStateEssentials(suite.ctx)
	is := suite.getInsuranceState(suite.ctx)
	pairedChunksInt := sdk.NewInt(int64(len(env.pairedChunks)))
	// 1 chunk size * number of paired chunks (=3) tokens are liquidated
	currentLiquidatedTokens := types.ChunkSize.Mul(pairedChunksInt)
	currentInsuranceTokens := oneInsurance.Amount.Mul(pairedChunksInt)
	{
		suite.True(nase.Equal(types.NetAmountStateEssentials{
			MintRate:                    sdk.OneDec(),
			LsTokensTotalSupply:         currentLiquidatedTokens,
			NetAmount:                   currentLiquidatedTokens.ToDec(),
			TotalLiquidTokens:           currentLiquidatedTokens,
			RewardModuleAccBalance:      sdk.ZeroInt(),
			FeeRate:                     sdk.ZeroDec(),
			UtilizationRatio:            sdk.MustNewDecFromStr("0.005999999856000003"),
			RemainingChunkSlots:         sdk.NewInt(47),
			DiscountRate:                sdk.ZeroDec(),
			TotalDelShares:              currentLiquidatedTokens.ToDec(),
			TotalRemainingRewards:       sdk.ZeroDec(),
			TotalChunksBalance:          sdk.ZeroInt(),
			TotalUnbondingChunksBalance: sdk.ZeroInt(),
			NumPairedChunks:             sdk.NewInt(3),
		}), "no epoch(=1 block in test) processed yet, so there are no mint rate change and remaining rewards yet")
		// check insurnaceState
		suite.Equal(insuranceState{
			TotalInsuranceTokens:               oneInsurance.Amount.Mul(sdk.NewInt(int64(len(env.insurances)))),
			TotalPairedInsuranceTokens:         currentInsuranceTokens,
			TotalUnpairingInsuranceTokens:      sdk.ZeroInt(),
			TotalRemainingInsuranceCommissions: sdk.ZeroDec(),
		}, is)
	}

	suite.ctx = suite.advanceHeight(suite.ctx, 1, "")
	beforeNas := nase
	nase = suite.app.LiquidStakingKeeper.GetNetAmountStateEssentials(suite.ctx)
	{
		suite.Equal(
			"80999676001295325000.000000000000000000",
			nase.TotalRemainingRewards.Sub(beforeNas.TotalRemainingRewards).String(),
		)
		suite.Equal("0.999892012094645400", nase.MintRate.String())
	}

	beforeNas = nase
	beforeIs := is
	suite.ctx = suite.advanceEpoch(suite.ctx)
	suite.ctx = suite.advanceHeight(suite.ctx, 1, "delegation reward are distributed to insurance and reward module")
	nase = suite.app.LiquidStakingKeeper.GetNetAmountStateEssentials(suite.ctx)
	is = suite.getInsuranceState(suite.ctx)
	{
		suite.Equal(
			"-80999676001295325000.000000000000000000",
			nase.TotalRemainingRewards.Sub(beforeNas.TotalRemainingRewards).String(),
		)
		suite.Equal(
			"161999352002591325000",
			nase.RewardModuleAccBalance.Sub(beforeNas.RewardModuleAccBalance).String(),
			"delegation reward are distributed to reward module",
		)
		suite.Equal(
			"17999928000287925000.000000000000000000",
			is.TotalRemainingInsuranceCommissions.Sub(beforeIs.TotalRemainingInsuranceCommissions).String(),
			"delegation reward are distributed to insurance",
		)
		suite.Equal("0.999784047509547900", nase.MintRate.String())
		suite.True(nase.MintRate.LT(beforeNas.MintRate), "mint rate decreased because of reward distribution")
	}
}

func (suite *KeeperTestSuite) TestLiquidUnstakeWithAdvanceBlocks() {
	fixedInsuranceFeeRate := TenPercentFeeRate
	env := suite.setupLiquidStakeTestingEnv(
		testingEnvOptions{
			"TestLiquidUnstakeWithAdvanceBlocks",
			3,
			TenPercentFeeRate,
			nil,
			onePower,
			nil,
			10,
			fixedInsuranceFeeRate,
			nil,
			3,
			types.ChunkSize.MulRaw(500),
		},
	)
	oneChunk, oneInsurance := suite.app.LiquidStakingKeeper.GetMinimumRequirements(suite.ctx)
	pairedChunksInt := sdk.NewInt(int64(len(env.pairedChunks)))
	mostExpensivePairedChunk := suite.getMostExpensivePairedChunk(env.pairedChunks)
	nase := suite.app.LiquidStakingKeeper.GetNetAmountStateEssentials(suite.ctx)
	is := suite.getInsuranceState(suite.ctx)
	// 1 chunk size * number of paired chunks (=3) tokens are liquidated
	currentLiquidatedTokens := types.ChunkSize.Mul(pairedChunksInt)
	currentInsuranceTokens := oneInsurance.Amount.Mul(pairedChunksInt)
	{
		suite.True(nase.Equal(types.NetAmountStateEssentials{
			MintRate:                    sdk.OneDec(),
			LsTokensTotalSupply:         currentLiquidatedTokens,
			NetAmount:                   currentLiquidatedTokens.ToDec(),
			TotalLiquidTokens:           currentLiquidatedTokens,
			RewardModuleAccBalance:      sdk.ZeroInt(),
			FeeRate:                     sdk.ZeroDec(),
			UtilizationRatio:            sdk.MustNewDecFromStr("0.005999999856000003"),
			RemainingChunkSlots:         sdk.NewInt(47),
			DiscountRate:                sdk.ZeroDec(),
			TotalDelShares:              currentLiquidatedTokens.ToDec(),
			TotalRemainingRewards:       sdk.ZeroDec(),
			TotalChunksBalance:          sdk.ZeroInt(),
			TotalUnbondingChunksBalance: sdk.ZeroInt(),
			NumPairedChunks:             sdk.NewInt(3),
		}), "no epoch(=1 block in test) processed yet, so there are no mint rate change and remaining rewards yet")
		suite.Equal(insuranceState{
			TotalInsuranceTokens:               oneInsurance.Amount.Mul(sdk.NewInt(int64(len(env.insurances)))),
			TotalPairedInsuranceTokens:         currentInsuranceTokens,
			TotalUnpairingInsuranceTokens:      sdk.ZeroInt(),
			TotalRemainingInsuranceCommissions: sdk.ZeroDec(),
		}, is)
	}
	// advance 1 block(= epoch period in test environment) so reward is accumulated which means mint rate is changed
	suite.ctx = suite.advanceHeight(suite.ctx, 1, "")

	beforeNas := nase
	nase = suite.app.LiquidStakingKeeper.GetNetAmountStateEssentials(suite.ctx)
	{
		suite.Equal(
			"80999676001295325000.000000000000000000",
			nase.TotalRemainingRewards.Sub(beforeNas.TotalRemainingRewards).String(),
			"one epoch(=1 block in test) passed, so remaining rewards must be increased",
		)
		suite.Equal("80999676001295325000.000000000000000000", nase.NetAmount.Sub(beforeNas.NetAmount).String(), "net amount must be increased by not claimed rewards")
		suite.Equal("0.999892012094645400", nase.MintRate.String(), "mint rate increased because of reward accumulation")
	}

	undelegator := env.delegators[0]
	// Queue liquid unstake 1 chunk
	beforeBondDenomBalance := suite.app.BankKeeper.GetBalance(suite.ctx, undelegator, env.bondDenom)
	beforeLiquidBondDenomBalance := suite.app.BankKeeper.GetBalance(suite.ctx, undelegator, env.liquidBondDenom)
	msg := types.NewMsgLiquidUnstake(undelegator.String(), oneChunk)
	lsTokensToEscrow := nase.MintRate.Mul(oneChunk.Amount.ToDec()).TruncateInt()
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
	suite.ctx = suite.advanceEpoch(suite.ctx)
	suite.ctx = suite.advanceHeight(suite.ctx, 1, "The actual unstaking started\nThe insurance commission and reward are claimed")
	beforeNas = nase
	beforeIs := is
	nase = suite.app.LiquidStakingKeeper.GetNetAmountStateEssentials(suite.ctx)
	is = suite.getInsuranceState(suite.ctx)

	// Check NetAmounState changed right
	{
		suite.Equal(
			beforeNas.TotalDelShares.Sub(nase.TotalDelShares).TruncateInt().String(),
			oneChunk.Amount.String(),
			"unstaking 1 chunk is started which means undelegate is already triggered so total del shares must be decreased by 1 chunk amount",
		)
		suite.Equal(
			nase.LsTokensTotalSupply.String(),
			beforeNas.LsTokensTotalSupply.String(),
			"unstaking is not finished so ls tokens total supply must not be changed",
		)
		suite.Equal(
			nase.TotalUnbondingChunksBalance.String(),
			oneChunk.Amount.String(),
			"unstaking 1 chunk is started which means undelegate is already triggered",
		)
		suite.True(nase.TotalRemainingRewards.IsZero(), "all rewards are claimed")
		// there is a diff because of truncation
		suite.Equal(
			"161999352002591325000",
			nase.RewardModuleAccBalance.Sub(beforeNas.RewardModuleAccBalance).String(),
			fmt.Sprintf("before unstaking triggered there are collecting reward process "+
				"so reward module got %d chunk's rewards", pairedChunksInt.Int64()),
		)
		totalUnpairingInsCommissions := suite.calcTotalInsuranceCommissions(types.INSURANCE_STATUS_UNPAIRING)
		suite.Equal(
			"5999976000095975000",
			totalUnpairingInsCommissions.String(),
		)
		suite.Equal(
			oneInsurance.Amount.String(),
			is.TotalUnpairingInsuranceTokens.Sub(beforeIs.TotalUnpairingInsuranceTokens).String(),
			"",
		)
		suite.True(nase.MintRate.LT(beforeNas.MintRate), "mint rate decreased because of reward is accumulated")
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

	suite.ctx = suite.advanceEpoch(suite.ctx)
	suite.ctx = suite.advanceHeight(suite.ctx, 1, "The insurance commission and reward are claimed\nThe unstaking is completed")

	beforeNas = nase
	nase = suite.app.LiquidStakingKeeper.GetNetAmountStateEssentials(suite.ctx)
	afterBondDenomBalance := suite.app.BankKeeper.GetBalance(suite.ctx, undelegator, env.bondDenom).Amount
	// Get bondDeno balance of undelegator
	{
		suite.Equal(beforeNas.TotalDelShares.String(), nase.TotalDelShares.String())
		suite.Equal(beforeNas.TotalLiquidTokens.String(), nase.TotalLiquidTokens.String())
		suite.Equal(
			beforeNas.TotalUnbondingChunksBalance.Sub(oneChunk.Amount).String(),
			nase.TotalUnbondingChunksBalance.String(),
			"unstaking(=unbonding) is finished",
		)
		suite.True(nase.LsTokensTotalSupply.LT(beforeNas.LsTokensTotalSupply), "ls tokens are burned")
		suite.True(nase.TotalRemainingRewards.IsZero(), "all rewards are claimed")
		suite.Equal(
			"80999514002915550000",
			nase.RewardModuleAccBalance.Sub(beforeNas.RewardModuleAccBalance).String(),
			"reward module account balance must be increased",
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
		TenPercentFeeRate,
		nil,
	)
	suite.fundAccount(suite.ctx, fundingAccount, types.ChunkSize.MulRaw(500))
	oneChunk, oneInsurance := suite.app.LiquidStakingKeeper.GetMinimumRequirements(suite.ctx)
	providers, providerBalances := suite.AddTestAddrsWithFunding(fundingAccount, 10, oneInsurance.Amount)
	suite.provideInsurances(suite.ctx, providers, valAddrs, providerBalances, sdk.ZeroDec(), nil)
	delegators, delegatorBalances := suite.AddTestAddrsWithFunding(fundingAccount, 3, oneChunk.Amount)
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
		TenPercentFeeRate,
		nil,
	)
	suite.fundAccount(suite.ctx, fundingAccount, types.ChunkSize.MulRaw(500))
	_, minimumCoverage := suite.app.LiquidStakingKeeper.GetMinimumRequirements(suite.ctx)
	providers, balances := suite.AddTestAddrsWithFunding(fundingAccount, 10, minimumCoverage.Amount)
	insurances := suite.provideInsurances(suite.ctx, providers, valAddrs, balances, sdk.ZeroDec(), nil)

	provider := providers[0]
	insurance := insurances[0]
	remainingCommissions := sdk.NewInt(100)
	suite.fundAccount(suite.ctx, insurance.FeePoolAddress(), remainingCommissions)
	escrowed := suite.app.BankKeeper.GetBalance(suite.ctx, insurance.DerivedAddress(), suite.denom)
	beforeProviderBalance := suite.app.BankKeeper.GetBalance(suite.ctx, provider, suite.denom)
	msg := types.NewMsgCancelProvideInsurance(provider.String(), insurance.Id)
	canceledInsurance, err := suite.app.LiquidStakingKeeper.DoCancelProvideInsurance(suite.ctx, msg)
	suite.NoError(err)
	suite.True(insurance.Equal(canceledInsurance))
	afterProviderBalance := suite.app.BankKeeper.GetBalance(suite.ctx, provider, suite.denom)
	suite.True(afterProviderBalance.Amount.Equal(beforeProviderBalance.Amount.Add(escrowed.Amount).Add(remainingCommissions)), "provider should get back escrowed amount and remaining commissions")
	suite.mustPassInvariants()
}

func (suite *KeeperTestSuite) TestDoCancelProvideInsuranceFail() {
	env := suite.setupLiquidStakeTestingEnv(
		testingEnvOptions{
			"TestDoCancelProvideInsuranceFail",
			3,
			TenPercentFeeRate,
			nil,
			onePower,
			nil,
			3,
			TenPercentFeeRate,
			nil,
			1,
			types.ChunkSize.MulRaw(500),
		},
	)
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
	env := suite.setupLiquidStakeTestingEnv(
		testingEnvOptions{
			"TestDoWithdrawInsurance",
			3,
			TenPercentFeeRate,
			nil,
			onePower,
			nil,
			3,
			TenPercentFeeRate,
			nil,
			3,
			types.ChunkSize.MulRaw(500),
		},
	)

	toBeWithdrawnInsurance, _ := suite.app.LiquidStakingKeeper.GetInsurance(suite.ctx, env.insurances[0].Id)
	_, _, err := suite.app.LiquidStakingKeeper.DoWithdrawInsurance(
		suite.ctx,
		types.NewMsgWithdrawInsurance(
			toBeWithdrawnInsurance.ProviderAddress,
			toBeWithdrawnInsurance.Id,
		),
	)
	suite.NoError(err)
	suite.ctx = suite.advanceEpoch(suite.ctx)
	suite.ctx = suite.advanceHeight(suite.ctx, 1, "queued withdraw insurance request is handled and there are no additional insurances yet so unpairing triggered")

	suite.ctx = suite.advanceHeight(suite.ctx, 1, "")

	suite.ctx = suite.advanceEpoch(suite.ctx)
	suite.ctx = suite.advanceHeight(suite.ctx, 1, "unpairing is done")

	unpairedInsurance, _ := suite.app.LiquidStakingKeeper.GetInsurance(suite.ctx, env.insurances[0].Id)
	suite.Equal(types.INSURANCE_STATUS_UNPAIRED, unpairedInsurance.Status)

	beforeProviderBalance := suite.app.BankKeeper.GetBalance(suite.ctx, unpairedInsurance.GetProvider(), suite.denom)
	unpairedInsuranceBalance := suite.app.BankKeeper.GetBalance(suite.ctx, unpairedInsurance.DerivedAddress(), suite.denom)
	unpairedInsuranceCommission := suite.app.BankKeeper.GetBalance(suite.ctx, unpairedInsurance.FeePoolAddress(), suite.denom)
	_, _, err = suite.app.LiquidStakingKeeper.DoWithdrawInsurance(
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
		TenPercentFeeRate,
		nil,
	)
	suite.fundAccount(suite.ctx, fundingAccount, types.ChunkSize.MulRaw(500))
	_, oneInsurance := suite.app.LiquidStakingKeeper.GetMinimumRequirements(suite.ctx)
	providers, providerBalances := suite.AddTestAddrsWithFunding(fundingAccount, 3, oneInsurance.Amount.Add(sdk.NewInt(100)))
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
		_, _, err := suite.app.LiquidStakingKeeper.DoWithdrawInsurance(suite.ctx, tc.msg)
		if tc.expectedErr == nil {
			suite.NoError(err)
		}
		suite.ErrorContains(err, tc.expectedErr.Error())
	}
	suite.mustPassInvariants()
}

func (suite *KeeperTestSuite) TestDoWithdrawInsuranceCommission() {
	fixedInsuranceFeeRate := TenPercentFeeRate
	env := suite.setupLiquidStakeTestingEnv(
		testingEnvOptions{
			"TestDoWithdrawInsuranceCommission",
			3,
			TenPercentFeeRate,
			nil,
			onePower,
			nil,
			3,
			fixedInsuranceFeeRate,
			nil,
			3,
			types.ChunkSize.MulRaw(500),
		},
	)

	provider := env.providers[0]
	targetInsurance := env.insurances[0]
	beforeInsuranceCommission := suite.app.BankKeeper.GetBalance(suite.ctx, targetInsurance.FeePoolAddress(), suite.denom)
	suite.ctx = suite.advanceHeight(suite.ctx, 1, "")
	afterInsuranceCommission := suite.app.BankKeeper.GetBalance(suite.ctx, targetInsurance.FeePoolAddress(), suite.denom)
	suite.Equal(
		afterInsuranceCommission.String(),
		beforeInsuranceCommission.String(),
		"epoch is not reached yet so no insurance commission is distributed",
	)

	suite.ctx = suite.advanceEpoch(suite.ctx)
	suite.ctx = suite.advanceHeight(suite.ctx, 1, "cumulated delegation reward is distributed to withdraw fee pool")
	afterInsuranceCommission = suite.app.BankKeeper.GetBalance(suite.ctx, targetInsurance.FeePoolAddress(), suite.denom)
	suite.Equal(
		"5999976000095975000acanto",
		afterInsuranceCommission.String(),
		"cumulated delegation reward is distributed to withdraw fee pool",
	)

	beforeProviderBalance := suite.app.BankKeeper.GetBalance(suite.ctx, provider, suite.denom)
	// withdraw insurance commission
	_, err := suite.app.LiquidStakingKeeper.DoWithdrawInsuranceCommission(
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
		TenPercentFeeRate,
		nil,
	)
	suite.fundAccount(suite.ctx, fundingAccount, types.ChunkSize.MulRaw(500))
	_, oneInsurance := suite.app.LiquidStakingKeeper.GetMinimumRequirements(suite.ctx)
	providers, providerBalances := suite.AddTestAddrsWithFunding(fundingAccount, 3, oneInsurance.Amount.Add(sdk.NewInt(100)))
	insurances := suite.provideInsurances(
		suite.ctx,
		providers,
		valAddrs,
		providerBalances,
		TenPercentFeeRate,
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
		_, err := suite.app.LiquidStakingKeeper.DoWithdrawInsuranceCommission(suite.ctx, tc.msg)
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
		TenPercentFeeRate,
		nil,
	)
	suite.fundAccount(suite.ctx, fundingAccount, types.ChunkSize.MulRaw(500))
	_, oneInsurance := suite.app.LiquidStakingKeeper.GetMinimumRequirements(suite.ctx)
	providers, _ := suite.AddTestAddrsWithFunding(fundingAccount, 3, oneInsurance.Amount.Add(sdk.NewInt(100)))
	insurances := suite.provideInsurances(
		suite.ctx,
		providers,
		validators,
		[]sdk.Coin{oneInsurance, oneInsurance, oneInsurance},
		TenPercentFeeRate,
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
		TenPercentFeeRate,
		nil,
	)
	suite.fundAccount(suite.ctx, fundingAccount, types.ChunkSize.MulRaw(500))
	_, oneInsurance := suite.app.LiquidStakingKeeper.GetMinimumRequirements(suite.ctx)
	providers, providerBalances := suite.AddTestAddrsWithFunding(fundingAccount, 3, oneInsurance.Amount.Add(sdk.NewInt(100)))
	insurances := suite.provideInsurances(
		suite.ctx,
		providers,
		valAddrs,
		providerBalances,
		TenPercentFeeRate,
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
	env := suite.setupLiquidStakeTestingEnv(
		testingEnvOptions{
			"TestRankInsurances",
			3,
			TenPercentFeeRate,
			nil,
			onePower,
			nil,
			3,
			TenPercentFeeRate,
			nil,
			3,
			types.ChunkSize.MulRaw(500),
		},
	)
	_, oneInsurance := suite.app.LiquidStakingKeeper.GetMinimumRequirements(suite.ctx)
	var idsOfPairedInsurances []uint64
	for _, insurance := range env.insurances {
		idsOfPairedInsurances = append(idsOfPairedInsurances, insurance.Id)
	}

	// INITIAL STATE: all paired chunks are working fine and there are no additional insurances yet
	newlyRankedInInsurances, rankOutInsurances := suite.app.LiquidStakingKeeper.RankInsurances(suite.ctx)
	suite.Len(newlyRankedInInsurances, 0)
	suite.Len(rankOutInsurances, 0)

	suite.ctx = suite.advanceHeight(suite.ctx, 1, "")

	// Cheap insurances which are competitive than current paired insurances are provided
	otherProviders, otherProviderBalances := suite.AddTestAddrsWithFunding(fundingAccount, 3, oneInsurance.Amount)
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

	newlyRankedInInsurances, rankOutInsurances = suite.app.LiquidStakingKeeper.RankInsurances(suite.ctx)
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
	env := suite.setupLiquidStakeTestingEnv(
		testingEnvOptions{
			"TestEndBlocker",
			3,
			TenPercentFeeRate,
			nil,
			onePower,
			nil,
			3,
			TenPercentFeeRate,
			nil,
			3,
			types.ChunkSize.MulRaw(500),
		},
	)

	// Queue withdraw insurance request
	toBeWithdrawnInsurance, _ := suite.app.LiquidStakingKeeper.GetInsurance(suite.ctx, env.insurances[0].Id)
	chunkToBeUnpairing, _ := suite.app.LiquidStakingKeeper.GetChunk(suite.ctx, toBeWithdrawnInsurance.ChunkId)
	_, _, err := suite.app.LiquidStakingKeeper.DoWithdrawInsurance(
		suite.ctx,
		types.NewMsgWithdrawInsurance(
			toBeWithdrawnInsurance.ProviderAddress,
			toBeWithdrawnInsurance.Id,
		),
	)
	suite.NoError(err)
	suite.ctx = suite.advanceEpoch(suite.ctx)
	suite.ctx = suite.advanceHeight(suite.ctx, 1, "queued withdraw insurance request is handled and there are no additional insurances yet so unpairing triggered")
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

	suite.ctx = suite.advanceHeight(suite.ctx, 1, "")

	suite.ctx = suite.advanceEpoch(suite.ctx)
	suite.ctx = suite.advanceHeight(suite.ctx, 1, "withdrawal and unbonding of chunkToBeUnpairing is finished")
	withdrawnInsurance, _ := suite.app.LiquidStakingKeeper.GetInsurance(suite.ctx, toBeWithdrawnInsurance.Id)
	pairingChunk, _ := suite.app.LiquidStakingKeeper.GetChunk(suite.ctx, chunkToBeUnpairing.Id)
	{
		suite.Equal(types.CHUNK_STATUS_PAIRING, pairingChunk.Status)
		suite.Equal(uint64(0), pairingChunk.UnpairingInsuranceId)
		suite.Equal(types.INSURANCE_STATUS_UNPAIRED, withdrawnInsurance.Status)
	}

	suite.ctx = suite.advanceHeight(suite.ctx, 1, "")

	_, oneInsurance := suite.app.LiquidStakingKeeper.GetMinimumRequirements(suite.ctx)
	newValAddrs, _ := suite.CreateValidators(
		[]int64{onePower, onePower, onePower},
		TenPercentFeeRate,
		nil,
	)
	newProviders, newProviderBalances := suite.AddTestAddrsWithFunding(fundingAccount, 3, oneInsurance.Amount)
	newInsurances := suite.provideInsurances(
		suite.ctx,
		newProviders,
		newValAddrs,
		newProviderBalances,
		sdk.NewDecWithPrec(1, 2), // much cheaper than current paired insurances
		nil,
	)
	suite.ctx = suite.advanceEpoch(suite.ctx)
	suite.ctx = suite.advanceHeight(suite.ctx, 1, "pairing chunk is paired now") // PROBLEM
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
		suite.app.LiquidStakingKeeper.IterateAllChunks(suite.ctx, func(chunk types.Chunk) bool {
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
			return false
		})
	}

	suite.ctx = suite.advanceHeight(suite.ctx, 1, "")

	pairedInsurances := newInsurances
	newProviders, newProviderBalances = suite.AddTestAddrsWithFunding(fundingAccount, 3, oneInsurance.Amount)
	newInsurances = suite.provideInsurances(
		suite.ctx,
		newProviders,
		newValAddrs,
		newProviderBalances,
		sdk.NewDecWithPrec(1, 3), // much cheaper than current paired insurances
		nil,
	)
	suite.ctx = suite.advanceEpoch(suite.ctx)
	suite.ctx = suite.advanceHeight(suite.ctx, 1, "all paired chunks are started to be re-paired with new insurances")
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

		suite.app.LiquidStakingKeeper.IterateAllChunks(suite.ctx, func(chunk types.Chunk) bool {
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
			return false
		})
	}

}

// TestPairedChunkTombstonedAndRedelegated tests the case where a one chunk is re-paired
// after paired insurance was tombstoned
func (suite *KeeperTestSuite) TestPairedChunkTombstonedAndRedelegated() {
	env := suite.setupLiquidStakeTestingEnv(
		testingEnvOptions{
			"TestPairedChunkTombstonedAndRedelegated",
			3,
			TenPercentFeeRate,
			nil,
			onePower,
			nil,
			10,
			TenPercentFeeRate,
			nil,
			3,
			types.ChunkSize.MulRaw(500),
		},
	)
	suite.ctx = suite.advanceHeight(suite.ctx, 1, "liquid staking started")

	toBeTombstonedValidator := env.valAddrs[0]
	toBeTombstonedValidatorPubKey := env.pubKeys[0]
	toBeTombstonedChunk := env.pairedChunks[0]
	ins, _ := suite.app.LiquidStakingKeeper.GetInsurance(suite.ctx, toBeTombstonedChunk.PairedInsuranceId)
	// 7% + 3.75% = 10.75%
	// After tombstone, it still pass the line (5.75%) which means
	// The chunk will not be unpairing because of IsEnoughToCoverSlash check
	suite.fundAccount(suite.ctx, ins.DerivedAddress(), types.ChunkSize.ToDec().Mul(sdk.NewDecWithPrec(375, 2)).Ceil().TruncateInt())
	selfDelegationToken := suite.app.StakingKeeper.TokensFromConsensusPower(suite.ctx, onePower)
	// handle a signature to set signing info
	suite.app.SlashingKeeper.HandleValidatorSignature(
		suite.ctx,
		toBeTombstonedValidatorPubKey.Address(),
		selfDelegationToken.Int64(),
		true,
	)

	val := suite.app.StakingKeeper.Validator(suite.ctx, toBeTombstonedValidator)
	del, _ := suite.app.StakingKeeper.GetDelegation(
		suite.ctx,
		toBeTombstonedChunk.DerivedAddress(),
		toBeTombstonedValidator,
	)
	valTokensBeforeTombstoned := val.GetTokens()
	delTokens := val.TokensFromShares(del.GetShares())

	suite.tombstone(suite.ctx, toBeTombstonedValidator, toBeTombstonedValidatorPubKey, suite.ctx.BlockHeight()-1)
	suite.ctx = suite.advanceHeight(suite.ctx, 1, "validator is tombstoned now")
	fmt.Println(suite.app.LiquidStakingKeeper.GetNetAmountStateEssentials(suite.ctx))

	{
		valTombstoned := suite.app.StakingKeeper.Validator(suite.ctx, toBeTombstonedValidator)
		suite.Equal(stakingtypes.Unbonding, valTombstoned.GetStatus())
		valTokensAfterTombstoned := valTombstoned.GetTokens()
		delTokensAfterTombstoned := valTombstoned.TokensFromShares(del.GetShares())
		valTokensDiff := valTokensBeforeTombstoned.Sub(valTokensAfterTombstoned)

		suite.Equal("12500050000000000000000", valTokensDiff.String())
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

	suite.ctx = suite.advanceEpoch(suite.ctx)
	suite.ctx = suite.advanceHeight(suite.ctx, 1, "re-delegation happened in this epoch")
	fmt.Println(suite.app.LiquidStakingKeeper.GetNetAmountStateEssentials(suite.ctx))
	passedRewardsEpochAfterTombstoned := int64(1)

	// check chunk is started to be re-paired with new insurances
	// and chunk delegation token value is recovered or not
	tombstonedChunk, _ := suite.app.LiquidStakingKeeper.GetChunk(suite.ctx, toBeTombstonedChunk.Id)
	{
		valTombstoned := suite.app.StakingKeeper.Validator(suite.ctx, toBeTombstonedValidator)
		suite.Equal(stakingtypes.Unbonded, valTombstoned.GetStatus())
		suite.Equal(
			env.insurances[4].Id,
			tombstonedChunk.PairedInsuranceId,
			"insurances[3] cannot be ranked in because it points to the tombstoned validator, so next insurance is ranked in",
		)
		suite.Equal(types.CHUNK_STATUS_PAIRED, tombstonedChunk.Status)
		suite.Equal(toBeTombstonedChunk.PairedInsuranceId, tombstonedChunk.UnpairingInsuranceId)
		unpairingIns, _ := suite.app.LiquidStakingKeeper.GetInsurance(suite.ctx, tombstonedChunk.UnpairingInsuranceId)
		suite.Equal(
			"2999988000047975000",
			suite.app.BankKeeper.GetBalance(suite.ctx, unpairingIns.FeePoolAddress(), env.bondDenom).Amount.String(),
			fmt.Sprintf(
				"tombstoned insurance got commission for %d reward epochs",
				suite.rewardEpochCount-passedRewardsEpochAfterTombstoned,
			),
		)
		// Tombstoned validator got only 1 reward epoch commission because it is tombstoned before epoch is passed.
		// So other validator's delegation rewards are increased by the amount of tombstoned validator's delegation reward.
		suite.Equal(
			"11999952000191975000",
			suite.app.BankKeeper.GetBalance(suite.ctx, env.insurances[1].FeePoolAddress(), env.bondDenom).Amount.String(),
			fmt.Sprintf(
				"normal insurance got (commission for %d reward epochs + "+
					"tombstoned delegation reward / number of valid delegations x 2) "+
					"which means unit delegation reward is increased temporarily.\n"+
					"this is temporary because in this liquidstaking epoch, re-delegation happened so "+
					"every delegation reward will be same from now.",
				suite.rewardEpochCount,
			),
		)
	}
	newInsurance, _ := suite.app.LiquidStakingKeeper.GetInsurance(suite.ctx, tombstonedChunk.PairedInsuranceId)
	dstVal := suite.app.StakingKeeper.Validator(suite.ctx, newInsurance.GetValidator())
	// re-delegation obj must exist
	_, found := suite.app.StakingKeeper.GetRedelegation(
		suite.ctx,
		tombstonedChunk.DerivedAddress(),
		toBeTombstonedValidator,
		newInsurance.GetValidator(),
	)
	{
		suite.False(found, "srcVal was un-bonded validator, so re-delegation obj doesn't exist")
		del, _ = suite.app.StakingKeeper.GetDelegation(
			suite.ctx,
			tombstonedChunk.DerivedAddress(),
			newInsurance.GetValidator(),
		)
		afterCovered := dstVal.TokensFromShares(del.GetShares())
		suite.True(afterCovered.GTE(types.ChunkSize.ToDec()))
	}
	suite.ctx = suite.advanceHeight(suite.ctx, 1, "delegation rewards are accumulated")
	fmt.Println(suite.app.LiquidStakingKeeper.GetNetAmountStateEssentials(suite.ctx))

	suite.ctx = suite.advanceEpoch(suite.ctx)
	suite.ctx = suite.advanceHeight(suite.ctx, 1, "unpairing insurance because of tombstoned is unpaired now")
	fmt.Println(suite.app.LiquidStakingKeeper.GetNetAmountStateEssentials(suite.ctx))

	{
		unpairedInsurance, _ := suite.app.LiquidStakingKeeper.GetInsurance(suite.ctx, tombstonedChunk.UnpairingInsuranceId)
		unpairedInsuranceVal, _ := suite.app.StakingKeeper.GetValidator(suite.ctx, unpairedInsurance.GetValidator())
		suite.Equal(types.INSURANCE_STATUS_UNPAIRED, unpairedInsurance.Status)
		suite.Error(
			suite.app.LiquidStakingKeeper.ValidateValidator(suite.ctx, unpairedInsuranceVal),
			"validator of unpaired insurance is tombstoned",
		)
	}
}

// TestRedelegateToSlashedValidator tests scenario where validator got slashed and re-delegated to slashed validator.
// And during re-delegation period, evidence before re-delegation start height was discovered, so src validator is tombstoned.
// 1. v1 - c1 - (i2, x) and v2 - c2 - (i1, x)
// 1-1. i1 is more expensive than i2 (remember, chunk /w most expensive insurance is unpaired first when unstake)
// 1-2. i1, i2 are above minimum requirement, so it will not be unpaired when epoch because of lack of balance.
// 2. v2 slashed, so i1 will cover v2's slashing penalty
// 3. unstake c2
// 3-1. we can process queued unstake c2 because insurance still has enough balance and paired.
// NOW V2 have slashing history, but have no chunk
// 4. v1 - c1 - (i2, x) and v2 - x - (x, x)
// 5. provide cheap insurance i3 for v2
// RE-DELEGATE v1 -> v2
// 6. v1 - x - (x, x) and v2 - c1 - (i3, i2)
// 7. Found evidence of double signing of v1 before re-delegation start height, so v1 is tombstoned.
// E1-------(v1-double-signed)-------E2(re-delegate v1->v2)-------(found evidence for double-sign)-------E3(i2 must cover v1's double-sign slashing)
// 8. i2 should cover v1's slashing penalty for re-delegation.
// 9. After all, c1 should not get damaged.
func (suite *KeeperTestSuite) TestRedelegateToSlashedValidator() {
	initialHeight := int64(1)
	suite.ctx = suite.ctx.WithBlockHeight(initialHeight) // make sure we start with clean height
	suite.fundAccount(suite.ctx, fundingAccount, types.ChunkSize.MulRaw(500))
	valNum := 2
	addrs, _ := suite.AddTestAddrsWithFunding(fundingAccount, valNum, suite.app.StakingKeeper.TokensFromConsensusPower(suite.ctx, 200))
	valAddrs := simapp.ConvertAddrsToValAddrs(addrs)
	v1 := valAddrs[0]
	v2 := valAddrs[1]
	pubKeys := suite.createTestPubKeys(valNum)
	v1PubKey := pubKeys[0]
	v2PubKey := pubKeys[1]
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
	providers, providerBalances := suite.AddTestAddrsWithFunding(fundingAccount, 2, oneInsurance.Amount.MulRaw(2))
	suite.provideInsurances(
		suite.ctx,
		providers,
		// We will make v1 - i2 and v2 - i1
		[]sdk.ValAddress{v2, v1},
		providerBalances,
		sdk.ZeroDec(),
		[]sdk.Dec{TenPercentFeeRate, FivePercentFeeRate},
	)
	delegators, delegatorBalances := suite.AddTestAddrsWithFunding(fundingAccount, 2, oneChunk.Amount)
	chunks := suite.liquidStakes(suite.ctx, delegators, delegatorBalances)
	suite.Len(chunks, 2, "2 chunks are created")
	// v1 - c1 - i2, v2 - c2 - i1
	c1 := chunks[0]
	c2 := chunks[1]
	i2, _ := suite.app.LiquidStakingKeeper.GetInsurance(suite.ctx, c1.PairedInsuranceId)
	suite.Equal(FivePercentFeeRate, i2.FeeRate)
	suite.Equal(v1, i2.GetValidator())
	i1, _ := suite.app.LiquidStakingKeeper.GetInsurance(suite.ctx, c2.PairedInsuranceId)
	suite.Equal(TenPercentFeeRate, i1.FeeRate)
	suite.Equal(v2, i1.GetValidator())

	// Rewards are accumulated
	// If we do not accumulate rewards, then it fails when we queue unstake will be failed.
	// It because during slashing period, net amount is decreased, and minted ls token is same so the mint rate
	// goes high. That means liquid staker don't have ls token to liquid unstake.
	suite.ctx = suite.advanceHeight(suite.ctx, 100, "v1 - c1 - (i1, x) and v2 - c2 - (i2, x)")

	downValAddr := v2
	// let's downtime slashing v2
	{
		downValPubKey := v2PubKey
		epoch := suite.app.LiquidStakingKeeper.GetEpoch(suite.ctx)
		epochTime := suite.ctx.BlockTime().Add(epoch.Duration)
		called := 0
		for {
			// downtime 5 times, so insurance can cover this penalty
			if called == 5 {
				break
			}
			validator, _ := suite.app.StakingKeeper.GetValidatorByConsAddr(suite.ctx, sdk.GetConsAddress(downValPubKey))
			suite.downTimeSlashing(
				suite.ctx,
				downValPubKey,
				validator.GetConsensusPower(suite.app.StakingKeeper.PowerReduction(suite.ctx)),
				called,
				time.Second,
			)
			suite.unjail(suite.ctx, downValAddr, downValPubKey, time.Second)
			called++

			if suite.ctx.BlockTime().After(epochTime) {
				break
			}
		}
	}

	// liquid unstake c2
	{
		_, _, err := suite.app.LiquidStakingKeeper.QueueLiquidUnstake(
			suite.ctx,
			types.NewMsgLiquidUnstake(
				delegators[1].String(),
				sdk.NewCoin(suite.denom, oneChunk.Amount),
			),
		)
		suite.NoError(err)
	}
	// Trigger and finish unbonding.
	{
		i1BalBeforeCoverPenalty := suite.app.BankKeeper.GetBalance(suite.ctx, i1.DerivedAddress(), suite.denom)
		suite.ctx = suite.advanceEpoch(suite.ctx)
		suite.ctx = suite.advanceHeight(suite.ctx, 1, "unbonding chunk triggered "+
			"and slashing penalty is covered by a paired insurance")
		i1Bal := suite.app.BankKeeper.GetBalance(suite.ctx, i1.DerivedAddress(), suite.denom)
		suite.True(
			i1BalBeforeCoverPenalty.Amount.GT(i1Bal.Amount),
			"i1 covered penalty of v2, so unbonding chunk is successfully triggered",
		)
		i1, _ = suite.app.LiquidStakingKeeper.GetInsurance(suite.ctx, i1.Id)
		suite.Equal(types.INSURANCE_STATUS_UNPAIRING, i1.Status, "i1 is unpairing")
		_, found := suite.app.LiquidStakingKeeper.GetUnpairingForUnstakingChunkInfo(suite.ctx, c2.Id)
		suite.True(found, "unpairing info is created")

		suite.ctx = suite.advanceEpoch(suite.ctx)
		suite.ctx = suite.advanceHeight(suite.ctx, 1, "v1 - c1 - (i2, x) and v2 - x - (x, x)")
		_, found = suite.app.LiquidStakingKeeper.GetUnpairingForUnstakingChunkInfo(suite.ctx, c2.Id)
		suite.False(found, "unstaking is finished, so unpairing info is deleted")
		i1, _ = suite.app.LiquidStakingKeeper.GetInsurance(suite.ctx, i1.Id)
		suite.Equal(types.INSURANCE_STATUS_PAIRING, i1.Status, "i1 is unpaired, but it is still valid")
	}

	chunks = suite.app.LiquidStakingKeeper.GetAllChunks(suite.ctx)
	suite.Len(chunks, 1, "one chunk is left")
	leftChunk := chunks[0]
	suite.Equal(c1.Id, leftChunk.Id)
	suite.Equal(i2.Id, leftChunk.PairedInsuranceId, "c1 - i2 is left")

	anotherProviders, anotherProviderBalances := suite.AddTestAddrsWithFunding(fundingAccount, 1, oneInsurance.Amount.MulRaw(2))
	insurances := suite.provideInsurances(
		suite.ctx, anotherProviders,
		[]sdk.ValAddress{downValAddr},
		anotherProviderBalances, sdk.ZeroDec(),
		[]sdk.Dec{sdk.ZeroDec()}, // very attractive fee rate
	)
	i3 := insurances[0]
	suite.ctx = suite.advanceEpoch(suite.ctx)
	suite.ctx = suite.advanceHeight(suite.ctx, 1, "re-delegation is started")
	// Check re-delegation is started or not
	reDelStartedHeight := suite.ctx.BlockHeight()
	{
		leftChunk, _ = suite.app.LiquidStakingKeeper.GetChunk(suite.ctx, leftChunk.Id)
		suite.Equal(types.CHUNK_STATUS_PAIRED, leftChunk.Status)
		suite.NotEqual(i2.Id, leftChunk.PairedInsuranceId)
		suite.Equal(i2.Id, leftChunk.UnpairingInsuranceId, "i3 is new insurance and i2 is unpairing insurance")
		suite.Equal(i3.Id, leftChunk.PairedInsuranceId, "i3 is newly paired by ranking mechanism")
		srcInsurance, _ := suite.app.LiquidStakingKeeper.GetInsurance(suite.ctx, leftChunk.UnpairingInsuranceId)
		dstInsurance, _ := suite.app.LiquidStakingKeeper.GetInsurance(suite.ctx, leftChunk.PairedInsuranceId)
		red, found := suite.app.StakingKeeper.GetRedelegation(suite.ctx, leftChunk.DerivedAddress(), srcInsurance.GetValidator(), dstInsurance.GetValidator())
		suite.True(found)
		suite.Len(red.Entries, 1)
		entry := red.Entries[0]
		suite.True(entry.InitialBalance.GTE(types.ChunkSize))
		suite.True(
			entry.SharesDst.GTE(types.ChunkSize.ToDec()),
			"dst validator have history of slashing, so sharesDst should be greater than chunk size",
		)
		dstVal := suite.app.StakingKeeper.Validator(suite.ctx, dstInsurance.GetValidator())
		suite.Equal(v2, dstVal.GetOperator(), "v1 -> v2")
		tokenValue := dstVal.TokensFromShares(entry.SharesDst) // If we truncate it, then this value is less than chunk size
		suite.True(
			tokenValue.GTE(types.ChunkSize.ToDec()),
			"token value must not be less than chunk size, it because the slashing penalty already handled by insurance",
		)
	}

	c1, _ = suite.app.LiquidStakingKeeper.GetChunk(suite.ctx, c1.Id)
	i3, _ = suite.app.LiquidStakingKeeper.GetInsurance(suite.ctx, c1.PairedInsuranceId)
	del, _ := suite.app.StakingKeeper.GetDelegation(suite.ctx, c1.DerivedAddress(), i3.GetValidator())
	tokenValue := suite.app.StakingKeeper.Validator(suite.ctx, i3.GetValidator()).TokensFromShares(del.Shares)
	// let's tombstone slashing v1
	suite.tombstone(suite.ctx, v1, v1PubKey, reDelStartedHeight-1)
	redel, found := suite.app.StakingKeeper.GetRedelegation(suite.ctx, c1.DerivedAddress(), i2.GetValidator(), i3.GetValidator())
	suite.True(found)
	suite.Len(redel.Entries, 1)
	del, _ = suite.app.StakingKeeper.GetDelegation(suite.ctx, c1.DerivedAddress(), i3.GetValidator())
	suite.True(
		del.Shares.LT(redel.Entries[0].SharesDst),
		"because of v1's slashing, del shares is decreased",
	)
	tokenValue = suite.app.StakingKeeper.Validator(suite.ctx, i3.GetValidator()).TokensFromShares(del.Shares)
	suite.True(
		tokenValue.LT(types.ChunkSize.ToDec()),
		"because of v1's slashing, token value of del shares is decreased",
	)

	i2Bal := suite.app.BankKeeper.GetBalance(suite.ctx, i2.DerivedAddress(), suite.denom)
	i3Bal := suite.app.BankKeeper.GetBalance(suite.ctx, i3.DerivedAddress(), suite.denom)
	suite.ctx = suite.advanceEpoch(suite.ctx)
	suite.ctx = suite.advanceHeight(suite.ctx, 1, "slashed penalty of re-delegation will be covered by i2")
	// Check slashed penalty is covered by i2
	{
		i2BalAfter := suite.app.BankKeeper.GetBalance(suite.ctx, i2.DerivedAddress(), suite.denom)
		i3BalAfter := suite.app.BankKeeper.GetBalance(suite.ctx, i3.DerivedAddress(), suite.denom)

		suite.True(i3Bal.Equal(i3BalAfter), "i3 did not pay any penalty for v1(=srcVal)'s) tombstone slashing")
		suite.True(i2BalAfter.IsLT(i2Bal), "i2 did pay penalty of re-delegation for v1(=srcVal)'s tombstone slashing")
		// Let's see its covered by i2 correctly
		del, _ = suite.app.StakingKeeper.GetDelegation(suite.ctx, c1.DerivedAddress(), i3.GetValidator())
		tokenValue = suite.app.StakingKeeper.Validator(suite.ctx, i3.GetValidator()).TokensFromShares(del.Shares)
		suite.True(
			tokenValue.GTE(types.ChunkSize.ToDec()),
			"token value must not be less than chunk size, it because the slashing penalty already covered by insurance",
		)
	}
}

// TestRedelegateFromSlashedToSlashed tests re-delegation from slashed validator to slashed validator.
// And during re-delegation period, evidence before re-delegation start height was discovered, so src validator is tombstoned.
// 1. v1 - c1 - (i2, x) and v2 - c2 - (i1, x)
// 1-1. i1 is more expensive than i2 (remember, chunk /w most expensive insurance is unpaired first when unstake)
// 1-2. i1, i2 are above minimum requirement, so it will not be unpaired at epoch because of in-sufficient collateral.
// 2. v1 slashed, so i2 will cover v1's slashing penalty
// 2. v2 slashed, so i1 will cover v2's slashing penalty
// 3. unstake c2
// 3-1. we can process queued unstake c2 because insurance still has enough balance and paired.
// NOW V2 have slashing history, but have no chunk
// 4. v1 - c1 - (i2, x) and v2 - x - (x, x)
// 5. provide cheap insurance i3 for v2
// RE-DELEGATE v1 -> v2
// 6. v1 - x - (x, x) and v2 - c1 - (i3, i2)
// 7. Found evidence of double signing of v1 before re-delegation start height, so v1 is tombstoned.
// 8. i2 should cover v1's slashing penalty for re-delegation.
// 9. After all, c1 should not get damaged.
func (suite *KeeperTestSuite) TestRedelegateFromSlashedToSlashed() {
	initialHeight := int64(1)
	suite.ctx = suite.ctx.WithBlockHeight(initialHeight) // make sure we start with clean height
	suite.fundAccount(suite.ctx, fundingAccount, types.ChunkSize.MulRaw(500))
	valNum := 2
	addrs, _ := suite.AddTestAddrsWithFunding(fundingAccount, valNum, suite.app.StakingKeeper.TokensFromConsensusPower(suite.ctx, 200))
	valAddrs := simapp.ConvertAddrsToValAddrs(addrs)
	v1 := valAddrs[0]
	v2 := valAddrs[1]
	pubKeys := suite.createTestPubKeys(valNum)
	v1PubKey := pubKeys[0]
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
	// 14% of chunk size are provided as collateral
	providers, providerBalances := suite.AddTestAddrsWithFunding(fundingAccount, 2, oneInsurance.Amount.MulRaw(2))
	suite.provideInsurances(
		suite.ctx,
		providers,
		// We will make v1 - i2 and v2 - i1
		[]sdk.ValAddress{v2, v1},
		providerBalances,
		sdk.ZeroDec(),
		[]sdk.Dec{TenPercentFeeRate, FivePercentFeeRate},
	)
	delegators, delegatorBalances := suite.AddTestAddrsWithFunding(fundingAccount, 2, oneChunk.Amount)
	chunks := suite.liquidStakes(suite.ctx, delegators, delegatorBalances)
	suite.Len(chunks, 2, "2 chunks are created")
	// v1 - c1 - i2, v2 - c2 - i1
	c1 := chunks[0]
	c2 := chunks[1]
	i2, _ := suite.app.LiquidStakingKeeper.GetInsurance(suite.ctx, c1.PairedInsuranceId)
	suite.Equal(FivePercentFeeRate, i2.FeeRate)
	suite.Equal(v1, i2.GetValidator())
	i1, _ := suite.app.LiquidStakingKeeper.GetInsurance(suite.ctx, c2.PairedInsuranceId)
	suite.Equal(TenPercentFeeRate, i1.FeeRate)
	suite.Equal(v2, i1.GetValidator())

	// Rewards are accumulated
	// If we do not accumulate rewards, then it fails when we queue unstake will be failed.
	// It because during slashing period, net amount is decreased, and minted ls token is same so the mint rate
	// goes high. That means liquid staker don't have ls token to liquid unstake.
	suite.ctx = suite.advanceHeight(suite.ctx, 200, "v1 - c1 - (i1, x) and v2 - c2 - (i2, x)")

	// let's downtime slashing v1 and v2
	{
		for i, downValPubKey := range pubKeys {
			downValAddr := valAddrs[i]
			epoch := suite.app.LiquidStakingKeeper.GetEpoch(suite.ctx)
			epochTime := suite.ctx.BlockTime().Add(epoch.Duration)
			called := 0
			for {
				// downtime 5 times, so insurance can cover this penalty
				if called == 5 {
					break
				}
				validator, _ := suite.app.StakingKeeper.GetValidatorByConsAddr(suite.ctx, sdk.GetConsAddress(downValPubKey))
				suite.downTimeSlashing(
					suite.ctx,
					downValPubKey,
					validator.GetConsensusPower(suite.app.StakingKeeper.PowerReduction(suite.ctx)),
					called,
					time.Second,
				)
				suite.unjail(suite.ctx, downValAddr, downValPubKey, time.Second)
				called++

				if suite.ctx.BlockTime().After(epochTime) {
					break
				}
			}
		}
	}

	// liquid unstake c2
	{
		_, _, err := suite.app.LiquidStakingKeeper.QueueLiquidUnstake(
			suite.ctx,
			types.NewMsgLiquidUnstake(
				delegators[1].String(),
				sdk.NewCoin(suite.denom, oneChunk.Amount),
			),
		)
		suite.NoError(err)
	}
	// Trigger and finish unbonding.
	{
		i1BalBeforeCoverPenalty := suite.app.BankKeeper.GetBalance(suite.ctx, i1.DerivedAddress(), suite.denom)
		suite.ctx = suite.advanceEpoch(suite.ctx)
		suite.ctx = suite.advanceHeight(suite.ctx, 1, "unbonding chunk triggered "+
			"and slashing penalty is covered by a paired insurance")
		i1Bal := suite.app.BankKeeper.GetBalance(suite.ctx, i1.DerivedAddress(), suite.denom)
		suite.True(
			i1BalBeforeCoverPenalty.Amount.GT(i1Bal.Amount),
			"i1 covered penalty of v2, so unbonding chunk is successfully triggered",
		)
		i1, _ = suite.app.LiquidStakingKeeper.GetInsurance(suite.ctx, i1.Id)
		suite.Equal(types.INSURANCE_STATUS_UNPAIRING, i1.Status, "i1 is unpairing")
		_, found := suite.app.LiquidStakingKeeper.GetUnpairingForUnstakingChunkInfo(suite.ctx, c2.Id)
		suite.True(found, "unpairing info is created")

		suite.ctx = suite.advanceEpoch(suite.ctx)
		suite.ctx = suite.advanceHeight(suite.ctx, 1, "v1 - c1 - (i2, x) and v2 - x - (x, x)")
		_, found = suite.app.LiquidStakingKeeper.GetUnpairingForUnstakingChunkInfo(suite.ctx, c2.Id)
		suite.False(found, "unstaking is finished, so unpairing info is deleted")
		i1, _ = suite.app.LiquidStakingKeeper.GetInsurance(suite.ctx, i1.Id)
		suite.Equal(types.INSURANCE_STATUS_PAIRING, i1.Status, "i1 is unpaired, but it is still valid insurance.")
	}

	chunks = suite.app.LiquidStakingKeeper.GetAllChunks(suite.ctx)
	suite.Len(chunks, 1, "one chunk is left")
	leftChunk := chunks[0]
	suite.Equal(c1.Id, leftChunk.Id)
	suite.Equal(i2.Id, leftChunk.PairedInsuranceId, "c1 - i2 is left")

	anotherProviders, anotherProviderBalances := suite.AddTestAddrsWithFunding(fundingAccount, 1, oneInsurance.Amount)
	insurances := suite.provideInsurances(
		suite.ctx, anotherProviders,
		[]sdk.ValAddress{v2},
		anotherProviderBalances, sdk.ZeroDec(),
		[]sdk.Dec{sdk.ZeroDec()}, // very attractive fee rate
	)
	i3 := insurances[0]
	suite.ctx = suite.advanceEpoch(suite.ctx)
	suite.ctx = suite.advanceHeight(suite.ctx, 1, "re-delegation is started")
	// Check re-delegation is started or not
	reDelStartedHeight := suite.ctx.BlockHeight()
	{
		leftChunk, _ = suite.app.LiquidStakingKeeper.GetChunk(suite.ctx, leftChunk.Id)
		suite.NotEqual(i2.Id, leftChunk.PairedInsuranceId)
		suite.Equal(i2.Id, leftChunk.UnpairingInsuranceId, "i3 is new insurance and i2 is unpairing insurance")
		suite.Equal(i3.Id, leftChunk.PairedInsuranceId, "i3 is newly paired by ranking mechanism")
		srcInsurance, _ := suite.app.LiquidStakingKeeper.GetInsurance(suite.ctx, leftChunk.UnpairingInsuranceId)
		dstInsurance, _ := suite.app.LiquidStakingKeeper.GetInsurance(suite.ctx, leftChunk.PairedInsuranceId)
		red, found := suite.app.StakingKeeper.GetRedelegation(suite.ctx, leftChunk.DerivedAddress(), srcInsurance.GetValidator(), dstInsurance.GetValidator())
		suite.True(found)
		suite.Len(red.Entries, 1)
		entry := red.Entries[0]
		suite.True(entry.InitialBalance.GTE(types.ChunkSize))
		suite.True(
			entry.SharesDst.GTE(types.ChunkSize.ToDec()),
			"dst validator have history of slashing, so sharesDst should be greater than chunk size",
		)
		dstVal := suite.app.StakingKeeper.Validator(suite.ctx, dstInsurance.GetValidator())
		suite.Equal(v2, dstVal.GetOperator(), "v1 -> v2")
		tokenValue := dstVal.TokensFromShares(entry.SharesDst) // If we truncate it, then this value is less than chunk size
		suite.True(
			tokenValue.GTE(types.ChunkSize.ToDec()),
			"token value must not be less than chunk size, it because the slashing penalty already handled by insurance",
		)
	}

	c1, _ = suite.app.LiquidStakingKeeper.GetChunk(suite.ctx, c1.Id)
	i3, _ = suite.app.LiquidStakingKeeper.GetInsurance(suite.ctx, c1.PairedInsuranceId)
	del, _ := suite.app.StakingKeeper.GetDelegation(suite.ctx, c1.DerivedAddress(), i3.GetValidator())
	tokenValue := suite.app.StakingKeeper.Validator(suite.ctx, i3.GetValidator()).TokensFromShares(del.Shares)
	// let's tombstone slashing v1
	suite.tombstone(suite.ctx, v1, v1PubKey, reDelStartedHeight-1)
	redel, found := suite.app.StakingKeeper.GetRedelegation(suite.ctx, c1.DerivedAddress(), i2.GetValidator(), i3.GetValidator())
	suite.True(found)
	suite.Len(redel.Entries, 1)
	del, _ = suite.app.StakingKeeper.GetDelegation(suite.ctx, c1.DerivedAddress(), i3.GetValidator())
	suite.True(
		del.Shares.LT(redel.Entries[0].SharesDst),
		"because of v1's slashing, del shares is decreased",
	)
	tokenValue = suite.app.StakingKeeper.Validator(suite.ctx, i3.GetValidator()).TokensFromShares(del.Shares)
	suite.True(
		tokenValue.LT(types.ChunkSize.ToDec()),
		"because of v1's slashing, token value of del shares is decreased",
	)

	i2Bal := suite.app.BankKeeper.GetBalance(suite.ctx, i2.DerivedAddress(), suite.denom)
	i3Bal := suite.app.BankKeeper.GetBalance(suite.ctx, i3.DerivedAddress(), suite.denom)
	suite.ctx = suite.advanceEpoch(suite.ctx)
	suite.ctx = suite.advanceHeight(suite.ctx, 1, "slashed penalty of re-delegation will be covered by i2")
	// Check slashed penalty is covered by i2
	{
		i2BalAfter := suite.app.BankKeeper.GetBalance(suite.ctx, i2.DerivedAddress(), suite.denom)
		i3BalAfter := suite.app.BankKeeper.GetBalance(suite.ctx, i3.DerivedAddress(), suite.denom)

		suite.True(i3Bal.Equal(i3BalAfter), "i3 did not pay any penalty for v1(=srcVal)'s) tombstone slashing")
		suite.True(i2BalAfter.IsLT(i2Bal), "i2 did pay penalty of re-delegation for v1(=srcVal)'s tombstone slashing")
		// Let's see its covered by i2 correctly
		del, _ = suite.app.StakingKeeper.GetDelegation(suite.ctx, c1.DerivedAddress(), i3.GetValidator())
		tokenValue = suite.app.StakingKeeper.Validator(suite.ctx, i3.GetValidator()).TokensFromShares(del.Shares)
		suite.True(
			tokenValue.GTE(types.ChunkSize.ToDec()),
			"token value must not be less than chunk size, it because the slashing penalty already covered by insurance",
		)
	}

}

// TestUnpairingInsuranceCoversSlashingBeforeRedelegationHeight tests scenario where
// unpairing insurance covers slashing penalty happened before re-delegation height.
func (suite *KeeperTestSuite) TestUnpairingInsuranceCoversSlashingBeforeRedelegationHeight() {
	// validator - chunk - (paired insurance, unpairing insurance)
	// v1 - c1 - (i1, x), v2 - x - (x, x)
	// provide insurance i2 with lower fee which direct v2
	// reach epoch - checkpoint1
	// begin re-delegation => v1 - x - (x, x), v2 - c1 - (i2, i1)
	// recognized double-sign slashing for before checkpoint1
	// i1 should cover that slashing penalty
	env := suite.setupLiquidStakeTestingEnv(
		testingEnvOptions{
			"TestUnpairingInsuranceCoversSlashingBeforeRedelegationHeight",
			2,
			TenPercentFeeRate,
			nil,
			onePower,
			nil,
			1,
			TenPercentFeeRate,
			nil,
			1,
			types.ChunkSize.MulRaw(500),
		},
	)
	chunk := env.pairedChunks[0]
	srcValAddr := env.valAddrs[0]
	srcValPubKey := env.pubKeys[0]
	unpairingInsurance := env.insurances[0]
	suite.Equal(srcValAddr, env.insurances[0].GetValidator())

	dstValAddr := env.valAddrs[1]
	onePercentFeeRate := sdk.NewDecWithPrec(1, 2)
	suite.True(onePercentFeeRate.LT(unpairingInsurance.FeeRate))

	_, oneInsurance := suite.app.LiquidStakingKeeper.GetMinimumRequirements(suite.ctx)
	providers, providerBals := suite.AddTestAddrsWithFunding(fundingAccount, 1, oneInsurance.Amount)
	// provide insurance with lower fee
	suite.provideInsurances(suite.ctx, providers, []sdk.ValAddress{env.valAddrs[1]}, providerBals, onePercentFeeRate, nil)
	suite.ctx = suite.advanceEpoch(suite.ctx)
	suite.ctx = suite.advanceHeight(suite.ctx, 1, "checkpoint1: re-delegation")

	checkPoint1 := suite.ctx.BlockHeight()
	// Check state is correct before got slashed
	{
		redelegation, found := suite.app.StakingKeeper.GetRedelegation(suite.ctx, chunk.DerivedAddress(), srcValAddr, dstValAddr)
		suite.True(found)
		suite.Len(redelegation.Entries, 1)
		suite.Equal(srcValAddr.String(), redelegation.ValidatorSrcAddress)
		suite.Equal(dstValAddr.String(), redelegation.ValidatorDstAddress)
		suite.Equal(checkPoint1, redelegation.Entries[0].CreationHeight)
		suite.Equal(types.ChunkSize.ToDec().String(), redelegation.Entries[0].SharesDst.String())
		del := suite.app.StakingKeeper.Delegation(suite.ctx, chunk.DerivedAddress(), dstValAddr)
		suite.Equal(types.ChunkSize.ToDec().String(), del.GetShares().String())
	}

	beforeSlashedDelShares := suite.app.StakingKeeper.Delegation(suite.ctx, chunk.DerivedAddress(), dstValAddr).GetShares()
	beforeSlashedVal := suite.app.StakingKeeper.Validator(suite.ctx, dstValAddr)

	// double-sign slashing happened before checkPoint1
	suite.tombstone(suite.ctx, srcValAddr, srcValPubKey, checkPoint1-1)

	slashingParams := suite.app.SlashingKeeper.GetParams(suite.ctx)
	expectedPenalty := slashingParams.SlashFractionDoubleSign.Mul(types.ChunkSize.ToDec()).TruncateInt()
	afterSlashedDelShares := suite.app.StakingKeeper.Delegation(suite.ctx, chunk.DerivedAddress(), dstValAddr).GetShares()
	afterSlashedVal := suite.app.StakingKeeper.Validator(suite.ctx, dstValAddr)
	// Slashing re-delegation calls unbond internally which deducts tokens and del shares also from Validator
	{
		suite.True(afterSlashedDelShares.LT(beforeSlashedDelShares))
		suite.True(afterSlashedVal.GetDelegatorShares().LT(beforeSlashedVal.GetDelegatorShares()))
		suite.Equal(
			expectedPenalty.String(),
			beforeSlashedVal.GetDelegatorShares().Sub(afterSlashedVal.GetDelegatorShares()).TruncateInt().String(),
		)
		suite.True(afterSlashedVal.GetTokens().LT(beforeSlashedVal.GetTokens()))
		suite.Equal(
			expectedPenalty.String(),
			beforeSlashedVal.GetTokens().Sub(afterSlashedVal.GetTokens()).String(),
		)
	}

	unpairingInsBalBeforeCover := suite.app.BankKeeper.GetBalance(suite.ctx, unpairingInsurance.DerivedAddress(), oneInsurance.Denom)
	suite.ctx = suite.advanceEpoch(suite.ctx)
	suite.ctx = suite.advanceHeight(suite.ctx, 1, "unpairing insurance should covers slashing penalty")
	unpairingInsBalAfterCover := suite.app.BankKeeper.GetBalance(suite.ctx, unpairingInsurance.DerivedAddress(), oneInsurance.Denom)

	afterCoverDelShares := suite.app.StakingKeeper.Delegation(suite.ctx, chunk.DerivedAddress(), dstValAddr).GetShares()
	afterCoverVal := suite.app.StakingKeeper.Validator(suite.ctx, dstValAddr)
	// Check state is correct after slashing penalty covered by unpairing insurance
	{
		suite.True(afterCoverDelShares.Equal(beforeSlashedDelShares))
		suite.True(afterCoverVal.GetDelegatorShares().Equal(beforeSlashedVal.GetDelegatorShares()))
		suite.True(afterCoverVal.GetTokens().Equal(beforeSlashedVal.GetTokens()))
		suite.True(unpairingInsBalAfterCover.IsLT(unpairingInsBalBeforeCover))
		suite.Equal(
			expectedPenalty.String(),
			unpairingInsBalBeforeCover.Sub(unpairingInsBalAfterCover).Amount.String(),
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
			types.ChunkSize.MulRaw(500),
		},
	)
	suite.ctx = suite.advanceHeight(suite.ctx, 1, "liquid staking started")
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
	// 7% + 3.75% = 10.75%
	// After tombstone, it still pass the line (5.75%) which means
	// The chunk will not be unpairing because of IsEnoughToCoverSlash check
	suite.fundAccount(suite.ctx, pairedInsurance.DerivedAddress(), types.ChunkSize.ToDec().Mul(sdk.NewDecWithPrec(375, 2)).Ceil().TruncateInt())

	selfDelegationToken := suite.app.StakingKeeper.TokensFromConsensusPower(suite.ctx, onePower)
	// handle a signature to set signing info
	suite.app.SlashingKeeper.HandleValidatorSignature(
		suite.ctx,
		toBeTombstonedValidatorPubKey.Address(),
		selfDelegationToken.Int64(),
		true,
	)

	val := suite.app.StakingKeeper.Validator(suite.ctx, toBeTombstonedValidator)
	pairedInsuranceBalance := suite.app.BankKeeper.GetBalance(suite.ctx, pairedInsurance.DerivedAddress(), env.bondDenom)
	suite.tombstone(suite.ctx, toBeTombstonedValidator, toBeTombstonedValidatorPubKey, suite.ctx.BlockHeight()-1)

	suite.ctx = suite.advanceHeight(suite.ctx, 1, "one block passed afetr validator is tombstoned because of double signing")
	fmt.Println(suite.app.LiquidStakingKeeper.GetNetAmountStateEssentials(suite.ctx))

	suite.ctx = suite.advanceEpoch(suite.ctx)
	suite.ctx = suite.advanceHeight(suite.ctx, 1, "chunk started to be unpairing")
	passedRewardsEpochAfterTombstoned := int64(2)

	pairedInsuranceBalanceAfterCoveringSlash := suite.app.BankKeeper.GetBalance(suite.ctx, pairedInsurance.DerivedAddress(), env.bondDenom)
	fmt.Println(suite.app.LiquidStakingKeeper.GetNetAmountStateEssentials(suite.ctx))
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
			"2999988000047975000",
			suite.app.BankKeeper.GetBalance(suite.ctx, tombstonedInsurance.FeePoolAddress(), env.bondDenom).Amount.String(),
			fmt.Sprintf(
				"after tombstoned, tombstoned insurance couldn't get commissions corresponding %d * unit commission",
				passedRewardsEpochAfterTombstoned,
			),
		)
		suite.Equal(
			"11999952000191975000",
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
		suite.True(
			unbondingDelegation.Entries[0].InitialBalance.GTE(types.ChunkSize),
			"there were no candidate insurance to pair, so unbonding of chunk started",
		)
	}

	suite.ctx = suite.advanceHeight(suite.ctx, 1, "")
	fmt.Println(suite.app.LiquidStakingKeeper.GetNetAmountStateEssentials(suite.ctx))

	suite.ctx = suite.advanceEpoch(suite.ctx)
	suite.ctx = suite.advanceHeight(suite.ctx, 1, "unpairing of chunk is finished")

	{
		tombstonedChunkAfterUnpairing, _ := suite.app.LiquidStakingKeeper.GetChunk(suite.ctx, toBeTombstonedChunk.Id)
		suite.Equal(types.CHUNK_STATUS_PAIRING, tombstonedChunkAfterUnpairing.Status)
		suite.True(
			suite.app.BankKeeper.GetBalance(suite.ctx, tombstonedChunk.DerivedAddress(), env.bondDenom).Amount.GTE(types.ChunkSize),
			"chunk's balance must be gte chunk size",
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
			// insurance 0,3,6, will be invalid insurances (all pointing v1)
			// and insurance 10, 11, 13 will be replaced. (pointing v2, v3, v2)
			14,
			sdk.NewDecWithPrec(10, 2),
			nil,
			9,
			types.ChunkSize.MulRaw(500),
		},
	)
	suite.ctx = suite.advanceHeight(suite.ctx, 1, "liquid staking started")
	fmt.Println(suite.app.LiquidStakingKeeper.GetNetAmountStateEssentials(suite.ctx))

	toBeTombstonedValidator := env.valAddrs[0]
	toBeTombstonedValidatorPubKey := env.pubKeys[0]
	toBeTombstonedChunks := []types.Chunk{env.pairedChunks[0], env.pairedChunks[3], env.pairedChunks[6]}
	pairedInsurances := []types.Insurance{env.insurances[0], env.insurances[3], env.insurances[6]}
	toBeNewlyRankedInsurances := []types.Insurance{env.insurances[10], env.insurances[11], env.insurances[13]}
	{
		// 0, 3, 6 are paired currently but will be unpaired because it points to toBeTombstonedValidator
		// 0, 3, 6 must have 5.75% chunkSize as balance after tombstoned to be re-delegated, please check IsEnoughToCoverSlash.
		for i := 0; i < len(pairedInsurances); i++ {
			suite.Equal(pairedInsurances[i].Id, toBeTombstonedChunks[i].PairedInsuranceId)
			suite.Equal(toBeTombstonedValidator, pairedInsurances[i].GetValidator())
			// 7% + 3.75% = 10.75%
			// After 5% slashing => 5.75%
			suite.fundAccount(suite.ctx, pairedInsurances[i].DerivedAddress(), types.ChunkSize.ToDec().Mul(sdk.NewDecWithPrec(375, 2)).Ceil().TruncateInt())
		}
		// 10, 11, 13 are not paired currently but will be paired because it points to valid validator
		for i := 0; i < len(toBeNewlyRankedInsurances); i++ {
			suite.NotEqual(toBeTombstonedValidator, toBeNewlyRankedInsurances[i].GetValidator())
		}
	}
	targetInsurancesBalance := suite.app.BankKeeper.GetBalance(suite.ctx, pairedInsurances[0].DerivedAddress(), env.bondDenom).Amount

	// Tombstone validator
	{
		selfDelegationToken := suite.app.StakingKeeper.TokensFromConsensusPower(suite.ctx, onePower)
		// handle a signature to set signing info
		suite.app.SlashingKeeper.HandleValidatorSignature(
			suite.ctx,
			toBeTombstonedValidatorPubKey.Address(),
			selfDelegationToken.Int64(),
			true,
		)
		suite.tombstone(suite.ctx, toBeTombstonedValidator, toBeTombstonedValidatorPubKey, suite.ctx.BlockHeight()-1)
	}

	suite.ctx = suite.advanceHeight(suite.ctx, 1, "one block passed after validator is tombstoned because of double signing")
	fmt.Println(suite.app.LiquidStakingKeeper.GetNetAmountStateEssentials(suite.ctx))

	suite.ctx = suite.advanceEpoch(suite.ctx)
	suite.ctx = suite.advanceHeight(suite.ctx, 1, "re-pairing of chunks is finished")

	val := suite.app.StakingKeeper.Validator(suite.ctx, toBeTombstonedValidator)
	suite.Equal(stakingtypes.Unbonded, val.GetStatus())

	// check re-delegations are created
	{
		for i, pairedInsuranceBeforeTombstoned := range pairedInsurances {
			tombstonedInsurance, _ := suite.app.LiquidStakingKeeper.GetInsurance(suite.ctx, pairedInsuranceBeforeTombstoned.Id)
			suite.Equal(types.INSURANCE_STATUS_UNPAIRING, tombstonedInsurance.Status)
			chunk, _ := suite.app.LiquidStakingKeeper.GetChunk(suite.ctx, tombstonedInsurance.ChunkId)
			suite.Equal(types.CHUNK_STATUS_PAIRED, chunk.Status) // problem point
			suite.Equal(tombstonedInsurance.Id, chunk.UnpairingInsuranceId)
			newInsurance, _ := suite.app.LiquidStakingKeeper.GetInsurance(suite.ctx, toBeNewlyRankedInsurances[i].Id)
			suite.Equal(types.INSURANCE_STATUS_PAIRED, newInsurance.Status)
			suite.Equal(newInsurance.Id, chunk.PairedInsuranceId)

			// check re-delegation happened or not
			_, found := suite.app.StakingKeeper.GetRedelegation(
				suite.ctx,
				chunk.DerivedAddress(),
				tombstonedInsurance.GetValidator(),
				newInsurance.GetValidator(),
			)
			suite.False(found, "because src validator is tombstoned, there are no re-delegation. it was immediately re-delegated")
			del, _ := suite.app.StakingKeeper.GetDelegation(
				suite.ctx,
				chunk.DerivedAddress(),
				newInsurance.GetValidator(),
			)
			dstVal := suite.app.StakingKeeper.Validator(suite.ctx, newInsurance.GetValidator())
			suite.True(dstVal.TokensFromShares(del.GetShares()).GTE(types.ChunkSize.ToDec()))
		}
	}

	suite.ctx = suite.advanceHeight(suite.ctx, 1, "")

	suite.ctx = suite.advanceEpoch(suite.ctx)
	suite.ctx = suite.advanceHeight(suite.ctx, 1, "un-pairing insurances are unpaired")
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
				).Amount.LT(targetInsurancesBalance),
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
			types.ChunkSize.MulRaw(500),
		},
	)
	suite.ctx = suite.advanceHeight(suite.ctx, 1, "liquid staking started")
	fmt.Println(suite.app.LiquidStakingKeeper.GetNetAmountStateEssentials(suite.ctx))
	toBeTombstonedValidator := env.valAddrs[0]
	toBeTombstonedValidatorPubKey := env.pubKeys[0]
	toBeTombstonedChunks := []types.Chunk{env.pairedChunks[0], env.pairedChunks[3], env.pairedChunks[6]}
	pairedInsurances := []types.Insurance{env.insurances[0], env.insurances[3], env.insurances[6]}
	{
		for i := 0; i < len(pairedInsurances); i++ {
			suite.Equal(pairedInsurances[i].Id, toBeTombstonedChunks[i].PairedInsuranceId)
			suite.Equal(toBeTombstonedValidator, pairedInsurances[i].GetValidator())
			// 7% + 3.75% = 10.75%
			// After tombstone, it still pass the line (5.75%) which means
			// The chunk will not be unpairing because of IsEnoughToCoverSlash check
			suite.fundAccount(suite.ctx, pairedInsurances[i].DerivedAddress(), types.ChunkSize.ToDec().Mul(sdk.NewDecWithPrec(375, 2)).Ceil().TruncateInt())

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
	var pairedInsuranceBalances []sdk.Coin
	for _, pairedInsurance := range pairedInsurances {
		pairedInsuranceBalances = append(
			pairedInsuranceBalances,
			suite.app.BankKeeper.GetBalance(suite.ctx, pairedInsurance.DerivedAddress(), env.bondDenom),
		)
	}
	val := suite.app.StakingKeeper.Validator(suite.ctx, toBeTombstonedValidator)
	suite.tombstone(suite.ctx, toBeTombstonedValidator, toBeTombstonedValidatorPubKey, suite.ctx.BlockHeight()-1)
	suite.ctx = suite.advanceHeight(suite.ctx, 1, "one block passed after validator is tombstoned because of double signing")
	fmt.Println(suite.app.LiquidStakingKeeper.GetNetAmountStateEssentials(suite.ctx))

	suite.ctx = suite.advanceEpoch(suite.ctx)
	suite.ctx = suite.advanceHeight(suite.ctx, 1, "chunks started to be unpairing")

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
			suite.True(
				unbondingDelegation.Entries[0].InitialBalance.GTE(types.ChunkSize),
			)
		}
	}

	suite.ctx = suite.advanceHeight(suite.ctx, 1, "")
	fmt.Println(suite.app.LiquidStakingKeeper.GetNetAmountStateEssentials(suite.ctx))

	suite.ctx = suite.advanceEpoch(suite.ctx)
	suite.ctx = suite.advanceHeight(suite.ctx, 1, "unpairing of chunk is finished")

	{
		for i, toBeTombstonedChunk := range toBeTombstonedChunks {
			tombstonedChunkAfterUnpairing, _ := suite.app.LiquidStakingKeeper.GetChunk(suite.ctx, toBeTombstonedChunk.Id)
			suite.Equal(types.CHUNK_STATUS_PAIRING, tombstonedChunkAfterUnpairing.Status)
			suite.True(
				suite.app.BankKeeper.GetBalance(suite.ctx, tombstonedChunks[i].DerivedAddress(), env.bondDenom).Amount.GTE(types.ChunkSize),
				"chunk's balance must be GTE chunk size",
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

func (suite *KeeperTestSuite) TestUnpairingForUnstakingChunkTombstoned() {
	env := suite.setupLiquidStakeTestingEnv(
		testingEnvOptions{
			"TestUnpairingForUnstakingChunkTombstoned",
			3,
			TenPercentFeeRate,
			nil,
			onePower,
			nil,
			3,
			TenPercentFeeRate,
			nil,
			3,
			types.ChunkSize.MulRaw(500),
		},
	)
	numPassedRewardEpochsBeforeUnstaked := int64(0)
	suite.ctx = suite.advanceHeight(suite.ctx, 1, "liquid staking started")
	numPassedRewardEpochsBeforeUnstaked++
	fmt.Println(suite.app.LiquidStakingKeeper.GetNetAmountStateEssentials(suite.ctx))

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

	suite.ctx = suite.advanceEpoch(suite.ctx)
	suite.ctx = suite.advanceHeight(suite.ctx, 1, "unstaking started")
	numPassedRewardEpochsBeforeUnstaked++
	fmt.Println(suite.app.LiquidStakingKeeper.GetNetAmountStateEssentials(suite.ctx))

	var pairedInsuranceBalanceAfterUnstakingStarted sdk.Coin
	var pairedInsuranceCommissionAfterUnstakingStarted sdk.Coin
	{
		// check whether liquid unstaking started or not
		chunk, _ := suite.app.LiquidStakingKeeper.GetChunk(suite.ctx, toBeUnstakedChunk.Id)
		suite.Equal(types.CHUNK_STATUS_UNPAIRING_FOR_UNSTAKING, chunk.Status)
		info, _ := suite.app.LiquidStakingKeeper.GetUnpairingForUnstakingChunkInfo(suite.ctx, chunk.Id)
		suite.Equal(chunk.Id, info.ChunkId)
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

	suite.ctx = suite.advanceHeight(suite.ctx, 1, "")
	fmt.Println(suite.app.LiquidStakingKeeper.GetNetAmountStateEssentials(suite.ctx))

	selfDelegationToken := suite.app.StakingKeeper.TokensFromConsensusPower(suite.ctx, onePower)
	// handle a signature to set signing info
	suite.app.SlashingKeeper.HandleValidatorSignature(
		suite.ctx,
		toBeTombstonedValidatorPubKey.Address(),
		selfDelegationToken.Int64(),
		true,
	)
	suite.tombstone(suite.ctx, toBeTombstonedValidator, toBeTombstonedValidatorPubKey, suite.ctx.BlockHeight()-1)
	suite.ctx = suite.advanceHeight(suite.ctx, 1, "one block passed afetr validator is tombstoned because of double signing")
	fmt.Println(suite.app.LiquidStakingKeeper.GetNetAmountStateEssentials(suite.ctx))

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

	suite.ctx = suite.advanceEpoch(suite.ctx)
	suite.ctx = suite.advanceHeight(suite.ctx, 1, "epoch reached after validator is tombstoned because of double signing")
	fmt.Println(suite.app.LiquidStakingKeeper.GetNetAmountStateEssentials(suite.ctx))

	{
		_, found := suite.app.LiquidStakingKeeper.GetChunk(suite.ctx, toBeUnstakedChunk.Id)
		suite.False(found, "liquid unstaking of chunk is finished")
		undelegatorBalance := suite.app.BankKeeper.GetBalance(suite.ctx, undelegator, env.bondDenom)
		suite.Equal(
			types.ChunkSize.String(),
			undelegatorBalance.Sub(undelegatorInitialBalance).Amount.String(),
			"because insuracne covered penalty, undelegator get all unstaked amount",
		)
		insurance, _ := suite.app.LiquidStakingKeeper.GetInsurance(suite.ctx, toBeUnstakedChunk.PairedInsuranceId)
		suite.Equal(types.INSURANCE_STATUS_UNPAIRED, insurance.Status)
		balance := suite.app.BankKeeper.GetBalance(suite.ctx, insurance.DerivedAddress(), env.bondDenom)
		suite.Equal(
			penalty.String(),
			pairedInsuranceBalanceAfterUnstakingStarted.Sub(balance).Amount.String(),
			"insurance covered penalty after epoch reached",
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
			expectedPenaltyPercent: 59,
		},
		{
			name:                   "block time is 1 second including tombstone",
			blockTime:              time.Second,
			includeTombstone:       true,
			expectedPenaltyPercent: 61,
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
		suite.Run(tc.name, func() {
			// Must call this to refresh state
			suite.SetupTest()
			initialHeight := int64(1)
			suite.ctx = suite.ctx.WithBlockHeight(initialHeight) // make sure we start with clean height
			suite.fundAccount(suite.ctx, fundingAccount, types.ChunkSize.MulRaw(500))
			valNum := 2
			delAddrs, _ := suite.AddTestAddrsWithFunding(fundingAccount, valNum, suite.app.StakingKeeper.TokensFromConsensusPower(suite.ctx, 200))
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
			providers, providerBalances := suite.AddTestAddrsWithFunding(fundingAccount, 2, oneInsurance.Amount)
			suite.provideInsurances(suite.ctx, providers, valAddrs, providerBalances, TenPercentFeeRate, nil)
			delegators, delegatorBalances := suite.AddTestAddrsWithFunding(fundingAccount, 2, oneChunk.Amount)
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
				suite.tombstone(suite.ctx, downValAddr, downValPubKey, suite.ctx.BlockHeight()-1)
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
			suite.ctx = suite.advanceEpoch(suite.ctx)
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
			suite.ctx = suite.advanceEpoch(suite.ctx)
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
		})
	}
}

func (suite *KeeperTestSuite) TestDynamicFee() {
	fixedInsuranceFeeRate := TenPercentFeeRate
	tcs := []struct {
		name                        string
		numVals                     int
		numPairedChunks             int
		numInsurances               int
		fundingAccountBalance       sdk.Int
		unitDelegationReward        string
		u                           sdk.Dec
		dynamicFeeRate              sdk.Dec
		uAfterEpoch                 sdk.Dec
		dynamicFeeRateAfterEpoch    sdk.Dec
		rewardPoolBalanceAfterEpoch sdk.Int
	}{
		{
			"almost max fee rate",
			3,
			3,
			10,
			types.ChunkSize.MulRaw(32),
			"29999880000479750000",
			sdk.MustNewDecFromStr("0.093749964843763184"),
			sdk.MustNewDecFromStr("0.249998593750527360"),
			sdk.MustNewDecFromStr("0.093756620650551778"),
			sdk.MustNewDecFromStr("0.250264826022071120"),
			sdk.MustNewDecFromStr("60720863401021956426").TruncateInt(),
		},
		{
			"about +1% from softcap",
			3,
			2,
			10,
			types.ChunkSize.MulRaw(30),
			"44999730001619750000",
			sdk.MustNewDecFromStr("0.066666640000010667"),
			sdk.MustNewDecFromStr("0.041666600000026668"),
			sdk.MustNewDecFromStr("0.066665751123684568"),
			sdk.MustNewDecFromStr("0.041664377809211420"),
			sdk.MustNewDecFromStr("77622532705413356534").TruncateInt(),
		},
	}

	for _, tc := range tcs {
		suite.Run(tc.name, func() {
			suite.SetupTest()
			suite.setupLiquidStakeTestingEnv(
				testingEnvOptions{
					tc.name,
					tc.numVals,
					TenPercentFeeRate,
					nil,
					onePower,
					nil,
					tc.numInsurances,
					fixedInsuranceFeeRate,
					nil,
					tc.numPairedChunks,
					tc.fundingAccountBalance,
				},
			)
			{
				nase := suite.app.LiquidStakingKeeper.GetNetAmountStateEssentials(suite.ctx)
				fmt.Println(nase)
				// Check current state before reaching epoch
				suite.Equal(
					tc.u.String(),
					nase.UtilizationRatio.String(),
				)
				suite.Equal(
					tc.dynamicFeeRate.String(),
					nase.FeeRate.String(),
				)
			}
			beforeNas := suite.app.LiquidStakingKeeper.GetNetAmountStateEssentials(suite.ctx)
			suite.ctx = suite.advanceEpoch(suite.ctx)
			suite.ctx = suite.advanceHeight(suite.ctx, 1, "got rewards and dynamic fee is charged")
			nase := suite.app.LiquidStakingKeeper.GetNetAmountStateEssentials(suite.ctx)
			suite.True(
				nase.RewardModuleAccBalance.GT(beforeNas.RewardModuleAccBalance),
				"reward module account's balance increased",
			)
			suite.Equal(tc.rewardPoolBalanceAfterEpoch.String(), nase.RewardModuleAccBalance.String())
		})
	}
}

func (suite *KeeperTestSuite) TestCalcDiscountRate() {
	suite.setupLiquidStakeTestingEnv(
		testingEnvOptions{
			"TestCalcDiscountRate",
			3,
			TenPercentFeeRate,
			nil,
			onePower,
			nil,
			3,
			TenPercentFeeRate,
			nil,
			1,
			types.ChunkSize.MulRaw(500),
		},
	)
	tcs := []struct {
		name                 string
		numRewardEpochs      int
		expectedDiscountRate sdk.Dec
	}{
		{
			"a lot of rewards but cannot exceed MaximumDiscountRate",
			100,
			types.DefaultMaximumDiscountRate,
		},
		{
			"small reward",
			10,
			sdk.MustNewDecFromStr("0.003229497673565564"),
		},
	}

	for _, tc := range tcs {
		suite.Run(tc.name, func() {
			cachedCtx, _ := suite.ctx.CacheContext()
			cachedCtx = suite.advanceHeight(cachedCtx, tc.numRewardEpochs-1, fmt.Sprintf("let's pass %d reward epoch", tc.numRewardEpochs))
			cachedCtx = suite.advanceEpoch(cachedCtx) // reward is accumulated to reward pool
			cachedCtx = suite.advanceHeight(cachedCtx, 1, "liquid staking endblocker is triggered")
			nase := suite.app.LiquidStakingKeeper.GetNetAmountStateEssentials(cachedCtx)
			suite.Equal(tc.expectedDiscountRate.String(), nase.DiscountRate.String())
		})
	}

}

// TestDoClaimDiscountedReward tests success cases.
func (suite *KeeperTestSuite) TestDoClaimDiscountedReward() {
	env := suite.setupLiquidStakeTestingEnv(
		testingEnvOptions{
			"TestDoClaimDiscountedReward",
			3,
			TenPercentFeeRate,
			nil,
			onePower,
			nil,
			10,
			TenPercentFeeRate,
			nil,
			10,
			types.ChunkSize.MulRaw(500),
		},
	)
	type expected struct {
		discountRate                string
		lsTokenToGetAll             string
		claimAmount                 string
		claimAmountBiggerThanReward bool
		beforeMintRate              string
		beforeTokenBal              string
		beforeLsTokenBal            string
		afterMintRate               string
		afterTokenBal               string
		afterLsTokenBal             string
		increasedMintRate           string
		decreasedLsTokenBal         string
	}

	liquidBondDenom := suite.app.LiquidStakingKeeper.GetLiquidBondDenom(suite.ctx)
	tcs := []struct {
		name            string
		numRewardEpochs int
		expected
		msg *types.MsgClaimDiscountedReward
	}{
		{
			"discounted little and claim all",
			100,
			expected{
				"0.003229532439457230",
				"8047756399199309411842",
				"251622622449703783661134",
				true,
				"0.996770467560542769",
				"0",
				"250000000000000000000000",
				"0.996780897440320276",
				"8099990280011661750000",
				"241952243600800690588158",
				"0.000010429879777507",
				"8047756399199309411842",
			},
			types.NewMsgClaimDiscountedReward(
				env.delegators[0].String(),
				sdk.NewCoin(liquidBondDenom, types.ChunkSize),
				sdk.MustNewDecFromStr("0.002"),
			),
		},
		{
			"discounted little and claim little",
			100,
			expected{
				"0.003229532439457230",
				"8047756399199309411842",
				"1006",
				false,
				"0.996770467560542769",
				"0",
				"250000000000000000000000",
				"0.996770467560542769",
				"1006",
				"249999999999999999999000",
				"0.000000000000000000",
				"1000",
			},
			types.NewMsgClaimDiscountedReward(
				env.delegators[0].String(),
				sdk.NewCoin(liquidBondDenom, sdk.NewInt(1000)),
				sdk.MustNewDecFromStr("0.002"),
			),
		},
		{
			"discounted a lot and claim all",
			1000,
			expected{
				"0.030000000000000000",
				"76104134710420714265643",
				"266082464206197591875507",
				true,
				"0.968616851665805890",
				"0",
				"250000000000000000000000",
				"0.969558346115831714",
				"80999902800116637750000",
				"173895865289579285734357",
				"0.000941494450025824",
				"76104134710420714265643",
			},
			types.NewMsgClaimDiscountedReward(
				env.delegators[0].String(),
				sdk.NewCoin(liquidBondDenom, types.ChunkSize),
				sdk.MustNewDecFromStr("0.03"),
			),
		},
		{
			"discounted a lot and claim little",
			1000,
			expected{
				"0.030000000000000000",
				"76104134710420714265643",
				"106432985",
				false,
				"0.968616851665805890",
				"0",
				"250000000000000000000000",
				"0.968616851665805892",
				"106432985",
				"249999999999999900000000",
				"0.000000000000000002",
				"100000000",
			},
			types.NewMsgClaimDiscountedReward(
				env.delegators[0].String(),
				sdk.NewCoin(liquidBondDenom, sdk.NewInt(100000000)),
				sdk.MustNewDecFromStr("0.03"),
			),
		},
	}

	for _, tc := range tcs {
		suite.Run(tc.name, func() {
			cachedCtx, _ := suite.ctx.CacheContext()
			cachedCtx = suite.advanceHeight(cachedCtx, tc.numRewardEpochs-1, fmt.Sprintf("pass %d reward epoch", tc.numRewardEpochs))
			cachedCtx = suite.advanceEpoch(cachedCtx) // reward is accumulated to reward pool
			cachedCtx = suite.advanceHeight(cachedCtx, 1, "liquid staking endblocker is triggered")
			requester := tc.msg.GetRequestser()
			nase := suite.app.LiquidStakingKeeper.GetNetAmountStateEssentials(cachedCtx)
			suite.Equal(tc.expected.discountRate, nase.DiscountRate.String())
			discountedMintRate := suite.app.LiquidStakingKeeper.GetNetAmountStateEssentials(cachedCtx).MintRate.Mul(
				sdk.OneDec().Sub(nase.DiscountRate),
			)
			claimableAmt := suite.app.BankKeeper.GetBalance(cachedCtx, types.RewardPool, suite.denom)
			lsTokenToGetAll := claimableAmt.Amount.ToDec().Mul(discountedMintRate).Ceil().TruncateInt()
			claimAmt := tc.msg.Amount.Amount.ToDec().Quo(discountedMintRate).TruncateInt()
			suite.Equal(tc.lsTokenToGetAll, lsTokenToGetAll.String())
			suite.Equal(tc.claimAmount, claimAmt.String())
			suite.Equal(tc.claimAmountBiggerThanReward, claimAmt.GT(claimableAmt.Amount))
			suite.Equal(tc.expected.beforeTokenBal, suite.app.BankKeeper.GetBalance(cachedCtx, requester, suite.denom).Amount.String())
			beforeLsTokenBal := suite.app.BankKeeper.GetBalance(cachedCtx, requester, liquidBondDenom).Amount
			suite.Equal(tc.expected.beforeLsTokenBal, beforeLsTokenBal.String())
			beforeMintRate := suite.app.LiquidStakingKeeper.GetNetAmountStateEssentials(cachedCtx).MintRate
			suite.Equal(tc.expected.beforeMintRate, beforeMintRate.String())
			_, _, err := suite.app.LiquidStakingKeeper.DoClaimDiscountedReward(cachedCtx, tc.msg)
			suite.NoError(err)
			suite.Equal(tc.afterTokenBal, suite.app.BankKeeper.GetBalance(cachedCtx, requester, suite.denom).Amount.String())
			afterLsTokenBal := suite.app.BankKeeper.GetBalance(cachedCtx, requester, liquidBondDenom).Amount
			suite.Equal(tc.expected.afterLsTokenBal, afterLsTokenBal.String())
			afterMintRate := suite.app.LiquidStakingKeeper.GetNetAmountStateEssentials(cachedCtx).MintRate
			suite.Equal(tc.expected.afterMintRate, afterMintRate.String())
			suite.Equal(tc.expected.increasedMintRate, afterMintRate.Sub(beforeMintRate).String())
			suite.Equal(tc.expected.decreasedLsTokenBal, beforeLsTokenBal.Sub(afterLsTokenBal).String())
		})
	}
}

func (suite *KeeperTestSuite) TestDoClaimDiscountedRewardFail() {
	env := suite.setupLiquidStakeTestingEnv(
		testingEnvOptions{
			"TestDoClaimDiscountedRewardFail",
			3,
			TenPercentFeeRate,
			nil,
			onePower,
			nil,
			3,
			TenPercentFeeRate,
			nil,
			1,
			types.ChunkSize.MulRaw(500),
		},
	)
	suite.ctx = suite.advanceHeight(suite.ctx, 99, "pass 100 reward epoch")
	suite.ctx = suite.advanceEpoch(suite.ctx) // reward is accumulated to reward pool
	suite.ctx = suite.advanceHeight(suite.ctx, 1, "liquid staking endblocker is triggered")

	tcs := []struct {
		name        string
		msg         *types.MsgClaimDiscountedReward
		expectedErr error
	}{
		{
			name: "invalid denom",
			msg: types.NewMsgClaimDiscountedReward(
				env.delegators[0].String(),
				sdk.NewCoin("invalidDenom", sdk.NewInt(100)),
				sdk.MustNewDecFromStr("0.00001"),
			),
			expectedErr: types.ErrInvalidLiquidBondDenom,
		},
		{
			name: "mint rate is lower than min mint rate",
			msg: types.NewMsgClaimDiscountedReward(
				env.delegators[0].String(),
				sdk.NewCoin(types.DefaultLiquidBondDenom, sdk.NewInt(100)),
				sdk.MustNewDecFromStr("0.5"),
			),
			expectedErr: types.ErrDiscountRateTooLow,
		},
		{
			name: "requester does not have msg.Amount",
			msg: types.NewMsgClaimDiscountedReward(
				sdk.AccAddress(ed25519.GenPrivKey().PubKey().Address()).String(),
				sdk.NewCoin(types.DefaultLiquidBondDenom, sdk.TokensFromConsensusPower(10_000, ethermint.PowerReduction)),
				sdk.MustNewDecFromStr("0.00000001"),
			),
			expectedErr: sdkerrors.ErrInsufficientFunds,
		},
	}

	for _, tc := range tcs {
		suite.Run(tc.name, func() {
			_, _, err := suite.app.LiquidStakingKeeper.DoClaimDiscountedReward(suite.ctx, tc.msg)
			suite.ErrorContains(err, tc.expectedErr.Error())
		})
	}
}

// TestChunkPositiveBalanceBeforeEpoch tests scenario where someone sends coins to chunk.
// This is a special case because the chunk is not a normal account.
func (suite *KeeperTestSuite) TestChunkPositiveBalanceBeforeEpoch() {
	env := suite.setupLiquidStakeTestingEnv(
		testingEnvOptions{
			"TestChunkPositiveBalanceBeforeEpoch",
			3,
			TenPercentFeeRate,
			nil,
			onePower,
			nil,
			3,
			TenPercentFeeRate,
			nil,
			3,
			types.ChunkSize.MulRaw(500),
		},
	)
	suite.ctx = suite.advanceHeight(suite.ctx, 1, "pass 1 reward epoch")
	// send some coins to chunk before epoch
	coin := sdk.NewCoin(suite.denom, sdk.NewInt(100))
	suite.NoError(
		suite.app.BankKeeper.SendCoins(
			suite.ctx,
			fundingAccount,
			env.pairedChunks[0].DerivedAddress(),
			sdk.NewCoins(coin),
		),
	)

	suite.ctx = suite.advanceEpoch(suite.ctx)
	suite.ctx = suite.advanceHeight(suite.ctx, 1, "liquid staking endblocker is triggered")

	originReardModuleAccBalance, _ := sdk.NewIntFromString("161999352002591325000")
	nase := suite.app.LiquidStakingKeeper.GetNetAmountStateEssentials(suite.ctx)
	{
		additionalCommissions := coin.Amount.ToDec().Mul(TenPercentFeeRate).TruncateInt()
		suite.Equal(
			coin.Sub(sdk.NewCoin(suite.denom, additionalCommissions)).Amount.String(),
			nase.RewardModuleAccBalance.Sub(originReardModuleAccBalance).String(),
			"reward module account balance should be increased by 90% of the sent coins",
		)
	}
}

// TestRePairChunkWhichGotWithdrawInsuranceRequest tests scenario where a chunk starts unpairing
// at epoch by handling withdraw insurance request and then re-pair at current epoch with another insurance.
func (suite *KeeperTestSuite) TestRePairChunkWhichGotWithdrawInsuranceRequest() {
	env := suite.setupLiquidStakeTestingEnv(
		testingEnvOptions{
			"TestRePairChunkWhichGotWithdrawInsuranceRequest",
			3,
			TenPercentFeeRate,
			nil,
			onePower,
			nil,
			3,
			TenPercentFeeRate,
			nil,
			1,
			types.ChunkSize.MulRaw(500),
		},
	)
	chunkBeforeRePair := env.pairedChunks[0]
	toBeWithdrawn := env.insurances[0].Id
	_, req, err := suite.app.LiquidStakingKeeper.DoWithdrawInsurance(
		suite.ctx,
		types.NewMsgWithdrawInsurance(
			env.providers[0].String(),
			env.insurances[0].Id,
		),
	)
	suite.NoError(err)
	suite.Equal(toBeWithdrawn, req.InsuranceId)

	suite.ctx = suite.advanceEpoch(suite.ctx)
	suite.ctx = suite.advanceHeight(suite.ctx, 1, "withdraw insurance started")
	chunkAfterRePair, _ := suite.app.LiquidStakingKeeper.GetChunk(suite.ctx, chunkBeforeRePair.Id)
	suite.NotEqual(toBeWithdrawn, chunkAfterRePair.PairedInsuranceId)
	suite.Equal(toBeWithdrawn, chunkAfterRePair.UnpairingInsuranceId)

	suite.ctx = suite.advanceEpoch(suite.ctx)
	suite.ctx = suite.advanceHeight(suite.ctx, 1, "liquid staking endblocker is triggered")

	withdrawnInsurance, _ := suite.app.LiquidStakingKeeper.GetInsurance(suite.ctx, toBeWithdrawn)
	chunkAfterRePair, _ = suite.app.LiquidStakingKeeper.GetChunk(suite.ctx, chunkBeforeRePair.Id)
	suite.Equal(types.INSURANCE_STATUS_UNPAIRED, withdrawnInsurance.Status)
	suite.Equal(types.Empty, chunkAfterRePair.UnpairingInsuranceId, "unpairing insurance id should be cleared")
}

// TestTargetChunkGotBothUnstakeAndWithdrawInsuranceReqs tests scenario where a chunk got both
// unstake and withdraw insurance requests at the same epoch.
func (suite *KeeperTestSuite) TestTargetChunkGotBothUnstakeAndWithdrawInsuranceReqs() {
	env := suite.setupLiquidStakeTestingEnv(
		testingEnvOptions{
			"TestTargetChunkGotBothUnstakeAndWithdrawInsuranceReqs",
			3,
			TenPercentFeeRate,
			nil,
			onePower,
			nil,
			1,
			TenPercentFeeRate,
			nil,
			1,
			types.ChunkSize.MulRaw(500),
		},
	)
	_, _, err := suite.app.LiquidStakingKeeper.QueueLiquidUnstake(
		suite.ctx,
		types.NewMsgLiquidUnstake(
			env.delegators[0].String(),
			sdk.NewCoin(suite.denom, types.ChunkSize),
		),
	)
	suite.NoError(err)
	_, _, err = suite.app.LiquidStakingKeeper.DoWithdrawInsurance(
		suite.ctx,
		types.NewMsgWithdrawInsurance(
			env.providers[0].String(),
			env.insurances[0].Id,
		),
	)
	suite.NoError(err)
	suite.ctx = suite.advanceEpoch(suite.ctx)
	suite.ctx = suite.advanceHeight(suite.ctx, 1, "liquid staking endblocker is triggered")

	chunk, _ := suite.app.LiquidStakingKeeper.GetChunk(suite.ctx, env.pairedChunks[0].Id)
	suite.Equal(types.CHUNK_STATUS_UNPAIRING_FOR_UNSTAKING, chunk.Status)
	insurance, _ := suite.app.LiquidStakingKeeper.GetInsurance(suite.ctx, env.insurances[0].Id)
	suite.Equal(types.INSURANCE_STATUS_UNPAIRING_FOR_WITHDRAWAL, insurance.Status)

	beforeDelegatorBalance := suite.app.BankKeeper.GetBalance(suite.ctx, env.delegators[0], suite.denom)

	suite.ctx = suite.advanceEpoch(suite.ctx)
	suite.ctx = suite.advanceHeight(suite.ctx, 1, "unstaking and withdrawal end")

	afterDelegatorBalance := suite.app.BankKeeper.GetBalance(suite.ctx, env.delegators[0], suite.denom)
	oneChunk, oneInsurance := suite.app.LiquidStakingKeeper.GetMinimumRequirements(suite.ctx)
	unpairedInsuranceBal := suite.app.BankKeeper.GetBalance(suite.ctx, env.insurances[0].DerivedAddress(), suite.denom)
	suite.True(
		unpairedInsuranceBal.IsGTE(oneInsurance),
		"unpaired insurance got its coins back",
	)
	suite.Equal(
		beforeDelegatorBalance.Add(oneChunk).String(),
		afterDelegatorBalance.String(),
	)
}

// TestOnlyOnePairedChunkGotDamagedSoNoChunksAvailableToUnstake tests scenario where
// matched chunk with unstaking request got damaged during period(queued ~ epoch), so it cannot be unstaked at epoch.
// The unstake request will be canceled and the coins will be returned to the delegator.
func (suite *KeeperTestSuite) TestOnlyOnePairedChunkGotDamagedSoNoChunksAvailableToUnstake() {
	initialHeight := int64(1)
	suite.ctx = suite.ctx.WithBlockHeight(initialHeight) // make sure we start with clean height
	suite.fundAccount(suite.ctx, fundingAccount, types.ChunkSize.MulRaw(500))
	valNum := 2
	addrs, _ := suite.AddTestAddrsWithFunding(fundingAccount, valNum, suite.app.StakingKeeper.TokensFromConsensusPower(suite.ctx, 200))
	valAddrs := simapp.ConvertAddrsToValAddrs(addrs)
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

	// Let's create 1 chunk and 1 insurance
	oneChunk, oneInsurance := suite.app.LiquidStakingKeeper.GetMinimumRequirements(suite.ctx)
	providers, providerBalances := suite.AddTestAddrsWithFunding(fundingAccount, 1, oneInsurance.Amount)
	suite.provideInsurances(suite.ctx, providers, valAddrs, providerBalances, TenPercentFeeRate, nil)
	delegators, delegatorBalances := suite.AddTestAddrsWithFunding(fundingAccount, 1, oneChunk.Amount)
	suite.liquidStakes(suite.ctx, delegators, delegatorBalances)
	suite.ctx = suite.ctx.WithBlockHeight(suite.ctx.BlockHeight() + 1)
	staking.EndBlocker(suite.ctx, suite.app.StakingKeeper)

	// Queue liquid unstake before huge slashing started
	_, infos, err := suite.app.LiquidStakingKeeper.QueueLiquidUnstake(
		suite.ctx,
		types.NewMsgLiquidUnstake(
			delegators[0].String(),
			sdk.NewCoin(suite.denom, oneChunk.Amount),
		),
	)
	suite.NoError(err)
	delBal := suite.app.BankKeeper.GetBalance(suite.ctx, delegators[0], types.DefaultLiquidBondDenom)
	suite.Equal(
		"0",
		delBal.Amount.String(),
		"delegator's lstoken is escrowed",
	)

	downValAddr := valAddrs[0]
	downValPubKey := pubKeys[0]
	// toBeUnpairedChunk := pairedChunks[0]

	epoch := suite.app.LiquidStakingKeeper.GetEpoch(suite.ctx)
	epochTime := suite.ctx.BlockTime().Add(epoch.Duration)
	called := 0
	// huge slashing started
	for {
		validator, _ := suite.app.StakingKeeper.GetValidatorByConsAddr(suite.ctx, sdk.GetConsAddress(downValPubKey))
		suite.downTimeSlashing(
			suite.ctx,
			downValPubKey,
			validator.GetConsensusPower(suite.app.StakingKeeper.PowerReduction(suite.ctx)),
			called,
			time.Second,
		)
		suite.unjail(suite.ctx, downValAddr, downValPubKey, time.Second)
		called++

		if suite.ctx.BlockTime().After(epochTime) {
			break
		}
	}
	suite.ctx = suite.advanceEpoch(suite.ctx)
	staking.EndBlocker(suite.ctx, suite.app.StakingKeeper)
	liquidstakingkeeper.EndBlocker(suite.ctx, suite.app.LiquidStakingKeeper)
	_, found := suite.app.LiquidStakingKeeper.GetUnpairingForUnstakingChunkInfo(suite.ctx, infos[0].ChunkId)
	suite.False(
		found,
		"When unstake request is queued, matched chunk was fine but after validator got slashed, "+
			"chunk is not available to unstake and info is deleted.",
	)
	delBal = suite.app.BankKeeper.GetBalance(suite.ctx, delegators[0], types.DefaultLiquidBondDenom)
	suite.Equal(
		infos[0].EscrowedLstokens.String(),
		delBal.String(),
		"delegator's get back escrowed ls tokens, because unstake is not processed",
	)
}

// TestPenaltyCoverageInSameValidatorRePairing tests whether unpairing insurance which have same validator with
// paired insurance and ranked-out from previous epoch cover penalty before the epoch or not.
func (suite *KeeperTestSuite) TestPenaltyCoverageInSameValidatorRePairing() {
	env := suite.setupLiquidStakeTestingEnv(
		testingEnvOptions{
			"TestTargetChunkGotBothUnstakeAndWithdrawInsuranceReqs",
			2,
			TenPercentFeeRate,
			nil,
			onePower,
			nil,
			2,
			sdk.ZeroDec(),
			[]sdk.Dec{TenPercentFeeRate, FivePercentFeeRate},
			2,
			types.ChunkSize.MulRaw(500),
		},
	)
	_, oneInsurance := suite.app.LiquidStakingKeeper.GetMinimumRequirements(suite.ctx)
	providers, bals := suite.AddTestAddrsWithFunding(fundingAccount, 1, oneInsurance.Amount)
	// provide insurance which have same validator and have cheaper fee rate
	insurances := suite.provideInsurances(suite.ctx, providers, []sdk.ValAddress{env.valAddrs[0]}, bals, OnePercentFeeRate, nil)
	newIns := insurances[0]

	chunk := env.pairedChunks[1]
	outIns := env.insurances[0]
	suite.Equal(outIns.Id, chunk.PairedInsuranceId)

	suite.ctx = suite.advanceHeight(suite.ctx, 5, "5 block passed")
	// before re-delegation
	evidenceHeight := suite.ctx.BlockHeight()
	suite.ctx = suite.advanceHeight(suite.ctx, 5, "5 block passed")

	suite.ctx = suite.advanceEpoch(suite.ctx)
	suite.ctx = suite.advanceHeight(suite.ctx, 1, "re-pairing")
	chunk, _ = suite.app.LiquidStakingKeeper.GetChunk(suite.ctx, chunk.Id)
	suite.Equal(newIns.Id, chunk.PairedInsuranceId)

	suite.ctx = suite.advanceHeight(suite.ctx, 10, "10 blocks passed")
	// tombstoned after re-delegation
	suite.tombstone(suite.ctx, env.valAddrs[0], env.pubKeys[0], evidenceHeight)

	// penalty must be covered by outIns
	suite.ctx = suite.advanceEpoch(suite.ctx)
	suite.ctx = suite.advanceHeight(suite.ctx, 1, "unpairing insurance will cover double sign slashing penalty")
}

func (suite *KeeperTestSuite) TestGetAllRePairableChunksAndOutInsurances() {
	// create 3 paired chunks
	// create 2 unpairing for unstaking chunks
	// create 2 unpairing chunk wihtout unbonding delegation obj
	// create 1 unpairing chunk with unbonding delegation obj
	env := suite.setupLiquidStakeTestingEnv(
		testingEnvOptions{
			"TestGetAllRePairableChunksAndOutInsurances",
			8,
			TenPercentFeeRate,
			nil,
			onePower,
			nil,
			8,
			TenPercentFeeRate,
			nil,
			8,
			types.ChunkSize.MulRaw(500),
		},
	)
	// 1, 2, 3: paired chunks - Pairable
	paraibles := []uint64{1, 2, 3}
	bondDenom := suite.app.StakingKeeper.BondDenom(suite.ctx)
	// 7, 8: unpairing for unstaking chunks - Not pairable
	notPairables := []uint64{7, 8}
	for i := 6; i <= 7; i++ {
		suite.app.LiquidStakingKeeper.QueueLiquidUnstake(
			suite.ctx, types.NewMsgLiquidUnstake(env.delegators[i].String(), sdk.NewCoin(bondDenom, types.ChunkSize)),
		)
	}
	infos := suite.app.LiquidStakingKeeper.GetAllUnpairingForUnstakingChunkInfos(suite.ctx)
	suite.Require().Equal(2, len(infos))
	suite.Equal(uint64(7), infos[0].ChunkId)
	suite.Equal(uint64(8), infos[1].ChunkId)

	// 5, 6: Unpairing chunk without unbonding delegation obj - Pairable
	paraibles = append(paraibles, 5, 6)
	for i := 5; i <= 6; i++ {
		suite.app.LiquidStakingKeeper.DoWithdrawInsurance(
			suite.ctx, types.NewMsgWithdrawInsurance(env.insurances[i-1].ProviderAddress, env.insurances[i-1].Id),
		)
	}
	reqs := suite.app.LiquidStakingKeeper.GetAllWithdrawInsuranceRequests(suite.ctx)
	suite.Equal(uint64(5), reqs[0].InsuranceId)
	suite.Equal(uint64(6), reqs[1].InsuranceId)

	// 4: Unpairing chunk with unbonding delegation obj -> Not Pairable
	// Damaged chunk situation
	notPairables = append(notPairables, 4)
	tombstoneValAddr := env.valAddrs[3]
	tombstonePubKey := env.pubKeys[3]
	notEnoughIns := env.insurances[3]

	_, oneIns := suite.app.LiquidStakingKeeper.GetMinimumRequirements(suite.ctx)

	suite.ctx = suite.advanceHeight(suite.ctx, 5, "5 block passed")
	suite.tombstone(suite.ctx, tombstoneValAddr, tombstonePubKey, suite.ctx.BlockHeight())
	// Let's make notEnoughIns to cover tombstone
	suite.app.BankKeeper.SendCoins(suite.ctx, notEnoughIns.DerivedAddress(), types.RewardPool, sdk.NewCoins(oneIns))
	suite.ctx = suite.advanceEpoch(suite.ctx)
	suite.ctx = suite.advanceHeight(suite.ctx, 1, "testing env is set finally")

	expectedRepairableChunkIds := make(map[uint64]bool)
	notRepairableChunkIds := make(map[uint64]bool)
	{
		for _, id := range paraibles {
			expectedRepairableChunkIds[id] = true
		}
		for _, id := range notPairables {
			notRepairableChunkIds[id] = true
		}
	}
	expectedOutInsurances := make(map[uint64]bool)
	{
		expectedOutInsurances[5] = true
		expectedOutInsurances[6] = true
	}
	rePairableChunks, outInsurances, _ := suite.app.LiquidStakingKeeper.GetAllRePairableChunksAndOutInsurances(suite.ctx)
	for _, chunk := range rePairableChunks {
		suite.True(expectedRepairableChunkIds[chunk.Id])
		suite.False(notRepairableChunkIds[chunk.Id])
	}
	for _, ins := range outInsurances {
		suite.True(expectedOutInsurances[ins.Id])
	}
}

func (suite *KeeperTestSuite) TestCalcCeiledPenalty() {
	env := suite.setupLiquidStakeTestingEnv(
		testingEnvOptions{
			"TestTargetChunkGotBothUnstakeAndWithdrawInsuranceReqs",
			1,
			TenPercentFeeRate,
			nil,
			onePower,
			nil,
			2,
			sdk.ZeroDec(),
			[]sdk.Dec{TenPercentFeeRate, FivePercentFeeRate},
			2,
			types.ChunkSize.MulRaw(500),
		},
	)
	toBeTombstonedValidator := env.valAddrs[0]
	toBeTombstonedValidatorPubKey := env.pubKeys[0]

	// Make tombstoned validator
	{
		selfDelegationToken := suite.app.StakingKeeper.TokensFromConsensusPower(suite.ctx, onePower)
		// handle a signature to set signing info
		suite.app.SlashingKeeper.HandleValidatorSignature(
			suite.ctx,
			toBeTombstonedValidatorPubKey.Address(),
			selfDelegationToken.Int64(),
			true,
		)
		suite.tombstone(suite.ctx, toBeTombstonedValidator, toBeTombstonedValidatorPubKey, suite.ctx.BlockHeight()-1)
	}

	validator, _ := suite.app.StakingKeeper.GetValidator(suite.ctx, toBeTombstonedValidator)
	del, _ := suite.app.StakingKeeper.GetDelegation(suite.ctx, env.pairedChunks[0].DerivedAddress(), validator.GetOperator())

	tokens := validator.TokensFromShares(del.GetShares()).Ceil().TruncateInt()
	penaltyAmt := types.ChunkSize.Sub(tokens)
	suite.Equal("12500000000000000000000", penaltyAmt.String())
	// penalty value was exactly 5% of chunk size tokens, but what if we delegate additionally with this token?

	// Mimic CalcCeiledPenalty to see what happens if we delegate with penaltyAmt
	penaltyShares, _ := validator.SharesFromTokens(penaltyAmt)
	suite.Equal("13157894736842105263157.894736842105263157", penaltyShares.String())
	sharesToToken := validator.TokensFromShares(penaltyShares)
	suite.Equal(
		"12499999999999999999999.999999999999999999", sharesToToken.String(),
		"if we delegate with penaltyAmt additionally, then the actual token value of added can be less than penaltyAmt",
	)

	// Now let's use CalcCeiledPenalty
	result := suite.app.LiquidStakingKeeper.CalcCeiledPenalty(validator, penaltyAmt)
	suite.Equal("12500000000000000000001", result.String())
	suite.True(
		result.GT(penaltyAmt),
		"to cover penalty fully by delegate more to chunk, must be greater than penaltyAmt",
	)
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
	ctx sdk.Context, valAddr sdk.ValAddress, valPubKey cryptotypes.PubKey, evidenceHeight int64,
) {
	validator := suite.app.StakingKeeper.Validator(ctx, valAddr)
	power := validator.GetConsensusPower(suite.app.StakingKeeper.PowerReduction(ctx))
	evidence := &evidencetypes.Equivocation{
		Height:           evidenceHeight,
		Time:             ctx.BlockTime(),
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
	jailDuration := suite.app.SlashingKeeper.GetParams(ctx).DowntimeJailDuration
	blockNum := int64(jailDuration / blockTime)
	suite.ctx = ctx.WithBlockHeight(
		ctx.BlockHeight() + blockNum,
	).WithBlockTime(
		ctx.BlockTime().Add(jailDuration),
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

func (suite *KeeperTestSuite) calcTotalInsuranceCommissions(status types.InsuranceStatus) (totalCommission sdk.Int) {
	totalCommission = sdk.ZeroInt()
	suite.app.LiquidStakingKeeper.IterateAllInsurances(suite.ctx, func(insurance types.Insurance) bool {
		if insurance.Status == status {
			commission := suite.app.BankKeeper.GetBalance(suite.ctx, insurance.FeePoolAddress(), suite.denom)
			totalCommission = totalCommission.Add(commission.Amount)
		}
		return false
	})
	return
}
