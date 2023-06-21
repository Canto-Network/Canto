package keeper_test

import (
	"time"

	"github.com/Canto-Network/Canto/v6/x/liquidstaking/keeper"
	"github.com/Canto-Network/Canto/v6/x/liquidstaking/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/cosmos-sdk/x/staking"
)

func (suite *KeeperTestSuite) TestNetAmountInvariant() {
	env := suite.setupLiquidStakeTestingEnv(
		testingEnvOptions{
			"TestNetAmountInvariant",
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
	_, broken := keeper.ChunksInvariant(suite.app.LiquidStakingKeeper)(suite.ctx)
	suite.False(broken, "completely normal")

	suite.ctx = suite.advanceHeight(suite.ctx, 29, "rewards accumulated")
	suite.ctx = suite.advanceEpoch(suite.ctx)
	suite.ctx = suite.advanceHeight(suite.ctx, 1, "module epoch reached")

	nas := suite.app.LiquidStakingKeeper.GetNetAmountState(suite.ctx)
	oneChunk, oneInsurance := suite.app.LiquidStakingKeeper.GetMinimumRequirements(suite.ctx)
	suite.True(nas.Equal(types.NetAmountState{
		MintRate:                           sdk.MustNewDecFromStr("0.990373683313988266"),
		LsTokensTotalSupply:                types.ChunkSize,
		NetAmount:                          sdk.MustNewDecFromStr("252429970840349915725000"),
		TotalLiquidTokens:                  types.ChunkSize,
		RewardModuleAccBalance:             sdk.MustNewDecFromStr("2429970840349915725000").TruncateInt(),
		FeeRate:                            sdk.ZeroDec(),
		UtilizationRatio:                   sdk.MustNewDecFromStr("0.001999951953154277"),
		RemainingChunkSlots:                sdk.NewInt(49),
		NumPairedChunks:                    sdk.NewInt(1),
		DiscountRate:                       sdk.MustNewDecFromStr("0.009719883361399663"),
		TotalDelShares:                     types.ChunkSize.ToDec(),
		TotalRemainingRewards:              sdk.ZeroDec(),
		TotalChunksBalance:                 sdk.ZeroInt(),
		TotalUnbondingChunksBalance:        sdk.ZeroInt(),
		TotalInsuranceTokens:               oneInsurance.Amount,
		TotalPairedInsuranceTokens:         oneInsurance.Amount,
		TotalUnpairingInsuranceTokens:      sdk.ZeroInt(),
		TotalRemainingInsuranceCommissions: sdk.MustNewDecFromStr("269996760038879525000"),
	}))

	// forcefully make net amount zero
	{
		cachedCtx, _ := suite.ctx.CacheContext()
		completionTime, err := suite.app.StakingKeeper.Undelegate(
			cachedCtx,
			env.pairedChunks[0].DerivedAddress(),
			env.insurances[0].GetValidator(),
			types.ChunkSize.ToDec(),
		)
		// change completion time to duration from cachedCtx.BlockTime()
		suite.NoError(err)
		cachedCtx = cachedCtx.WithBlockHeight(
			cachedCtx.BlockHeight() + 1000,
		).WithBlockTime(
			completionTime.Add(time.Hour),
		)
		staking.EndBlocker(cachedCtx, suite.app.StakingKeeper)

		oneChunkCoins := sdk.NewCoins(oneChunk)
		reward := sdk.NewCoins(sdk.NewCoin(suite.denom, nas.RewardModuleAccBalance))
		inputs := []banktypes.Input{
			banktypes.NewInput(env.pairedChunks[0].DerivedAddress(), oneChunkCoins),
			banktypes.NewInput(types.RewardPool, reward),
		}
		outputs := []banktypes.Output{
			banktypes.NewOutput(sdk.AccAddress("1"), oneChunkCoins),
			banktypes.NewOutput(sdk.AccAddress("1"), reward),
		}

		suite.NoError(suite.app.BankKeeper.InputOutputCoins(cachedCtx, inputs, outputs))
		_, broken = keeper.NetAmountInvariant(suite.app.LiquidStakingKeeper)(cachedCtx)
		suite.True(broken, "net amount is zero")
	}

	// forcefully burn all ls tokens
	{
		cachedCtx, _ := suite.ctx.CacheContext()
		lsTokenBal := suite.app.BankKeeper.GetBalance(cachedCtx, env.delegators[0], types.DefaultLiquidBondDenom)
		suite.NoError(suite.app.BankKeeper.SendCoinsFromAccountToModule(
			cachedCtx,
			env.delegators[0],
			types.ModuleName,
			sdk.NewCoins(lsTokenBal),
		))
		suite.NoError(suite.app.BankKeeper.BurnCoins(
			cachedCtx,
			types.ModuleName,
			sdk.NewCoins(lsTokenBal),
		))
		_, broken = keeper.NetAmountInvariant(suite.app.LiquidStakingKeeper)(cachedCtx)
		suite.True(broken, "ls token is zero, but total liquid tokens is not zero")
	}
}

func (suite *KeeperTestSuite) TestChunksInvariant() {
	env := suite.setupLiquidStakeTestingEnv(
		testingEnvOptions{
			"TestChunksInvariant",
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
	_, broken := keeper.ChunksInvariant(suite.app.LiquidStakingKeeper)(suite.ctx)
	suite.False(broken, "completely normal")

	// 1: PAIRED CHUNK
	var origin, mutated types.Chunk = env.pairedChunks[0], env.pairedChunks[0]
	// forcefully change status of chunk as invalid
	{
		mutated.PairedInsuranceId = types.Empty
		suite.app.LiquidStakingKeeper.SetChunk(suite.ctx, mutated)
		_, broken = keeper.ChunksInvariant(suite.app.LiquidStakingKeeper)(suite.ctx)
		suite.True(broken, "paired chunk must have valid paired insurance id")
		// recover
		suite.app.LiquidStakingKeeper.SetChunk(suite.ctx, origin)
	}

	originIns, _ := suite.app.LiquidStakingKeeper.GetInsurance(suite.ctx, origin.PairedInsuranceId)
	// delete paired insurance
	{
		suite.app.LiquidStakingKeeper.DeleteInsurance(suite.ctx, originIns.Id)
		_, broken = keeper.ChunksInvariant(suite.app.LiquidStakingKeeper)(suite.ctx)
		suite.True(broken, "paired insurance must exist in store")
		// recover
		suite.app.LiquidStakingKeeper.SetInsurance(suite.ctx, originIns)
		suite.mustPassInvariants()
	}

	// forcefully change status of insurance as invalid
	{
		mutatedIns := originIns
		mutatedIns.Status = types.INSURANCE_STATUS_UNSPECIFIED
		suite.app.LiquidStakingKeeper.SetInsurance(suite.ctx, mutatedIns)
		_, broken = keeper.ChunksInvariant(suite.app.LiquidStakingKeeper)(suite.ctx)
		suite.True(broken, "insurance must have valid status")
		// recover
		suite.app.LiquidStakingKeeper.SetInsurance(suite.ctx, originIns)
		suite.mustPassInvariants()
	}

	originDel, _ := suite.app.StakingKeeper.GetDelegation(suite.ctx, origin.DerivedAddress(), originIns.GetValidator())
	// forcefully delete delegation obj of paired chunk
	{
		suite.app.StakingKeeper.RemoveDelegation(suite.ctx, originDel)
		_, broken = keeper.ChunksInvariant(suite.app.LiquidStakingKeeper)(suite.ctx)
		suite.True(broken, "delegation must exist in store")
		// recover
		suite.app.StakingKeeper.SetDelegation(suite.ctx, originDel)
		suite.mustPassInvariants()
	}

	// forcefully delegation shares as invalid
	{
		mutatedDel := originDel
		mutatedDel.Shares = mutatedDel.Shares.Sub(sdk.OneDec())
		suite.app.StakingKeeper.SetDelegation(suite.ctx, mutatedDel)
		_, broken = keeper.ChunksInvariant(suite.app.LiquidStakingKeeper)(suite.ctx)
		suite.True(broken, "delegation must have valid shares")
		// recover
		suite.app.StakingKeeper.SetDelegation(suite.ctx, originDel)
		suite.mustPassInvariants()
	}
	suite.ctx = suite.advanceEpoch(suite.ctx)
	suite.ctx = suite.advanceHeight(suite.ctx, 1, "")

	// 2: UNPAIRING CHUNK
	// first, create unpairing chunk
	insToBeWithdrawn, _, err := suite.app.LiquidStakingKeeper.DoWithdrawInsurance(
		suite.ctx,
		types.NewMsgWithdrawInsurance(env.insurances[2].ProviderAddress, env.insurances[2].Id),
	)
	suite.NoError(err)
	suite.ctx = suite.advanceEpoch(suite.ctx)
	suite.ctx = suite.advanceHeight(suite.ctx, 1, "start withdrawing insurance")

	origin, _ = suite.app.LiquidStakingKeeper.GetChunk(suite.ctx, insToBeWithdrawn.ChunkId)
	suite.checkUnpairingAndUnpairingForUnstakingChunks(suite.ctx, origin)

	// 3: PAIRING
	suite.ctx = suite.advanceEpoch(suite.ctx)
	suite.ctx = suite.advanceHeight(suite.ctx, 1, "unpairing finished")
	origin, _ = suite.app.LiquidStakingKeeper.GetChunk(suite.ctx, origin.Id)
	suite.Equal(
		types.CHUNK_STATUS_PAIRING, origin.Status,
		"after unpairing finished, chunk's status must be pairing",
	)
	// forcefully change paired insurance id of pairing chunk
	{
		mutated := origin
		mutated.PairedInsuranceId = 5
		suite.app.LiquidStakingKeeper.SetChunk(suite.ctx, mutated)
		_, broken = keeper.ChunksInvariant(suite.app.LiquidStakingKeeper)(suite.ctx)
		suite.True(broken, "pairing chunk must not have paired insurance id")
		// recover
		suite.app.LiquidStakingKeeper.SetChunk(suite.ctx, origin)
		suite.mustPassInvariants()
	}

	chunkBal := suite.app.BankKeeper.GetBalance(suite.ctx, origin.DerivedAddress(), suite.denom)
	suite.True(chunkBal.Amount.GTE(types.ChunkSize))
	// forcefully change chunk's balance
	{
		oneToken := sdk.NewCoins(sdk.NewCoin(suite.denom, sdk.OneInt()))
		suite.app.BankKeeper.SendCoins(
			suite.ctx,
			origin.DerivedAddress(),
			sdk.AccAddress(env.valAddrs[0]),
			oneToken,
		)
		_, broken = keeper.ChunksInvariant(suite.app.LiquidStakingKeeper)(suite.ctx)
		suite.True(broken, "chunk must have valid balance")
		// recover
		suite.app.BankKeeper.SendCoins(
			suite.ctx,
			sdk.AccAddress(env.valAddrs[0]),
			origin.DerivedAddress(),
			oneToken,
		)
		suite.mustPassInvariants()
	}

	// 4: UNPAIRING FOR UNSTAKING CHUNK
	// first, create unpairing for unstaking chunk
	toBeUnstakedChunks, _, err := suite.app.LiquidStakingKeeper.QueueLiquidUnstake(
		suite.ctx,
		types.NewMsgLiquidUnstake(
			env.delegators[0].String(),
			sdk.NewCoin(suite.denom, types.ChunkSize),
		),
	)
	suite.NoError(err)
	suite.ctx = suite.advanceEpoch(suite.ctx)
	suite.ctx = suite.advanceHeight(suite.ctx, 1, "unstaking chunk started")

	origin, _ = suite.app.LiquidStakingKeeper.GetChunk(suite.ctx, toBeUnstakedChunks[0].Id)
	suite.checkUnpairingAndUnpairingForUnstakingChunks(suite.ctx, origin)
}

func (suite *KeeperTestSuite) TestInsurancesInvariant() {
	env := suite.setupLiquidStakeTestingEnv(
		testingEnvOptions{
			"TestInsurancesInvariant",
			3,
			TenPercentFeeRate,
			nil,
			onePower,
			nil,
			3,
			TenPercentFeeRate,
			nil,
			2,
			types.ChunkSize.MulRaw(500),
		},
	)
	_, broken := keeper.ChunksInvariant(suite.app.LiquidStakingKeeper)(suite.ctx)
	suite.False(broken, "completely normal")

	// 1: PAIRING INSURANCE
	// first, create pairing insurance
	origin, _ := suite.app.LiquidStakingKeeper.GetInsurance(suite.ctx, env.insurances[2].Id)
	suite.Equal(types.INSURANCE_STATUS_PAIRING, origin.Status)
	// forcefully change status of pairing insurance
	{
		mutated := origin
		mutated.Status = types.INSURANCE_STATUS_UNSPECIFIED
		suite.app.LiquidStakingKeeper.SetInsurance(suite.ctx, mutated)
		_, broken := keeper.InsurancesInvariant(suite.app.LiquidStakingKeeper)(suite.ctx)
		suite.True(broken, "pairing insurance must have valid status")
		// recover
		suite.app.LiquidStakingKeeper.SetInsurance(suite.ctx, origin)
		suite.mustPassInvariants()
	}

	// 2: PAIRED INSURANCE
	origin, _ = suite.app.LiquidStakingKeeper.GetInsurance(suite.ctx, env.insurances[0].Id)
	suite.Equal(types.INSURANCE_STATUS_PAIRED, origin.Status)
	// forcefully change status of paired insurance
	{
		mutated := origin
		mutated.Status = types.INSURANCE_STATUS_UNSPECIFIED
		suite.app.LiquidStakingKeeper.SetInsurance(suite.ctx, mutated)
		_, broken := keeper.InsurancesInvariant(suite.app.LiquidStakingKeeper)(suite.ctx)
		suite.True(broken, "paired insurance must have valid status")
		// recover
		suite.app.LiquidStakingKeeper.SetInsurance(suite.ctx, origin)
		suite.mustPassInvariants()
	}

	// forcefully change paired chunk id
	{
		mutated := origin
		mutated.ChunkId = types.Empty
		suite.app.LiquidStakingKeeper.SetInsurance(suite.ctx, mutated)
		_, broken := keeper.InsurancesInvariant(suite.app.LiquidStakingKeeper)(suite.ctx)
		suite.True(broken, "paired insurance must have valid chunk id")
		// recover
		suite.app.LiquidStakingKeeper.SetInsurance(suite.ctx, origin)
		suite.mustPassInvariants()
	}

	// forcefully change paired chunk's status
	{
		originChunk, _ := suite.app.LiquidStakingKeeper.GetChunk(suite.ctx, origin.ChunkId)
		mutated := originChunk
		mutated.Status = types.CHUNK_STATUS_UNSPECIFIED
		suite.app.LiquidStakingKeeper.SetChunk(suite.ctx, mutated)
		_, broken := keeper.InsurancesInvariant(suite.app.LiquidStakingKeeper)(suite.ctx)
		suite.True(broken, "paired insurance must have valid chunk's status")
		// recover
		suite.app.LiquidStakingKeeper.SetChunk(suite.ctx, originChunk)
		suite.mustPassInvariants()
	}

	// 3: UNPAIRING INSURANCE
	toBeUnstakedChunks, _, err := suite.app.LiquidStakingKeeper.QueueLiquidUnstake(
		suite.ctx, types.NewMsgLiquidUnstake(
			env.delegators[0].String(),
			sdk.NewCoin(suite.denom, types.ChunkSize),
		))
	suite.NoError(err)
	suite.ctx = suite.advanceEpoch(suite.ctx)
	suite.ctx = suite.advanceHeight(suite.ctx, 1, "unstaking chunk started")

	originChunk, _ := suite.app.LiquidStakingKeeper.GetChunk(suite.ctx, toBeUnstakedChunks[0].Id)
	origin, _ = suite.app.LiquidStakingKeeper.GetInsurance(suite.ctx, originChunk.UnpairingInsuranceId)
	suite.NotEqual(types.Empty, origin.ChunkId)
	// forcefully empty chunk id
	{
		mutated := origin
		mutated.ChunkId = types.Empty
		suite.app.LiquidStakingKeeper.SetInsurance(suite.ctx, mutated)
		_, broken := keeper.InsurancesInvariant(suite.app.LiquidStakingKeeper)(suite.ctx)
		suite.True(broken, "unpairing insurance must have valid chunk id")
		// recover
		suite.app.LiquidStakingKeeper.SetInsurance(suite.ctx, origin)
		suite.mustPassInvariants()
	}

	// forcefully delete chunk
	{
		suite.app.LiquidStakingKeeper.DeleteChunk(suite.ctx, origin.ChunkId)
		_, broken := keeper.InsurancesInvariant(suite.app.LiquidStakingKeeper)(suite.ctx)
		suite.True(broken, "unpairing insurance must have valid chunk id")
		// recover
		suite.app.LiquidStakingKeeper.SetChunk(suite.ctx, originChunk)
		suite.mustPassInvariants()
	}

	// 4: UNPAIRED INSURANCE
	suite.ctx = suite.advanceEpoch(suite.ctx)
	suite.ctx = suite.advanceHeight(suite.ctx, 1, "unstaking chunk finished")

	origin, _ = suite.app.LiquidStakingKeeper.GetInsurance(suite.ctx, originChunk.UnpairingInsuranceId)
	suite.Equal(types.INSURANCE_STATUS_UNPAIRED, origin.Status)

	// forcefully change chunk id of unpaired insurance
	{
		mutated := origin
		mutated.ChunkId = 2
		suite.app.LiquidStakingKeeper.SetInsurance(suite.ctx, mutated)
		_, broken := keeper.InsurancesInvariant(suite.app.LiquidStakingKeeper)(suite.ctx)
		suite.True(broken, "unpaired insurance must have valid status")
		// recover
		suite.app.LiquidStakingKeeper.SetInsurance(suite.ctx, origin)
		suite.mustPassInvariants()
	}

	// 5: UNPAIRING FOR WITHDRAWAL
	origin, _, err = suite.app.LiquidStakingKeeper.DoWithdrawInsurance(
		suite.ctx,
		types.NewMsgWithdrawInsurance(
			env.providers[0].String(),
			env.insurances[0].Id,
		),
	)
	suite.NoError(err)
	suite.ctx = suite.advanceEpoch(suite.ctx)
	suite.ctx = suite.advanceHeight(suite.ctx, 1, "unpairing for withdrawal started")

	origin, _ = suite.app.LiquidStakingKeeper.GetInsurance(suite.ctx, origin.Id)
	suite.Equal(types.INSURANCE_STATUS_UNPAIRING_FOR_WITHDRAWAL, origin.Status)
	originChunk, _ = suite.app.LiquidStakingKeeper.GetChunk(suite.ctx, origin.ChunkId)

	// forcefully empty chunk id
	{
		mutated := origin
		mutated.ChunkId = types.Empty
		suite.app.LiquidStakingKeeper.SetInsurance(suite.ctx, mutated)
		_, broken := keeper.InsurancesInvariant(suite.app.LiquidStakingKeeper)(suite.ctx)
		suite.True(broken, "unpairing for withdrawal insurance must have valid chunk id")
		// recover
		suite.app.LiquidStakingKeeper.SetInsurance(suite.ctx, origin)
		suite.mustPassInvariants()
	}

	// forcefully delete chunk
	{
		suite.app.LiquidStakingKeeper.DeleteChunk(suite.ctx, origin.ChunkId)
		_, broken := keeper.InsurancesInvariant(suite.app.LiquidStakingKeeper)(suite.ctx)
		suite.True(broken, "unpairing for withdrawal insurance must have chunk")
		// recover
		suite.app.LiquidStakingKeeper.SetChunk(suite.ctx, originChunk)
		suite.mustPassInvariants()
	}

	// forcefully change status of chunk as invalid
	{
		mutated := originChunk
		mutated.Status = types.CHUNK_STATUS_PAIRING
		suite.app.LiquidStakingKeeper.SetChunk(suite.ctx, mutated)
		_, broken := keeper.InsurancesInvariant(suite.app.LiquidStakingKeeper)(suite.ctx)
		suite.True(broken, "unpairing for withdrawal insurance must have valid unpairing insurance id")
		// recover
		suite.app.LiquidStakingKeeper.SetChunk(suite.ctx, originChunk)
		suite.mustPassInvariants()
	}
}

func (suite *KeeperTestSuite) TestUnpairingForUnstakingChunkInfosInvariant() {
	env := suite.setupLiquidStakeTestingEnv(
		testingEnvOptions{
			"TestUnpairingForUnstakingChunkInfosInvariant",
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
	_, broken := keeper.ChunksInvariant(suite.app.LiquidStakingKeeper)(suite.ctx)
	suite.False(broken, "completely normal")

	// 1: Unstake
	_, infos, err := suite.app.LiquidStakingKeeper.QueueLiquidUnstake(
		suite.ctx,
		types.NewMsgLiquidUnstake(
			env.delegators[0].String(),
			sdk.NewCoin(suite.denom, types.ChunkSize),
		),
	)
	suite.NoError(err)
	chunkToBeUnstaked, _ := suite.app.LiquidStakingKeeper.GetChunk(suite.ctx, infos[0].ChunkId)
	suite.Equal(types.CHUNK_STATUS_PAIRED, chunkToBeUnstaked.Status)
	// forcefully delete chunk
	{
		suite.app.LiquidStakingKeeper.DeleteChunk(suite.ctx, infos[0].ChunkId)
		_, broken := keeper.UnpairingForUnstakingChunkInfosInvariant(suite.app.LiquidStakingKeeper)(suite.ctx)
		suite.True(broken, "unstaking chunk must have chunk")
		// recover
		suite.app.LiquidStakingKeeper.SetChunk(suite.ctx, chunkToBeUnstaked)
		suite.mustPassInvariants()
	}

	suite.ctx = suite.advanceEpoch(suite.ctx)
	suite.ctx = suite.advanceHeight(suite.ctx, 1, "unstaking chunk started")

	chunkToBeUnstaked, _ = suite.app.LiquidStakingKeeper.GetChunk(suite.ctx, infos[0].ChunkId)
	suite.Equal(types.CHUNK_STATUS_UNPAIRING_FOR_UNSTAKING, chunkToBeUnstaked.Status)
}

func (suite *KeeperTestSuite) TestWithdrawInsuranceRequestsInvariant() {
	env := suite.setupLiquidStakeTestingEnv(
		testingEnvOptions{
			"TestWithdrawInsuranceRequestsInvariant",
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
	_, broken := keeper.ChunksInvariant(suite.app.LiquidStakingKeeper)(suite.ctx)
	suite.False(broken, "completely normal")

	_, req, err := suite.app.LiquidStakingKeeper.DoWithdrawInsurance(
		suite.ctx,
		types.NewMsgWithdrawInsurance(
			env.providers[0].String(),
			env.insurances[0].Id,
		),
	)
	suite.NoError(err)
	origin, _ := suite.app.LiquidStakingKeeper.GetInsurance(suite.ctx, req.InsuranceId)
	suite.Equal(types.INSURANCE_STATUS_PAIRED, origin.Status)

	// forcefully delete insurance
	{
		suite.app.LiquidStakingKeeper.DeleteInsurance(suite.ctx, req.InsuranceId)
		_, broken := keeper.WithdrawInsuranceRequestsInvariant(suite.app.LiquidStakingKeeper)(suite.ctx)
		suite.True(broken, "withdraw insurance request must have insurance")
		// recover
		suite.app.LiquidStakingKeeper.SetInsurance(suite.ctx, origin)
		suite.mustPassInvariants()
	}
}

func (suite *KeeperTestSuite) checkUnpairingAndUnpairingForUnstakingChunks(
	ctx sdk.Context,
	origin types.Chunk,
) {
	// forcefully change status of chunk as invalid
	{
		mutated := origin
		mutated.UnpairingInsuranceId = types.Empty
		suite.app.LiquidStakingKeeper.SetChunk(suite.ctx, mutated)
		_, broken := keeper.ChunksInvariant(suite.app.LiquidStakingKeeper)(suite.ctx)
		suite.True(broken, "unpairing chunk must have valid unpairing insurance id")
		// recover
		suite.app.LiquidStakingKeeper.SetChunk(suite.ctx, origin)
		suite.mustPassInvariants()
	}

	originIns, _ := suite.app.LiquidStakingKeeper.GetInsurance(ctx, origin.UnpairingInsuranceId)
	// forcefully delete unpairing insurance
	{
		suite.app.LiquidStakingKeeper.DeleteInsurance(ctx, originIns.Id)
		_, broken := keeper.ChunksInvariant(suite.app.LiquidStakingKeeper)(ctx)
		suite.True(broken, "unpairing insurance must exist in store")
		// recover
		suite.app.LiquidStakingKeeper.SetInsurance(ctx, originIns)
		suite.mustPassInvariants()
	}

	ubd, _ := suite.app.StakingKeeper.GetUnbondingDelegation(ctx, origin.DerivedAddress(), originIns.GetValidator())
	// forcefully delete unbonding delegation obj of unpairing chunk
	{
		suite.app.StakingKeeper.RemoveUnbondingDelegation(ctx, ubd)
		_, broken := keeper.ChunksInvariant(suite.app.LiquidStakingKeeper)(ctx)
		suite.True(broken, "unbonding delegation must exist in store")
		// recover
		suite.app.StakingKeeper.SetUnbondingDelegation(ctx, ubd)
		suite.mustPassInvariants()
	}

	// forcefully add unbonding entry
	{
		ubd.Entries = append(ubd.Entries, ubd.Entries[0])
		suite.app.StakingKeeper.SetUnbondingDelegation(ctx, ubd)
		_, broken := keeper.ChunksInvariant(suite.app.LiquidStakingKeeper)(ctx)
		suite.True(broken, "chunk's unbonding delegation must have one entry")
		// recover
		ubd.Entries = ubd.Entries[:len(ubd.Entries)-1]
		suite.app.StakingKeeper.SetUnbondingDelegation(ctx, ubd)
		suite.mustPassInvariants()
	}

	// forcefully change initial balance of unbonding entry
	{
		ubd.Entries[0].InitialBalance = ubd.Entries[0].InitialBalance.Sub(sdk.OneInt())
		suite.app.StakingKeeper.SetUnbondingDelegation(ctx, ubd)
		_, broken := keeper.ChunksInvariant(suite.app.LiquidStakingKeeper)(ctx)
		suite.True(broken, "chunk's unbonding delegation's entry must have valid initial balance")
		// recover
		ubd.Entries[0].InitialBalance = ubd.Entries[0].InitialBalance.Add(sdk.OneInt())
		suite.app.StakingKeeper.SetUnbondingDelegation(ctx, ubd)
		suite.mustPassInvariants()
	}
}
