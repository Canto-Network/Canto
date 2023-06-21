package ante_test

import (
	"strconv"
	"testing"

	"github.com/Canto-Network/Canto/v6/app"
	"github.com/Canto-Network/Canto/v6/app/ante"
	"github.com/Canto-Network/Canto/v6/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/cosmos/cosmos-sdk/x/params/types/proposal"
	feemarkettypes "github.com/evmos/ethermint/x/feemarket/types"
	"github.com/stretchr/testify/suite"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
)

var proposer = sdk.AccAddress("test1")

type SlashingParamChangeAnteTestSuite struct {
	suite.Suite

	app *app.Canto
	ctx sdk.Context
}

func (s *SlashingParamChangeAnteTestSuite) SetupTest() {
	s.app = app.Setup(false, feemarkettypes.DefaultGenesisState())
	s.ctx = s.app.BaseApp.NewContext(false, tmproto.Header{
		Height:  1,
		ChainID: "canto_9001-1",
	})
}

func TestSlashingParamChangeAnteSuite(t *testing.T) {
	suite.Run(t, new(SlashingParamChangeAnteTestSuite))
}

func (s *SlashingParamChangeAnteTestSuite) TestSlashingParamChangeProposal() {
	s.SetupTest()
	params := s.app.SlashingKeeper.GetParams(s.ctx)
	tests := []struct {
		desc                 string
		createSubmitProposal func() *proposal.ParameterChangeProposal
		expectedError        string
	}{
		{
			"SignedBlocksWindow cannot be decreased",
			func() *proposal.ParameterChangeProposal {
				smaller := strconv.FormatInt(params.GetSignedBlocksWindow()-1, 10)
				signedBlocksWindow := proposal.NewParamChange("slashing", "SignedBlocksWindow", smaller)
				return proposal.NewParameterChangeProposal("", "", []proposal.ParamChange{signedBlocksWindow})
			},
			types.ErrInvalidSignedBlocksWindow.Error(),
		},
		{
			"SignedBlocksWindow can be increased",
			func() *proposal.ParameterChangeProposal {
				smaller := strconv.FormatInt(params.GetSignedBlocksWindow()+1, 10)
				signedBlocksWindow := proposal.NewParamChange("slashing", "SignedBlocksWindow", smaller)
				return proposal.NewParameterChangeProposal("", "", []proposal.ParamChange{signedBlocksWindow})
			},
			"",
		},
		{
			"MinSignedPerWindow cannot be decreased",
			func() *proposal.ParameterChangeProposal {
				smaller := params.MinSignedPerWindow.Sub(sdk.OneDec()).String()
				minSignedPerWindow := proposal.NewParamChange("slashing", "MinSignedPerWindow", smaller)
				return proposal.NewParameterChangeProposal("", "", []proposal.ParamChange{minSignedPerWindow})
			},
			types.ErrInvalidMinSignedPerWindow.Error(),
		},
		{
			"MinSignedPerWindow can be increased",
			func() *proposal.ParameterChangeProposal {
				smaller := params.MinSignedPerWindow.Add(sdk.OneDec()).String()
				minSignedPerWindow := proposal.NewParamChange("slashing", "MinSignedPerWindow", smaller)
				return proposal.NewParameterChangeProposal("", "", []proposal.ParamChange{minSignedPerWindow})
			},
			"",
		},
		{
			"DowntimeJailDuration cannot be decreased",
			func() *proposal.ParameterChangeProposal {
				smaller := strconv.FormatInt(int64(params.DowntimeJailDuration)-1, 10)
				downtimeJailDuration := proposal.NewParamChange("slashing", "DowntimeJailDuration", smaller)
				return proposal.NewParameterChangeProposal("", "", []proposal.ParamChange{downtimeJailDuration})
			},
			types.ErrInvalidDowntimeJailDuration.Error(),
		},
		{
			"DowntimeJailDuration can be increased",
			func() *proposal.ParameterChangeProposal {
				smaller := strconv.FormatInt(int64(params.DowntimeJailDuration)+1, 10)
				downtimeJailDuration := proposal.NewParamChange("slashing", "DowntimeJailDuration", smaller)
				return proposal.NewParameterChangeProposal("", "", []proposal.ParamChange{downtimeJailDuration})
			},
			"",
		},
		{
			"SlashFractionDoubleSign cannot be increased",
			func() *proposal.ParameterChangeProposal {
				smaller := params.SlashFractionDoubleSign.Add(sdk.OneDec()).String()
				slashFractionDoubleSign := proposal.NewParamChange("slashing", "SlashFractionDoubleSign", smaller)
				return proposal.NewParameterChangeProposal("", "", []proposal.ParamChange{slashFractionDoubleSign})
			},
			types.ErrInvalidSlashFractionDoubleSign.Error(),
		},
		{
			"SlashFractionDoubleSign can be decreased",
			func() *proposal.ParameterChangeProposal {
				smaller := params.SlashFractionDoubleSign.Sub(sdk.OneDec()).String()
				slashFractionDoubleSign := proposal.NewParamChange("slashing", "SlashFractionDoubleSign", smaller)
				return proposal.NewParameterChangeProposal("", "", []proposal.ParamChange{slashFractionDoubleSign})
			},
			"",
		},
		{
			"SlashFractionDowntime cannot be increased",
			func() *proposal.ParameterChangeProposal {
				smaller := params.SlashFractionDowntime.Add(sdk.OneDec()).String()
				slashFractionDowntime := proposal.NewParamChange("slashing", "SlashFractionDowntime", smaller)
				return proposal.NewParameterChangeProposal("", "", []proposal.ParamChange{slashFractionDowntime})
			},
			types.ErrInvalidSlashFractionDowntime.Error(),
		},
	}

	decorator := ante.NewSlashingParamChangeLimitDecorator(s.app.AppCodec(), &s.app.SlashingKeeper)

	for _, tc := range tests {
		s.Run(tc.desc, func() {
			msg, err := govtypes.NewMsgSubmitProposal(
				tc.createSubmitProposal(),
				sdk.NewCoins(sdk.NewCoin(s.app.StakingKeeper.BondDenom(s.ctx), sdk.NewInt(10000))),
				proposer,
			)
			s.Require().NoError(err)
			err = decorator.ValidateMsgs(s.ctx, []sdk.Msg{msg})
			if tc.expectedError == "" {
				s.Require().NoError(err)
			} else {
				s.Require().ErrorContains(err, tc.expectedError)
			}
		})

	}
}
