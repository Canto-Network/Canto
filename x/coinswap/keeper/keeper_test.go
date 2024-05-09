package keeper_test

import (
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/suite"

	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cometbft/cometbft/crypto/tmhash"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	tmversion "github.com/cometbft/cometbft/proto/tendermint/version"
	"github.com/cometbft/cometbft/version"

	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/evmos/ethermint/crypto/ethsecp256k1"

	"github.com/Canto-Network/Canto/v7/app"
	"github.com/Canto-Network/Canto/v7/x/coinswap/keeper"
	"github.com/Canto-Network/Canto/v7/x/coinswap/types"
)

var (
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
	validator   stakingtypes.Validator
}

func (suite *TestSuite) SetupTest() {
	suite.app = setupWithGenesisAccounts()

	// account key
	priv, err := ethsecp256k1.GenerateKey()
	suite.Require().NoError(err)
	address := common.BytesToAddress(priv.PubKey().Address().Bytes())

	// consensus key
	pubKey := ed25519.GenPrivKey().PubKey()
	consAddress := sdk.ConsAddress(pubKey.Address())

	suite.ctx = suite.app.BaseApp.NewContextLegacy(false, tmproto.Header{
		Height:          1,
		ChainID:         "canto_9001-1",
		Time:            time.Now().UTC(),
		ProposerAddress: consAddress.Bytes(),

		Version: tmversion.Consensus{
			Block: version.BlockProtocol,
		},
		LastBlockId: tmproto.BlockID{
			Hash: tmhash.Sum([]byte("block_id")),
			PartSetHeader: tmproto.PartSetHeader{
				Total: 11,
				Hash:  tmhash.Sum([]byte("partset_header")),
			},
		},
		AppHash:            tmhash.Sum([]byte("app")),
		DataHash:           tmhash.Sum([]byte("data")),
		EvidenceHash:       tmhash.Sum([]byte("evidence")),
		ValidatorsHash:     tmhash.Sum([]byte("validators")),
		NextValidatorsHash: tmhash.Sum([]byte("next_validators")),
		ConsensusHash:      tmhash.Sum([]byte("consensus")),
		LastResultsHash:    tmhash.Sum([]byte("last_result")),
	})

	queryHelper := baseapp.NewQueryServerTestHelper(suite.ctx, suite.app.InterfaceRegistry())
	types.RegisterQueryServer(queryHelper, suite.app.CoinswapKeeper)
	queryClient := types.NewQueryClient(queryHelper)

	suite.keeper = suite.app.CoinswapKeeper
	suite.queryClient = queryClient

	sdk.SetCoinDenomRegex(func() string {
		return `[a-zA-Z][a-zA-Z0-9/\-]{2,127}`
	})

	// Set Validator
	valAddr := sdk.ValAddress(address.Bytes())
	validator, err := stakingtypes.NewValidator(valAddr.String(), pubKey, stakingtypes.Description{})
	suite.Require().NoError(err)

	valbz, err := suite.app.StakingKeeper.ValidatorAddressCodec().StringToBytes(validator.GetOperator())
	suite.Require().NoError(err)
	suite.app.StakingKeeper.SetValidator(suite.ctx, validator)
	suite.app.StakingKeeper.Hooks().AfterValidatorCreated(suite.ctx, valbz)
	err = suite.app.StakingKeeper.SetValidatorByConsAddr(suite.ctx, validator)
	suite.Require().NoError(err)
	suite.validator = validator
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(TestSuite))
}

// Commit commits and starts a new block with an updated context.
func (suite *TestSuite) Commit() {
	suite.CommitAfter(time.Second * 0)
}

func (suite *TestSuite) CommitAfter(t time.Duration) {
	header := suite.ctx.BlockHeader()
	suite.app.FinalizeBlock(&abci.RequestFinalizeBlock{
		Height: header.Height,
	})

	suite.app.Commit()

	header.Height += 1
	header.Time = header.Time.Add(t)
	suite.app.FinalizeBlock(&abci.RequestFinalizeBlock{
		Height: header.Height,
		Time:   header.Time,
	})

	// update ctx
	suite.ctx = suite.app.BaseApp.NewContextLegacy(false, header)
}

func (suite *TestSuite) TestParams() {
	cases := []struct {
		params types.Params
	}{
		{types.DefaultParams()},
		{
			params: types.Params{
				Fee:                    sdkmath.LegacyNewDec(0),
				PoolCreationFee:        sdk.Coin{sdk.DefaultBondDenom, sdkmath.ZeroInt()},
				TaxRate:                sdkmath.LegacyNewDec(0),
				MaxStandardCoinPerPool: sdkmath.NewInt(10_000_000_000),
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
	amountInitStandard, _ := sdkmath.NewIntFromString("30000000000000000000")
	amountInitBTC, _ := sdkmath.NewIntFromString("3000000000")

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

	var amount sdkmath.Int

	cases := []struct {
		name           string
		malleate       func()
		expectedAmount sdkmath.Int
	}{
		{
			"AmountOf doesn't work for some denoms",
			func() {
				amount = MaxSwapAmount.AmountOf(searchDenom)
			},
			sdkmath.NewInt(1000000),
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
			sdkmath.NewInt(1000000),
		},
	}
	for _, tc := range cases {
		tc.malleate()
		suite.Equal(tc.expectedAmount, amount, tc.name)
	}
}

func (suite *TestSuite) TestLiquidity() {
	params := types.Params{
		Fee:                    sdkmath.LegacyNewDec(0),
		PoolCreationFee:        sdk.Coin{sdk.DefaultBondDenom, sdkmath.ZeroInt()},
		TaxRate:                sdkmath.LegacyNewDec(0),
		MaxStandardCoinPerPool: sdkmath.NewInt(10_000_000_000),
		MaxSwapAmount:          sdk.NewCoins(sdk.NewInt64Coin(denomBTC, 10_000_000)),
	}
	suite.app.CoinswapKeeper.SetParams(suite.ctx, params)

	// Test add liquidity with non-whitelisted denom
	// Fail to create a pool
	ethAmt, _ := sdkmath.NewIntFromString("100")
	standardAmt, _ := sdkmath.NewIntFromString("1000000000")
	depositCoin := sdk.NewCoin(denomETH, ethAmt)
	minReward := sdkmath.NewInt(1)
	deadline := time.Now().Add(1 * time.Minute)

	msg := types.NewMsgAddLiquidity(depositCoin, standardAmt, minReward, deadline.Unix(), addrSender1.String())
	_, err := suite.app.CoinswapKeeper.AddLiquidity(suite.ctx, msg)
	suite.Error(err)

	// test add liquidity with exceeding standard coin limit
	// Expected behavior: fails to create to pool
	btcAmt, _ := sdkmath.NewIntFromString("100")
	standardAmt, _ = sdkmath.NewIntFromString("15000000000")
	depositCoin = sdk.NewCoin(denomBTC, btcAmt)

	msg = types.NewMsgAddLiquidity(depositCoin, standardAmt, minReward, deadline.Unix(), addrSender1.String())
	_, err = suite.app.CoinswapKeeper.AddLiquidity(suite.ctx, msg)
	suite.Error(err)

	// Test add liquidity
	// Deposit: 100btc, 8000000000stake
	// Pool created and mint 8000000000lpt-1
	// Expected pool balance: 100btc, 8000000000stake
	standardAmt, _ = sdkmath.NewIntFromString("8000000000")
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
		sdk.NewCoin(denomStandard, sdkmath.NewIntWithDecimal(8000, 6)),
	)
	suite.Equal(expCoins.Sort().String(), reservePoolBalances.Sort().String())

	expCoins = sdk.NewCoins(
		sdk.NewInt64Coin(denomBTC, 2999999900),
		sdk.NewCoin(denomStandard, sdkmath.NewIntWithDecimal(3, 19).Sub(sdkmath.NewIntWithDecimal(8000, 6))),
		sdk.NewCoin(lptDenom, sdkmath.NewIntWithDecimal(8000, 6)),
	)
	suite.Equal(expCoins.Sort().String(), sender1Balances.Sort().String())

	// test add liquidity (pool exists)
	// Deposit try: 200btc, 8000000000stake
	// Actual deposit: 26btc, 2000000000stake
	// Mint: 2000000000lpt-1
	// Expected pool balance: 126btc, 10000000000stake
	expLptDenom, _ := suite.app.CoinswapKeeper.GetLptDenomFromDenoms(suite.ctx, denomBTC, denomStandard)
	suite.Require().Equal(expLptDenom, lptDenom)

	btcAmt, _ = sdkmath.NewIntFromString("200")
	standardAmt, _ = sdkmath.NewIntFromString("8000000000")
	depositCoin = sdk.NewCoin(denomBTC, btcAmt)
	minReward = sdkmath.NewInt(1)
	deadline = time.Now().Add(1 * time.Minute)

	msg = types.NewMsgAddLiquidity(depositCoin, standardAmt, minReward, deadline.Unix(), addrSender2.String())
	_, err = suite.app.CoinswapKeeper.AddLiquidity(suite.ctx, msg)
	suite.NoError(err)

	reservePoolBalances = suite.app.BankKeeper.GetAllBalances(suite.ctx, poolAddr)
	sender2Balances := suite.app.BankKeeper.GetAllBalances(suite.ctx, addrSender2)
	suite.Equal("10000000000", suite.app.BankKeeper.GetSupply(suite.ctx, lptDenom).Amount.String())

	expCoins = sdk.NewCoins(
		sdk.NewInt64Coin(denomBTC, 126),
		sdk.NewCoin(denomStandard, sdkmath.NewIntWithDecimal(10000, 6)),
	)
	suite.Equal(expCoins.Sort().String(), reservePoolBalances.Sort().String())

	expCoins = sdk.NewCoins(
		sdk.NewInt64Coin(denomBTC, 2999999974),
		sdk.NewCoin(denomStandard, sdkmath.NewIntWithDecimal(3, 19).Sub(sdkmath.NewIntWithDecimal(2000, 6))),
		sdk.NewCoin(lptDenom, sdkmath.NewIntWithDecimal(2000, 6)),
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
	withdraw, _ := sdkmath.NewIntFromString("8000000000")
	msgRemove := types.NewMsgRemoveLiquidity(
		sdkmath.NewInt(1),
		sdk.NewCoin(lptDenom, withdraw),
		sdkmath.NewInt(1),
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
		sdk.NewCoin(denomStandard, sdkmath.NewIntWithDecimal(3, 19)),
	)
	suite.Equal(expCoins.Sort().String(), sender1Balances.Sort().String())

	expCoins = sdk.NewCoins(
		sdk.NewInt64Coin(denomBTC, 26),
		sdk.NewCoin(denomStandard, sdkmath.NewIntWithDecimal(2000, 6)),
	)
	suite.Equal(expCoins.Sort().String(), reservePoolBalances.String())

	// Test remove liquidity (overdraft)
	// Expected behavior: fails to withdraw
	withdraw = sdkmath.NewIntWithDecimal(8000, 6)
	msgRemove = types.NewMsgRemoveLiquidity(
		sdkmath.NewInt(1),
		sdk.NewCoin(lptDenom, withdraw),
		sdkmath.NewInt(1),
		suite.ctx.BlockHeader().Time.Unix(),
		addrSender2.String(),
	)

	_, err = suite.app.CoinswapKeeper.RemoveLiquidity(suite.ctx, msgRemove)
	suite.Error(err)

	// Test remove liquidity (remove all)
	// Expected pool coin supply: 0
	// Expected reserve balance: 0btc, 0stake
	withdraw = sdkmath.NewIntWithDecimal(2000, 6)
	msgRemove = types.NewMsgRemoveLiquidity(
		sdkmath.NewInt(1),
		sdk.NewCoin(lptDenom, withdraw),
		sdkmath.NewInt(1),
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
		sdk.NewCoin(denomStandard, sdkmath.NewIntWithDecimal(3, 19)),
	)
	suite.Equal(expCoins.Sort().String(), sender2Balances.Sort().String())
	suite.Equal("", reservePoolBalances.String())

	// Test add liquidity (pool exists but empty)
	// Deposit: 200btc, 8000000000stake
	// Pool created and mint 8000000000lpt-1
	// Expected pool balance: 200btc, 8000000000stake
	standardAmt, _ = sdkmath.NewIntFromString("8000000000")
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
		sdk.NewCoin(denomStandard, sdkmath.NewIntWithDecimal(8000, 6)),
	)
	suite.Equal(expCoins.Sort().String(), reservePoolBalances.Sort().String())
}
