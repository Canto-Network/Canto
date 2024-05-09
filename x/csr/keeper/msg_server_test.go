package keeper_test

import (
	"math"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	govtypesv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/Canto-Network/Canto/v7/testutil"
	csrtypes "github.com/Canto-Network/Canto/v7/x/csr/types"
)

func (suite *KeeperTestSuite) TestMsgUpdateParamsByProposal() {
	suite.SetupTest()

	// change mindeposit for denom
	govParams, err := suite.app.GovKeeper.Params.Get(suite.ctx)
	suite.Require().NoError(err)
	govParams.MinDeposit = []sdk.Coin{sdk.NewCoin(suite.denom, sdkmath.NewInt(1))}
	err = suite.app.GovKeeper.Params.Set(suite.ctx, govParams)
	suite.Require().NoError(err)

	// create account and deligate to validator
	_, proposer := GenerateKey()
	initAmount := sdkmath.NewInt(int64(math.Pow10(18)) * 2)
	initBalance := sdk.NewCoins(sdk.NewCoin(suite.denom, initAmount))
	testutil.FundAccount(suite.app.BankKeeper, suite.ctx, proposer, initBalance)
	shares, err := suite.app.StakingKeeper.Delegate(suite.ctx, proposer, sdk.DefaultPowerReduction, stakingtypes.Unbonded, suite.validator, true)
	suite.Require().NoError(err)
	suite.Require().True(shares.GT(sdkmath.LegacyNewDec(0)))

	// submit proposal for change params
	changeParams := csrtypes.Params{
		EnableCsr: false,
		CsrShares: sdkmath.LegacyNewDecWithPrec(20, 2),
	}
	msg := &csrtypes.MsgUpdateParams{
		Authority: authtypes.NewModuleAddress(govtypes.ModuleName).String(),
		Params:    changeParams,
	}
	proposal, err := suite.app.GovKeeper.SubmitProposal(suite.ctx, []sdk.Msg{msg}, "", "test", "description", proposer, false)
	suite.Require().NoError(err)
	suite.Commit()

	ok, err := suite.app.GovKeeper.AddDeposit(suite.ctx, proposal.Id, proposer, govParams.MinDeposit)
	suite.Require().NoError(err)
	suite.Require().True(ok)
	suite.Commit()

	err = suite.app.GovKeeper.AddVote(suite.ctx, proposal.Id, proposer, govtypesv1.NewNonSplitVoteOption(govtypesv1.OptionYes), "")
	suite.Require().NoError(err)
	suite.CommitAfter(*govParams.VotingPeriod)

	proposal, err = suite.app.GovKeeper.Proposals.Get(suite.ctx, proposal.Id)
	suite.Require().NoError(err)
	suite.Require().Equal(govtypesv1.ProposalStatus_PROPOSAL_STATUS_PASSED, proposal.Status)
	suite.Require().Equal(suite.app.CSRKeeper.GetParams(suite.ctx), changeParams)
}
