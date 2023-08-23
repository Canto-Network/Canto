package ante_test

import (
	"fmt"
	"github.com/Canto-Network/Canto/v7/app/ante"
	authzante "github.com/Canto-Network/Canto/v7/app/ante/cosmos"
	"github.com/Canto-Network/Canto/v7/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"time"
)

func (suite *AnteTestSuite) TestValCommissionChange() {
	suite.SetupTest(false)
	epoch := suite.app.LiquidStakingKeeper.GetEpoch(suite.ctx)
	validators := suite.app.StakingKeeper.GetAllValidators(suite.ctx)

	tests := []struct {
		desc          string
		setEpoch      func(ctx sdk.Context)
		createMsg     func() *stakingtypes.MsgEditValidator
		expectedError string
	}{
		{
			"Pass: No fee rate change",
			nil,
			func() *stakingtypes.MsgEditValidator {
				minSelfDelegation := validators[0].GetMinSelfDelegation()
				feeRate := validators[0].GetCommission()
				return stakingtypes.NewMsgEditValidator(
					validators[0].GetOperator(), stakingtypes.Description{}, &feeRate, &minSelfDelegation,
				)
			},
			"",
		},
		{
			"Pass: 23 hours and 49 minutes left to the next epoch",
			func(ctx sdk.Context) {
				// 23 hours and 49 minutes left to the next epoch
				nextEpochTime := ctx.BlockTime().Add(23*time.Hour + 49*time.Minute)
				epoch.StartTime = nextEpochTime.Add(-epoch.Duration)
				suite.app.LiquidStakingKeeper.SetEpoch(suite.ctx, epoch)
			},
			func() *stakingtypes.MsgEditValidator {
				minSelfDelegation := validators[0].GetMinSelfDelegation()
				feeRateChanged := validators[0].GetCommission().Add(sdk.NewDecWithPrec(1, 2))
				return stakingtypes.NewMsgEditValidator(
					validators[0].GetOperator(), stakingtypes.Description{}, &feeRateChanged, &minSelfDelegation,
				)
			},
			"",
		},
		{
			"Pass: 1 hour left to the next epoch",
			func(ctx sdk.Context) {
				// 23 hours and 49 minutes left to the next epoch
				nextEpochTime := ctx.BlockTime().Add(time.Hour)
				epoch.StartTime = nextEpochTime.Add(-epoch.Duration)
				suite.app.LiquidStakingKeeper.SetEpoch(suite.ctx, epoch)
			},
			func() *stakingtypes.MsgEditValidator {
				minSelfDelegation := validators[0].GetMinSelfDelegation()
				feeRateChanged := validators[0].GetCommission().Add(sdk.NewDecWithPrec(1, 2))
				return stakingtypes.NewMsgEditValidator(
					validators[0].GetOperator(), stakingtypes.Description{}, &feeRateChanged, &minSelfDelegation,
				)
			},
			"",
		},
		{
			"Fail: 23 hours and 51 minutes left to the next epoch (1 minute over)",
			func(ctx sdk.Context) {
				// 23 hours and 51 minutes left to the next epoch
				nextEpochTime := ctx.BlockTime().Add(23*time.Hour + 51*time.Minute)
				epoch.StartTime = nextEpochTime.Add(-epoch.Duration)
				suite.app.LiquidStakingKeeper.SetEpoch(suite.ctx, epoch)
			},
			func() *stakingtypes.MsgEditValidator {
				minSelfDelegation := validators[0].GetMinSelfDelegation()
				feeRateChanged := validators[0].GetCommission().Add(sdk.NewDecWithPrec(1, 2))
				return stakingtypes.NewMsgEditValidator(
					validators[0].GetOperator(), stakingtypes.Description{}, &feeRateChanged, &minSelfDelegation,
				)
			},
			"1m0s left",
		},
		{
			"Fail: 3 days left to the next epoch",
			func(ctx sdk.Context) {
				// 23 hours and 51 minutes left to the next epoch
				nextEpochTime := ctx.BlockTime().Add(3 * 24 * time.Hour)
				epoch.StartTime = nextEpochTime.Add(-epoch.Duration)
				suite.app.LiquidStakingKeeper.SetEpoch(suite.ctx, epoch)
			},
			func() *stakingtypes.MsgEditValidator {
				minSelfDelegation := validators[0].GetMinSelfDelegation()
				feeRateChanged := validators[0].GetCommission().Add(sdk.NewDecWithPrec(1, 2))
				return stakingtypes.NewMsgEditValidator(
					validators[0].GetOperator(), stakingtypes.Description{}, &feeRateChanged, &minSelfDelegation,
				)
			},
			"48h10m0s left",
		},
	}

	vccld := ante.NewValCommissionChangeLimitDecorator(&suite.app.LiquidStakingKeeper, &suite.app.StakingKeeper, suite.app.AppCodec())
	anteHandler := sdk.ChainAnteDecorators(vccld)
	for _, tc := range tests {
		suite.Run(tc.desc, func() {
			if tc.setEpoch != nil {
				tc.setEpoch(suite.ctx)
			}
			tx, err := createTx(suite.priv, []sdk.Msg{tc.createMsg()}...)
			suite.Require().NoError(err)
			_, err = anteHandler(suite.ctx, tx, false)
			if tc.expectedError != "" {
				suite.ErrorIs(err, types.ErrChangingValCommissionForbidden)
				suite.ErrorContains(err, tc.expectedError)
			} else {
				suite.Require().NoError(err)
			}
		})
	}
}

// authz, nested, multi msg cases
func (suite *AnteTestSuite) TestValCommissionChangeAuthzCases() {
	suite.SetupTest(false)

	validators := suite.app.StakingKeeper.GetAllValidators(suite.ctx)
	minSelfDelegation := validators[0].GetMinSelfDelegation()
	feeRate := validators[0].GetCommission()
	notLimitedMsgNoFeeRateChange := stakingtypes.NewMsgEditValidator(
		validators[0].GetOperator(), stakingtypes.Description{}, &feeRate, &minSelfDelegation,
	)
	feeRateChanged := feeRate.Add(sdk.NewDecWithPrec(1, 2))
	limitedMsgFeeRateChange := stakingtypes.NewMsgEditValidator(
		validators[0].GetOperator(), stakingtypes.Description{}, &feeRateChanged, &minSelfDelegation,
	)

	coins := sdk.NewCoins(sdk.NewCoin(suite.app.StakingKeeper.BondDenom(suite.ctx), sdk.NewInt(10000)))
	normMsg := &banktypes.MsgSend{
		FromAddress: suite.addr.String(),
		ToAddress:   suite.addr.String(),
		Amount:      coins,
	}

	wrapAuthzMsg := func(msg sdk.Msg) *authz.MsgExec {
		v := authz.NewMsgExec(suite.addr, []sdk.Msg{msg})
		return &v
	}
	authzMsgNoFeeRateChange := authz.NewMsgExec(suite.addr, []sdk.Msg{notLimitedMsgNoFeeRateChange})
	authzMsgFeeRateChange := authz.NewMsgExec(suite.addr, []sdk.Msg{limitedMsgFeeRateChange})
	authzMultiMsgNoFeeRateChange := authz.NewMsgExec(suite.addr, []sdk.Msg{normMsg, notLimitedMsgNoFeeRateChange})
	authzMultiMsgFeeRateChange := authz.NewMsgExec(suite.addr, []sdk.Msg{normMsg, limitedMsgFeeRateChange})

	// set epoch time to 2 days before the next epoch which means
	// fee rate change msg cannot pass time limit validation
	epoch := suite.app.LiquidStakingKeeper.GetEpoch(suite.ctx)
	nextEpochTime := suite.ctx.BlockTime().Add(2 * 24 * time.Hour)
	epoch.StartTime = nextEpochTime.Add(-epoch.Duration)
	suite.app.LiquidStakingKeeper.SetEpoch(suite.ctx, epoch)

	tests := []struct {
		desc          string
		msgs          []sdk.Msg
		expectedError error
	}{
		{
			"normal msg",
			[]sdk.Msg{normMsg},
			nil,
		},
		{
			"pass: no fee rate change msg",
			[]sdk.Msg{notLimitedMsgNoFeeRateChange},
			nil,
		},
		{
			"fail: limited msg",
			[]sdk.Msg{limitedMsgFeeRateChange},
			types.ErrChangingValCommissionForbidden,
		},
		{
			"pass: multi msg",
			[]sdk.Msg{normMsg, notLimitedMsgNoFeeRateChange},
			nil,
		},
		{
			"fail: multi msg including limited msg",
			[]sdk.Msg{normMsg, limitedMsgFeeRateChange},
			types.ErrChangingValCommissionForbidden,
		},
		{
			"pass: multi msg 2",
			[]sdk.Msg{notLimitedMsgNoFeeRateChange, normMsg},
			nil,
		},
		{
			"fail: multi msg including limited msg 2",
			[]sdk.Msg{limitedMsgFeeRateChange, normMsg},
			types.ErrChangingValCommissionForbidden,
		},
		{
			"pass: authz msg",
			[]sdk.Msg{&authzMsgNoFeeRateChange},
			nil,
		},
		{
			"fail: authz msg including limited msg",
			[]sdk.Msg{&authzMsgFeeRateChange},
			types.ErrChangingValCommissionForbidden,
		},
		{
			"pass: authz msg 2",
			[]sdk.Msg{normMsg, &authzMsgNoFeeRateChange},
			nil,
		},
		{
			"fail: authz msg including limited msg 2",
			[]sdk.Msg{normMsg, &authzMsgFeeRateChange},
			types.ErrChangingValCommissionForbidden,
		},
		{
			"pass: authz msg 3",
			[]sdk.Msg{&authzMsgNoFeeRateChange, normMsg},
			nil,
		},
		{
			"fail: authz msg including limited msg 3",
			[]sdk.Msg{&authzMsgFeeRateChange, normMsg},
			types.ErrChangingValCommissionForbidden,
		},
		{
			"pass: authz msg 4",
			[]sdk.Msg{&authzMultiMsgNoFeeRateChange},
			nil,
		},
		{
			"fail: authz msg including limited msg 4",
			[]sdk.Msg{&authzMultiMsgFeeRateChange},
			types.ErrChangingValCommissionForbidden,
		},
		{
			"pass: authz msg 5",
			[]sdk.Msg{normMsg, &authzMultiMsgNoFeeRateChange},
			nil,
		},
		{
			"fail: authz msg including limited msg 5",
			[]sdk.Msg{normMsg, &authzMultiMsgFeeRateChange},
			types.ErrChangingValCommissionForbidden,
		},
		{
			"fail: nested authz msg-1",
			[]sdk.Msg{
				wrapAuthzMsg(
					wrapAuthzMsg(
						&authzMsgFeeRateChange),
				),
			},
			types.ErrChangingValCommissionForbidden,
		},
		{
			"fail: nested authz msg-2",
			[]sdk.Msg{
				wrapAuthzMsg(
					wrapAuthzMsg(
						wrapAuthzMsg(
							wrapAuthzMsg(
								&authzMsgFeeRateChange),
						),
					),
				),
			},
			types.ErrChangingValCommissionForbidden,
		},
		{
			"pass: nested authz msg",
			[]sdk.Msg{
				wrapAuthzMsg(
					wrapAuthzMsg(
						wrapAuthzMsg(
							wrapAuthzMsg(
								normMsg),
						),
					),
				),
			},
			nil,
		},
		{
			"pass: nested authz msg 2",
			[]sdk.Msg{
				wrapAuthzMsg(
					wrapAuthzMsg(
						wrapAuthzMsg(
							wrapAuthzMsg(
								&authzMsgNoFeeRateChange),
						),
					),
				),
			},
			nil,
		},
		{
			"fail: nested authz msg exceeding limit",
			[]sdk.Msg{
				wrapAuthzMsg(
					wrapAuthzMsg(
						wrapAuthzMsg(
							wrapAuthzMsg(
								wrapAuthzMsg(
									wrapAuthzMsg(
										&authzMsgFeeRateChange),
								),
							),
						),
					),
				),
			},
			fmt.Errorf("found more nested msgs than permited. Limit is : %d", authzante.MaxNestedMsgs),
		},
		{
			"fail: nested authz msg exceeding limit 2",
			[]sdk.Msg{
				wrapAuthzMsg(
					wrapAuthzMsg(
						wrapAuthzMsg(
							wrapAuthzMsg(
								wrapAuthzMsg(
									wrapAuthzMsg(
										normMsg),
								),
							),
						),
					),
				),
			},
			fmt.Errorf("found more nested msgs than permited. Limit is : %d", authzante.MaxNestedMsgs),
		},
	}

	vcd := ante.NewValCommissionChangeLimitDecorator(&suite.app.LiquidStakingKeeper, &suite.app.StakingKeeper, suite.app.AppCodec())
	anteHandler := sdk.ChainAnteDecorators(vcd)
	for _, tc := range tests {
		suite.Run(tc.desc, func() {
			tx, err := createTx(suite.priv, tc.msgs...)
			suite.Require().NoError(err)
			_, err = anteHandler(suite.ctx, tx, false)
			if tc.expectedError != nil {
				suite.ErrorContains(err, tc.expectedError.Error())
			} else {
				suite.Require().NoError(err)
			}
		})
	}
}
