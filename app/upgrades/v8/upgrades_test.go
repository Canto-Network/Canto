package v8_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	upgradetypes "cosmossdk.io/x/upgrade/types"
	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/evmos/ethermint/crypto/ethsecp256k1"
	feemarkettypes "github.com/evmos/ethermint/x/feemarket/types"

	abci "github.com/cometbft/cometbft/abci/types"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"

	chain "github.com/Canto-Network/Canto/v7/app"
	v8 "github.com/Canto-Network/Canto/v7/app/upgrades/v8"
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

	s.ctx = s.app.BaseApp.NewContextLegacy(false, cmtproto.Header{
		ChainID:         "canto_9001-1",
		Height:          1,
		Time:            time.Date(2023, 5, 9, 8, 0, 0, 0, time.UTC),
		ProposerAddress: s.consAddress.Bytes(),
	})

	// In the upgrade to comsos-sdk v0.47.x, the migration from subspace
	// to ParamsStore of the consensus module should be performed. Since
	// the upgrade of binaries does not happen in unittest, it is necessary
	// to set the required values into subspace to prevent panic.
	bz := s.app.AppCodec().MustMarshalJSON(chain.DefaultConsensusParams.Block)
	sp, ok := s.app.ParamsKeeper.GetSubspace(baseapp.Paramspace)
	s.Require().True(ok)
	err = sp.Update(s.ctx, baseapp.ParamStoreKeyBlockParams, bz)
	s.Require().NoError(err)

	// Set Validator
	valAddr := sdk.ValAddress(s.consAddress.Bytes())
	validator, err := stakingtypes.NewValidator(valAddr.String(), priv.PubKey(), stakingtypes.Description{})
	s.Require().NoError(err)
	validator = stakingkeeper.TestingUpdateValidator(s.app.StakingKeeper, s.ctx, validator, true)
	valbz, err := s.app.StakingKeeper.ValidatorAddressCodec().StringToBytes(validator.GetOperator())
	s.Require().NoError(err)
	s.app.StakingKeeper.Hooks().AfterValidatorCreated(s.ctx, valbz)
	err = s.app.StakingKeeper.SetValidatorByConsAddr(s.ctx, validator)
	s.Require().NoError(err)
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
			"v8 upgrade min_commission_rate",
			func() {},
			func() {
				params, err := s.app.StakingKeeper.GetParams(s.ctx)
				s.Require().NoError(err)
				s.Require().Equal(v8.MinCommissionRate, params.MinCommissionRate)
			},
			true,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.title, func() {
			s.SetupTest()
			tc.before()

			// proceed with the block until the test upgrade height
			for s.ctx.BlockHeight() < testUpgradeHeight {
				err := s.Commit()
				s.Require().NoError(err)
			}

			// simulate binary upgrade
			plan := upgradetypes.Plan{Name: v8.UpgradeName, Height: testUpgradeHeight}
			err := s.app.UpgradeKeeper.ScheduleUpgrade(s.ctx, plan)
			s.Require().NoError(err)
			_, err = s.app.UpgradeKeeper.GetUpgradePlan(s.ctx)
			s.Require().NoError(err)

			// execute upgrade handler
			err = s.Commit()
			s.Require().NoError(err)

			tc.after()
		})
	}
}

func (s *UpgradeTestSuite) Commit() error {
	header := s.ctx.BlockHeader()
	if _, err := s.app.FinalizeBlock(&abci.RequestFinalizeBlock{
		Height: header.Height,
	}); err != nil {
		return err
	}

	if _, err := s.app.Commit(); err != nil {
		return err
	}

	header.Height += 1
	if _, err := s.app.FinalizeBlock(&abci.RequestFinalizeBlock{
		Height: header.Height,
		Time:   header.Time,
	}); err != nil {
		return err
	}

	// update ctx
	s.ctx = s.app.BaseApp.NewContextLegacy(false, header)

	return nil
}
