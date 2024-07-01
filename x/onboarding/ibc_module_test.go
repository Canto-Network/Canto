package onboarding_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/ethereum/go-ethereum/common"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	ibcgotesting "github.com/cosmos/ibc-go/v8/testing"

	"github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
	clienttypes "github.com/cosmos/ibc-go/v8/modules/core/02-client/types"

	"github.com/Canto-Network/Canto/v7/app"
	"github.com/Canto-Network/Canto/v7/contracts"
	ibctesting "github.com/Canto-Network/Canto/v7/ibc/testing"
	coinswaptypes "github.com/Canto-Network/Canto/v7/x/coinswap/types"
	inflationtypes "github.com/Canto-Network/Canto/v7/x/inflation/types"
	onboardingtest "github.com/Canto-Network/Canto/v7/x/onboarding/testutil"
)

var (
	ibcBase     = "ibc/C053D637CCA2A2BA030E2C5EE1B28A16F71CCB0E45E8BE52766DC1B241B77878"
	metadataIbc = banktypes.Metadata{
		Description: "IBC voucher (channel 0)",
		Base:        ibcBase,
		// NOTE: Denom units MUST be increasing
		DenomUnits: []*banktypes.DenomUnit{
			{
				Denom:    ibcBase,
				Exponent: 0,
			},
		},
		Name:    "Ibc Token channel-0",
		Symbol:  "ibcToken-0",
		Display: ibcBase,
	}
)

type TransferTestSuite struct {
	suite.Suite

	coordinator *ibctesting.Coordinator

	// testing chains used for convenience and readability
	chainA *ibctesting.TestChain
	chainB *ibctesting.TestChain
}

func (suite *TransferTestSuite) SetupTest() {
	suite.coordinator = ibctesting.NewCoordinator(suite.T(), 1, 1)
	suite.chainA = suite.coordinator.GetChain(ibctesting.GetChainID(2))
	suite.chainB = suite.coordinator.GetChain(ibctesting.GetChainIDCanto(1))
}

func NewTransferPath(chainA, chainB *ibctesting.TestChain) *ibctesting.Path {
	path := ibctesting.NewPath(chainA, chainB)
	path.EndpointA.ChannelConfig.PortID = ibctesting.TransferPort
	path.EndpointB.ChannelConfig.PortID = ibctesting.TransferPort
	path.EndpointA.ChannelConfig.Version = types.Version
	path.EndpointB.ChannelConfig.Version = types.Version

	return path
}

// constructs a send from chainA to chainB on the established channel/connection
// and sends the same coin again from chainA to chainB.
func (suite *TransferTestSuite) TestHandleMsgTransfer() {
	// setup path between chainA and chainB
	path := NewTransferPath(suite.chainA, suite.chainB)
	suite.coordinator.Setup(path)

	// Fund chainB
	coins := sdk.NewCoins(
		sdk.NewCoin(ibcBase, sdkmath.NewInt(1000000000000)),
		sdk.NewCoin("acanto", sdkmath.NewInt(10000000000)),
	)
	err := suite.chainB.App.(*app.Canto).BankKeeper.MintCoins(suite.chainB.GetContext(), inflationtypes.ModuleName, coins)
	suite.Require().NoError(err)
	err = suite.chainB.App.(*app.Canto).BankKeeper.SendCoinsFromModuleToAccount(suite.chainB.GetContext(), inflationtypes.ModuleName, suite.chainB.SenderAccount.GetAddress(), coins)
	suite.Require().NoError(err)

	// Setup params for coinswap
	coinswapKeeper := suite.chainB.App.(*app.Canto).GetCoinswapKeeper()
	coinswapKeeper.SetStandardDenom(suite.chainB.GetContext(), "acanto")
	params := coinswapKeeper.GetParams(suite.chainB.GetContext())

	params.MaxSwapAmount = sdk.NewCoins(sdk.NewCoin(ibcBase, sdkmath.NewInt(10000000)))
	coinswapKeeper.SetParams(suite.chainB.GetContext(), params)

	middlewareParams := suite.chainB.App.(*app.Canto).GetOnboardingKeeper().GetParams(suite.chainB.GetContext())
	middlewareParams.AutoSwapThreshold = sdkmath.NewInt(4000000)
	suite.chainB.App.(*app.Canto).GetOnboardingKeeper().SetParams(suite.chainB.GetContext(), middlewareParams)

	erc20Keeper := suite.chainB.App.(*app.Canto).GetErc20Keeper()
	pair, err := erc20Keeper.RegisterCoin(suite.chainB.GetContext(), metadataIbc)
	suite.Require().NoError(err)

	// Pool creation
	msgAddLiquidity := coinswaptypes.MsgAddLiquidity{
		MaxToken:         sdk.NewCoin(ibcBase, sdkmath.NewInt(10000000000)),
		ExactStandardAmt: sdkmath.NewInt(10000000000),
		MinLiquidity:     sdkmath.NewInt(1),
		Deadline:         time.Now().Add(time.Minute * 10).Unix(),
		Sender:           suite.chainB.SenderAccount.GetAddress().String(),
	}

	_, err = coinswapKeeper.AddLiquidity(suite.chainB.GetContext(), &msgAddLiquidity)
	suite.Require().NoError(err)

	timeoutHeight := clienttypes.NewHeight(10, 100)

	amount, ok := sdkmath.NewIntFromString("9223372036854775808") // 2^63 (one above int64)
	suite.Require().True(ok)
	coinToSendToB := sdk.NewCoin(sdk.DefaultBondDenom, amount)

	// send coins from chainA to chainB
	// auto swap and auto convert should happen
	msg := types.NewMsgTransfer(path.EndpointA.ChannelConfig.PortID, path.EndpointA.ChannelID, coinToSendToB, suite.chainA.SenderAccount.GetAddress().String(), suite.chainB.SenderAccount.GetAddress().String(), timeoutHeight, 0, "")
	res, err := suite.chainA.SendMsgs(msg)
	suite.Require().NoError(err) // message committed

	packet, err := ibctesting.ParsePacketFromEvents(res.GetEvents())
	suite.Require().NoError(err)

	voucherDenomTrace := types.ParseDenomTrace(types.GetPrefixedDenom(packet.GetDestPort(), packet.GetDestChannel(), sdk.DefaultBondDenom))

	// check balances on chainB before the IBC transfer
	balanceVoucherBefore := suite.chainB.App.(*app.Canto).BankKeeper.GetBalance(suite.chainB.GetContext(), suite.chainB.SenderAccount.GetAddress(), voucherDenomTrace.IBCDenom())
	balanceCantoBefore := suite.chainB.App.(*app.Canto).BankKeeper.GetBalance(suite.chainB.GetContext(), suite.chainB.SenderAccount.GetAddress(), "acanto")
	balanceErc20Before := erc20Keeper.BalanceOf(suite.chainB.GetContext(), contracts.ERC20MinterBurnerDecimalsContract.ABI, pair.GetERC20Contract(), common.BytesToAddress(suite.chainB.SenderAccount.GetAddress().Bytes()))

	// relay send
	res, err = onboardingtest.RelayPacket(path, packet)
	suite.Require().NoError(err) // relay committed

	events := res.GetEvents()
	var sdkEvents []sdk.Event
	for _, event := range events {
		sdkEvents = append(sdkEvents, sdk.Event(event))
	}
	attrs := onboardingtest.ExtractAttributes(onboardingtest.FindEvent(sdkEvents, "swap"))
	swapAmount, ok := sdkmath.NewIntFromString(attrs["amount"])
	if !ok {
		swapAmount = sdkmath.ZeroInt()
	}

	// check balances on chainB after the IBC transfer
	balanceVoucher := suite.chainB.App.(*app.Canto).BankKeeper.GetBalance(suite.chainB.GetContext(), suite.chainB.SenderAccount.GetAddress(), voucherDenomTrace.IBCDenom())
	balanceCanto := suite.chainB.App.(*app.Canto).BankKeeper.GetBalance(suite.chainB.GetContext(), suite.chainB.SenderAccount.GetAddress(), "acanto")
	balanceErc20 := erc20Keeper.BalanceOf(suite.chainB.GetContext(), contracts.ERC20MinterBurnerDecimalsContract.ABI, pair.GetERC20Contract(), common.BytesToAddress(suite.chainB.SenderAccount.GetAddress().Bytes()))

	coinSentFromAToB := types.GetTransferCoin(path.EndpointB.ChannelConfig.PortID, path.EndpointB.ChannelID, sdk.DefaultBondDenom, amount)

	// Check that the IBC voucher balance is same
	suite.Require().Equal(balanceVoucherBefore, balanceVoucher)

	// check whether the canto is swapped and the amount is greater than the threshold
	if balanceCantoBefore.Amount.LT(middlewareParams.AutoSwapThreshold) {
		suite.Require().Equal(balanceCanto.Amount, balanceCantoBefore.Amount.Add(middlewareParams.AutoSwapThreshold))
	} else {
		suite.Require().Equal(balanceCanto.Amount, balanceCantoBefore.Amount)
	}

	// Check that the convert is successful
	before := sdkmath.NewIntFromBigInt(balanceErc20Before)
	suite.Require().True(before.IsZero())
	suite.Require().Equal(coinSentFromAToB.Amount.Sub(swapAmount), sdkmath.NewIntFromBigInt(balanceErc20))

	// IBC transfer to blocked address
	blockedAddr := "canto10d07y265gmmuvt4z0w9aw880jnsr700jg5j4zm" // gov module
	coinToSendToB = suite.chainA.GetSimApp().BankKeeper.GetBalance(suite.chainA.GetContext(), suite.chainA.SenderAccount.GetAddress(), sdk.DefaultBondDenom)
	msg = types.NewMsgTransfer(path.EndpointA.ChannelConfig.PortID, path.EndpointA.ChannelID, coinToSendToB, suite.chainA.SenderAccount.GetAddress().String(), blockedAddr, timeoutHeight, 0, "")

	res, err = suite.chainA.SendMsgs(msg)
	suite.Require().NoError(err) // message committed

	packet, err = ibctesting.ParsePacketFromEvents(res.GetEvents())
	suite.Require().NoError(err)

	// relay send
	res, err = onboardingtest.RelayPacket(path, packet)
	suite.Require().NoError(err)
	ack, err := ibcgotesting.ParseAckFromEvents(res.GetEvents())
	suite.Require().NoError(err)
	suite.Require().Equal(ack, []byte(`{"error":"ABCI code: 4: error handling packet: see events for details"}`))

	// Send again from chainA to chainB
	// auto swap should not happen
	// auto convert all transferred IBC vouchers to ERC20
	coinToSendToB = suite.chainA.GetSimApp().BankKeeper.GetBalance(suite.chainA.GetContext(), suite.chainA.SenderAccount.GetAddress(), sdk.DefaultBondDenom)
	balanceVoucherBefore = suite.chainB.App.(*app.Canto).BankKeeper.GetBalance(suite.chainB.GetContext(), suite.chainB.SenderAccount.GetAddress(), voucherDenomTrace.IBCDenom())
	balanceCantoBefore = suite.chainB.App.(*app.Canto).BankKeeper.GetBalance(suite.chainB.GetContext(), suite.chainB.SenderAccount.GetAddress(), "acanto")
	balanceErc20Before = erc20Keeper.BalanceOf(suite.chainB.GetContext(), contracts.ERC20MinterBurnerDecimalsContract.ABI, pair.GetERC20Contract(), common.BytesToAddress(suite.chainB.SenderAccount.GetAddress().Bytes()))

	msg = types.NewMsgTransfer(path.EndpointA.ChannelConfig.PortID, path.EndpointA.ChannelID, coinToSendToB, suite.chainA.SenderAccount.GetAddress().String(), suite.chainB.SenderAccount.GetAddress().String(), timeoutHeight, 0, "")

	res, err = suite.chainA.SendMsgs(msg)
	suite.Require().NoError(err) // message committed

	packet, err = ibctesting.ParsePacketFromEvents(res.GetEvents())
	suite.Require().NoError(err)

	// relay send
	err = path.RelayPacket(packet)
	suite.Require().NoError(err) // relay committed

	events = res.GetEvents()
	sdkEvents = nil
	for _, event := range events {
		sdkEvents = append(sdkEvents, sdk.Event(event))
	}
	attrs = onboardingtest.ExtractAttributes(onboardingtest.FindEvent(sdkEvents, "swap"))
	swapAmount, ok = sdkmath.NewIntFromString(attrs["amount"])
	if !ok {
		swapAmount = sdkmath.ZeroInt()
	}

	coinSentFromAToB = types.GetTransferCoin(path.EndpointB.ChannelConfig.PortID, path.EndpointB.ChannelID, sdk.DefaultBondDenom, coinToSendToB.Amount)
	balanceVoucher = suite.chainB.App.(*app.Canto).BankKeeper.GetBalance(suite.chainB.GetContext(), suite.chainB.SenderAccount.GetAddress(), voucherDenomTrace.IBCDenom())
	balanceCanto = suite.chainB.App.(*app.Canto).BankKeeper.GetBalance(suite.chainB.GetContext(), suite.chainB.SenderAccount.GetAddress(), "acanto")
	balanceErc20 = erc20Keeper.BalanceOf(suite.chainB.GetContext(), contracts.ERC20MinterBurnerDecimalsContract.ABI, pair.GetERC20Contract(), common.BytesToAddress(suite.chainB.SenderAccount.GetAddress().Bytes()))

	suite.Require().Equal(balanceCantoBefore, balanceCanto)
	suite.Require().Equal(balanceVoucherBefore, balanceVoucher)
	suite.Require().Equal(sdkmath.NewIntFromBigInt(balanceErc20Before).Add(coinSentFromAToB.Amount).Sub(swapAmount), sdkmath.NewIntFromBigInt(balanceErc20))

}

func TestTransferTestSuite(t *testing.T) {
	suite.Run(t, new(TransferTestSuite))
}
