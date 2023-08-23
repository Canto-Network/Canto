package v8_test

import (
	"testing"
	"time"

	chain "github.com/Canto-Network/Canto/v7/app"
	v8 "github.com/Canto-Network/Canto/v7/app/upgrades/v8"
	"github.com/Canto-Network/Canto/v7/x/liquidstaking/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	"github.com/evmos/ethermint/crypto/ethsecp256k1"
	feemarkettypes "github.com/evmos/ethermint/x/feemarket/types"
	"github.com/stretchr/testify/suite"
	abci "github.com/tendermint/tendermint/abci/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
)

type UpgradeTestSuite struct {
	suite.Suite
	ctx         sdk.Context
	app         *chain.Canto
	consAddress sdk.ConsAddress
}

func (s *UpgradeTestSuite) SetupTest() {

	// consensus key
	priv, err := ethsecp256k1.GenerateKey()
	s.Require().NoError(err)
	s.consAddress = sdk.ConsAddress(priv.PubKey().Address())

	s.app = chain.Setup(false, feemarkettypes.DefaultGenesisState())

	s.ctx = s.app.BaseApp.NewContext(false, tmproto.Header{
		ChainID:         "canto_9001-1",
		Height:          1,
		Time:            time.Date(2023, 5, 9, 8, 0, 0, 0, time.UTC),
		ProposerAddress: s.consAddress.Bytes(),
	})

	// Set Validator
	valAddr := sdk.ValAddress(s.consAddress.Bytes())
	validator, err := stakingtypes.NewValidator(valAddr, priv.PubKey(), stakingtypes.Description{})
	s.NoError(err)
	validator = stakingkeeper.TestingUpdateValidator(s.app.StakingKeeper, s.ctx, validator, true)
	s.app.StakingKeeper.AfterValidatorCreated(s.ctx, validator.GetOperator())
	err = s.app.StakingKeeper.SetValidatorByConsAddr(s.ctx, validator)
	s.NoError(err)

}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(UpgradeTestSuite))
}

const testUpgradeHeight = 10

func (s *UpgradeTestSuite) TestUpgradeV8() {
	testCases := []struct {
		title   string
		before  func()
		after   func()
		expPass bool
	}{
		{
			"v8 upgrade liquidstaking",
			func() {},
			func() {
				params := s.app.LiquidStakingKeeper.GetParams(s.ctx)
				s.Require().EqualValues(
					params.DynamicFeeRate.R0, types.DefaultR0)
				s.Require().EqualValues(
					params.DynamicFeeRate.USoftCap, types.DefaultUSoftCap)
				s.Require().EqualValues(
					params.DynamicFeeRate.UHardCap, types.DefaultUHardCap)
				s.Require().EqualValues(
					params.DynamicFeeRate.UOptimal, types.DefaultUOptimal)
				s.Require().EqualValues(
					params.DynamicFeeRate.Slope1, types.DefaultSlope1)
				s.Require().EqualValues(
					params.DynamicFeeRate.Slope2, types.DefaultSlope2)
				s.Require().EqualValues(
					params.DynamicFeeRate.MaxFeeRate, types.DefaultMaxFee)
				s.Require().EqualValues(
					params.MaximumDiscountRate, types.DefaultMaximumDiscountRate)
			},
			true,
		},
	}
	for _, tc := range testCases {
		s.Run(tc.title, func() {
			s.SetupTest()

			tc.before()

			s.ctx = s.ctx.WithBlockHeight(testUpgradeHeight - 1)
			plan := upgradetypes.Plan{Name: v8.UpgradeName, Height: testUpgradeHeight}
			err := s.app.UpgradeKeeper.ScheduleUpgrade(s.ctx, plan)
			s.Require().NoError(err)
			_, exists := s.app.UpgradeKeeper.GetUpgradePlan(s.ctx)
			s.Require().True(exists)

			s.ctx = s.ctx.WithBlockHeight(testUpgradeHeight)
			s.Require().NotPanics(func() {
				s.ctx.WithProposer(s.consAddress)
				s.app.BeginBlocker(s.ctx, abci.RequestBeginBlock{})
			})

			tc.after()
		})
	}
}
