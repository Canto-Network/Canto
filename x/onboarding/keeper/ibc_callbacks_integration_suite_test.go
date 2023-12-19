package keeper_test

import (
	"strconv"
	"testing"
	"time"

	abci "github.com/cometbft/cometbft/abci/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/suite"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"

	transfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
	clienttypes "github.com/cosmos/ibc-go/v8/modules/core/02-client/types"
	channeltypes "github.com/cosmos/ibc-go/v8/modules/core/04-channel/types"

	"github.com/Canto-Network/Canto/v7/app"
	ibcgotesting "github.com/Canto-Network/Canto/v7/ibc/testing"
	coinswaptypes "github.com/Canto-Network/Canto/v7/x/coinswap/types"
	erc20types "github.com/Canto-Network/Canto/v7/x/erc20/types"
	inflationtypes "github.com/Canto-Network/Canto/v7/x/inflation/types"
	onboardingtest "github.com/Canto-Network/Canto/v7/x/onboarding/testutil"
	"github.com/Canto-Network/Canto/v7/x/onboarding/types"
)

type IBCTestingSuite struct {
	suite.Suite
	coordinator *ibcgotesting.Coordinator

	// testing chains used for convenience and readability
	cantoChain      *ibcgotesting.TestChain
	IBCGravityChain *ibcgotesting.TestChain
	IBCCosmosChain  *ibcgotesting.TestChain

	pathGravitycanto  *ibcgotesting.Path
	pathCosmoscanto   *ibcgotesting.Path
	pathGravityCosmos *ibcgotesting.Path
}

var s *IBCTestingSuite

func TestIBCTestingSuite(t *testing.T) {
	s = new(IBCTestingSuite)
	suite.Run(t, s)

	// Run Ginkgo integration tests
	RegisterFailHandler(Fail)
	RunSpecs(t, "Keeper Suite")
}

func (suite *IBCTestingSuite) SetupTest() {
	// initializes 3 test chains
	suite.coordinator = ibcgotesting.NewCoordinator(suite.T(), 1, 2)
	suite.cantoChain = suite.coordinator.GetChain(ibcgotesting.GetChainIDCanto(1))
	suite.IBCGravityChain = suite.coordinator.GetChain(ibcgotesting.GetChainID(2))
	suite.IBCCosmosChain = suite.coordinator.GetChain(ibcgotesting.GetChainID(3))
	suite.coordinator.CommitNBlocks(suite.cantoChain, 2)
	suite.coordinator.CommitNBlocks(suite.IBCGravityChain, 2)
	suite.coordinator.CommitNBlocks(suite.IBCCosmosChain, 2)

	// Mint coins on the gravity side which we'll use to unlock our acanto
	coinUsdc := sdk.NewCoin("uUSDC", sdkmath.NewIntWithDecimal(10000, 6))
	coinUsdt := sdk.NewCoin("uUSDT", sdkmath.NewIntWithDecimal(10000, 6))
	coins := sdk.NewCoins(coinUsdc, coinUsdt)
	err := suite.IBCGravityChain.GetSimApp().BankKeeper.MintCoins(suite.IBCGravityChain.GetContext(), minttypes.ModuleName, coins)
	suite.Require().NoError(err)
	err = suite.IBCGravityChain.GetSimApp().BankKeeper.SendCoinsFromModuleToAccount(suite.IBCGravityChain.GetContext(), minttypes.ModuleName, suite.IBCGravityChain.SenderAccount.GetAddress(), coins)
	suite.Require().NoError(err)

	// Mint coins on the cosmos side which we'll use to unlock our acanto
	coinAtom := sdk.NewCoin("uatom", sdkmath.NewIntWithDecimal(10000, 6))
	coins = sdk.NewCoins(coinAtom)
	err = suite.IBCCosmosChain.GetSimApp().BankKeeper.MintCoins(suite.IBCCosmosChain.GetContext(), minttypes.ModuleName, coins)
	suite.Require().NoError(err)
	err = suite.IBCCosmosChain.GetSimApp().BankKeeper.SendCoinsFromModuleToAccount(suite.IBCCosmosChain.GetContext(), minttypes.ModuleName, suite.IBCCosmosChain.SenderAccount.GetAddress(), coins)
	suite.Require().NoError(err)

	params := types.DefaultParams()
	params.EnableOnboarding = true
	suite.cantoChain.App.(*app.Canto).OnboardingKeeper.SetParams(suite.cantoChain.GetContext(), params)

	// Setup the paths between the chains
	suite.pathGravitycanto = ibcgotesting.NewTransferPath(suite.IBCGravityChain, suite.cantoChain) // clientID, connectionID, channelID empty
	suite.pathCosmoscanto = ibcgotesting.NewTransferPath(suite.IBCCosmosChain, suite.cantoChain)
	suite.pathGravityCosmos = ibcgotesting.NewTransferPath(suite.IBCCosmosChain, suite.IBCGravityChain)
	suite.coordinator.Setup(suite.pathGravitycanto) // clientID, connectionID, channelID filled
	suite.coordinator.Setup(suite.pathCosmoscanto)
	suite.coordinator.Setup(suite.pathGravityCosmos)
	suite.Require().Equal("07-tendermint-0", suite.pathGravitycanto.EndpointA.ClientID)
	suite.Require().Equal("connection-0", suite.pathGravitycanto.EndpointA.ConnectionID)
	suite.Require().Equal("channel-0", suite.pathGravitycanto.EndpointA.ChannelID)

	// Set the proposer address for the current header
	// It because EVMKeeper.GetCoinbaseAddress requires ProposerAddress in block header
	suite.cantoChain.CurrentHeader.ProposerAddress = suite.cantoChain.LastHeader.ValidatorSet.Proposer.Address
	suite.IBCGravityChain.CurrentHeader.ProposerAddress = suite.IBCGravityChain.LastHeader.ValidatorSet.Proposer.Address
	suite.IBCCosmosChain.CurrentHeader.ProposerAddress = suite.IBCCosmosChain.LastHeader.ValidatorSet.Proposer.Address
}

// FundCantoChain mints coins and sends them to the cantoChain sender account
func (suite *IBCTestingSuite) FundCantoChain(coins sdk.Coins) {
	err := suite.cantoChain.App.(*app.Canto).BankKeeper.MintCoins(suite.cantoChain.GetContext(), inflationtypes.ModuleName, coins)
	suite.Require().NoError(err)
	err = suite.cantoChain.App.(*app.Canto).BankKeeper.SendCoinsFromModuleToAccount(suite.cantoChain.GetContext(), inflationtypes.ModuleName, suite.cantoChain.SenderAccount.GetAddress(), coins)
	suite.Require().NoError(err)
}

// setupRegisterCoin deploys an erc20 contract and creates the token pair
func (suite *IBCTestingSuite) setupRegisterCoin(metadata banktypes.Metadata) *erc20types.TokenPair {
	err := suite.cantoChain.App.(*app.Canto).BankKeeper.MintCoins(suite.cantoChain.GetContext(), inflationtypes.ModuleName, sdk.Coins{sdk.NewInt64Coin(metadata.Base, 1)})
	suite.Require().NoError(err)

	pair, err := suite.cantoChain.App.(*app.Canto).Erc20Keeper.RegisterCoin(suite.cantoChain.GetContext(), metadata)
	suite.Require().NoError(err)
	return pair
}

// CreatePool creates a pool with acanto and the given denom
func (suite *IBCTestingSuite) CreatePool(denom string) {
	coincanto := sdk.NewCoin("acanto", sdkmath.NewIntWithDecimal(10000, 18))
	coinIBC := sdk.NewCoin(denom, sdkmath.NewIntWithDecimal(10000, 6))
	coins := sdk.NewCoins(coincanto, coinIBC)
	suite.FundCantoChain(coins)

	coinswapKeeper := suite.cantoChain.App.(*app.Canto).CoinswapKeeper
	coinswapKeeper.SetStandardDenom(suite.cantoChain.GetContext(), "acanto")
	coinswapParams := coinswapKeeper.GetParams(suite.cantoChain.GetContext())
	coinswapParams.MaxSwapAmount = sdk.NewCoins(sdk.NewCoin(denom, sdkmath.NewIntWithDecimal(10, 6)))
	coinswapKeeper.SetParams(suite.cantoChain.GetContext(), coinswapParams)

	// Create a message to add liquidity to the pool
	msgAddLiquidity := coinswaptypes.MsgAddLiquidity{
		MaxToken:         sdk.NewCoin(denom, sdkmath.NewIntWithDecimal(10000, 6)),
		ExactStandardAmt: sdkmath.NewIntWithDecimal(10000, 18),
		MinLiquidity:     sdkmath.NewInt(1),
		Deadline:         time.Now().Add(time.Minute * 10).Unix(),
		Sender:           suite.cantoChain.SenderAccount.GetAddress().String(),
	}

	// Add liquidity to the pool
	suite.cantoChain.App.(*app.Canto).CoinswapKeeper.AddLiquidity(suite.cantoChain.GetContext(), &msgAddLiquidity)
}

var (
	timeoutHeight   = clienttypes.NewHeight(1000, 1000)
	uusdcDenomtrace = transfertypes.DenomTrace{
		Path:      "transfer/channel-0",
		BaseDenom: "uUSDC",
	}
	uusdcIbcdenom = uusdcDenomtrace.IBCDenom()

	uusdtDenomtrace = transfertypes.DenomTrace{
		Path:      "transfer/channel-0",
		BaseDenom: "uUSDT",
	}
	uusdtIbcdenom = uusdtDenomtrace.IBCDenom()

	uatomDenomtrace = transfertypes.DenomTrace{
		Path:      "transfer/channel-1",
		BaseDenom: "uatom",
	}
	uatomIbcdenom = uatomDenomtrace.IBCDenom()
)

// SendAndReceiveMessage sends a transfer message from the origin chain to the destination chain
func (suite *IBCTestingSuite) SendAndReceiveMessage(path *ibcgotesting.Path, origin *ibcgotesting.TestChain, coin string, amount int64, sender string, receiver string, seq uint64) *abci.ExecTxResult {
	// Send coin from A to B
	transferMsg := transfertypes.NewMsgTransfer(path.EndpointA.ChannelConfig.PortID, path.EndpointA.ChannelID, sdk.NewCoin(coin, sdkmath.NewInt(amount)), sender, receiver, timeoutHeight, 0, "")
	_, err := origin.SendMsgs(transferMsg)
	suite.Require().NoError(err) // message committed

	// Recreate the packet that was sent
	transfer := transfertypes.NewFungibleTokenPacketData(coin, strconv.Itoa(int(amount)), sender, receiver, "")
	packet := channeltypes.NewPacket(transfer.GetBytes(), seq, path.EndpointA.ChannelConfig.PortID, path.EndpointA.ChannelID, path.EndpointB.ChannelConfig.PortID, path.EndpointB.ChannelID, timeoutHeight, 0)

	// patched RelayPacket call to get res
	res, err := onboardingtest.RelayPacket(path, packet)

	suite.Require().NoError(err)
	return res
}
