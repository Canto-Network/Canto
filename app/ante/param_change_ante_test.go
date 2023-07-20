package ante_test

import (
	"strconv"

	"github.com/Canto-Network/Canto/v6/app/ante"
	"github.com/Canto-Network/Canto/v6/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/cosmos/cosmos-sdk/x/params/types/proposal"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

// TODO: Advanced test cases (e.g. nested param change proposals)
// Authz and multi msg cases
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
				unbondingTime := proposal.NewParamChange("staking", "BondDenom", "adoge")
				return proposal.NewParameterChangeProposal("tc11", "tc11", []proposal.ParamChange{unbondingTime})
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
