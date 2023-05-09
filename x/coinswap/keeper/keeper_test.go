package keeper_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/tendermint/tendermint/crypto/tmhash"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	"github.com/Canto-Network/Canto/v6/app"
	"github.com/Canto-Network/Canto/v6/x/coinswap/keeper"
	"github.com/Canto-Network/Canto/v6/x/coinswap/types"
)

const (
	denomStandard = sdk.DefaultBondDenom
	denomBTC      = "btc"
	denomETH      = "eth"
)

var (
	addrSender1 sdk.AccAddress
	addrSender2 sdk.AccAddress
)

// test that the params can be properly set and retrieved
type TestSuite struct {
	suite.Suite

	ctx         sdk.Context
	app         *app.Canto
	keeper      keeper.Keeper
	queryClient types.QueryClient
	msgServer   types.MsgServer
}

func (suite *TestSuite) SetupTest() {
	app := setupWithGenesisAccounts()
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

	queryHelper := baseapp.NewQueryServerTestHelper(ctx, app.InterfaceRegistry())
	types.RegisterQueryServer(queryHelper, app.CoinswapKeeper)
	queryClient := types.NewQueryClient(queryHelper)

	suite.app = app
	suite.ctx = ctx
	suite.keeper = app.CoinswapKeeper
	suite.queryClient = queryClient

	sdk.SetCoinDenomRegex(func() string {
		return `[a-zA-Z][a-zA-Z0-9/\-]{2,127}`
	})
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(TestSuite))
}

func (suite *TestSuite) TestParams() {
	cases := []struct {
		params types.Params
	}{
		{types.DefaultParams()},
		{
			params: types.Params{
				Fee:                    sdk.NewDec(0),
				PoolCreationFee:        sdk.Coin{sdk.DefaultBondDenom, sdk.ZeroInt()},
				TaxRate:                sdk.NewDec(0),
				MaxStandardCoinPerPool: sdk.NewInt(10_000_000_000),
				MaxSwapAmount: sdk.NewCoins(
					sdk.NewInt64Coin("usdc", 10_000_000),
					sdk.NewInt64Coin("usdt", 10_000_000),
					sdk.NewInt64Coin("eth", 100_000),
				),
			},
		},
	}
	for _, tc := range cases {
		suite.app.CoinswapKeeper.SetParams(suite.ctx, tc.params)

		feeParam := suite.app.CoinswapKeeper.GetParams(suite.ctx)
		suite.Equal(tc.params.Fee, feeParam.Fee)
	}
}

func setupWithGenesisAccounts() *app.Canto {
	amountInitStandard, _ := sdk.NewIntFromString("30000000000000000000")
	amountInitBTC, _ := sdk.NewIntFromString("3000000000")

	addrSender1 = sdk.AccAddress(tmhash.SumTruncated([]byte("addrSender1")))
	addrSender2 = sdk.AccAddress(tmhash.SumTruncated([]byte("addrSender2")))
	acc1Balances := banktypes.Balance{
		Address: addrSender1.String(),
		Coins: sdk.NewCoins(
			sdk.NewCoin(denomStandard, amountInitStandard),
			sdk.NewCoin(denomBTC, amountInitBTC),
		),
	}

	acc2Balances := banktypes.Balance{
		Address: addrSender2.String(),
		Coins: sdk.NewCoins(
			sdk.NewCoin(denomStandard, amountInitStandard),
			sdk.NewCoin(denomBTC, amountInitBTC),
		),
	}

	acc1 := &authtypes.BaseAccount{
		Address: addrSender1.String(),
	}
	acc2 := &authtypes.BaseAccount{
		Address: addrSender2.String(),
	}

	genAccs := []authtypes.GenesisAccount{acc1, acc2}
	app := app.SetupWithGenesisAccounts(genAccs, acc1Balances, acc2Balances)
	return app
}

func (suite *TestSuite) TestAmountOf() {

	MaxSwapAmount := sdk.NewCoins(
		sdk.NewInt64Coin("ibc/FBEEDF2F566CF2568921399BD092363FCC45EB53278A3A09318C4348AAE2B27F", 1000000),
		sdk.NewInt64Coin("ibc/4B32742658E7D16C1F77468D0DC35178731D694DEB17378242647EA02622EF64", 1000000),
	)

	searchDenom := "ibc/FBEEDF2F566CF2568921399BD092363FCC45EB53278A3A09318C4348AAE2B27F"

	var amount sdk.Int

	cases := []struct {
		name           string
		malleate       func()
		expectedAmount sdk.Int
	}{
		{
			"AmountOf doesn't work for some denoms",
			func() {
				amount = MaxSwapAmount.AmountOf(searchDenom)
			},
			sdk.NewInt(1000000),
		},
		{
			"manual search for denom works",
			func() {

				for _, coin := range MaxSwapAmount {
					if coin.Denom == searchDenom {
						amount = coin.Amount
						break
					}
				}
			},
			sdk.NewInt(1000000),
		},
	}
	for _, tc := range cases {
		tc.malleate()
		suite.Equal(tc.expectedAmount, amount, tc.name)
	}
}

func (suite *TestSuite) TestLiquidity() {
	params := types.Params{
		Fee:                    sdk.NewDec(0),
		PoolCreationFee:        sdk.Coin{sdk.DefaultBondDenom, sdk.ZeroInt()},
		TaxRate:                sdk.NewDec(0),
		MaxStandardCoinPerPool: sdk.NewInt(10_000_000_000),
		MaxSwapAmount:          sdk.NewCoins(sdk.NewInt64Coin(denomBTC, 10_000_000)),
	}
	suite.app.CoinswapKeeper.SetParams(suite.ctx, params)

	// Test add liquidity with non-whitelisted denom
	// Fail to create a pool
	ethAmt, _ := sdk.NewIntFromString("100")
	standardAmt, _ := sdk.NewIntFromString("1000000000")
	depositCoin := sdk.NewCoin(denomETH, ethAmt)
	minReward := sdk.NewInt(1)
	deadline := time.Now().Add(1 * time.Minute)

	msg := types.NewMsgAddLiquidity(depositCoin, standardAmt, minReward, deadline.Unix(), addrSender1.String())
	_, err := suite.app.CoinswapKeeper.AddLiquidity(suite.ctx, msg)
	suite.Error(err)

	// test add liquidity with exceeding standard coin limit
	// Expected behavior: fails to create to pool
	btcAmt, _ := sdk.NewIntFromString("100")
	standardAmt, _ = sdk.NewIntFromString("15000000000")
	depositCoin = sdk.NewCoin(denomBTC, btcAmt)

	msg = types.NewMsgAddLiquidity(depositCoin, standardAmt, minReward, deadline.Unix(), addrSender1.String())
	_, err = suite.app.CoinswapKeeper.AddLiquidity(suite.ctx, msg)
	suite.Error(err)

	// Test add liquidity
	// Deposit: 100btc, 8000000000stake
	// Pool created and mint 8000000000lpt-1
	// Expected pool balance: 100btc, 8000000000stake
	standardAmt, _ = sdk.NewIntFromString("8000000000")
	msg = types.NewMsgAddLiquidity(depositCoin, standardAmt, minReward, deadline.Unix(), addrSender1.String())
	_, err = suite.app.CoinswapKeeper.AddLiquidity(suite.ctx, msg)
	suite.NoError(err)

	poolId := types.GetPoolId(denomBTC)
	pool, has := suite.app.CoinswapKeeper.GetPool(suite.ctx, poolId)
	suite.Require().True(has)

	poolAddr, err := sdk.AccAddressFromBech32(pool.EscrowAddress)
	suite.Require().NoError(err)

	lptDenom := pool.LptDenom

	reservePoolBalances := suite.app.BankKeeper.GetAllBalances(suite.ctx, poolAddr)
	sender1Balances := suite.app.BankKeeper.GetAllBalances(suite.ctx, addrSender1)
	suite.Equal("8000000000", suite.app.BankKeeper.GetSupply(suite.ctx, lptDenom).Amount.String())

	expCoins := sdk.NewCoins(
		sdk.NewInt64Coin(denomBTC, 100),
		sdk.NewCoin(denomStandard, sdk.NewIntWithDecimal(8000, 6)),
	)
	suite.Equal(expCoins.Sort().String(), reservePoolBalances.Sort().String())

	expCoins = sdk.NewCoins(
		sdk.NewInt64Coin(denomBTC, 2999999900),
		sdk.NewCoin(denomStandard, sdk.NewIntWithDecimal(3, 19).Sub(sdk.NewIntWithDecimal(8000, 6))),
		sdk.NewCoin(lptDenom, sdk.NewIntWithDecimal(8000, 6)),
	)
	suite.Equal(expCoins.Sort().String(), sender1Balances.Sort().String())

	// test add liquidity (pool exists)
	// Deposit try: 200btc, 8000000000stake
	// Actual deposit: 26btc, 2000000000stake
	// Mint: 2000000000lpt-1
	// Expected pool balance: 126btc, 10000000000stake
	expLptDenom, _ := suite.app.CoinswapKeeper.GetLptDenomFromDenoms(suite.ctx, denomBTC, denomStandard)
	suite.Require().Equal(expLptDenom, lptDenom)

	btcAmt, _ = sdk.NewIntFromString("200")
	standardAmt, _ = sdk.NewIntFromString("8000000000")
	depositCoin = sdk.NewCoin(denomBTC, btcAmt)
	minReward = sdk.NewInt(1)
	deadline = time.Now().Add(1 * time.Minute)

	msg = types.NewMsgAddLiquidity(depositCoin, standardAmt, minReward, deadline.Unix(), addrSender2.String())
	_, err = suite.app.CoinswapKeeper.AddLiquidity(suite.ctx, msg)
	suite.NoError(err)

	reservePoolBalances = suite.app.BankKeeper.GetAllBalances(suite.ctx, poolAddr)
	sender2Balances := suite.app.BankKeeper.GetAllBalances(suite.ctx, addrSender2)
	suite.Equal("10000000000", suite.app.BankKeeper.GetSupply(suite.ctx, lptDenom).Amount.String())

	expCoins = sdk.NewCoins(
		sdk.NewInt64Coin(denomBTC, 126),
		sdk.NewCoin(denomStandard, sdk.NewIntWithDecimal(10000, 6)),
	)
	suite.Equal(expCoins.Sort().String(), reservePoolBalances.Sort().String())

	expCoins = sdk.NewCoins(
		sdk.NewInt64Coin(denomBTC, 2999999974),
		sdk.NewCoin(denomStandard, sdk.NewIntWithDecimal(3, 19).Sub(sdk.NewIntWithDecimal(2000, 6))),
		sdk.NewCoin(lptDenom, sdk.NewIntWithDecimal(2000, 6)),
	)
	suite.Equal(expCoins.Sort().String(), sender2Balances.Sort().String())

	// Test add liquidity when the pool is maxed
	// Expected behavior: fails to deposit
	_, err = suite.app.CoinswapKeeper.AddLiquidity(suite.ctx, msg)
	suite.Error(err)

	// Test remove liquidity (remove part)
	// Withdraw 8000*10^6 pool coin
	// Expected return: 8000*10^6 standard coin, 100btc
	// Expected pool reserve: 2000*10^6 standard coin, 26btc
	withdraw, _ := sdk.NewIntFromString("8000000000")
	msgRemove := types.NewMsgRemoveLiquidity(
		sdk.NewInt(1),
		sdk.NewCoin(lptDenom, withdraw),
		sdk.NewInt(1),
		suite.ctx.BlockHeader().Time.Unix(),
		addrSender1.String(),
	)

	_, err = suite.app.CoinswapKeeper.RemoveLiquidity(suite.ctx, msgRemove)
	suite.NoError(err)

	reservePoolBalances = suite.app.BankKeeper.GetAllBalances(suite.ctx, poolAddr)
	sender1Balances = suite.app.BankKeeper.GetAllBalances(suite.ctx, addrSender1)
	suite.Equal("2000000000", suite.app.BankKeeper.GetSupply(suite.ctx, lptDenom).Amount.String())

	expCoins = sdk.NewCoins(
		sdk.NewInt64Coin(denomBTC, 3000000000),
		sdk.NewCoin(denomStandard, sdk.NewIntWithDecimal(3, 19)),
	)
	suite.Equal(expCoins.Sort().String(), sender1Balances.Sort().String())

	expCoins = sdk.NewCoins(
		sdk.NewInt64Coin(denomBTC, 26),
		sdk.NewCoin(denomStandard, sdk.NewIntWithDecimal(2000, 6)),
	)
	suite.Equal(expCoins.Sort().String(), reservePoolBalances.String())

	// Test remove liquidity (overdraft)
	// Expected behavior: fails to withdraw
	withdraw = sdk.NewIntWithDecimal(8000, 6)
	msgRemove = types.NewMsgRemoveLiquidity(
		sdk.NewInt(1),
		sdk.NewCoin(lptDenom, withdraw),
		sdk.NewInt(1),
		suite.ctx.BlockHeader().Time.Unix(),
		addrSender2.String(),
	)

	_, err = suite.app.CoinswapKeeper.RemoveLiquidity(suite.ctx, msgRemove)
	suite.Error(err)

	// Test remove liquidity (remove all)
	// Expected pool coin supply: 0
	// Expected reserve balance: 0btc, 0stake
	withdraw = sdk.NewIntWithDecimal(2000, 6)
	msgRemove = types.NewMsgRemoveLiquidity(
		sdk.NewInt(1),
		sdk.NewCoin(lptDenom, withdraw),
		sdk.NewInt(1),
		suite.ctx.BlockHeader().Time.Unix(),
		addrSender2.String(),
	)

	_, err = suite.app.CoinswapKeeper.RemoveLiquidity(suite.ctx, msgRemove)
	suite.NoError(err)

	reservePoolBalances = suite.app.BankKeeper.GetAllBalances(suite.ctx, poolAddr)
	sender2Balances = suite.app.BankKeeper.GetAllBalances(suite.ctx, addrSender2)
	suite.Equal("0", suite.app.BankKeeper.GetSupply(suite.ctx, lptDenom).Amount.String())

	expCoins = sdk.NewCoins(
		sdk.NewInt64Coin(denomBTC, 3000000000),
		sdk.NewCoin(denomStandard, sdk.NewIntWithDecimal(3, 19)),
	)
	suite.Equal(expCoins.Sort().String(), sender2Balances.Sort().String())
	suite.Equal("", reservePoolBalances.String())

	// Test add liquidity (pool exists but empty)
	// Deposit: 200btc, 8000000000stake
	// Pool created and mint 8000000000lpt-1
	// Expected pool balance: 200btc, 8000000000stake
	standardAmt, _ = sdk.NewIntFromString("8000000000")
	msg = types.NewMsgAddLiquidity(depositCoin, standardAmt, minReward, deadline.Unix(), addrSender1.String())
	_, err = suite.app.CoinswapKeeper.AddLiquidity(suite.ctx, msg)
	suite.NoError(err)

	poolId = types.GetPoolId(denomBTC)
	pool, has = suite.app.CoinswapKeeper.GetPool(suite.ctx, poolId)
	suite.Require().True(has)

	poolAddr, err = sdk.AccAddressFromBech32(pool.EscrowAddress)
	suite.Require().NoError(err)

	lptDenom = pool.LptDenom

	reservePoolBalances = suite.app.BankKeeper.GetAllBalances(suite.ctx, poolAddr)
	sender1Balances = suite.app.BankKeeper.GetAllBalances(suite.ctx, addrSender1)
	suite.Equal("8000000000", suite.app.BankKeeper.GetSupply(suite.ctx, lptDenom).Amount.String())

	expCoins = sdk.NewCoins(
		sdk.NewInt64Coin(denomBTC, 200),
		sdk.NewCoin(denomStandard, sdk.NewIntWithDecimal(8000, 6)),
	)
	suite.Equal(expCoins.Sort().String(), reservePoolBalances.Sort().String())
}
