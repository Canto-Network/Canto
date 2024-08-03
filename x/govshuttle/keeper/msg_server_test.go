package keeper_test

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"math/big"

	"github.com/ethereum/go-ethereum/common"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	govtypesv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/evmos/ethermint/crypto/ethsecp256k1"
	"github.com/evmos/ethermint/server/config"
	evmtypes "github.com/evmos/ethermint/x/evm/types"

	"github.com/Canto-Network/Canto/v8/contracts"
	"github.com/Canto-Network/Canto/v8/testutil"
	govshuttletypes "github.com/Canto-Network/Canto/v8/x/govshuttle/types"
)

type ProposalResult struct {
	Id         *big.Int         `json:"id"`
	Title      string           `json:"title"`
	Desc       string           `json:"desc"`
	Targets    []common.Address `json:"targets"`
	Values     []*big.Int       `json:"values"`
	Signatures []string         `json:"signatures"`
	Calldatas  [][]byte         `json:"calldatas"`
}

func (suite *KeeperTestSuite) TestMsgExecutionByProposal() {
	suite.SetupTest()

	// get denom
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

	// deligate to validator
	initAmount := sdkmath.NewInt(int64(math.Pow10(18)) * 2)
	initBalance := sdk.NewCoins(sdk.NewCoin(denom, initAmount))
	testutil.FundAccount(suite.app.BankKeeper, suite.ctx, proposer, initBalance)
	shares, err := suite.app.StakingKeeper.Delegate(suite.ctx, proposer, sdk.DefaultPowerReduction, stakingtypes.Unbonded, suite.validator, true)
	suite.Require().NoError(err)
	suite.Require().True(shares.GT(sdkmath.LegacyNewDec(0)))

	testCases := []struct {
		name                    string
		msg                     sdk.Msg
		checkFunc               func(uint64, sdk.Msg)
		expectSubmitProposalErr bool
		expectProposalFailed    bool
	}{
		{
			"fail - MsgLendingMarketProposal - authority check",
			&govshuttletypes.MsgLendingMarketProposal{
				Authority:   "canto1yrmjye0zyfvr0lthc6fwq7qlwg9e8muftxa630",
				Title:       "lending market proposal test",
				Description: "lending market proposal test description",
				Metadata: &govshuttletypes.LendingMarketMetadata{
					Account:    []string{"0x20F72265e2225837fd77C692e0781f720B93eF89", "0xf6Db2570A2417188a5788D6d5Fd9faAa5B1fE555"},
					PropId:     0,
					Values:     []uint64{1234, 5678},
					Calldatas:  []string{hex.EncodeToString([]byte("calldata1")), hex.EncodeToString([]byte("calldata2"))},
					Signatures: []string{"sig1", "sig2"},
				},
			},
			func(uint64, sdk.Msg) {},
			true,
			false,
		},
		{
			"fail - MsgLendingMarketProposal - validate basic logic",
			&govshuttletypes.MsgLendingMarketProposal{
				Authority:   authtypes.NewModuleAddress(govtypes.ModuleName).String(),
				Title:       "lending market proposal test",
				Description: "lending market proposal test description",
				Metadata: &govshuttletypes.LendingMarketMetadata{
					Account:    []string{"0x20F72265e2225837fd77C692e0781f720B93eF89", "0xf6Db2570A2417188a5788D6d5Fd9faAa5B1fE555"},
					PropId:     0,
					Values:     []uint64{1234},
					Calldatas:  []string{hex.EncodeToString([]byte("calldata1")), hex.EncodeToString([]byte("calldata2"))},
					Signatures: []string{"sig1", "sig2"},
				},
			},
			func(uint64, sdk.Msg) {},
			false,
			true,
		},
		{
			"ok - MsgLendingMarketProposal",
			&govshuttletypes.MsgLendingMarketProposal{
				Authority:   authtypes.NewModuleAddress(govtypes.ModuleName).String(),
				Title:       "lending market proposal test",
				Description: "lending market proposal test description",
				Metadata: &govshuttletypes.LendingMarketMetadata{
					Account:    []string{"0x20F72265e2225837fd77C692e0781f720B93eF89", "0xf6Db2570A2417188a5788D6d5Fd9faAa5B1fE555"},
					PropId:     0,
					Values:     []uint64{1234, 5678},
					Calldatas:  []string{hex.EncodeToString([]byte("calldata1")), hex.EncodeToString([]byte("calldata2"))},
					Signatures: []string{"sig1", "sig2"},
				},
			},
			func(propId uint64, msg sdk.Msg) {
				proposalMsg, ok := msg.(*govshuttletypes.MsgLendingMarketProposal)
				suite.Require().True(ok)

				targets := []common.Address{}
				for _, acc := range proposalMsg.Metadata.Account {
					targets = append(targets, common.HexToAddress(acc))
				}

				values := []*big.Int{}
				for _, value := range proposalMsg.Metadata.Values {
					values = append(values, big.NewInt(int64(value)))
				}

				calldatas := [][]byte{}
				for _, calldata := range proposalMsg.Metadata.Calldatas {
					c, err := hex.DecodeString(calldata)
					suite.Require().NoError(err)

					calldatas = append(calldatas, c)
				}

				suite.checkQueryPropResult(
					propId,
					ProposalResult{
						Id:         big.NewInt(int64(propId)),
						Title:      proposalMsg.Title,
						Desc:       proposalMsg.Description,
						Targets:    targets,
						Values:     values,
						Signatures: proposalMsg.Metadata.Signatures,
						Calldatas:  calldatas,
					},
				)
			},
			false,
			false,
		},
		{
			"fail - MsgTreasuryProposal - authority check",
			&govshuttletypes.MsgTreasuryProposal{
				Authority:   "canto1yrmjye0zyfvr0lthc6fwq7qlwg9e8muftxa630",
				Title:       "treasury proposal test",
				Description: "treasury proposal test description",
				Metadata: &govshuttletypes.TreasuryProposalMetadata{
					PropID:    0,
					Recipient: "0x20F72265e2225837fd77C692e0781f720B93eF89",
					Amount:    1234,
					Denom:     "acanto",
				},
			},
			func(uint64, sdk.Msg) {},
			true,
			false,
		},
		{
			"fail - MsgTreasuryProposal - validate basic logic",
			&govshuttletypes.MsgTreasuryProposal{
				Authority:   authtypes.NewModuleAddress(govtypes.ModuleName).String(),
				Title:       "treasury proposal test",
				Description: "treasury proposal test description",
				Metadata: &govshuttletypes.TreasuryProposalMetadata{
					PropID:    0,
					Recipient: "0x20F72265e2225837fd77C692e0781f720B93eF89",
					Amount:    1234,
					Denom:     "canto2",
				},
			},
			func(uint64, sdk.Msg) {},
			false,
			true,
		},
		{
			"ok - MsgTreasuryProposal",
			&govshuttletypes.MsgTreasuryProposal{
				Authority:   authtypes.NewModuleAddress(govtypes.ModuleName).String(),
				Title:       "treasury proposal test",
				Description: "treasury proposal test description",
				Metadata: &govshuttletypes.TreasuryProposalMetadata{
					PropID:    0,
					Recipient: "0x20F72265e2225837fd77C692e0781f720B93eF89",
					Amount:    1234,
					Denom:     "canto",
				},
			},
			func(propId uint64, msg sdk.Msg) {
				proposalMsg, ok := msg.(*govshuttletypes.MsgTreasuryProposal)
				suite.Require().True(ok)

				targets := []common.Address{common.HexToAddress(proposalMsg.Metadata.Recipient)}
				values := []*big.Int{big.NewInt(int64(proposalMsg.Metadata.Amount))}
				signatures := []string{proposalMsg.Metadata.Denom}
				calldatas := [][]byte{}

				suite.checkQueryPropResult(
					propId,
					ProposalResult{
						Id:         big.NewInt(int64(propId)),
						Title:      proposalMsg.Title,
						Desc:       proposalMsg.Description,
						Targets:    targets,
						Values:     values,
						Signatures: signatures,
						Calldatas:  calldatas,
					},
				)
			},
			false,
			false,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			// submit proposal
			proposal, err := suite.app.GovKeeper.SubmitProposal(suite.ctx, []sdk.Msg{tc.msg}, "", "test", "description", proposer, false)
			if tc.expectSubmitProposalErr {
				suite.Require().Error(err)
			} else {
				suite.Require().NoError(err)
				suite.Commit()

				ok, err := suite.app.GovKeeper.AddDeposit(suite.ctx, proposal.Id, proposer, govParams.MinDeposit)
				suite.Require().NoError(err)
				suite.Require().True(ok)
				suite.Commit()

				err = suite.app.GovKeeper.AddVote(suite.ctx, proposal.Id, proposer, govtypesv1.NewNonSplitVoteOption(govtypesv1.OptionYes), "")
				suite.Require().NoError(err)
				suite.CommitAfter(*govParams.VotingPeriod)

				proposal, err := suite.app.GovKeeper.Proposals.Get(suite.ctx, proposal.Id)
				suite.Require().NoError(err)
				if tc.expectProposalFailed {
					suite.Require().Equal(govtypesv1.ProposalStatus_PROPOSAL_STATUS_FAILED, proposal.Status)
				} else {
					suite.Require().Equal(govtypesv1.ProposalStatus_PROPOSAL_STATUS_PASSED, proposal.Status)

					// check proposal result
					propId, err := suite.app.GovKeeper.ProposalID.Peek(suite.ctx)
					suite.Require().NoError(err)
					tc.checkFunc(propId, tc.msg)
				}
			}
		})
	}
}

func (suite *KeeperTestSuite) checkQueryPropResult(propId uint64, expectedResult ProposalResult) {
	// make calldata
	data, err := contracts.ProposalStoreContract.ABI.Pack("QueryProp", big.NewInt(int64(propId)))
	suite.Require().NoError(err)

	// get port contract address
	portAddr, ok := suite.app.GovshuttleKeeper.GetPort(suite.ctx)
	suite.Require().True(ok)

	txArgs := map[string]interface{}{
		"to":   portAddr,
		"data": fmt.Sprintf("0x%x", data),
	}
	txArgsJson, err := json.Marshal(txArgs)
	suite.Require().NoError(err)

	// query to contract
	req := &evmtypes.EthCallRequest{
		Args:   txArgsJson,
		GasCap: config.DefaultGasCap,
	}
	rpcRes, err := suite.app.EvmKeeper.EthCall(suite.ctx, req)
	suite.Require().NoError(err)

	queryRes, err := contracts.ProposalStoreContract.ABI.Unpack("QueryProp", rpcRes.Ret)
	suite.Require().NoError(err)

	// marshal and unmarshal to get ProposalResult
	var res ProposalResult
	b, err := json.Marshal(queryRes[0])
	suite.Require().NoError(err)
	json.Unmarshal(b, &res)

	suite.Require().Equal(
		expectedResult,
		res,
	)
}
