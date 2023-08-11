package keeper_test

import (
	"github.com/Canto-Network/Canto/v7/x/liquidstaking/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (suite *KeeperTestSuite) TestGRPCParams() {
	resp, err := suite.app.LiquidStakingKeeper.Params(sdk.WrapSDKContext(suite.ctx), &types.QueryParamsRequest{})
	suite.NoError(err)
	suite.Equal(suite.app.LiquidStakingKeeper.GetParams(suite.ctx), resp.Params)
}

func (suite *KeeperTestSuite) TestGRPCEpoch() {
	resp, err := suite.app.LiquidStakingKeeper.Epoch(sdk.WrapSDKContext(suite.ctx), &types.QueryEpochRequest{})
	suite.NoError(err)
	suite.Equal(suite.app.LiquidStakingKeeper.GetEpoch(suite.ctx), resp.Epoch)
}

func (suite *KeeperTestSuite) TestGRPCChunks() {
	suite.setupLiquidStakeTestingEnv(testingEnvOptions{
		desc:                  "",
		numVals:               3,
		fixedValFeeRate:       TenPercentFeeRate,
		valFeeRates:           nil,
		fixedPower:            1,
		powers:                nil,
		numInsurances:         3,
		fixedInsuranceFeeRate: TenPercentFeeRate,
		insuranceFeeRates:     nil,
		numPairedChunks:       3,
		fundingAccountBalance: types.ChunkSize.MulRaw(1000),
	})

	for _, tc := range []struct {
		name      string
		req       *types.QueryChunksRequest
		expectErr bool
		postRun   func(response *types.QueryChunksResponse)
	}{
		{
			"nil request",
			nil,
			true,
			nil,
		},
		{
			"query all",
			&types.QueryChunksRequest{},
			false,
			func(response *types.QueryChunksResponse) {
				suite.Len(response.Chunks, 3)
			},
		},
		{
			"query only paired chunks",
			&types.QueryChunksRequest{
				Status: types.CHUNK_STATUS_PAIRED,
			},
			false,
			func(response *types.QueryChunksResponse) {
				suite.Len(response.Chunks, 3)
			},
		},
		{
			"query only pairing chunks",
			&types.QueryChunksRequest{
				Status: types.CHUNK_STATUS_PAIRING,
			},
			false,
			func(response *types.QueryChunksResponse) {
				suite.Len(response.Chunks, 0)
			},
		},
	} {
		suite.Run(tc.name, func() {
			resp, err := suite.app.LiquidStakingKeeper.Chunks(sdk.WrapSDKContext(suite.ctx), tc.req)
			if tc.expectErr {
				suite.Error(err)
				return
			}
			suite.NoError(err)
			if tc.postRun != nil {
				tc.postRun(resp)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestGRPCChunk() {
	env := suite.setupLiquidStakeTestingEnv(testingEnvOptions{
		desc:                  "",
		numVals:               3,
		fixedValFeeRate:       TenPercentFeeRate,
		valFeeRates:           nil,
		fixedPower:            1,
		powers:                nil,
		numInsurances:         3,
		fixedInsuranceFeeRate: TenPercentFeeRate,
		insuranceFeeRates:     nil,
		numPairedChunks:       3,
		fundingAccountBalance: types.ChunkSize.MulRaw(1000),
	})

	for _, tc := range []struct {
		name      string
		req       *types.QueryChunkRequest
		expectErr bool
		postRun   func(response *types.QueryChunkResponse)
	}{
		{
			"nil request",
			nil,
			true,
			nil,
		},
		{
			"invalid request",
			&types.QueryChunkRequest{},
			true,
			nil,
		},
		{
			"query by id",
			&types.QueryChunkRequest{
				Id: 1,
			},
			false,
			func(response *types.QueryChunkResponse) {
				chunk := env.pairedChunks[0]
				suite.True(chunk.Equal(response.Chunk))
			},
		},
		{
			"query by non-existing id",
			&types.QueryChunkRequest{
				Id: types.Empty,
			},
			true,
			nil,
		},
	} {
		suite.Run(tc.name, func() {
			resp, err := suite.app.LiquidStakingKeeper.Chunk(sdk.WrapSDKContext(suite.ctx), tc.req)
			if tc.expectErr {
				suite.Error(err)
				return
			}
			suite.NoError(err)
			if tc.postRun != nil {
				tc.postRun(resp)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestGRPCInsurances() {
	env := suite.setupLiquidStakeTestingEnv(testingEnvOptions{
		desc:                  "",
		numVals:               3,
		fixedValFeeRate:       TenPercentFeeRate,
		valFeeRates:           nil,
		fixedPower:            1,
		powers:                nil,
		numInsurances:         5,
		fixedInsuranceFeeRate: TenPercentFeeRate,
		insuranceFeeRates:     nil,
		numPairedChunks:       3,
		fundingAccountBalance: types.ChunkSize.MulRaw(1000),
	})

	for _, tc := range []struct {
		name      string
		req       *types.QueryInsurancesRequest
		expectErr bool
		postRun   func(response *types.QueryInsurancesResponse)
	}{
		{
			"nil request",
			nil,
			true,
			nil,
		},
		{
			"query all",
			&types.QueryInsurancesRequest{},
			false,
			func(response *types.QueryInsurancesResponse) {
				suite.Len(response.Insurances, 5)
			},
		},
		{
			"query only paired insurances",
			&types.QueryInsurancesRequest{
				Status: types.INSURANCE_STATUS_PAIRED,
			},
			false,
			func(response *types.QueryInsurancesResponse) {
				suite.Len(response.Insurances, 3)
			},
		},
		{
			"query only pairing insurances",
			&types.QueryInsurancesRequest{
				Status: types.INSURANCE_STATUS_PAIRING,
			},
			false,
			func(response *types.QueryInsurancesResponse) {
				suite.Len(response.Insurances, 2)
			},
		},
		{
			"query by provider address",
			&types.QueryInsurancesRequest{
				ProviderAddress: env.providers[0].String(),
			},
			false,
			func(response *types.QueryInsurancesResponse) {
				suite.Len(response.Insurances, 1)
			},
		},
		{
			"query by validator address",
			&types.QueryInsurancesRequest{
				ValidatorAddress: env.valAddrs[0].String(),
			},
			false,
			func(response *types.QueryInsurancesResponse) {
				suite.Len(response.Insurances, 2)
			},
		},
	} {
		suite.Run(tc.name, func() {
			resp, err := suite.app.LiquidStakingKeeper.Insurances(sdk.WrapSDKContext(suite.ctx), tc.req)
			if tc.expectErr {
				suite.Error(err)
				return
			}
			suite.NoError(err)
			if tc.postRun != nil {
				tc.postRun(resp)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestGRPCInsurance() {
	env := suite.setupLiquidStakeTestingEnv(testingEnvOptions{
		desc:                  "",
		numVals:               3,
		fixedValFeeRate:       TenPercentFeeRate,
		valFeeRates:           nil,
		fixedPower:            1,
		powers:                nil,
		numInsurances:         3,
		fixedInsuranceFeeRate: TenPercentFeeRate,
		insuranceFeeRates:     nil,
		numPairedChunks:       3,
		fundingAccountBalance: types.ChunkSize.MulRaw(1000),
	})

	for _, tc := range []struct {
		name      string
		req       *types.QueryInsuranceRequest
		expectErr bool
		postRun   func(response *types.QueryInsuranceResponse)
	}{
		{
			"nil request",
			nil,
			true,
			nil,
		},
		{
			"invalid request",
			&types.QueryInsuranceRequest{},
			true,
			nil,
		},
		{
			"query by id",
			&types.QueryInsuranceRequest{
				Id: 1,
			},
			false,
			func(response *types.QueryInsuranceResponse) {
				suite.True(env.insurances[0].Equal(response.Insurance))
			},
		},
		{
			"query by non-existing id",
			&types.QueryInsuranceRequest{
				Id: types.Empty,
			},
			true,
			nil,
		},
	} {
		suite.Run(tc.name, func() {
			resp, err := suite.app.LiquidStakingKeeper.Insurance(sdk.WrapSDKContext(suite.ctx), tc.req)
			if tc.expectErr {
				suite.Error(err)
				return
			}
			suite.NoError(err)
			if tc.postRun != nil {
				tc.postRun(resp)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestGRPCWithdrawInsuranceRequests() {
	env := suite.setupLiquidStakeTestingEnv(testingEnvOptions{
		desc:                  "",
		numVals:               3,
		fixedValFeeRate:       TenPercentFeeRate,
		valFeeRates:           nil,
		fixedPower:            1,
		powers:                nil,
		numInsurances:         5,
		fixedInsuranceFeeRate: TenPercentFeeRate,
		insuranceFeeRates:     nil,
		numPairedChunks:       3,
		fundingAccountBalance: types.ChunkSize.MulRaw(1000),
	})
	// 3 providers requests withdraw.
	// 3 withdraw insurance requests will be queued.
	for i := 0; i < 3; i++ {
		suite.app.LiquidStakingKeeper.DoWithdrawInsurance(
			suite.ctx,
			types.NewMsgWithdrawInsurance(
				env.providers[i].String(),
				env.insurances[i].Id,
			),
		)
	}
	for _, tc := range []struct {
		name      string
		req       *types.QueryWithdrawInsuranceRequestsRequest
		expectErr bool
		postRun   func(response *types.QueryWithdrawInsuranceRequestsResponse)
	}{
		{
			"nil request",
			nil,
			true,
			nil,
		},
		{
			"query all",
			&types.QueryWithdrawInsuranceRequestsRequest{},
			false,
			func(response *types.QueryWithdrawInsuranceRequestsResponse) {
				// Only paired or unpairing insurances can have withdraw requests.
				suite.Len(response.WithdrawInsuranceRequests, 3)
			},
		},
		{
			"query by provider  address",
			&types.QueryWithdrawInsuranceRequestsRequest{
				ProviderAddress: env.providers[0].String(),
			},
			false,
			func(response *types.QueryWithdrawInsuranceRequestsResponse) {
				suite.Len(response.WithdrawInsuranceRequests, 1)
				insurance, found := suite.app.LiquidStakingKeeper.GetInsurance(suite.ctx, response.WithdrawInsuranceRequests[0].InsuranceId)
				suite.True(found)
				suite.True(insurance.Equal(env.insurances[0]))
			},
		},
	} {
		suite.Run(tc.name, func() {
			resp, err := suite.app.LiquidStakingKeeper.WithdrawInsuranceRequests(sdk.WrapSDKContext(suite.ctx), tc.req)
			if tc.expectErr {
				suite.Error(err)
				return
			}
			suite.NoError(err)
			if tc.postRun != nil {
				tc.postRun(resp)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestGRPCWithdrawInsuranceRequest() {
	env := suite.setupLiquidStakeTestingEnv(testingEnvOptions{
		desc:                  "",
		numVals:               3,
		fixedValFeeRate:       TenPercentFeeRate,
		valFeeRates:           nil,
		fixedPower:            1,
		powers:                nil,
		numInsurances:         5,
		fixedInsuranceFeeRate: TenPercentFeeRate,
		insuranceFeeRates:     nil,
		numPairedChunks:       3,
		fundingAccountBalance: types.ChunkSize.MulRaw(1000),
	})
	// 3 providers requests withdraw.
	// 3 withdraw insurance requests will be queued.
	for i := 0; i < 3; i++ {
		suite.app.LiquidStakingKeeper.DoWithdrawInsurance(
			suite.ctx,
			types.NewMsgWithdrawInsurance(
				env.providers[i].String(),
				env.insurances[i].Id,
			),
		)
	}
	for _, tc := range []struct {
		name      string
		req       *types.QueryWithdrawInsuranceRequestRequest
		expectErr bool
		postRun   func(response *types.QueryWithdrawInsuranceRequestResponse)
	}{
		{
			"nil request",
			nil,
			true,
			nil,
		},
		{
			"invalid request",
			&types.QueryWithdrawInsuranceRequestRequest{},
			true,
			nil,
		},
		{
			"query by insurance id",
			&types.QueryWithdrawInsuranceRequestRequest{
				Id: 1,
			},
			false,
			func(response *types.QueryWithdrawInsuranceRequestResponse) {
				_, found := suite.app.LiquidStakingKeeper.GetWithdrawInsuranceRequest(
					suite.ctx, response.WithdrawInsuranceRequest.InsuranceId,
				)
				suite.True(found)
			},
		},
		{
			"query by non-existing insurance id",
			&types.QueryWithdrawInsuranceRequestRequest{
				Id: types.Empty,
			},
			true,
			nil,
		},
	} {
		suite.Run(tc.name, func() {
			resp, err := suite.app.LiquidStakingKeeper.WithdrawInsuranceRequest(sdk.WrapSDKContext(suite.ctx), tc.req)
			if tc.expectErr {
				suite.Error(err)
				return
			}
			suite.NoError(err)
			if tc.postRun != nil {
				tc.postRun(resp)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestGRPCUnpairingForUnstakingChunkInfos() {
	env := suite.setupLiquidStakeTestingEnv(testingEnvOptions{
		desc:                  "",
		numVals:               3,
		fixedValFeeRate:       TenPercentFeeRate,
		valFeeRates:           nil,
		fixedPower:            1,
		powers:                nil,
		numInsurances:         3,
		fixedInsuranceFeeRate: TenPercentFeeRate,
		insuranceFeeRates:     nil,
		numPairedChunks:       3,
		fundingAccountBalance: types.ChunkSize.MulRaw(1000),
	})
	// 3 delegators requests liquid unstake.
	// 3 unpairing for unstaking requests will be queued.
	for i := 0; i < len(env.pairedChunks); i++ {
		_, _, err := suite.app.LiquidStakingKeeper.QueueLiquidUnstake(
			suite.ctx,
			types.NewMsgLiquidUnstake(
				env.delegators[i].String(),
				sdk.NewCoin(suite.denom, types.ChunkSize),
			),
		)
		suite.NoError(err)
	}
	for _, tc := range []struct {
		name      string
		req       *types.QueryUnpairingForUnstakingChunkInfosRequest
		expectErr bool
		postRun   func(response *types.QueryUnpairingForUnstakingChunkInfosResponse)
	}{
		{
			"nil request",
			nil,
			true,
			nil,
		},
		{
			"query all",
			&types.QueryUnpairingForUnstakingChunkInfosRequest{},
			false,
			func(response *types.QueryUnpairingForUnstakingChunkInfosResponse) {
				suite.Len(response.UnpairingForUnstakingChunkInfos, len(env.pairedChunks))
			},
		},
		{
			"query queued info by delegator address",
			&types.QueryUnpairingForUnstakingChunkInfosRequest{
				DelegatorAddress: env.delegators[0].String(),
			},
			false,
			func(response *types.QueryUnpairingForUnstakingChunkInfosResponse) {
				suite.Len(response.UnpairingForUnstakingChunkInfos, 1)
			},
		},
		{
			"query info by delegator address",
			&types.QueryUnpairingForUnstakingChunkInfosRequest{
				DelegatorAddress: env.delegators[0].String(),
				Queued:           true,
			},
			false,
			func(response *types.QueryUnpairingForUnstakingChunkInfosResponse) {
				suite.Len(response.UnpairingForUnstakingChunkInfos, 1)
				chunk, found := suite.app.LiquidStakingKeeper.GetChunk(suite.ctx, response.UnpairingForUnstakingChunkInfos[0].ChunkId)
				suite.True(found)
				suite.True(chunk.Equal(env.pairedChunks[2]))
			},
		},
	} {
		suite.Run(tc.name, func() {
			resp, err := suite.app.LiquidStakingKeeper.UnpairingForUnstakingChunkInfos(sdk.WrapSDKContext(suite.ctx), tc.req)
			if tc.expectErr {
				suite.Error(err)
				return
			}
			suite.NoError(err)
			if tc.postRun != nil {
				tc.postRun(resp)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestGRPCUnpairingForUnstakingChunkInfo() {
	env := suite.setupLiquidStakeTestingEnv(testingEnvOptions{
		desc:                  "",
		numVals:               3,
		fixedValFeeRate:       TenPercentFeeRate,
		valFeeRates:           nil,
		fixedPower:            1,
		powers:                nil,
		numInsurances:         3,
		fixedInsuranceFeeRate: TenPercentFeeRate,
		insuranceFeeRates:     nil,
		numPairedChunks:       3,
		fundingAccountBalance: types.ChunkSize.MulRaw(1000),
	})
	// 3 delegators requests liquid unstake.
	// 3 unpairing for unstaking requests will be queued.
	for i := 0; i < len(env.pairedChunks); i++ {
		_, _, err := suite.app.LiquidStakingKeeper.QueueLiquidUnstake(
			suite.ctx,
			types.NewMsgLiquidUnstake(
				env.delegators[i].String(),
				sdk.NewCoin(suite.denom, types.ChunkSize),
			),
		)
		suite.NoError(err)
	}
	for _, tc := range []struct {
		name      string
		req       *types.QueryUnpairingForUnstakingChunkInfoRequest
		expectErr bool
		postRun   func(response *types.QueryUnpairingForUnstakingChunkInfoResponse)
	}{
		{
			"nil request",
			nil,
			true,
			nil,
		},
		{
			"invalid request",
			&types.QueryUnpairingForUnstakingChunkInfoRequest{},
			true,
			nil,
		},
		{
			"query by chunk id",
			&types.QueryUnpairingForUnstakingChunkInfoRequest{
				Id: env.pairedChunks[0].Id,
			},
			false,
			func(response *types.QueryUnpairingForUnstakingChunkInfoResponse) {
				chunk, found := suite.app.LiquidStakingKeeper.GetChunk(suite.ctx, response.UnpairingForUnstakingChunkInfo.ChunkId)
				suite.True(found)
				suite.True(chunk.Equal(env.pairedChunks[0]))
			},
		},
		{
			"query by non-existing chunk id",
			&types.QueryUnpairingForUnstakingChunkInfoRequest{
				Id: types.Empty,
			},
			true,
			nil,
		},
	} {
		suite.Run(tc.name, func() {
			resp, err := suite.app.LiquidStakingKeeper.UnpairingForUnstakingChunkInfo(sdk.WrapSDKContext(suite.ctx), tc.req)
			if tc.expectErr {
				suite.Error(err)
				return
			}
			suite.NoError(err)
			if tc.postRun != nil {
				tc.postRun(resp)
			}
		})
	}

}

func (suite *KeeperTestSuite) TestGRPCRedelegationInfos() {
	env := suite.setupLiquidStakeTestingEnv(testingEnvOptions{
		desc:                  "",
		numVals:               3,
		fixedValFeeRate:       TenPercentFeeRate,
		valFeeRates:           nil,
		fixedPower:            1,
		powers:                nil,
		numInsurances:         3,
		fixedInsuranceFeeRate: TenPercentFeeRate,
		insuranceFeeRates:     nil,
		numPairedChunks:       3,
		fundingAccountBalance: types.ChunkSize.MulRaw(1000),
	})
	onePercentFeeRate := sdk.MustNewDecFromStr("0.01")
	newVals, _ := suite.CreateValidators(
		[]int64{onePower, onePower, onePower},
		onePercentFeeRate,
		nil,
	)
	_, oneInsurnace := suite.app.LiquidStakingKeeper.GetMinimumRequirements(suite.ctx)
	providers, bals := suite.AddTestAddrsWithFunding(fundingAccount, 3, oneInsurnace.Amount)
	// newly provided 3 insurances are more cheaper than the existing ones.
	suite.provideInsurances(suite.ctx, providers, newVals, bals, onePercentFeeRate, nil)

	suite.ctx = suite.advanceEpoch(suite.ctx)
	suite.ctx = suite.advanceHeight(suite.ctx, 1, "re-delegations is started")

	for _, tc := range []struct {
		name      string
		req       *types.QueryRedelegationInfosRequest
		expectErr bool
		postRun   func(response *types.QueryRedelegationInfosResponse)
	}{
		{
			"nil request",
			nil,
			true,
			nil,
		},
		{
			"query all",
			&types.QueryRedelegationInfosRequest{},
			false,
			func(response *types.QueryRedelegationInfosResponse) {
				suite.Len(response.RedelegationInfos, len(env.pairedChunks))
			},
		},
	} {
		suite.Run(tc.name, func() {
			resp, err := suite.app.LiquidStakingKeeper.RedelegationInfos(sdk.WrapSDKContext(suite.ctx), tc.req)
			if tc.expectErr {
				suite.Error(err)
				return
			}
			suite.NoError(err)
			if tc.postRun != nil {
				tc.postRun(resp)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestGRPCRedelegationInfo() {
	env := suite.setupLiquidStakeTestingEnv(testingEnvOptions{
		desc:                  "",
		numVals:               3,
		fixedValFeeRate:       TenPercentFeeRate,
		valFeeRates:           nil,
		fixedPower:            1,
		powers:                nil,
		numInsurances:         3,
		fixedInsuranceFeeRate: TenPercentFeeRate,
		insuranceFeeRates:     nil,
		numPairedChunks:       3,
		fundingAccountBalance: types.ChunkSize.MulRaw(1000),
	})
	onePercentFeeRate := sdk.MustNewDecFromStr("0.01")
	newVals, _ := suite.CreateValidators(
		[]int64{onePower, onePower, onePower},
		onePercentFeeRate,
		nil,
	)
	_, oneInsurnace := suite.app.LiquidStakingKeeper.GetMinimumRequirements(suite.ctx)
	providers, bals := suite.AddTestAddrsWithFunding(fundingAccount, 3, oneInsurnace.Amount)
	// newly provided 3 insurances are more cheaper than the existing ones.
	suite.provideInsurances(suite.ctx, providers, newVals, bals, onePercentFeeRate, nil)

	suite.ctx = suite.advanceEpoch(suite.ctx)
	suite.ctx = suite.advanceHeight(suite.ctx, 1, "re-delegations is started")

	for _, tc := range []struct {
		name      string
		req       *types.QueryRedelegationInfoRequest
		expectErr bool
		postRun   func(response *types.QueryRedelegationInfoResponse)
	}{
		{
			"nil request",
			nil,
			true,
			nil,
		},
		{
			"invalid request",
			&types.QueryRedelegationInfoRequest{},
			true,
			nil,
		},
		{
			"query by chunk id",
			&types.QueryRedelegationInfoRequest{
				Id: env.pairedChunks[0].Id,
			},
			false,
			func(response *types.QueryRedelegationInfoResponse) {
				suite.Equal(response.RedelegationInfo.ChunkId, env.pairedChunks[0].Id)
			},
		},
		{
			"query by non-existing chunk id",
			&types.QueryRedelegationInfoRequest{
				Id: types.Empty,
			},
			true,
			nil,
		},
	} {
		suite.Run(tc.name, func() {
			resp, err := suite.app.LiquidStakingKeeper.RedelegationInfo(sdk.WrapSDKContext(suite.ctx), tc.req)
			if tc.expectErr {
				suite.Error(err)
				return
			}
			suite.NoError(err)
			if tc.postRun != nil {
				tc.postRun(resp)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestGRPCChunkSize() {
	for _, tc := range []struct {
		name      string
		req       *types.QueryChunkSizeRequest
		expectErr bool
		postRun   func(response *types.QueryChunkSizeResponse)
	}{
		{
			"nil request",
			nil,
			true,
			nil,
		},
		{
			"query",
			&types.QueryChunkSizeRequest{},
			false,
			func(response *types.QueryChunkSizeResponse) {
				suite.Equal(response.ChunkSize.Denom, suite.denom)
				suite.True(response.ChunkSize.Amount.Equal(types.ChunkSize))
			},
		},
	} {
		suite.Run(tc.name, func() {
			resp, err := suite.app.LiquidStakingKeeper.ChunkSize(sdk.WrapSDKContext(suite.ctx), tc.req)
			if tc.expectErr {
				suite.Error(err)
				return
			}
			suite.NoError(err)
			if tc.postRun != nil {
				tc.postRun(resp)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestGRPCMinimumCollateral() {
	for _, tc := range []struct {
		name      string
		req       *types.QueryMinimumCollateralRequest
		expectErr bool
		postRun   func(response *types.QueryMinimumCollateralResponse)
	}{
		{
			"nil request",
			nil,
			true,
			nil,
		},
		{
			"query",
			&types.QueryMinimumCollateralRequest{},
			false,
			func(response *types.QueryMinimumCollateralResponse) {
				_, oneInsurance := suite.app.LiquidStakingKeeper.GetMinimumRequirements(suite.ctx)
				suite.Equal(suite.denom, response.MinimumCollateral.Denom)
				suite.True(response.MinimumCollateral.Amount.Equal(oneInsurance.Amount.ToDec()))
			},
		},
	} {
		suite.Run(tc.name, func() {
			resp, err := suite.app.LiquidStakingKeeper.MinimumCollateral(sdk.WrapSDKContext(suite.ctx), tc.req)
			if tc.expectErr {
				suite.Error(err)
				return
			}
			suite.NoError(err)
			if tc.postRun != nil {
				tc.postRun(resp)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestGRPCStates() {
	for _, tc := range []struct {
		name      string
		req       *types.QueryStatesRequest
		expectErr bool
		postRun   func(response *types.QueryStatesResponse)
	}{
		{
			"nil request",
			nil,
			true,
			nil,
		},
		{
			"query net amount state",
			&types.QueryStatesRequest{},
			false,
			func(response *types.QueryStatesResponse) {
				suite.Equal(
					suite.app.LiquidStakingKeeper.GetNetAmountState(suite.ctx),
					response.NetAmountState,
				)
			},
		},
	} {
		suite.Run(tc.name, func() {
			resp, err := suite.app.LiquidStakingKeeper.States(sdk.WrapSDKContext(suite.ctx), tc.req)
			if tc.expectErr {
				suite.Error(err)
				return
			}
			suite.NoError(err)
			if tc.postRun != nil {
				tc.postRun(resp)
			}
		})
	}
}
