package keeper_test

import (
	"math"

	sdkmath "cosmossdk.io/math"
	"github.com/cometbft/cometbft/crypto/tmhash"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	govtypesv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/evmos/ethermint/crypto/ethsecp256k1"

	"github.com/Canto-Network/Canto/v7/testutil"
	"github.com/Canto-Network/Canto/v7/x/coinswap/types"
)

var (
	sender = sdk.AccAddress(tmhash.SumTruncated([]byte("sender"))).String()
)

func (suite *TestSuite) TestMsgSwapOrder_ValidateBasic() {
	msg := types.MsgSwapOrder{}
	suite.Require().Equal("/canto.coinswap.v1.MsgSwapOrder", sdk.MsgTypeURL(&msg))

	type fields struct {
		Input      types.Input
		Output     types.Output
		Deadline   int64
		IsBuyOrder bool
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{name: "invalid input sender", wantErr: true, fields: fields{IsBuyOrder: true, Deadline: 10, Input: types.Input{Address: "", Coin: buildCoin("stake", 1000)}, Output: types.Output{Address: sender, Coin: buildCoin("iris", 1000)}}},
		{name: "invalid input coin  denom", wantErr: true, fields: fields{IsBuyOrder: true, Deadline: 10, Input: types.Input{Address: sender, Coin: buildCoin("invalidstake", 1000)}, Output: types.Output{Address: sender, Coin: buildCoin("iris", 1000)}}},
		{name: "invalid input coin amount", wantErr: true, fields: fields{IsBuyOrder: true, Deadline: 10, Input: types.Input{Address: sender, Coin: buildCoin("stake", -1000)}, Output: types.Output{Address: sender, Coin: buildCoin("iris", 1000)}}},
		{name: "invalid output sender", wantErr: true, fields: fields{IsBuyOrder: true, Deadline: 10, Input: types.Input{Address: sender, Coin: buildCoin("stake", 1000)}, Output: types.Output{Address: "", Coin: buildCoin("iris", 1000)}}},
		{name: "invalid output coin denom", wantErr: true, fields: fields{IsBuyOrder: true, Deadline: 10, Input: types.Input{Address: sender, Coin: buildCoin("stake", 1000)}, Output: types.Output{Address: sender, Coin: buildCoin("131iris", 1000)}}},
		{name: "invalid output coin amount", wantErr: true, fields: fields{IsBuyOrder: true, Deadline: 10, Input: types.Input{Address: sender, Coin: buildCoin("stake", 1000)}, Output: types.Output{Address: sender, Coin: buildCoin("iris", -1000)}}},
	}
	for _, tt := range tests {
		suite.Run(tt.name, func() {
			msg := types.MsgSwapOrder{
				Input:      tt.fields.Input,
				Output:     tt.fields.Output,
				Deadline:   tt.fields.Deadline,
				IsBuyOrder: tt.fields.IsBuyOrder,
			}
			err := suite.app.CoinswapKeeper.Swap(suite.ctx, &msg)
			if tt.wantErr {
				suite.Require().Error(err)
			} else {
				suite.Require().NoError(err)
			}
		})
	}
}

func (suite *TestSuite) TestMsgAddLiquidity_ValidateBasic() {
	msg := types.MsgAddLiquidity{}
	suite.Require().Equal("/canto.coinswap.v1.MsgAddLiquidity", sdk.MsgTypeURL(&msg))

	type fields struct {
		MaxToken         sdk.Coin
		ExactStandardAmt sdkmath.Int
		MinLiquidity     sdkmath.Int
		Deadline         int64
		Sender           string
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name:    "invalid MaxToken denom",
			wantErr: true,
			fields: fields{
				MaxToken:         buildCoin("invalidstake", 1000),
				ExactStandardAmt: sdkmath.NewInt(100),
				MinLiquidity:     sdkmath.NewInt(100),
				Deadline:         1611213344,
				Sender:           sender,
			},
		},
		{
			name:    "invalid MaxToken amount",
			wantErr: true,
			fields: fields{
				MaxToken:         buildCoin("stake", -1000),
				ExactStandardAmt: sdkmath.NewInt(100),
				MinLiquidity:     sdkmath.NewInt(100),
				Deadline:         1611213344,
				Sender:           sender,
			},
		},
		{
			name:    "invalid ExactStandardAmt",
			wantErr: true,
			fields: fields{
				MaxToken:         buildCoin("stake", 1000),
				ExactStandardAmt: sdkmath.NewInt(-100),
				MinLiquidity:     sdkmath.NewInt(100),
				Deadline:         1611213344,
				Sender:           sender,
			},
		},
		{
			name:    "invalid MinLiquidity",
			wantErr: true,
			fields: fields{
				MaxToken:         buildCoin("stake", 1000),
				ExactStandardAmt: sdkmath.NewInt(100),
				MinLiquidity:     sdkmath.NewInt(-100),
				Deadline:         1611213344,
				Sender:           sender,
			},
		},
		{
			name:    "invalid Deadline",
			wantErr: true,
			fields: fields{
				MaxToken:         buildCoin("stake", 1000),
				ExactStandardAmt: sdkmath.NewInt(100),
				MinLiquidity:     sdkmath.NewInt(100),
				Deadline:         0,
				Sender:           sender,
			},
		},
		{
			name:    "invalid Sender",
			wantErr: true,
			fields: fields{
				MaxToken:         buildCoin("stake", 1000),
				ExactStandardAmt: sdkmath.NewInt(100),
				MinLiquidity:     sdkmath.NewInt(100),
				Deadline:         0,
				Sender:           "",
			},
		},
	}
	for _, tt := range tests {
		suite.Run(tt.name, func() {
			msg := types.MsgAddLiquidity{
				MaxToken:         tt.fields.MaxToken,
				ExactStandardAmt: tt.fields.ExactStandardAmt,
				MinLiquidity:     tt.fields.MinLiquidity,
				Deadline:         tt.fields.Deadline,
				Sender:           tt.fields.Sender,
			}
			res, err := suite.app.CoinswapKeeper.AddLiquidity(suite.ctx, &msg)
			if tt.wantErr {
				suite.Require().Error(err)
				suite.Require().False(res.IsValid())
			} else {
				suite.Require().NoError(err)
				suite.Require().True(res.IsValid())
			}
		})
	}
}

func (suite *TestSuite) TestMsgRemoveLiquidity_ValidateBasic() {
	msg := types.MsgRemoveLiquidity{}
	suite.Require().Equal("/canto.coinswap.v1.MsgRemoveLiquidity", sdk.MsgTypeURL(&msg))

	type fields struct {
		WithdrawLiquidity sdk.Coin
		MinToken          sdkmath.Int
		MinStandardAmt    sdkmath.Int
		Deadline          int64
		Sender            string
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name:    "right test case",
			wantErr: true,
			fields: fields{
				WithdrawLiquidity: buildCoin("stake", 1000),
				MinToken:          sdkmath.NewInt(100),
				MinStandardAmt:    sdkmath.NewInt(100),
				Deadline:          1611213344,
				Sender:            sender,
			},
		},
		{
			name:    "invalid WithdrawLiquidity denom",
			wantErr: true,
			fields: fields{
				WithdrawLiquidity: buildCoin("invalidstake", 1000),
				MinToken:          sdkmath.NewInt(100),
				MinStandardAmt:    sdkmath.NewInt(100),
				Deadline:          1611213344,
				Sender:            sender,
			},
		},
		{
			name:    "invalid MinToken",
			wantErr: true,
			fields: fields{
				WithdrawLiquidity: buildCoin("stake", -1000),
				MinToken:          sdkmath.NewInt(-100),
				MinStandardAmt:    sdkmath.NewInt(100),
				Deadline:          1611213344,
				Sender:            sender,
			},
		},
		{
			name:    "invalid MinStandardAmt",
			wantErr: true,
			fields: fields{
				WithdrawLiquidity: buildCoin("stake", 1000),
				MinToken:          sdkmath.NewInt(100),
				MinStandardAmt:    sdkmath.NewInt(-100),
				Deadline:          1611213344,
				Sender:            sender,
			},
		},
		{
			name:    "invalid Deadline",
			wantErr: true,
			fields: fields{
				WithdrawLiquidity: buildCoin("stake", 1000),
				MinToken:          sdkmath.NewInt(100),
				MinStandardAmt:    sdkmath.NewInt(100),
				Deadline:          0,
				Sender:            sender,
			},
		},
		{
			name:    "invalid Sender",
			wantErr: true,
			fields: fields{
				WithdrawLiquidity: buildCoin("stake", 1000),
				MinToken:          sdkmath.NewInt(100),
				MinStandardAmt:    sdkmath.NewInt(100),
				Deadline:          0,
				Sender:            "",
			},
		},
	}
	for _, tt := range tests {
		suite.Run(tt.name, func() {
			msg := types.MsgRemoveLiquidity{
				WithdrawLiquidity: tt.fields.WithdrawLiquidity,
				MinToken:          tt.fields.MinToken,
				MinStandardAmt:    tt.fields.MinStandardAmt,
				Deadline:          tt.fields.Deadline,
				Sender:            tt.fields.Sender,
			}
			res, err := suite.app.CoinswapKeeper.RemoveLiquidity(suite.ctx, &msg)
			if tt.wantErr {
				suite.Require().Error(err)
				suite.Require().True(res.IsValid())
			} else {
				suite.Require().NoError(err)
				suite.Require().False(res.IsValid())
			}
		})
	}
}

func (suite *TestSuite) TestMsgExecutionByProposal() {
	suite.SetupTest()

	// set denom
	stakingParams, err := suite.app.StakingKeeper.GetParams(suite.ctx)
	suite.Require().NoError(err)
	denom := stakingParams.BondDenom

	// change mindeposit for denom
	govParams, err := suite.app.GovKeeper.Params.Get(suite.ctx)
	suite.Require().NoError(err)
	govParams.MinDeposit = []sdk.Coin{sdk.NewCoin(denom, sdkmath.NewInt(1))}
	err = suite.app.GovKeeper.Params.Set(suite.ctx, govParams)
	suite.Require().NoError(err)

	// create account
	privKey, err := ethsecp256k1.GenerateKey()
	suite.Require().NoError(err)
	proposer := sdk.AccAddress(privKey.PubKey().Address().Bytes())

	// create account and deligate to validator
	initAmount := sdkmath.NewInt(int64(math.Pow10(18)) * 2)
	initBalance := sdk.NewCoins(sdk.NewCoin(denom, initAmount))
	testutil.FundAccount(suite.app.BankKeeper, suite.ctx, proposer, initBalance)
	shares, err := suite.app.StakingKeeper.Delegate(suite.ctx, proposer, sdk.DefaultPowerReduction, stakingtypes.Unbonded, suite.validator, true)
	suite.Require().NoError(err)
	suite.Require().True(shares.GT(sdkmath.LegacyNewDec(0)))

	testCases := []struct {
		name      string
		msg       sdk.Msg
		checkFunc func(uint64)
	}{
		{
			"ok - proposal MsgUpdateParams",
			&types.MsgUpdateParams{
				Authority: authtypes.NewModuleAddress(govtypes.ModuleName).String(),
				Params: types.Params{
					Fee:                    sdkmath.LegacyNewDec(0),
					PoolCreationFee:        sdk.Coin{denomStandard, sdkmath.ZeroInt()},
					TaxRate:                sdkmath.LegacyNewDec(0),
					MaxStandardCoinPerPool: sdkmath.NewInt(10_000_000_000),
					MaxSwapAmount: sdk.NewCoins(sdk.NewInt64Coin(denomBTC, 10_000_000),
						sdk.NewInt64Coin(denomETH, 10_000_000),
					),
				},
			},
			func(proposalId uint64) {
				changeParams := types.Params{
					Fee:                    sdkmath.LegacyNewDec(0),
					PoolCreationFee:        sdk.Coin{denomStandard, sdkmath.ZeroInt()},
					TaxRate:                sdkmath.LegacyNewDec(0),
					MaxStandardCoinPerPool: sdkmath.NewInt(10_000_000_000),
					MaxSwapAmount: sdk.NewCoins(sdk.NewInt64Coin(denomBTC, 10_000_000),
						sdk.NewInt64Coin(denomETH, 10_000_000),
					),
				}

				proposal, err := suite.app.GovKeeper.Proposals.Get(suite.ctx, proposalId)
				suite.Require().NoError(err)
				suite.Require().Equal(govtypesv1.ProposalStatus_PROPOSAL_STATUS_PASSED, proposal.Status)
				suite.Require().Equal(suite.app.CoinswapKeeper.GetParams(suite.ctx), changeParams)
			},
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			// submit proposal
			proposal, err := suite.app.GovKeeper.SubmitProposal(suite.ctx, []sdk.Msg{tc.msg}, "", "test", "description", proposer, false)
			suite.Require().NoError(err)
			suite.Commit()

			ok, err := suite.app.GovKeeper.AddDeposit(suite.ctx, proposal.Id, proposer, govParams.MinDeposit)
			suite.Require().NoError(err)
			suite.Require().True(ok)
			suite.Commit()

			err = suite.app.GovKeeper.AddVote(suite.ctx, proposal.Id, proposer, govtypesv1.NewNonSplitVoteOption(govtypesv1.OptionYes), "")
			suite.Require().NoError(err)
			suite.CommitAfter(*govParams.VotingPeriod)

			// check proposal result
			tc.checkFunc(proposal.Id)
		})
	}
}

func buildCoin(denom string, amt int64) sdk.Coin {
	return sdk.Coin{
		Denom:  denom,
		Amount: sdkmath.NewInt(amt),
	}
}
