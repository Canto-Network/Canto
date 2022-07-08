package keeper_test

import (
	"fmt"

	"github.com/Canto-Network/Canto-Testnet-v2/v1/x/inflation/types"
	ethermint "github.com/Canto-Network/ethermint-v2/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

func (suite *KeeperTestSuite) TestMintAndAllocateInflation() {
	testCases := []struct {
		name                string
		mintCoin            sdk.Coin
		malleate            func()
		expStakingRewardAmt sdk.Coin
		expCommunityPoolAmt sdk.DecCoins
		expPass             bool
	}{
		{
			"pass",
			sdk.NewCoin(denomMint, sdk.NewInt(1_000_000)),
			func() {},
			sdk.NewCoin(denomMint, sdk.NewInt(800_000)),
			sdk.NewDecCoins(sdk.NewDecCoin(denomMint, sdk.NewInt(200_000))),
			true,
		},
		{
			"pass - no coins minted ",
			sdk.NewCoin(denomMint, sdk.ZeroInt()),
			func() {},
			sdk.NewCoin(denomMint, sdk.ZeroInt()),
			sdk.DecCoins(nil),
			true,
		},
	}
	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.SetupTest() // reset

			tc.malleate()

			_, _, err := suite.app.InflationKeeper.MintAndAllocateInflation(suite.ctx, tc.mintCoin)

			// Get balances
			balanceModule := suite.app.BankKeeper.GetBalance(
				suite.ctx,
				suite.app.AccountKeeper.GetModuleAddress(types.ModuleName),
				denomMint,
			)

			feeCollector := suite.app.AccountKeeper.GetModuleAddress(authtypes.FeeCollectorName)
			balanceStakingRewards := suite.app.BankKeeper.GetBalance(
				suite.ctx,
				feeCollector,
				denomMint,
			)

			balanceCommunityPool := suite.app.DistrKeeper.GetFeePoolCommunityCoins(suite.ctx)

			if tc.expPass {
				suite.Require().NoError(err, tc.name)
				suite.Require().True(balanceModule.IsZero())
				suite.Require().Equal(tc.expStakingRewardAmt, balanceStakingRewards)
				suite.Require().Equal(tc.expCommunityPoolAmt, balanceCommunityPool)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestGetCirculatingSupplyAndProvision() {
	testCases := []struct {
		name             string
		bankSupply       int64
		malleate         func()
		expInflationRate sdk.Dec
		expProvision     sdk.Dec
	}{
		{
			"Low Circulating Supply",
			400_000_000,
			func() {
				suite.app.InflationKeeper.SetEpochMintProvision(suite.ctx, sdk.ZeroDec())
			},
			sdk.MustNewDecFromStr("0.007640821917808219"),
			sdk.MustNewDecFromStr("1115560000000000000000000000.000000000000000000"),
		},
		{
			"no epochs per period",
			400_000_000,
			func() {
				suite.app.InflationKeeper.SetEpochsPerPeriod(suite.ctx, 0)
			},
			sdk.ZeroDec(),
			sdk.ZeroDec(),
		},
		{
			"high supply",
			800_000_000,
			func() {},
			sdk.MustNewDecFromStr("0.007640821917808219"),
			sdk.MustNewDecFromStr("2231120000000000000000000000.000000000000000000"),
		},
	}
	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.SetupTest() // reset

			// Team allocation is only set on mainnet
			suite.ctx = suite.ctx.WithChainID("canto_9001-1")
			tc.malleate()

			// Mint coins to increase supply
			coin := sdk.NewCoin(types.DefaultInflationDenom, sdk.TokensFromConsensusPower(tc.bankSupply, ethermint.PowerReduction))
			decCoin := sdk.NewDecCoinFromCoin(coin)
			err := suite.app.InflationKeeper.MintCoins(suite.ctx, coin)
			suite.Require().NoError(err)

			circulatingSupply := s.app.InflationKeeper.GetCirculatingSupply(suite.ctx)

			suite.Require().Equal(decCoin.Amount, circulatingSupply)

			inflationRate, err := s.app.InflationKeeper.GetInflationRate(suite.ctx)
			suite.Require().NoError(err)
			suite.Require().Equal(tc.expInflationRate, inflationRate)
			periodProvision, err := s.app.InflationKeeper.CalculateEpochMintProvision(suite.ctx)
			suite.Require().NoError(err)
			suite.Require().Equal(tc.expProvision, periodProvision)
		})
	}
}
