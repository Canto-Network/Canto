package keeper_test

import (
	"fmt"
	"github.com/Canto-Network/Canto/v6/x/liquidstaking/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/simulation"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"math/rand"
)

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

	addrs, balances := suite.AddTestAddrs(types.MaxPairedChunks, minimumRequirement.Amount)

	// TC: There are no pairing insurances yet. Insurances must be provided to liquid stake
	acc1 := addrs[0]
	msg := types.NewMsgLiquidStake(acc1.String(), minimumRequirement)
	_, _, _, err := suite.app.LiquidStakingKeeper.DoLiquidStake(suite.ctx, msg)
	suite.ErrorIs(err, types.ErrNoPairingInsurance)

	providers, providerBalances := suite.AddTestAddrs(10, minimumCoverage.Amount)
	suite.provideInsurances(providers, valAddrs, providerBalances)

	// TC: Not enough amount to liquid stake
	// acc1 tries to liquid stake 2 * ChunkSize tokens, but he has only ChunkSize tokens
	msg = types.NewMsgLiquidStake(acc1.String(), minimumRequirement.AddAmount(types.ChunkSize))
	_, _, _, err = suite.app.LiquidStakingKeeper.DoLiquidStake(suite.ctx, msg)
	suite.ErrorIs(err, sdkerrors.ErrInsufficientFunds)

	// Pairs as many chunks as the MaxPairedChunks
	_ = suite.liquidStakes(addrs, balances)

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
	insurance, found := suite.app.LiquidStakingKeeper.GetInsurance(suite.ctx, chunk.InsuranceId)
	suite.True(found)

	_, err = suite.app.LiquidStakingKeeper.DoCancelInsuranceProvide(
		suite.ctx,
		types.NewMsgCancelInsuranceProvide(insurance.ProviderAddress, insurance.Id),
	)
	suite.ErrorIs(err, types.ErrPairingInsuranceNotFound, "only pairing insurances can be canceled")
}
