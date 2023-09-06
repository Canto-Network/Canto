package ante_test

import (
	"fmt"
	"strconv"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/cosmos/cosmos-sdk/x/params/types/proposal"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/Canto-Network/Canto/v7/app/ante"
	authzante "github.com/Canto-Network/Canto/v7/app/ante/cosmos"
	"github.com/Canto-Network/Canto/v7/types"
)

// single param change msg cases
func (suite *AnteTestSuite) TestParamChangeProposal() {
	suite.SetupTest(false)
	params := suite.app.SlashingKeeper.GetParams(suite.ctx)
	tests := []struct {
		desc                 string
		createSubmitProposal func() *proposal.ParameterChangeProposal
		expectedError        error
	}{
		{
			"SignedBlocksWindow cannot be decreased",
			func() *proposal.ParameterChangeProposal {
				smaller := strconv.FormatInt(params.GetSignedBlocksWindow()-1, 10)
				signedBlocksWindow := proposal.NewParamChange("slashing", "SignedBlocksWindow", smaller)
				return proposal.NewParameterChangeProposal("tc1", "tc1", []proposal.ParamChange{signedBlocksWindow})
			},
			types.ErrInvalidSignedBlocksWindow,
		},
		{
			"SignedBlocksWindow can be increased",
			func() *proposal.ParameterChangeProposal {
				smaller := strconv.FormatInt(params.GetSignedBlocksWindow()+1, 10)
				signedBlocksWindow := proposal.NewParamChange("slashing", "SignedBlocksWindow", smaller)
				return proposal.NewParameterChangeProposal("tc2", "tc2", []proposal.ParamChange{signedBlocksWindow})
			},
			nil,
		},
		{
			"MinSignedPerWindow cannot be decreased",
			func() *proposal.ParameterChangeProposal {
				smaller := params.MinSignedPerWindow.Sub(sdk.OneDec()).String()
				minSignedPerWindow := proposal.NewParamChange("slashing", "MinSignedPerWindow", smaller)
				return proposal.NewParameterChangeProposal("tc3", "tc3", []proposal.ParamChange{minSignedPerWindow})
			},
			types.ErrInvalidMinSignedPerWindow,
		},
		{
			"MinSignedPerWindow can be increased",
			func() *proposal.ParameterChangeProposal {
				smaller := params.MinSignedPerWindow.Add(sdk.OneDec()).String()
				minSignedPerWindow := proposal.NewParamChange("slashing", "MinSignedPerWindow", smaller)
				return proposal.NewParameterChangeProposal("tc4", "tc4", []proposal.ParamChange{minSignedPerWindow})
			},
			nil,
		},
		{
			"DowntimeJailDuration cannot be decreased",
			func() *proposal.ParameterChangeProposal {
				smaller := strconv.FormatInt(int64(params.DowntimeJailDuration)-1, 10)
				downtimeJailDuration := proposal.NewParamChange("slashing", "DowntimeJailDuration", smaller)
				return proposal.NewParameterChangeProposal("tc5", "tc5", []proposal.ParamChange{downtimeJailDuration})
			},
			types.ErrInvalidDowntimeJailDuration,
		},
		{
			"DowntimeJailDuration can be increased",
			func() *proposal.ParameterChangeProposal {
				smaller := strconv.FormatInt(int64(params.DowntimeJailDuration)+1, 10)
				downtimeJailDuration := proposal.NewParamChange("slashing", "DowntimeJailDuration", smaller)
				return proposal.NewParameterChangeProposal("tc6", "tc6", []proposal.ParamChange{downtimeJailDuration})
			},
			nil,
		},
		{
			"SlashFractionDoubleSign cannot be increased",
			func() *proposal.ParameterChangeProposal {
				smaller := params.SlashFractionDoubleSign.Add(sdk.OneDec()).String()
				slashFractionDoubleSign := proposal.NewParamChange("slashing", "SlashFractionDoubleSign", smaller)
				return proposal.NewParameterChangeProposal("tc7", "tc7", []proposal.ParamChange{slashFractionDoubleSign})
			},
			types.ErrInvalidSlashFractionDoubleSign,
		},
		{
			"SlashFractionDoubleSign can be decreased",
			func() *proposal.ParameterChangeProposal {
				smaller := params.SlashFractionDoubleSign.Sub(sdk.OneDec()).String()
				slashFractionDoubleSign := proposal.NewParamChange("slashing", "SlashFractionDoubleSign", smaller)
				return proposal.NewParameterChangeProposal("tc8", "tc8", []proposal.ParamChange{slashFractionDoubleSign})
			},
			nil,
		},
		{
			"SlashFractionDowntime cannot be increased",
			func() *proposal.ParameterChangeProposal {
				smaller := params.SlashFractionDowntime.Add(sdk.OneDec()).String()
				slashFractionDowntime := proposal.NewParamChange("slashing", "SlashFractionDowntime", smaller)
				return proposal.NewParameterChangeProposal("tc9", "tc9", []proposal.ParamChange{slashFractionDowntime})
			},
			types.ErrInvalidSlashFractionDowntime,
		},
		{
			"Changing Unbonding Time is not allowed",
			func() *proposal.ParameterChangeProposal {
				smaller := strconv.FormatInt(int64(stakingtypes.DefaultUnbondingTime)-1, 10)
				unbondingTime := proposal.NewParamChange("staking", "UnbondingTime", smaller)
				return proposal.NewParameterChangeProposal("tc10", "tc10", []proposal.ParamChange{unbondingTime})
			},
			types.ErrChangingUnbondingPeriodForbidden,
		},
		{
			"Changing BondDenom is not allowed",
			func() *proposal.ParameterChangeProposal {
				bondDenomChange := proposal.NewParamChange("staking", "BondDenom", "adoge")
				return proposal.NewParameterChangeProposal("tc11", "tc11", []proposal.ParamChange{bondDenomChange})
			},
			types.ErrChangingBondDenomForbidden,
		},
	}

	spcld := ante.NewParamChangeLimitDecorator(&suite.app.SlashingKeeper, suite.app.AppCodec())
	anteHandler := sdk.ChainAnteDecorators(spcld)
	for _, tc := range tests {
		suite.Run(tc.desc, func() {
			msg, err := govtypes.NewMsgSubmitProposal(
				tc.createSubmitProposal(),
				sdk.NewCoins(sdk.NewCoin(suite.app.StakingKeeper.BondDenom(suite.ctx), sdk.NewInt(10000))),
				suite.addr,
			)
			tx, err := createTx(suite.priv, []sdk.Msg{msg}...)
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

// authz, nested, multi msg cases
func (suite *AnteTestSuite) TestAuthzCases() {
	suite.SetupTest(false)

	bondDenomChange := proposal.NewParamChange("staking", "BondDenom", "adoge")
	coins := sdk.NewCoins(sdk.NewCoin(suite.app.StakingKeeper.BondDenom(suite.ctx), sdk.NewInt(10000)))
	normMsg := &banktypes.MsgSend{
		FromAddress: suite.addr.String(),
		ToAddress:   suite.addr.String(),
		Amount:      coins,
	}
	limitedMsg, err := govtypes.NewMsgSubmitProposal(
		proposal.NewParameterChangeProposal("tc11", "tc11", []proposal.ParamChange{bondDenomChange}),
		coins,
		suite.addr,
	)
	suite.NoError(err)

	wrapAuthzMsg := func(msg sdk.Msg) *authz.MsgExec {
		v := authz.NewMsgExec(suite.addr, []sdk.Msg{msg})
		return &v
	}
	authzMsg := authz.NewMsgExec(suite.addr, []sdk.Msg{limitedMsg})
	authzMultiMsg := authz.NewMsgExec(suite.addr, []sdk.Msg{normMsg, limitedMsg})

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
			"limited msg",
			[]sdk.Msg{limitedMsg},
			types.ErrChangingBondDenomForbidden,
		},
		{
			"multi msg-1",
			[]sdk.Msg{normMsg, limitedMsg},
			types.ErrChangingBondDenomForbidden,
		},
		{
			"multi msg-2",
			[]sdk.Msg{limitedMsg, normMsg},
			types.ErrChangingBondDenomForbidden,
		},
		{
			"authz msg-1",
			[]sdk.Msg{&authzMsg},
			types.ErrChangingBondDenomForbidden,
		},
		{
			"authz msg-2",
			[]sdk.Msg{normMsg, &authzMsg},
			types.ErrChangingBondDenomForbidden,
		},
		{
			"authz msg-3",
			[]sdk.Msg{&authzMsg, normMsg},
			types.ErrChangingBondDenomForbidden,
		},
		{
			"authz msg-4",
			[]sdk.Msg{&authzMultiMsg},
			types.ErrChangingBondDenomForbidden,
		},
		{
			"authz msg-5",
			[]sdk.Msg{normMsg, &authzMultiMsg},
			types.ErrChangingBondDenomForbidden,
		},
		{
			"nested authz msg-1",
			[]sdk.Msg{
				wrapAuthzMsg(
					wrapAuthzMsg(
						&authzMsg),
				),
			},
			types.ErrChangingBondDenomForbidden,
		},
		{
			"nested authz msg-2",
			[]sdk.Msg{
				wrapAuthzMsg(
					wrapAuthzMsg(
						wrapAuthzMsg(
							wrapAuthzMsg(
								&authzMsg),
						),
					),
				),
			},
			types.ErrChangingBondDenomForbidden,
		},
		{
			"nested authz msg-3",
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
			"nested authz msg-3",
			[]sdk.Msg{
				wrapAuthzMsg(
					wrapAuthzMsg(
						wrapAuthzMsg(
							wrapAuthzMsg(
								wrapAuthzMsg(
									wrapAuthzMsg(
										&authzMsg),
								),
							),
						),
					),
				),
			},
			fmt.Errorf("found more nested msgs than permited. Limit is : %d", authzante.MaxNestedMsgs),
		},
		{
			"nested authz msg-4",
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

	spcld := ante.NewParamChangeLimitDecorator(&suite.app.SlashingKeeper, suite.app.AppCodec())
	anteHandler := sdk.ChainAnteDecorators(spcld)
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
