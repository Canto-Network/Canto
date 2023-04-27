package keeper_test

import (
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/evmos/ethermint/crypto/ethsecp256k1"
	"github.com/stretchr/testify/mock"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"

	transfertypes "github.com/cosmos/ibc-go/v3/modules/apps/transfer/types"
	clienttypes "github.com/cosmos/ibc-go/v3/modules/core/02-client/types"
	channeltypes "github.com/cosmos/ibc-go/v3/modules/core/04-channel/types"
	ibcgotesting "github.com/cosmos/ibc-go/v3/testing"
	ibcmock "github.com/cosmos/ibc-go/v3/testing/mock"

	"github.com/Canto-Network/Canto/v6/contracts"
	"github.com/Canto-Network/Canto/v6/testutil"
	coinswaptypes "github.com/Canto-Network/Canto/v6/x/coinswap/types"
	erc20types "github.com/Canto-Network/Canto/v6/x/erc20/types"
	inflationtypes "github.com/Canto-Network/Canto/v6/x/inflation/types"
	"github.com/Canto-Network/Canto/v6/x/onboarding/keeper"
	onboardingtest "github.com/Canto-Network/Canto/v6/x/onboarding/testutil"
	"github.com/Canto-Network/Canto/v6/x/onboarding/types"
	vestingtypes "github.com/Canto-Network/Canto/v6/x/vesting/types"
)

var (
	metadataIbcUSDC = banktypes.Metadata{
		Description: "USDC IBC voucher (channel 0)",
		Base:        uusdcIbcdenom,
		// NOTE: Denom units MUST be increasing
		DenomUnits: []*banktypes.DenomUnit{
			{
				Denom:    uusdcIbcdenom,
				Exponent: 0,
			},
		},
		Name:    "USDC channel-0",
		Symbol:  "ibcUSDC-0",
		Display: uusdcIbcdenom,
	}

	metadataIbcUSDT = banktypes.Metadata{
		Description: "USDT IBC voucher (channel 0)",
		Base:        uusdtIbcdenom,
		// NOTE: Denom units MUST be increasing
		DenomUnits: []*banktypes.DenomUnit{
			{
				Denom:    uusdtIbcdenom,
				Exponent: 0,
			},
		},
		Name:    "USDT channel-0",
		Symbol:  "ibcUSDT-0",
		Display: uusdtIbcdenom,
	}
)

// setupRegisterCoin is a helper function for registering a new ERC20 token pair using the Erc20Keeper.
func (suite *KeeperTestSuite) setupRegisterCoin(metadata banktypes.Metadata) *erc20types.TokenPair {
	pair, err := suite.app.Erc20Keeper.RegisterCoin(suite.ctx, metadata)
	suite.Require().NoError(err)
	return pair
}

func (suite *KeeperTestSuite) TestOnRecvPacket() {
	// secp256k1 account
	secpPk := secp256k1.GenPrivKey()
	secpAddr := sdk.AccAddress(secpPk.PubKey().Address())
	secpAddrCosmos := sdk.MustBech32ifyAddressBytes(sdk.Bech32MainPrefix, secpAddr)

	// ethsecp256k1 account
	ethPk, err := ethsecp256k1.GenerateKey()
	suite.Require().Nil(err)
	ethsecpAddr := sdk.AccAddress(ethPk.PubKey().Address())
	ethsecpAddrcanto := sdk.AccAddress(ethPk.PubKey().Address()).String()

	// Setup Cosmos <=> canto IBC relayer
	denom := "uUSDC"
	ibcDenom := uusdcIbcdenom
	transferAmount := sdk.NewIntWithDecimal(25, 6)
	sourceChannel := "channel-0"
	cantoChannel := "channel-0"
	path := fmt.Sprintf("%s/%s", transfertypes.PortID, cantoChannel)

	timeoutHeight := clienttypes.NewHeight(0, 100)
	disabledTimeoutTimestamp := uint64(0)
	mockPacket := channeltypes.NewPacket(ibcgotesting.MockPacketData, 1, transfertypes.PortID, sourceChannel, transfertypes.PortID, cantoChannel, timeoutHeight, disabledTimeoutTimestamp)
	packet := mockPacket
	expAck := ibcmock.MockAcknowledgement

	testCases := []struct {
		name              string
		malleate          func()
		ackSuccess        bool
		receiverBalance   sdk.Coins
		expCantoBalance   sdk.Coin
		expVoucherBalance sdk.Coin
		expErc20Balance   sdk.Int
	}{
		{
			"fail - invalid sender - missing '1' ",
			func() {
				transfer := transfertypes.NewFungibleTokenPacketData(denom, "100", "canto", ethsecpAddrcanto)
				bz := transfertypes.ModuleCdc.MustMarshalJSON(&transfer)
				packet = channeltypes.NewPacket(bz, 100, transfertypes.PortID, sourceChannel, transfertypes.PortID, cantoChannel, timeoutHeight, 0)
			},
			false,
			sdk.NewCoins(sdk.NewCoin("acanto", sdk.ZeroInt())),
			sdk.NewCoin("acanto", sdk.ZeroInt()),
			sdk.NewCoin(uusdcIbcdenom, transferAmount),
			sdk.ZeroInt(),
		},
		{
			"fail - invalid sender - invalid bech32",
			func() {
				transfer := transfertypes.NewFungibleTokenPacketData(denom, "100", "badba1sv9m0g7ycejwr3s369km58h5qe7xj77hvcxrms", ethsecpAddrcanto)
				bz := transfertypes.ModuleCdc.MustMarshalJSON(&transfer)
				packet = channeltypes.NewPacket(bz, 100, transfertypes.PortID, sourceChannel, transfertypes.PortID, cantoChannel, timeoutHeight, 0)
			},
			false,
			sdk.NewCoins(sdk.NewCoin("acanto", sdk.ZeroInt())),
			sdk.NewCoin("acanto", sdk.ZeroInt()),
			sdk.NewCoin(uusdcIbcdenom, transferAmount),
			sdk.ZeroInt(),
		},
		{
			"continue - receiver is a module account",
			func() {
				distrAcc := suite.app.AccountKeeper.GetModuleAccount(suite.ctx, distrtypes.ModuleName)
				suite.Require().NotNil(distrAcc)
				addr := distrAcc.GetAddress().String()
				transfer := transfertypes.NewFungibleTokenPacketData(denom, "100", addr, addr)
				bz := transfertypes.ModuleCdc.MustMarshalJSON(&transfer)
				packet = channeltypes.NewPacket(bz, 100, transfertypes.PortID, sourceChannel, transfertypes.PortID, cantoChannel, timeoutHeight, 0)
			},
			true,
			sdk.NewCoins(sdk.NewCoin("acanto", sdk.ZeroInt())),
			sdk.NewCoin("acanto", sdk.ZeroInt()),
			sdk.NewCoin(uusdcIbcdenom, transferAmount),
			sdk.ZeroInt(),
		},
		{
			"continue - params disabled",
			func() {
				params := suite.app.OnboardingKeeper.GetParams(suite.ctx)
				params.EnableOnboarding = false
				suite.app.OnboardingKeeper.SetParams(suite.ctx, params)
			},
			true,
			sdk.NewCoins(sdk.NewCoin("acanto", sdk.ZeroInt())),
			sdk.NewCoin("acanto", sdk.ZeroInt()),
			sdk.NewCoin(uusdcIbcdenom, transferAmount),
			sdk.ZeroInt(),
		},
		{
			"swap fail / convert all transferred amount - no liquidity pool exists",
			func() {

				denom = "uUSDT"
				ibcDenom = uusdtIbcdenom

				transfer := transfertypes.NewFungibleTokenPacketData(denom, transferAmount.String(), secpAddrCosmos, ethsecpAddrcanto)
				bz := transfertypes.ModuleCdc.MustMarshalJSON(&transfer)
				packet = channeltypes.NewPacket(bz, 100, transfertypes.PortID, sourceChannel, transfertypes.PortID, cantoChannel, timeoutHeight, 0)
			},
			true,
			sdk.NewCoins(sdk.NewCoin("acanto", sdk.ZeroInt())),
			sdk.NewCoin("acanto", sdk.ZeroInt()),
			sdk.NewCoin(uusdtIbcdenom, sdk.ZeroInt()),
			transferAmount,
		},
		{
			"no swap / convert all transferred amount - no liquidity pool exists but acanto balance is already bigger than threshold",
			func() {

				denom = "uUSDT"
				ibcDenom = uusdtIbcdenom

				transfer := transfertypes.NewFungibleTokenPacketData(denom, transferAmount.String(), secpAddrCosmos, ethsecpAddrcanto)
				bz := transfertypes.ModuleCdc.MustMarshalJSON(&transfer)
				packet = channeltypes.NewPacket(bz, 100, transfertypes.PortID, sourceChannel, transfertypes.PortID, cantoChannel, timeoutHeight, 0)
			},
			true,
			sdk.NewCoins(sdk.NewCoin("acanto", sdk.NewIntWithDecimal(4, 18))),
			sdk.NewCoin("acanto", sdk.NewIntWithDecimal(4, 18)),
			sdk.NewCoin(uusdtIbcdenom, sdk.ZeroInt()),
			transferAmount,
		},
		{
			"swap fail / convert all transferred amount - not enough ibc coin to swap threshold",
			func() {
				denom = "uUSDC"
				ibcDenom = uusdcIbcdenom

				transferAmount = sdk.NewIntWithDecimal(1, 6)
				transfer := transfertypes.NewFungibleTokenPacketData(denom, transferAmount.String(), secpAddrCosmos, ethsecpAddrcanto)
				bz := transfertypes.ModuleCdc.MustMarshalJSON(&transfer)
				packet = channeltypes.NewPacket(bz, 100, transfertypes.PortID, sourceChannel, transfertypes.PortID, cantoChannel, timeoutHeight, 0)
			},
			true,
			sdk.NewCoins(sdk.NewCoin("acanto", sdk.ZeroInt())),
			sdk.NewCoin("acanto", sdk.ZeroInt()),
			sdk.NewCoin(uusdcIbcdenom, sdk.ZeroInt()),
			sdk.NewIntWithDecimal(1, 6),
		},
		{
			"no swap / convert all transferred amount - acanto balance is already bigger than threshold",
			func() {
				transferAmount = sdk.NewIntWithDecimal(25, 6)
				transfer := transfertypes.NewFungibleTokenPacketData(denom, transferAmount.String(), secpAddrCosmos, ethsecpAddrcanto)
				bz := transfertypes.ModuleCdc.MustMarshalJSON(&transfer)
				packet = channeltypes.NewPacket(bz, 100, transfertypes.PortID, sourceChannel, transfertypes.PortID, cantoChannel, timeoutHeight, 0)
			},
			true,
			sdk.NewCoins(sdk.NewCoin("acanto", sdk.NewIntWithDecimal(4, 18))),
			sdk.NewCoin("acanto", sdk.NewIntWithDecimal(4, 18)),
			sdk.NewCoin(uusdcIbcdenom, sdk.ZeroInt()),
			transferAmount,
		},
		{
			"no swap / no convert - unauthorized  channel",
			func() {
				cantoChannel = "channel-100"
				transferAmount = sdk.NewIntWithDecimal(25, 6)
				transfer := transfertypes.NewFungibleTokenPacketData(denom, transferAmount.String(), secpAddrCosmos, ethsecpAddrcanto)
				bz := transfertypes.ModuleCdc.MustMarshalJSON(&transfer)
				packet = channeltypes.NewPacket(bz, 100, transfertypes.PortID, sourceChannel, transfertypes.PortID, cantoChannel, timeoutHeight, 0)
			},
			true,
			sdk.NewCoins(sdk.NewCoin("acanto", sdk.ZeroInt())),
			sdk.NewCoin("acanto", sdk.ZeroInt()),
			sdk.NewCoin(uusdcIbcdenom, sdk.NewIntWithDecimal(25, 6)),
			sdk.ZeroInt(),
		},
		{
			"swap / convert remaining ibc token - swap and erc20 conversion are successful",
			func() {
				cantoChannel = "channel-0"
				transferAmount = sdk.NewIntWithDecimal(25, 6)
				transfer := transfertypes.NewFungibleTokenPacketData(denom, transferAmount.String(), secpAddrCosmos, ethsecpAddrcanto)
				bz := transfertypes.ModuleCdc.MustMarshalJSON(&transfer)
				packet = channeltypes.NewPacket(bz, 100, transfertypes.PortID, sourceChannel, transfertypes.PortID, cantoChannel, timeoutHeight, 0)
			},
			true,
			sdk.NewCoins(sdk.NewCoin("acanto", sdk.ZeroInt())),
			sdk.NewCoin("acanto", sdk.NewIntWithDecimal(4, 18)),
			sdk.NewCoin(uusdcIbcdenom, sdk.ZeroInt()),
			sdk.NewInt(20998399),
		},
		{
			"swap / convert remaining ibc token - swap and erc20 conversion are successful (acanto balance is positive but less than threshold)",
			func() {
				transferAmount = sdk.NewIntWithDecimal(25, 6)
				transfer := transfertypes.NewFungibleTokenPacketData(denom, transferAmount.String(), secpAddrCosmos, ethsecpAddrcanto)
				bz := transfertypes.ModuleCdc.MustMarshalJSON(&transfer)
				packet = channeltypes.NewPacket(bz, 100, transfertypes.PortID, sourceChannel, transfertypes.PortID, cantoChannel, timeoutHeight, 0)
			},
			true,
			sdk.NewCoins(sdk.NewCoin("acanto", sdk.NewIntWithDecimal(3, 18))),
			sdk.NewCoin("acanto", sdk.NewIntWithDecimal(7, 18)),
			sdk.NewCoin(uusdcIbcdenom, sdk.ZeroInt()),
			sdk.NewInt(20998399),
		},
		{
			"swap / convert remaining ibc token - swap and erc20 conversion are successful (ibc token balance is bigger than 0)",
			func() {
				cantoChannel = "channel-0"
				transferAmount = sdk.NewIntWithDecimal(25, 6)
				transfer := transfertypes.NewFungibleTokenPacketData(denom, transferAmount.String(), secpAddrCosmos, ethsecpAddrcanto)
				bz := transfertypes.ModuleCdc.MustMarshalJSON(&transfer)
				packet = channeltypes.NewPacket(bz, 100, transfertypes.PortID, sourceChannel, transfertypes.PortID, cantoChannel, timeoutHeight, 0)
			},
			true,
			sdk.NewCoins(sdk.NewCoin("acanto", sdk.ZeroInt()), sdk.NewCoin(uusdcIbcdenom, sdk.NewIntWithDecimal(1, 6))),
			sdk.NewCoin("acanto", sdk.NewIntWithDecimal(4, 18)),
			sdk.NewCoin(uusdcIbcdenom, sdk.NewIntWithDecimal(1, 6)),
			sdk.NewInt(20998399),
		},
		{
			"swap / convert remaining ibc token - receiver is a vesting account",
			func() {
				// Set vesting account
				bacc := authtypes.NewBaseAccount(secpAddr, nil, 0, 0)
				acc := vestingtypes.NewClawbackVestingAccount(bacc, secpAddr, nil, suite.ctx.BlockTime(), nil, nil)

				suite.app.AccountKeeper.SetAccount(suite.ctx, acc)

				transferAmount = sdk.NewIntWithDecimal(25, 6)
				transfer := transfertypes.NewFungibleTokenPacketData(denom, transferAmount.String(), secpAddrCosmos, ethsecpAddrcanto)
				bz := transfertypes.ModuleCdc.MustMarshalJSON(&transfer)
				packet = channeltypes.NewPacket(bz, 100, transfertypes.PortID, sourceChannel, transfertypes.PortID, cantoChannel, timeoutHeight, 0)
			},
			true,
			sdk.NewCoins(sdk.NewCoin("acanto", sdk.ZeroInt()), sdk.NewCoin(uusdcIbcdenom, sdk.NewIntWithDecimal(1, 6))),
			sdk.NewCoin("acanto", sdk.NewIntWithDecimal(4, 18)),
			sdk.NewCoin(uusdcIbcdenom, sdk.NewIntWithDecimal(1, 6)),
			sdk.NewInt(20998399),
		},
		{
			"swap / convert remaining ibc token - swap and erc20 conversion are successful (acanto and ibc token balance is bigger than 0)",
			func() {
				transferAmount = sdk.NewIntWithDecimal(25, 6)
				transfer := transfertypes.NewFungibleTokenPacketData(denom, transferAmount.String(), secpAddrCosmos, ethsecpAddrcanto)
				bz := transfertypes.ModuleCdc.MustMarshalJSON(&transfer)
				packet = channeltypes.NewPacket(bz, 100, transfertypes.PortID, sourceChannel, transfertypes.PortID, cantoChannel, timeoutHeight, 0)
			},
			true,
			sdk.NewCoins(sdk.NewCoin("acanto", sdk.NewIntWithDecimal(3, 18)), sdk.NewCoin(uusdcIbcdenom, sdk.NewIntWithDecimal(1, 6))),
			sdk.NewCoin("acanto", sdk.NewIntWithDecimal(7, 18)),
			sdk.NewCoin(uusdcIbcdenom, sdk.NewIntWithDecimal(1, 6)),
			sdk.NewInt(20998399),
		},
		{
			"swap fail / convert all transferred amount - required ibc token to swap exceeds max swap amount limit",
			func() {
				coinswapParams := suite.app.CoinswapKeeper.GetParams(suite.ctx)
				coinswapParams.MaxSwapAmount = sdk.NewCoins(sdk.NewCoin(uusdcIbcdenom, sdk.NewIntWithDecimal(5, 5)))
				suite.app.CoinswapKeeper.SetParams(suite.ctx, coinswapParams)

				transferAmount = sdk.NewIntWithDecimal(25, 6)
				transfer := transfertypes.NewFungibleTokenPacketData(denom, transferAmount.String(), secpAddrCosmos, ethsecpAddrcanto)
				bz := transfertypes.ModuleCdc.MustMarshalJSON(&transfer)
				packet = channeltypes.NewPacket(bz, 100, transfertypes.PortID, sourceChannel, transfertypes.PortID, cantoChannel, timeoutHeight, 0)
			},
			true,
			sdk.NewCoins(sdk.NewCoin("acanto", sdk.NewIntWithDecimal(3, 18))),
			sdk.NewCoin("acanto", sdk.NewIntWithDecimal(3, 18)),
			sdk.NewCoin(uusdcIbcdenom, sdk.ZeroInt()),
			transferAmount,
		},
		{
			"convert fail",
			func() {
				cantoChannel = "channel-0"
				transferAmount = sdk.NewIntWithDecimal(25, 6)
				transfer := transfertypes.NewFungibleTokenPacketData(denom, transferAmount.String(), secpAddrCosmos, ethsecpAddrcanto)
				bz := transfertypes.ModuleCdc.MustMarshalJSON(&transfer)
				packet = channeltypes.NewPacket(bz, 100, transfertypes.PortID, sourceChannel, transfertypes.PortID, cantoChannel, timeoutHeight, 0)

				pairID := suite.app.Erc20Keeper.GetTokenPairID(suite.ctx, metadataIbcUSDC.Base)
				pair, _ := suite.app.Erc20Keeper.GetTokenPair(suite.ctx, pairID)
				pair.ContractOwner = erc20types.OWNER_UNSPECIFIED
				suite.app.Erc20Keeper.SetTokenPair(suite.ctx, pair)

			},
			true,
			sdk.NewCoins(sdk.NewCoin("acanto", sdk.ZeroInt())),
			sdk.NewCoin("acanto", sdk.NewIntWithDecimal(4, 18)),
			sdk.NewCoin(uusdcIbcdenom, sdk.NewInt(20998399)),
			sdk.NewInt(0),
		},
	}
	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.SetupTest() // reset

			// Setup liquidity pool (acanto/uUSDC)
			suite.app.CoinswapKeeper.SetStandardDenom(suite.ctx, "acanto")
			coinswapParams := suite.app.CoinswapKeeper.GetParams(suite.ctx)
			coinswapParams.MaxSwapAmount = sdk.NewCoins(sdk.NewCoin(uusdcIbcdenom, sdk.NewIntWithDecimal(10, 6)))
			suite.app.CoinswapKeeper.SetParams(suite.ctx, coinswapParams)

			testutil.FundAccount(suite.app.BankKeeper, suite.ctx, secpAddr, sdk.NewCoins(sdk.NewCoin("acanto", sdk.NewIntWithDecimal(10000, 18)), sdk.NewCoin(uusdcIbcdenom, sdk.NewIntWithDecimal(10000, 6))))

			msgAddLiquidity := coinswaptypes.MsgAddLiquidity{
				MaxToken:         sdk.NewCoin(uusdcIbcdenom, sdk.NewIntWithDecimal(10000, 6)),
				ExactStandardAmt: sdk.NewIntWithDecimal(10000, 18),
				MinLiquidity:     sdk.NewInt(1),
				Deadline:         time.Now().Add(time.Minute * 10).Unix(),
				Sender:           secpAddr.String(),
			}

			suite.app.CoinswapKeeper.AddLiquidity(suite.ctx, &msgAddLiquidity)

			// Enable Onboarding
			params := suite.app.OnboardingKeeper.GetParams(suite.ctx)
			params.EnableOnboarding = true
			params.WhitelistedChannels = []string{"channel-0"}
			suite.app.OnboardingKeeper.SetParams(suite.ctx, params)

			// Deploy ERC20 Contract
			err := suite.app.BankKeeper.MintCoins(suite.ctx, inflationtypes.ModuleName, sdk.Coins{sdk.NewInt64Coin(metadataIbcUSDC.Base, 1)})
			suite.Require().NoError(err)
			usdcPair := suite.setupRegisterCoin(metadataIbcUSDC)
			suite.Require().NotNil(usdcPair)
			suite.app.Erc20Keeper.SetTokenPair(suite.ctx, *usdcPair)

			err = suite.app.BankKeeper.MintCoins(suite.ctx, inflationtypes.ModuleName, sdk.Coins{sdk.NewInt64Coin(metadataIbcUSDT.Base, 1)})
			suite.Require().NoError(err)
			usdtPair := suite.setupRegisterCoin(metadataIbcUSDT)
			suite.Require().NotNil(usdtPair)
			suite.app.Erc20Keeper.SetTokenPair(suite.ctx, *usdtPair)

			tc.malleate()

			// Set Denom Trace
			denomTrace := transfertypes.DenomTrace{
				Path:      path,
				BaseDenom: denom,
			}
			suite.app.TransferKeeper.SetDenomTrace(suite.ctx, denomTrace)

			// Set Cosmos Channel
			channel := channeltypes.Channel{
				State:          channeltypes.INIT,
				Ordering:       channeltypes.UNORDERED,
				Counterparty:   channeltypes.NewCounterparty(transfertypes.PortID, sourceChannel),
				ConnectionHops: []string{sourceChannel},
			}
			suite.app.IBCKeeper.ChannelKeeper.SetChannel(suite.ctx, transfertypes.PortID, cantoChannel, channel)

			// Set Next Sequence Send
			suite.app.IBCKeeper.ChannelKeeper.SetNextSequenceSend(suite.ctx, transfertypes.PortID, cantoChannel, 1)

			// Mock the Transferkeeper to always return nil on SendTransfer(), as this
			// method requires a successfull handshake with the counterparty chain.
			// This, however, exceeds the requirements of the unit tests.
			mockTransferKeeper := &MockTransferKeeper{
				Keeper: suite.app.BankKeeper,
			}

			mockTransferKeeper.On("GetDenomTrace", mock.Anything, mock.Anything).Return(denomTrace, true)
			mockTransferKeeper.On("SendTransfer", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

			sp, found := suite.app.ParamsKeeper.GetSubspace(types.ModuleName)
			suite.Require().True(found)
			suite.app.OnboardingKeeper = keeper.NewKeeper(sp, suite.app.AccountKeeper, suite.app.BankKeeper, suite.app.IBCKeeper.ChannelKeeper, mockTransferKeeper, suite.app.CoinswapKeeper, suite.app.Erc20Keeper)

			// Fund receiver account with canto, IBC vouchers
			testutil.FundAccount(suite.app.BankKeeper, suite.ctx, ethsecpAddr, tc.receiverBalance)
			// Fund receiver account with the transferred amount
			testutil.FundAccount(suite.app.BankKeeper, suite.ctx, ethsecpAddr, sdk.NewCoins(sdk.NewCoin(ibcDenom, transferAmount)))

			// Perform IBC callback
			ack := suite.app.OnboardingKeeper.OnRecvPacket(suite.ctx, packet, expAck)

			// Check acknowledgement
			if tc.ackSuccess {
				suite.Require().True(ack.Success(), string(ack.Acknowledgement()))
				suite.Require().Equal(expAck, ack)
			} else {
				suite.Require().False(ack.Success(), string(ack.Acknowledgement()))
			}

			// Check balances
			cantoBalance := suite.app.BankKeeper.GetBalance(suite.ctx, ethsecpAddr, "acanto")
			voucherBalance := suite.app.BankKeeper.GetBalance(suite.ctx, ethsecpAddr, ibcDenom)
			erc20balance := big.NewInt(0)

			if ibcDenom == uusdcIbcdenom {
				erc20balance = suite.app.Erc20Keeper.BalanceOf(suite.ctx, contracts.ERC20MinterBurnerDecimalsContract.ABI, usdcPair.GetERC20Contract(), common.BytesToAddress(ethsecpAddr.Bytes()))
			} else {
				erc20balance = suite.app.Erc20Keeper.BalanceOf(suite.ctx, contracts.ERC20MinterBurnerDecimalsContract.ABI, usdtPair.GetERC20Contract(), common.BytesToAddress(ethsecpAddr.Bytes()))
			}

			suite.Require().Equal(tc.expCantoBalance, cantoBalance)
			suite.Require().Equal(tc.expVoucherBalance, voucherBalance)
			suite.Require().Equal(tc.expErc20Balance.String(), erc20balance.String())

			events := suite.ctx.EventManager().Events()

			attrs := onboardingtest.ExtractAttributes(onboardingtest.FindEvent(events, "swap"))

			swappedAmount, ok := sdk.NewIntFromString(attrs["amount"])
			if !ok {
				swappedAmount = sdk.ZeroInt()
			}

			attrs = onboardingtest.ExtractAttributes(onboardingtest.FindEvent(events, "convert_coin"))

			if tc.expErc20Balance.IsPositive() {
				// Check that the amount of ERC20 tokens minted is equal to the difference between
				// the transferred amount and the swapped amount
				suite.Require().Equal(tc.expErc20Balance.String(), transferAmount.Sub(swappedAmount).String())
				suite.Require().Equal(tc.expErc20Balance.String(), attrs["amount"])
			} else {
				suite.Require().Equal(0, len(attrs))
			}
		})
	}
}
