package keeper_test

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Canto-Network/Canto/v7/x/coinswap/keeper"
	"github.com/Canto-Network/Canto/v7/x/coinswap/types"
)

func TestSwapSuite(t *testing.T) {
	suite.Run(t, new(TestSuite))
}

type Data struct {
	delta sdkmath.Int
	x     sdkmath.Int
	y     sdkmath.Int
	fee   sdkmath.LegacyDec
}
type SwapCase struct {
	data   Data
	expect sdkmath.Int
}

func (suite *TestSuite) TestGetInputPrice() {
	var datas = []SwapCase{{
		data:   Data{delta: sdkmath.NewInt(100), x: sdkmath.NewInt(1000), y: sdkmath.NewInt(1000), fee: sdkmath.LegacyNewDecWithPrec(3, 3)},
		expect: sdkmath.NewInt(90),
	}, {
		data:   Data{delta: sdkmath.NewInt(200), x: sdkmath.NewInt(1000), y: sdkmath.NewInt(1000), fee: sdkmath.LegacyNewDecWithPrec(3, 3)},
		expect: sdkmath.NewInt(166),
	}, {
		data:   Data{delta: sdkmath.NewInt(300), x: sdkmath.NewInt(1000), y: sdkmath.NewInt(1000), fee: sdkmath.LegacyNewDecWithPrec(3, 3)},
		expect: sdkmath.NewInt(230),
	}, {
		data:   Data{delta: sdkmath.NewInt(1000), x: sdkmath.NewInt(1000), y: sdkmath.NewInt(1000), fee: sdkmath.LegacyNewDecWithPrec(3, 3)},
		expect: sdkmath.NewInt(499),
	}, {
		data:   Data{delta: sdkmath.NewInt(1000), x: sdkmath.NewInt(1000), y: sdkmath.NewInt(1000), fee: sdkmath.LegacyZeroDec()},
		expect: sdkmath.NewInt(500),
	}}
	for _, tcase := range datas {
		data := tcase.data
		actual := keeper.GetInputPrice(data.delta, data.x, data.y, data.fee)
		suite.Equal(tcase.expect, actual)
	}
}

func (suite *TestSuite) TestGetOutputPrice() {
	var datas = []SwapCase{{
		data:   Data{delta: sdkmath.NewInt(100), x: sdkmath.NewInt(1000), y: sdkmath.NewInt(1000), fee: sdkmath.LegacyNewDecWithPrec(3, 3)},
		expect: sdkmath.NewInt(112),
	}, {
		data:   Data{delta: sdkmath.NewInt(200), x: sdkmath.NewInt(1000), y: sdkmath.NewInt(1000), fee: sdkmath.LegacyNewDecWithPrec(3, 3)},
		expect: sdkmath.NewInt(251),
	}, {
		data:   Data{delta: sdkmath.NewInt(300), x: sdkmath.NewInt(1000), y: sdkmath.NewInt(1000), fee: sdkmath.LegacyNewDecWithPrec(3, 3)},
		expect: sdkmath.NewInt(430),
	}, {
		data:   Data{delta: sdkmath.NewInt(300), x: sdkmath.NewInt(1000), y: sdkmath.NewInt(1000), fee: sdkmath.LegacyZeroDec()},
		expect: sdkmath.NewInt(429),
	}}
	for _, tcase := range datas {
		data := tcase.data
		actual := keeper.GetOutputPrice(data.delta, data.x, data.y, data.fee)
		suite.Equal(tcase.expect, actual)
	}
}

func (suite *TestSuite) TestSwap() {
	sender, reservePoolAddr := createReservePool(suite, denomBTC)

	poolId := types.GetPoolId(denomBTC)
	pool, has := suite.app.CoinswapKeeper.GetPool(suite.ctx, poolId)
	suite.Require().True(has)

	lptDenom := pool.LptDenom

	// swap buy order msg
	msg := types.NewMsgSwapOrder(
		types.Input{Coin: sdk.NewCoin(denomBTC, sdkmath.NewIntWithDecimal(100, 6)), Address: sender.String()},
		types.Output{Coin: sdk.NewCoin(denomStandard, sdkmath.NewIntWithDecimal(50, 6)), Address: sender.String()},
		time.Now().Add(1*time.Minute).Unix(),
		true,
	)
	// failed swap buy order because of exceeded maximum swap amount
	err := suite.app.CoinswapKeeper.Swap(suite.ctx, msg)
	suite.Error(err)

	// swap buy order msg
	msg = types.NewMsgSwapOrder(
		types.Input{Coin: sdk.NewCoin(denomBTC, sdkmath.NewIntWithDecimal(10, 6)), Address: sender.String()},
		types.Output{Coin: sdk.NewCoin(denomStandard, sdkmath.NewIntWithDecimal(5, 6)), Address: sender.String()},
		time.Now().Add(1*time.Minute).Unix(),
		true,
	)
	// first successful swap buy order
	err = suite.app.CoinswapKeeper.Swap(suite.ctx, msg)
	suite.NoError(err)
	reservePoolBalances := suite.app.BankKeeper.GetAllBalances(suite.ctx, reservePoolAddr)
	senderBalances := suite.app.BankKeeper.GetAllBalances(suite.ctx, sender)

	expCoins := sdk.NewCoins(
		sdk.NewInt64Coin(denomBTC, 10005002502),
		sdk.NewInt64Coin(denomStandard, 9995000000),
	)
	suite.Equal(expCoins.Sort().String(), reservePoolBalances.Sort().String())

	expCoins = sdk.NewCoins(
		sdk.NewInt64Coin(denomBTC, 19994997498),
		sdk.NewInt64Coin(denomStandard, 20005000000),
		sdk.NewInt64Coin(lptDenom, 10000000000),
	)
	suite.Equal(expCoins.Sort().String(), senderBalances.Sort().String())

	// second swap buy order
	err = suite.app.CoinswapKeeper.Swap(suite.ctx, msg)
	suite.NoError(err)
	reservePoolBalances = suite.app.BankKeeper.GetAllBalances(suite.ctx, reservePoolAddr)
	senderBalances = suite.app.BankKeeper.GetAllBalances(suite.ctx, sender)

	expCoins = sdk.NewCoins(
		sdk.NewInt64Coin(denomBTC, 10010010011),
		sdk.NewInt64Coin(denomStandard, 9990000000),
	)
	suite.Equal(expCoins.Sort().String(), reservePoolBalances.Sort().String())

	expCoins = sdk.NewCoins(
		sdk.NewInt64Coin(denomBTC, 19989989989),
		sdk.NewInt64Coin(denomStandard, 20010000000),
		sdk.NewInt64Coin(lptDenom, 10000000000),
	)
	suite.Equal(expCoins.Sort().String(), senderBalances.Sort().String())

	reservePoolBalances = suite.app.BankKeeper.GetAllBalances(suite.ctx, reservePoolAddr)

	// swap sell order msg
	msg = types.NewMsgSwapOrder(
		types.Input{Coin: sdk.NewCoin(denomStandard, sdkmath.NewIntWithDecimal(100, 6)), Address: sender.String()},
		types.Output{Coin: sdk.NewCoin(denomBTC, sdkmath.NewIntWithDecimal(100, 6)), Address: sender.String()},
		time.Now().Add(1*time.Minute).Unix(),
		false,
	)

	// first swap sell order
	// failed because of exceed the maximum swap amount
	err = suite.app.CoinswapKeeper.Swap(suite.ctx, msg)
	suite.Error(err)

	// swap sell order msg
	msg = types.NewMsgSwapOrder(
		types.Input{Coin: sdk.NewCoin(denomStandard, sdkmath.NewIntWithDecimal(5, 6)), Address: sender.String()},
		types.Output{Coin: sdk.NewCoin(denomBTC, sdkmath.NewIntWithDecimal(5, 6)), Address: sender.String()},
		time.Now().Add(1*time.Minute).Unix(),
		false,
	)

	// first successful swap sell order
	err = suite.app.CoinswapKeeper.Swap(suite.ctx, msg)
	suite.NoError(err)

	reservePoolBalances = suite.app.BankKeeper.GetAllBalances(suite.ctx, reservePoolAddr)
	senderBalances = suite.app.BankKeeper.GetAllBalances(suite.ctx, sender)
	expCoins = sdk.NewCoins(
		sdk.NewInt64Coin(denomBTC, 10005002503),
		sdk.NewInt64Coin(denomStandard, 9995000000),
	)
	suite.Equal(expCoins.Sort().String(), reservePoolBalances.Sort().String())

	expCoins = sdk.NewCoins(
		sdk.NewInt64Coin(denomBTC, 19994997497),
		sdk.NewInt64Coin(denomStandard, 20005000000),
		sdk.NewInt64Coin(lptDenom, 10000000000),
	)
	suite.Equal(expCoins.Sort().String(), senderBalances.Sort().String())

	// second successful swap sell order
	err = suite.app.CoinswapKeeper.Swap(suite.ctx, msg)
	suite.NoError(err)
	reservePoolBalances = suite.app.BankKeeper.GetAllBalances(suite.ctx, reservePoolAddr)
	senderBalances = suite.app.BankKeeper.GetAllBalances(suite.ctx, sender)

	expCoins = sdk.NewCoins(
		sdk.NewInt64Coin(denomBTC, 10000000002),
		sdk.NewInt64Coin(denomStandard, 10000000000),
	)
	suite.Equal(expCoins.Sort().String(), reservePoolBalances.Sort().String())

	expCoins = sdk.NewCoins(
		sdk.NewInt64Coin(denomBTC, 19999999998),
		sdk.NewInt64Coin(denomStandard, 20000000000),
		sdk.NewInt64Coin(lptDenom, 10000000000),
	)
	suite.Equal(expCoins.Sort().String(), senderBalances.Sort().String())
}

func createReservePool(suite *TestSuite, denom string) (sdk.AccAddress, sdk.AccAddress) {
	// Set parameters
	params := types.Params{
		Fee:                    sdkmath.LegacyNewDec(0),
		PoolCreationFee:        sdk.Coin{denomStandard, sdkmath.ZeroInt()},
		TaxRate:                sdkmath.LegacyNewDec(0),
		MaxStandardCoinPerPool: sdkmath.NewInt(10_000_000_000),
		MaxSwapAmount: sdk.NewCoins(sdk.NewInt64Coin(denomBTC, 10_000_000),
			sdk.NewInt64Coin(denomETH, 10_000_000),
		),
	}
	suite.app.CoinswapKeeper.SetParams(suite.ctx, params)

	amountInit, _ := sdkmath.NewIntFromString("30000000000")
	addrSender := sdk.AccAddress(getRandomString(20))
	_ = suite.app.AccountKeeper.NewAccountWithAddress(suite.ctx, addrSender)

	coins := sdk.NewCoins(
		sdk.NewCoin(denomStandard, amountInit),
		sdk.NewCoin(denom, amountInit),
	)

	err := suite.app.BankKeeper.MintCoins(suite.ctx, types.ModuleName, coins)
	suite.NoError(err)
	err = suite.app.BankKeeper.SendCoinsFromModuleToAccount(suite.ctx, types.ModuleName, addrSender, coins)
	suite.NoError(err)

	depositAmt, _ := sdkmath.NewIntFromString("10000000000")
	depositCoin := sdk.NewCoin(denom, depositAmt)

	standardAmt, _ := sdkmath.NewIntFromString("10000000000")
	minReward := sdkmath.NewInt(1)
	deadline := time.Now().Add(1 * time.Minute)
	msg := types.NewMsgAddLiquidity(depositCoin, standardAmt, minReward, deadline.Unix(), addrSender.String())
	_, err = suite.app.CoinswapKeeper.AddLiquidity(suite.ctx, msg)
	suite.NoError(err)

	poolId := types.GetPoolId(denom)
	pool, has := suite.app.CoinswapKeeper.GetPool(suite.ctx, poolId)
	suite.Require().True(has)
	reservePoolAddr := types.GetReservePoolAddr(pool.LptDenom)

	reservePoolBalances := suite.app.BankKeeper.GetAllBalances(suite.ctx, reservePoolAddr)
	senderBlances := suite.app.BankKeeper.GetAllBalances(suite.ctx, addrSender)
	suite.Equal("10000000000", suite.app.BankKeeper.GetSupply(suite.ctx, pool.LptDenom).Amount.String())

	expCoins := sdk.NewCoins(
		sdk.NewInt64Coin(denom, 10000000000),
		sdk.NewInt64Coin(denomStandard, 10000000000),
	)
	suite.Equal(expCoins.Sort().String(), reservePoolBalances.Sort().String())

	params = suite.app.CoinswapKeeper.GetParams(suite.ctx)
	expCoins = sdk.NewCoins(
		sdk.NewInt64Coin(denom, 20000000000),
		sdk.NewInt64Coin(denomStandard, 20000000000).Sub(params.PoolCreationFee),
		sdk.NewInt64Coin(pool.LptDenom, 10000000000),
	)
	suite.Equal(expCoins.Sort().String(), senderBlances.Sort().String())
	return addrSender, reservePoolAddr
}

func (suite *TestSuite) TestTradeInputForExactOutput() {
	sender, poolAddr := createReservePool(suite, denomBTC)

	outputCoin := sdk.NewCoin(denomStandard, sdkmath.NewInt(1000))
	inputCoin := sdk.NewCoin(denomBTC, sdkmath.NewInt(1000000))
	input := types.Input{
		Address: sender.String(),
		Coin:    inputCoin,
	}
	output := types.Output{
		Address: sender.String(),
		Coin:    outputCoin,
	}

	poolBalances := suite.app.BankKeeper.GetAllBalances(suite.ctx, poolAddr)
	senderBlances := suite.app.BankKeeper.GetAllBalances(suite.ctx, sender)

	initSupplyOutput := poolBalances.AmountOf(outputCoin.Denom)
	maxCnt := int(initSupplyOutput.Quo(outputCoin.Amount).Int64())

	for i := 1; i < 100; i++ {
		amt, err := suite.app.CoinswapKeeper.TradeInputForExactOutput(suite.ctx, input, output)
		if i == maxCnt {
			suite.Error(err)
			break
		}
		suite.NoError(err)

		bought := sdk.NewCoins(outputCoin)
		sold := sdk.NewCoins(sdk.NewCoin(denomBTC, amt))

		pb := poolBalances.Add(sold...).Sub(bought...)
		sb := senderBlances.Add(bought...).Sub(sold...)
		fmt.Println(pb.String())
		assertResult(suite, poolAddr, sender, pb, sb)

		poolBalances = pb
		senderBlances = sb
	}
}

func (suite *TestSuite) TestTradeExactInputForOutput() {
	sender, poolAddr := createReservePool(suite, denomBTC)

	outputCoin := sdk.NewCoin(denomStandard, sdkmath.NewInt(0))
	inputCoin := sdk.NewCoin(denomBTC, sdkmath.NewInt(10000))
	input := types.Input{
		Address: sender.String(),
		Coin:    inputCoin,
	}
	output := types.Output{
		Address: sender.String(),
		Coin:    outputCoin,
	}

	poolBalances := suite.app.BankKeeper.GetAllBalances(suite.ctx, poolAddr)
	senderBlances := suite.app.BankKeeper.GetAllBalances(suite.ctx, sender)

	for i := 1; i < 1000; i++ {
		amt, err := suite.app.CoinswapKeeper.TradeExactInputForOutput(suite.ctx, input, output)
		suite.NoError(err)

		sold := sdk.NewCoins(inputCoin)
		bought := sdk.NewCoins(sdk.NewCoin(denomStandard, amt))

		pb := poolBalances.Add(sold...).Sub(bought...)
		sb := senderBlances.Add(bought...).Sub(sold...)

		assertResult(suite, poolAddr, sender, pb, sb)

		poolBalances = pb
		senderBlances = sb
	}
}

func assertResult(suite *TestSuite, reservePoolAddr, sender sdk.AccAddress, expectPoolBalance, expectSenderBalance sdk.Coins) {
	reservePoolBalances := suite.app.BankKeeper.GetAllBalances(suite.ctx, reservePoolAddr)
	senderBlances := suite.app.BankKeeper.GetAllBalances(suite.ctx, sender)
	suite.Equal(expectPoolBalance.String(), reservePoolBalances.String())
	suite.Equal(expectSenderBalance.String(), senderBlances.String())
}

func getRandomString(l int) string {
	str := "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	bytes := []byte(str)
	var result []byte
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := 0; i < l; i++ {
		result = append(result, bytes[r.Intn(len(bytes))])
	}
	return string(result)
}
