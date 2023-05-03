package v6_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	"github.com/evmos/ethermint/crypto/ethsecp256k1"
	feemarkettypes "github.com/evmos/ethermint/x/feemarket/types"

	abci "github.com/tendermint/tendermint/abci/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	chain "github.com/Canto-Network/Canto/v6/app"
	v6 "github.com/Canto-Network/Canto/v6/app/upgrades/v6"
	coinswaptypes "github.com/Canto-Network/Canto/v6/x/coinswap/types"
	onboardingtypes "github.com/Canto-Network/Canto/v6/x/onboarding/types"
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

func (s *UpgradeTestSuite) TestUpgradeV6() {
	testCases := []struct {
		title   string
		before  func()
		after   func()
		expPass bool
	}{
		{
			"v6 upgrade onboarding & coinswap",
			func() {},
			func() {
				coinswapParams := s.app.CoinswapKeeper.GetParams(s.ctx)
				s.Require().EqualValues(
					coinswapParams.PoolCreationFee, coinswaptypes.DefaultPoolCreationFee)
				s.Require().EqualValues(
					coinswapParams.MaxSwapAmount, coinswaptypes.DefaultMaxSwapAmount)
				s.Require().EqualValues(
					coinswapParams.MaxStandardCoinPerPool, coinswaptypes.DefaultMaxStandardCoinPerPool)
				s.Require().EqualValues(
					coinswapParams.Fee, coinswaptypes.DefaultFee)
				s.Require().EqualValues(
					coinswapParams.TaxRate, coinswaptypes.DefaultTaxRate)

				onboardingParams := s.app.OnboardingKeeper.GetParams(s.ctx)
				s.Require().EqualValues(
					onboardingParams.AutoSwapThreshold, onboardingtypes.DefaultAutoSwapThreshold)
				s.Require().EqualValues(
					onboardingParams.WhitelistedChannels, onboardingtypes.DefaultWhitelistedChannels)
			},
			true,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.title, func() {
			s.SetupTest()

			tc.before()

			s.ctx = s.ctx.WithBlockHeight(testUpgradeHeight - 1)
			plan := upgradetypes.Plan{Name: v6.UpgradeName, Height: testUpgradeHeight}
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
