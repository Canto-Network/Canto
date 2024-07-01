package keeper_test

import (
	"fmt"
	"math"
	"math/big"

	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/runtime"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	govtypesv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/evmos/ethermint/crypto/ethsecp256k1"
	"github.com/evmos/ethermint/tests"
	"github.com/evmos/ethermint/x/evm/statedb"
	evmtypes "github.com/evmos/ethermint/x/evm/types"
	"github.com/stretchr/testify/mock"

	"github.com/Canto-Network/Canto/v7/testutil"
	"github.com/Canto-Network/Canto/v7/x/erc20/keeper"
	"github.com/Canto-Network/Canto/v7/x/erc20/types"
	inflationtypes "github.com/Canto-Network/Canto/v7/x/inflation/types"
)

const (
	contractMinterBurner = iota + 1
	contractDirectBalanceManipulation
	contractMaliciousDelayed
)

func (suite *KeeperTestSuite) setupRegisterERC20Pair(contractType int) common.Address {
	var contract common.Address
	// Deploy contract
	switch contractType {
	case contractDirectBalanceManipulation:
		contract = suite.DeployContractDirectBalanceManipulation(erc20Name, erc20Symbol)
	case contractMaliciousDelayed:
		contract = suite.DeployContractMaliciousDelayed(erc20Name, erc20Symbol)
	default:
		contract, _ = suite.DeployContract(erc20Name, erc20Symbol, erc20Decimals)
	}
	suite.Commit()

	_, err := suite.app.Erc20Keeper.RegisterERC20(suite.ctx, contract)
	suite.Require().NoError(err)
	return contract
}

func (suite *KeeperTestSuite) setupRegisterIBCVoucher() (banktypes.Metadata, *types.TokenPair) {
	suite.SetupTest()

	validMetadata := banktypes.Metadata{
		Description: "ATOM IBC voucher (channel 14)",
		Base:        ibcBase,
		// NOTE: Denom units MUST be increasing
		DenomUnits: []*banktypes.DenomUnit{
			{
				Denom:    ibcBase,
				Exponent: 0,
			},
		},
		Name:    "ATOM channel-14",
		Symbol:  "ibcATOM-14",
		Display: ibcBase,
	}

	err := suite.app.BankKeeper.MintCoins(suite.ctx, inflationtypes.ModuleName, sdk.Coins{sdk.NewInt64Coin(validMetadata.Base, 1)})
	suite.Require().NoError(err)

	pair, err := suite.app.Erc20Keeper.RegisterCoin(suite.ctx, validMetadata)
	suite.Require().NoError(err)
	suite.Commit()
	return validMetadata, pair
}

func (suite *KeeperTestSuite) TestMsgConvertCoin_ValidateBasic() {
	msg := types.MsgConvertCoin{}
	suite.Require().Equal("/canto.erc20.v1.MsgConvertCoin", sdk.MsgTypeURL(&msg))

	testCases := []struct {
		name       string
		coin       sdk.Coin
		receiver   string
		sender     string
		expectPass bool
	}{
		{
			"invalid denom",
			sdk.Coin{
				Denom:  "",
				Amount: sdkmath.NewInt(100),
			},
			"0x0000",
			tests.GenerateAddress().String(),
			false,
		},
		{
			"negative coin amount",
			sdk.Coin{
				Denom:  "coin",
				Amount: sdkmath.NewInt(-100),
			},
			"0x0000",
			tests.GenerateAddress().String(),
			false,
		},
		{
			"msg convert coin - invalid sender",
			sdk.NewCoin("coin", sdkmath.NewInt(100)),
			tests.GenerateAddress().String(),
			"cantoinvalid",
			false,
		},
		{
			"msg convert coin - invalid receiver",
			sdk.NewCoin("coin", sdkmath.NewInt(100)),
			"0x0000",
			sdk.AccAddress(tests.GenerateAddress().Bytes()).String(),
			false,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			msg := types.MsgConvertCoin{
				Coin:     tc.coin,
				Receiver: tc.receiver,
				Sender:   tc.sender,
			}
			_, err := suite.app.Erc20Keeper.ConvertCoin(suite.ctx, &msg)
			if tc.expectPass {
				suite.Require().NoError(err)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestMsgConvertERC20_ValidateBasic() {
	msg := types.MsgConvertERC20{}
	suite.Require().Equal("/canto.erc20.v1.MsgConvertERC20", sdk.MsgTypeURL(&msg))

	testCases := []struct {
		name       string
		amount     sdkmath.Int
		receiver   string
		contract   string
		sender     string
		expectPass bool
	}{
		{
			"invalid contract hex address",
			sdkmath.NewInt(100),
			sdk.AccAddress(tests.GenerateAddress().Bytes()).String(),
			sdk.AccAddress{}.String(),
			tests.GenerateAddress().String(),
			false,
		},
		{
			"negative coin amount",
			sdkmath.NewInt(-100),
			sdk.AccAddress(tests.GenerateAddress().Bytes()).String(),
			tests.GenerateAddress().String(),
			tests.GenerateAddress().String(),
			false,
		},
		{
			"invalid receiver address",
			sdkmath.NewInt(100),
			sdk.AccAddress{}.String(),
			tests.GenerateAddress().String(),
			tests.GenerateAddress().String(),
			false,
		},
		{
			"invalid sender address",
			sdkmath.NewInt(100),
			sdk.AccAddress(tests.GenerateAddress().Bytes()).String(),
			tests.GenerateAddress().String(),
			sdk.AccAddress{}.String(),
			false,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			msg := types.MsgConvertERC20{
				ContractAddress: tc.contract,
				Amount:          tc.amount,
				Receiver:        tc.receiver,
				Sender:          tc.sender,
			}
			_, err := suite.app.Erc20Keeper.ConvertERC20(suite.ctx, &msg)
			if tc.expectPass {
				suite.Require().NoError(err)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestConvertCoinNativeCoin() {
	testCases := []struct {
		name           string
		mint           int64
		burn           int64
		malleate       func(common.Address)
		extra          func()
		expPass        bool
		selfdestructed bool
	}{
		{
			"ok - sufficient funds",
			100,
			10,
			func(common.Address) {},
			func() {},
			true,
			false,
		},
		{
			"ok - equal funds",
			10,
			10,
			func(common.Address) {},
			func() {},
			true,
			false,
		},
		{
			"ok - suicided contract",
			10,
			10,
			func(erc20 common.Address) {
				stateDb := suite.StateDB()
				ok := stateDb.Suicide(erc20)
				suite.Require().True(ok)
				suite.Require().NoError(stateDb.Commit())
			},
			func() {},
			true,
			true,
		},
		{
			"fail - insufficient funds",
			0,
			10,
			func(common.Address) {},
			func() {},
			false,
			false,
		},
		{
			"fail - minting disabled",
			100,
			10,
			func(common.Address) {
				params := types.DefaultParams()
				params.EnableErc20 = false
				suite.app.Erc20Keeper.SetParams(suite.ctx, params)
			},
			func() {},
			false,
			false,
		},
		{
			"fail - deleted module account - force fail", 100, 10, func(common.Address) {},
			func() {
				acc := suite.app.AccountKeeper.GetAccount(suite.ctx, types.ModuleAddress.Bytes())
				suite.app.AccountKeeper.RemoveAccount(suite.ctx, acc)
			}, false, false,
		},
		{
			"fail - force evm fail", 100, 10, func(common.Address) {},
			func() {
				mockEVMKeeper := &MockEVMKeeper{}
				sp, found := suite.app.ParamsKeeper.GetSubspace(types.ModuleName)
				suite.Require().True(found)
				suite.app.Erc20Keeper = keeper.NewKeeper(runtime.NewKVStoreService(suite.app.GetKey("erc20")), suite.app.AppCodec(), sp, suite.app.AccountKeeper, suite.app.BankKeeper, mockEVMKeeper, authtypes.NewModuleAddress(govtypes.ModuleName).String())

				existingAcc := &statedb.Account{Nonce: uint64(1), Balance: common.Big1}
				balance := make([]uint8, 32)
				mockEVMKeeper.On("EstimateGas", mock.Anything, mock.Anything).Return(&evmtypes.EstimateGasResponse{Gas: uint64(200)}, nil)
				mockEVMKeeper.On("ApplyMessage", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&evmtypes.MsgEthereumTxResponse{Ret: balance}, nil).Once()
				mockEVMKeeper.On("ApplyMessage", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, fmt.Errorf("forced ApplyMessage error"))
				mockEVMKeeper.On("GetAccountWithoutBalance", mock.Anything, mock.Anything).Return(existingAcc, nil)
			}, false, false,
		},
		{
			"fail - force evm balance error", 100, 10, func(common.Address) {},
			func() {
				mockEVMKeeper := &MockEVMKeeper{}
				sp, found := suite.app.ParamsKeeper.GetSubspace(types.ModuleName)
				suite.Require().True(found)
				suite.app.Erc20Keeper = keeper.NewKeeper(runtime.NewKVStoreService(suite.app.GetKey("erc20")), suite.app.AppCodec(), sp, suite.app.AccountKeeper, suite.app.BankKeeper, mockEVMKeeper, authtypes.NewModuleAddress(govtypes.ModuleName).String())

				existingAcc := &statedb.Account{Nonce: uint64(1), Balance: common.Big1}
				balance := make([]uint8, 32)
				mockEVMKeeper.On("EstimateGas", mock.Anything, mock.Anything).Return(&evmtypes.EstimateGasResponse{Gas: uint64(200)}, nil)
				// first balance of
				mockEVMKeeper.On("ApplyMessage", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&evmtypes.MsgEthereumTxResponse{Ret: balance}, nil).Once()
				// convert coin
				mockEVMKeeper.On("ApplyMessage", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&evmtypes.MsgEthereumTxResponse{}, nil).Once()
				// second balance of
				mockEVMKeeper.On("ApplyMessage", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&evmtypes.MsgEthereumTxResponse{}, fmt.Errorf("third")).Once()
				// Extra call on test
				mockEVMKeeper.On("ApplyMessage", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&evmtypes.MsgEthereumTxResponse{}, nil)
				mockEVMKeeper.On("GetAccountWithoutBalance", mock.Anything, mock.Anything).Return(existingAcc, nil)
			}, false, false,
		},
		{
			"fail - force balance error", 100, 10, func(common.Address) {},
			func() {
				mockEVMKeeper := &MockEVMKeeper{}
				sp, found := suite.app.ParamsKeeper.GetSubspace(types.ModuleName)
				suite.Require().True(found)
				suite.app.Erc20Keeper = keeper.NewKeeper(runtime.NewKVStoreService(suite.app.GetKey("erc20")), suite.app.AppCodec(), sp, suite.app.AccountKeeper, suite.app.BankKeeper, mockEVMKeeper, authtypes.NewModuleAddress(govtypes.ModuleName).String())

				existingAcc := &statedb.Account{Nonce: uint64(1), Balance: common.Big1}
				balance := make([]uint8, 32)
				mockEVMKeeper.On("EstimateGas", mock.Anything, mock.Anything).Return(&evmtypes.EstimateGasResponse{Gas: uint64(200)}, nil)
				mockEVMKeeper.On("ApplyMessage", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&evmtypes.MsgEthereumTxResponse{Ret: balance}, nil).Times(4)
				mockEVMKeeper.On("GetAccountWithoutBalance", mock.Anything, mock.Anything).Return(existingAcc, nil)
			}, false, false,
		},
	}
	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.mintFeeCollector = true
			suite.SetupTest()
			metadata, pair := suite.setupRegisterCoin()
			suite.Require().NotNil(metadata)
			erc20 := pair.GetERC20Contract()
			tc.malleate(erc20)
			suite.Commit()

			ctx := sdk.WrapSDKContext(suite.ctx)
			coins := sdk.NewCoins(sdk.NewCoin(cosmosTokenBase, sdkmath.NewInt(tc.mint)))
			sender := sdk.AccAddress(suite.address.Bytes())
			msg := types.NewMsgConvertCoin(
				sdk.NewCoin(cosmosTokenBase, sdkmath.NewInt(tc.burn)),
				suite.address,
				sender,
			)

			suite.app.BankKeeper.MintCoins(suite.ctx, types.ModuleName, coins)
			suite.app.BankKeeper.SendCoinsFromModuleToAccount(suite.ctx, types.ModuleName, sender, coins)

			tc.extra()
			res, err := suite.app.Erc20Keeper.ConvertCoin(ctx, msg)
			expRes := &types.MsgConvertCoinResponse{}
			suite.Commit()
			balance := suite.BalanceOf(common.HexToAddress(pair.Erc20Address), suite.address)
			cosmosBalance := suite.app.BankKeeper.GetBalance(suite.ctx, sender, metadata.Base)

			if tc.expPass {
				suite.Require().NoError(err, tc.name)

				acc := suite.app.EvmKeeper.GetAccountWithoutBalance(suite.ctx, erc20)
				if tc.selfdestructed {
					suite.Require().Nil(acc, "expected contract to be destroyed")
				} else {
					suite.Require().NotNil(acc)
				}

				if tc.selfdestructed || !acc.IsContract() {
					id := suite.app.Erc20Keeper.GetTokenPairID(suite.ctx, erc20.String())
					_, found := suite.app.Erc20Keeper.GetTokenPair(suite.ctx, id)
					suite.Require().False(found)
				} else {
					suite.Require().Equal(expRes, res)
					suite.Require().Equal(cosmosBalance.Amount.Int64(), sdkmath.NewInt(tc.mint-tc.burn).Int64())
					suite.Require().Equal(balance.(*big.Int).Int64(), big.NewInt(tc.burn).Int64())
				}
			} else {
				suite.Require().Error(err, tc.name)
			}
		})
	}
	suite.mintFeeCollector = false
}

func (suite *KeeperTestSuite) TestConvertERC20NativeCoin() {
	testCases := []struct {
		name      string
		mint      int64
		burn      int64
		reconvert int64
		malleate  func()
		expPass   bool
	}{
		{"ok - sufficient funds", 100, 10, 5, func() {}, true},
		{"ok - equal funds", 10, 10, 10, func() {}, true},
		{"fail - insufficient funds", 10, 1, 5, func() {}, false},
		{"fail ", 10, 1, -5, func() {}, false},
		{
			"fail - deleted module account - force fail", 100, 10, 5,
			func() {
				acc := suite.app.AccountKeeper.GetAccount(suite.ctx, types.ModuleAddress.Bytes())
				suite.app.AccountKeeper.RemoveAccount(suite.ctx, acc)
			},
			false,
		},
		{
			"fail - force evm fail", 100, 10, 5,
			func() {
				mockEVMKeeper := &MockEVMKeeper{}
				sp, found := suite.app.ParamsKeeper.GetSubspace(types.ModuleName)
				suite.Require().True(found)
				suite.app.Erc20Keeper = keeper.NewKeeper(runtime.NewKVStoreService(suite.app.GetKey("erc20")), suite.app.AppCodec(), sp, suite.app.AccountKeeper, suite.app.BankKeeper, mockEVMKeeper, authtypes.NewModuleAddress(govtypes.ModuleName).String())

				existingAcc := &statedb.Account{Nonce: uint64(1), Balance: common.Big1}
				balance := make([]uint8, 32)
				mockEVMKeeper.On("EstimateGas", mock.Anything, mock.Anything).Return(&evmtypes.EstimateGasResponse{Gas: uint64(200)}, nil)
				mockEVMKeeper.On("ApplyMessage", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&evmtypes.MsgEthereumTxResponse{Ret: balance}, nil).Once()
				mockEVMKeeper.On("ApplyMessage", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, fmt.Errorf("forced ApplyMessage error"))
				mockEVMKeeper.On("GetAccountWithoutBalance", mock.Anything, mock.Anything).Return(existingAcc, nil)
			},
			false,
		},
		{
			"fail - force fail second balance", 100, 10, 5,
			func() {
				mockEVMKeeper := &MockEVMKeeper{}
				sp, found := suite.app.ParamsKeeper.GetSubspace(types.ModuleName)
				suite.Require().True(found)
				suite.app.Erc20Keeper = keeper.NewKeeper(runtime.NewKVStoreService(suite.app.GetKey("erc20")), suite.app.AppCodec(), sp, suite.app.AccountKeeper, suite.app.BankKeeper, mockEVMKeeper, authtypes.NewModuleAddress(govtypes.ModuleName).String())

				existingAcc := &statedb.Account{Nonce: uint64(1), Balance: common.Big1}
				balance := make([]uint8, 32)
				mockEVMKeeper.On("EstimateGas", mock.Anything, mock.Anything).Return(&evmtypes.EstimateGasResponse{Gas: uint64(200)}, nil)
				// first balance of
				mockEVMKeeper.On("ApplyMessage", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&evmtypes.MsgEthereumTxResponse{Ret: balance}, nil).Once()
				// convert coin
				mockEVMKeeper.On("ApplyMessage", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&evmtypes.MsgEthereumTxResponse{}, nil).Once()
				// second balance of
				mockEVMKeeper.On("ApplyMessage", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&evmtypes.MsgEthereumTxResponse{}, fmt.Errorf("third")).Once()
				// Extra call on test
				mockEVMKeeper.On("ApplyMessage", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&evmtypes.MsgEthereumTxResponse{}, nil)
				mockEVMKeeper.On("GetAccountWithoutBalance", mock.Anything, mock.Anything).Return(existingAcc, nil)
			},
			false,
		},
		{
			"fail - force fail second balance", 100, 10, 5,
			func() {
				mockEVMKeeper := &MockEVMKeeper{}
				sp, found := suite.app.ParamsKeeper.GetSubspace(types.ModuleName)
				suite.Require().True(found)
				suite.app.Erc20Keeper = keeper.NewKeeper(runtime.NewKVStoreService(suite.app.GetKey("erc20")), suite.app.AppCodec(), sp, suite.app.AccountKeeper, suite.app.BankKeeper, mockEVMKeeper, authtypes.NewModuleAddress(govtypes.ModuleName).String())

				existingAcc := &statedb.Account{Nonce: uint64(1), Balance: common.Big1}
				balance := make([]uint8, 32)
				mockEVMKeeper.On("EstimateGas", mock.Anything, mock.Anything).Return(&evmtypes.EstimateGasResponse{Gas: uint64(200)}, nil)
				// first balance of
				mockEVMKeeper.On("ApplyMessage", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&evmtypes.MsgEthereumTxResponse{Ret: balance}, nil).Once()
				// convert coin
				mockEVMKeeper.On("ApplyMessage", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&evmtypes.MsgEthereumTxResponse{}, nil).Once()
				// second balance of
				mockEVMKeeper.On("ApplyMessage", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&evmtypes.MsgEthereumTxResponse{Ret: balance}, nil).Once()
				// Extra call on test
				mockEVMKeeper.On("ApplyMessage", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&evmtypes.MsgEthereumTxResponse{}, nil)
				mockEVMKeeper.On("GetAccountWithoutBalance", mock.Anything, mock.Anything).Return(existingAcc, nil)
			},
			false,
		},
		{
			"fail - force fail unescrow", 100, 10, 5,
			func() {
				mockBankKeeper := &MockBankKeeper{}
				sp, found := suite.app.ParamsKeeper.GetSubspace(types.ModuleName)
				suite.Require().True(found)
				suite.app.Erc20Keeper = keeper.NewKeeper(runtime.NewKVStoreService(suite.app.GetKey("erc20")), suite.app.AppCodec(), sp, suite.app.AccountKeeper, mockBankKeeper, suite.app.EvmKeeper, authtypes.NewModuleAddress(govtypes.ModuleName).String())

				mockBankKeeper.On("SendCoinsFromModuleToAccount", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(fmt.Errorf("failed to unescrow"))
				mockBankKeeper.On("BlockedAddr", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(false)
				mockBankKeeper.On("GetBalance", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(sdk.Coin{Denom: "coin", Amount: sdkmath.OneInt()})
			},
			false,
		},
		{
			"fail - force fail balance after transfer", 100, 10, 5,
			func() {
				mockBankKeeper := &MockBankKeeper{}
				sp, found := suite.app.ParamsKeeper.GetSubspace(types.ModuleName)
				suite.Require().True(found)
				suite.app.Erc20Keeper = keeper.NewKeeper(runtime.NewKVStoreService(suite.app.GetKey("erc20")), suite.app.AppCodec(), sp, suite.app.AccountKeeper, mockBankKeeper, suite.app.EvmKeeper, authtypes.NewModuleAddress(govtypes.ModuleName).String())

				mockBankKeeper.On("SendCoinsFromModuleToAccount", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
				mockBankKeeper.On("BlockedAddr", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(false)
				mockBankKeeper.On("GetBalance", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(sdk.Coin{Denom: "acoin", Amount: sdkmath.OneInt()})
			},
			false,
		},
	}
	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.mintFeeCollector = true
			suite.SetupTest()
			metadata, pair := suite.setupRegisterCoin()
			suite.Require().NotNil(metadata)
			suite.Require().NotNil(pair)

			// Precondition: Convert Coin to ERC20
			coins := sdk.NewCoins(sdk.NewCoin(cosmosTokenBase, sdkmath.NewInt(tc.mint)))
			sender := sdk.AccAddress(suite.address.Bytes())
			suite.app.BankKeeper.MintCoins(suite.ctx, types.ModuleName, coins)
			suite.app.BankKeeper.SendCoinsFromModuleToAccount(suite.ctx, types.ModuleName, sender, coins)
			msg := types.NewMsgConvertCoin(
				sdk.NewCoin(cosmosTokenBase, sdkmath.NewInt(tc.burn)),
				suite.address,
				sender,
			)

			ctx := sdk.WrapSDKContext(suite.ctx)
			_, err := suite.app.Erc20Keeper.ConvertCoin(ctx, msg)
			suite.Require().NoError(err, tc.name)
			suite.Commit()
			balance := suite.BalanceOf(common.HexToAddress(pair.Erc20Address), suite.address)
			cosmosBalance := suite.app.BankKeeper.GetBalance(suite.ctx, sender, metadata.Base)
			suite.Require().Equal(cosmosBalance.Amount.Int64(), sdkmath.NewInt(tc.mint-tc.burn).Int64())
			suite.Require().Equal(balance, big.NewInt(tc.burn))

			// Convert ERC20s back to Coins
			ctx = sdk.WrapSDKContext(suite.ctx)
			contractAddr := common.HexToAddress(pair.Erc20Address)
			msgConvertERC20 := types.NewMsgConvertERC20(
				sdkmath.NewInt(tc.reconvert),
				sender,
				contractAddr,
				suite.address,
			)

			tc.malleate()
			res, err := suite.app.Erc20Keeper.ConvertERC20(ctx, msgConvertERC20)
			expRes := &types.MsgConvertERC20Response{}
			suite.Commit()
			balance = suite.BalanceOf(contractAddr, suite.address)
			cosmosBalance = suite.app.BankKeeper.GetBalance(suite.ctx, sender, pair.Denom)
			if tc.expPass {
				suite.Require().NoError(err, tc.name)
				suite.Require().Equal(expRes, res)
				suite.Require().Equal(cosmosBalance.Amount.Int64(), sdkmath.NewInt(tc.mint-tc.burn+tc.reconvert).Int64())
				suite.Require().Equal(balance.(*big.Int).Int64(), big.NewInt(tc.burn-tc.reconvert).Int64())
			} else {
				suite.Require().Error(err, tc.name)
			}
		})
	}
	suite.mintFeeCollector = false
}

func (suite *KeeperTestSuite) TestConvertERC20NativeERC20() {
	var contractAddr common.Address
	var coinName string

	testCases := []struct {
		name           string
		mint           int64
		transfer       int64
		malleate       func(common.Address)
		extra          func()
		contractType   int
		expPass        bool
		selfdestructed bool
	}{
		{
			"ok - sufficient funds",
			100,
			10,
			func(common.Address) {},
			func() {},
			contractMinterBurner,
			true,
			false,
		},
		{
			"ok - equal funds",
			10,
			10,
			func(common.Address) {},
			func() {},
			contractMinterBurner,
			true,
			false,
		},
		{
			"ok - equal funds",
			10,
			10,
			func(common.Address) {},
			func() {},
			contractMinterBurner,
			true,
			false,
		},
		{
			"ok - suicided contract",
			10,
			10,
			func(erc20 common.Address) {
				stateDb := suite.StateDB()
				ok := stateDb.Suicide(erc20)
				suite.Require().True(ok)
				suite.Require().NoError(stateDb.Commit())
			},
			func() {},
			contractMinterBurner,
			true,
			true,
		},
		{
			"fail - insufficient funds - callEVM",
			0,
			10,
			func(common.Address) {},
			func() {},
			contractMinterBurner,
			false,
			false,
		},
		{
			"fail - minting disabled",
			100,
			10,
			func(common.Address) {
				params := types.DefaultParams()
				params.EnableErc20 = false
				suite.app.Erc20Keeper.SetParams(suite.ctx, params)
			},
			func() {},
			contractMinterBurner,
			false,
			false,
		},
		{
			"fail - direct balance manipulation contract",
			100,
			10,
			func(common.Address) {},
			func() {},
			contractDirectBalanceManipulation,
			false,
			false,
		},
		{
			"fail - delayed malicious contract",
			10,
			10,
			func(common.Address) {},
			func() {},
			contractMaliciousDelayed,
			false,
			false,
		},
		{
			"fail - negative transfer contract",
			10,
			-10,
			func(common.Address) {},
			func() {},
			contractMinterBurner,
			false,
			false,
		},
		{
			"fail - no module address",
			100,
			10,
			func(common.Address) {
			},
			func() {
				acc := suite.app.AccountKeeper.GetAccount(suite.ctx, types.ModuleAddress.Bytes())
				suite.app.AccountKeeper.RemoveAccount(suite.ctx, acc)
			},
			contractMinterBurner,
			false,
			false,
		},
		{
			"fail - force evm fail",
			100,
			10,
			func(common.Address) {},
			func() {
				mockEVMKeeper := &MockEVMKeeper{}
				sp, found := suite.app.ParamsKeeper.GetSubspace(types.ModuleName)
				suite.Require().True(found)
				suite.app.Erc20Keeper = keeper.NewKeeper(runtime.NewKVStoreService(suite.app.GetKey("erc20")), suite.app.AppCodec(), sp, suite.app.AccountKeeper, suite.app.BankKeeper, mockEVMKeeper, authtypes.NewModuleAddress(govtypes.ModuleName).String())

				existingAcc := &statedb.Account{Nonce: uint64(1), Balance: common.Big1}
				balance := make([]uint8, 32)
				mockEVMKeeper.On("EstimateGas", mock.Anything, mock.Anything).Return(&evmtypes.EstimateGasResponse{Gas: uint64(200)}, nil)
				mockEVMKeeper.On("ApplyMessage", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&evmtypes.MsgEthereumTxResponse{Ret: balance}, nil).Once()
				mockEVMKeeper.On("ApplyMessage", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, fmt.Errorf("forced ApplyMessage error"))
				mockEVMKeeper.On("GetAccountWithoutBalance", mock.Anything, mock.Anything).Return(existingAcc, nil)
			},
			contractMinterBurner,
			false,
			false,
		},
		{
			"fail - force get balance fail",
			100,
			10,
			func(common.Address) {},
			func() {
				mockEVMKeeper := &MockEVMKeeper{}
				sp, found := suite.app.ParamsKeeper.GetSubspace(types.ModuleName)
				suite.Require().True(found)
				suite.app.Erc20Keeper = keeper.NewKeeper(runtime.NewKVStoreService(suite.app.GetKey("erc20")), suite.app.AppCodec(), sp, suite.app.AccountKeeper, suite.app.BankKeeper, mockEVMKeeper, authtypes.NewModuleAddress(govtypes.ModuleName).String())

				existingAcc := &statedb.Account{Nonce: uint64(1), Balance: common.Big1}
				balance := make([]uint8, 32)
				balance[31] = uint8(1)
				mockEVMKeeper.On("EstimateGas", mock.Anything, mock.Anything).Return(&evmtypes.EstimateGasResponse{Gas: uint64(200)}, nil)
				mockEVMKeeper.On("ApplyMessage", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&evmtypes.MsgEthereumTxResponse{Ret: balance}, nil).Twice()
				mockEVMKeeper.On("ApplyMessage", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, fmt.Errorf("forced balance error"))
				mockEVMKeeper.On("GetAccountWithoutBalance", mock.Anything, mock.Anything).Return(existingAcc, nil)
			},
			contractMinterBurner,
			false,
			false,
		},
		{
			"fail - force transfer unpack fail",
			100,
			10,
			func(common.Address) {},
			func() {
				mockEVMKeeper := &MockEVMKeeper{}
				sp, found := suite.app.ParamsKeeper.GetSubspace(types.ModuleName)
				suite.Require().True(found)
				suite.app.Erc20Keeper = keeper.NewKeeper(runtime.NewKVStoreService(suite.app.GetKey("erc20")), suite.app.AppCodec(), sp, suite.app.AccountKeeper, suite.app.BankKeeper, mockEVMKeeper, authtypes.NewModuleAddress(govtypes.ModuleName).String())

				existingAcc := &statedb.Account{Nonce: uint64(1), Balance: common.Big1}
				balance := make([]uint8, 32)
				mockEVMKeeper.On("EstimateGas", mock.Anything, mock.Anything).Return(&evmtypes.EstimateGasResponse{Gas: uint64(200)}, nil)
				mockEVMKeeper.On("ApplyMessage", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&evmtypes.MsgEthereumTxResponse{Ret: balance}, nil).Once()
				mockEVMKeeper.On("ApplyMessage", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&evmtypes.MsgEthereumTxResponse{}, nil)
				mockEVMKeeper.On("GetAccountWithoutBalance", mock.Anything, mock.Anything).Return(existingAcc, nil)
			},
			contractMinterBurner,
			false,
			false,
		},

		{
			"fail - force invalid transfer fail",
			100,
			10,
			func(common.Address) {},
			func() {
				mockEVMKeeper := &MockEVMKeeper{}
				sp, found := suite.app.ParamsKeeper.GetSubspace(types.ModuleName)
				suite.Require().True(found)
				suite.app.Erc20Keeper = keeper.NewKeeper(runtime.NewKVStoreService(suite.app.GetKey("erc20")), suite.app.AppCodec(), sp, suite.app.AccountKeeper, suite.app.BankKeeper, mockEVMKeeper, authtypes.NewModuleAddress(govtypes.ModuleName).String())

				existingAcc := &statedb.Account{Nonce: uint64(1), Balance: common.Big1}
				balance := make([]uint8, 32)
				mockEVMKeeper.On("EstimateGas", mock.Anything, mock.Anything).Return(&evmtypes.EstimateGasResponse{Gas: uint64(200)}, nil)
				mockEVMKeeper.On("ApplyMessage", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&evmtypes.MsgEthereumTxResponse{Ret: balance}, nil).Once()
				mockEVMKeeper.On("ApplyMessage", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&evmtypes.MsgEthereumTxResponse{Ret: balance}, nil)
				mockEVMKeeper.On("GetAccountWithoutBalance", mock.Anything, mock.Anything).Return(existingAcc, nil)
			},
			contractMinterBurner,
			false,
			false,
		},
		{
			"fail - force mint fail",
			100,
			10,
			func(common.Address) {},
			func() {
				mockBankKeeper := &MockBankKeeper{}
				sp, found := suite.app.ParamsKeeper.GetSubspace(types.ModuleName)
				suite.Require().True(found)
				suite.app.Erc20Keeper = keeper.NewKeeper(runtime.NewKVStoreService(suite.app.GetKey("erc20")), suite.app.AppCodec(), sp, suite.app.AccountKeeper, mockBankKeeper, suite.app.EvmKeeper, authtypes.NewModuleAddress(govtypes.ModuleName).String())

				mockBankKeeper.On("MintCoins", mock.Anything, mock.Anything, mock.Anything).Return(fmt.Errorf("failed to mint"))
				mockBankKeeper.On("SendCoinsFromModuleToAccount", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(fmt.Errorf("failed to unescrow"))
				mockBankKeeper.On("BlockedAddr", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(false)
				mockBankKeeper.On("GetBalance", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(sdk.Coin{Denom: "coin", Amount: sdkmath.OneInt()})
			},
			contractMinterBurner,
			false,
			false,
		},
		{
			"fail - force send minted fail",
			100,
			10,
			func(common.Address) {},
			func() {
				mockBankKeeper := &MockBankKeeper{}
				sp, found := suite.app.ParamsKeeper.GetSubspace(types.ModuleName)
				suite.Require().True(found)
				suite.app.Erc20Keeper = keeper.NewKeeper(runtime.NewKVStoreService(suite.app.GetKey("erc20")), suite.app.AppCodec(), sp, suite.app.AccountKeeper, mockBankKeeper, suite.app.EvmKeeper, authtypes.NewModuleAddress(govtypes.ModuleName).String())

				mockBankKeeper.On("MintCoins", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				mockBankKeeper.On("SendCoinsFromModuleToAccount", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(fmt.Errorf("failed to unescrow"))
				mockBankKeeper.On("BlockedAddr", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(false)
				mockBankKeeper.On("GetBalance", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(sdk.Coin{Denom: "coin", Amount: sdkmath.OneInt()})
			},
			contractMinterBurner,
			false,
			false,
		},
		{
			"fail - force bank balance fail",
			100,
			10,
			func(common.Address) {},
			func() {
				mockBankKeeper := &MockBankKeeper{}
				sp, found := suite.app.ParamsKeeper.GetSubspace(types.ModuleName)
				suite.Require().True(found)
				suite.app.Erc20Keeper = keeper.NewKeeper(runtime.NewKVStoreService(suite.app.GetKey("erc20")), suite.app.AppCodec(), sp, suite.app.AccountKeeper, mockBankKeeper, suite.app.EvmKeeper, authtypes.NewModuleAddress(govtypes.ModuleName).String())

				mockBankKeeper.On("MintCoins", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				mockBankKeeper.On("SendCoinsFromModuleToAccount", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
				mockBankKeeper.On("BlockedAddr", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(false)
				mockBankKeeper.On("GetBalance", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(sdk.Coin{Denom: coinName, Amount: sdkmath.NewInt(int64(10))})
			},
			contractMinterBurner,
			false,
			false,
		},
	}
	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.mintFeeCollector = true
			suite.SetupTest()

			contractAddr = suite.setupRegisterERC20Pair(tc.contractType)

			tc.malleate(contractAddr)
			suite.Require().NotNil(contractAddr)
			suite.Commit()

			coinName = types.CreateDenom(contractAddr.String())
			sender := sdk.AccAddress(suite.address.Bytes())
			msg := types.NewMsgConvertERC20(
				sdkmath.NewInt(tc.transfer),
				sender,
				contractAddr,
				suite.address,
			)

			suite.MintERC20Token(contractAddr, suite.address, suite.address, big.NewInt(tc.mint))
			suite.Commit()
			ctx := suite.ctx

			tc.extra()
			res, err := suite.app.Erc20Keeper.ConvertERC20(ctx, msg)

			expRes := &types.MsgConvertERC20Response{}
			suite.Commit()
			balance := suite.BalanceOf(contractAddr, suite.address)
			cosmosBalance := suite.app.BankKeeper.GetBalance(suite.ctx, sender, coinName)
			if tc.expPass {
				suite.Require().NoError(err, tc.name)

				acc := suite.app.EvmKeeper.GetAccountWithoutBalance(suite.ctx, contractAddr)
				if tc.selfdestructed {
					suite.Require().Nil(acc, "expected contract to be destroyed")
				} else {
					suite.Require().NotNil(acc)
				}

				if tc.selfdestructed || !acc.IsContract() {
					id := suite.app.Erc20Keeper.GetTokenPairID(suite.ctx, contractAddr.String())
					_, found := suite.app.Erc20Keeper.GetTokenPair(suite.ctx, id)
					suite.Require().False(found)
				} else {
					suite.Require().Equal(expRes, res)
					suite.Require().Equal(cosmosBalance.Amount, sdkmath.NewInt(tc.transfer))
					suite.Require().Equal(balance.(*big.Int).Int64(), big.NewInt(tc.mint-tc.transfer).Int64())
				}
			} else {
				suite.Require().Error(err, tc.name)
			}
		})
	}
	suite.mintFeeCollector = false
}

func (suite *KeeperTestSuite) TestConvertCoinNativeERC20() {
	var contractAddr common.Address

	testCases := []struct {
		name         string
		mint         int64
		convert      int64
		malleate     func(common.Address)
		extra        func()
		contractType int
		expPass      bool
	}{
		{
			"ok - sufficient funds",
			100,
			10,
			func(common.Address) {},
			func() {},
			contractMinterBurner,
			true,
		},
		{
			"ok - equal funds",
			100,
			100,
			func(common.Address) {},
			func() {},
			contractMinterBurner,
			true,
		},
		{
			"fail - insufficient funds",
			100,
			200,
			func(common.Address) {},
			func() {},
			contractMinterBurner,
			false,
		},
		{
			"fail - direct balance manipulation contract",
			100,
			10,
			func(common.Address) {},
			func() {},
			contractDirectBalanceManipulation,
			false,
		},
		{
			"fail - malicious delayed contract",
			100,
			10,
			func(common.Address) {},
			func() {},
			contractMaliciousDelayed,
			false,
		},
		{
			"fail - deleted module address - force fail",
			100,
			10,
			func(common.Address) {},
			func() {
				acc := suite.app.AccountKeeper.GetAccount(suite.ctx, types.ModuleAddress.Bytes())
				suite.app.AccountKeeper.RemoveAccount(suite.ctx, acc)
			},
			contractMinterBurner,
			false,
		},
		{
			"fail - force evm fail",
			100,
			10,
			func(common.Address) {},
			func() {
				mockEVMKeeper := &MockEVMKeeper{}
				sp, found := suite.app.ParamsKeeper.GetSubspace(types.ModuleName)
				suite.Require().True(found)
				suite.app.Erc20Keeper = keeper.NewKeeper(runtime.NewKVStoreService(suite.app.GetKey("erc20")), suite.app.AppCodec(), sp, suite.app.AccountKeeper, suite.app.BankKeeper, mockEVMKeeper, authtypes.NewModuleAddress(govtypes.ModuleName).String())

				existingAcc := &statedb.Account{Nonce: uint64(1), Balance: common.Big1}
				balance := make([]uint8, 32)
				mockEVMKeeper.On("EstimateGas", mock.Anything, mock.Anything).Return(&evmtypes.EstimateGasResponse{Gas: uint64(200)}, nil)
				mockEVMKeeper.On("ApplyMessage", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&evmtypes.MsgEthereumTxResponse{Ret: balance}, nil).Once()
				mockEVMKeeper.On("ApplyMessage", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, fmt.Errorf("forced ApplyMessage error"))
				mockEVMKeeper.On("GetAccountWithoutBalance", mock.Anything, mock.Anything).Return(existingAcc, nil)
			},
			contractMinterBurner,
			false,
		},
		{
			"fail - force invalid transfer",
			100,
			10,
			func(common.Address) {},
			func() {
				mockEVMKeeper := &MockEVMKeeper{}
				sp, found := suite.app.ParamsKeeper.GetSubspace(types.ModuleName)
				suite.Require().True(found)
				suite.app.Erc20Keeper = keeper.NewKeeper(runtime.NewKVStoreService(suite.app.GetKey("erc20")), suite.app.AppCodec(), sp, suite.app.AccountKeeper, suite.app.BankKeeper, mockEVMKeeper, authtypes.NewModuleAddress(govtypes.ModuleName).String())

				existingAcc := &statedb.Account{Nonce: uint64(1), Balance: common.Big1}
				balance := make([]uint8, 32)
				mockEVMKeeper.On("EstimateGas", mock.Anything, mock.Anything).Return(&evmtypes.EstimateGasResponse{Gas: uint64(200)}, nil)
				mockEVMKeeper.On("ApplyMessage", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&evmtypes.MsgEthereumTxResponse{Ret: balance}, nil).Once()
				mockEVMKeeper.On("ApplyMessage", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&evmtypes.MsgEthereumTxResponse{}, nil)
				mockEVMKeeper.On("GetAccountWithoutBalance", mock.Anything, mock.Anything).Return(existingAcc, nil)
			},
			contractMinterBurner,
			false,
		},
		{
			"fail - force fail second balance",
			100,
			10,
			func(common.Address) {},
			func() {
				mockEVMKeeper := &MockEVMKeeper{}
				sp, found := suite.app.ParamsKeeper.GetSubspace(types.ModuleName)
				suite.Require().True(found)
				suite.app.Erc20Keeper = keeper.NewKeeper(runtime.NewKVStoreService(suite.app.GetKey("erc20")), suite.app.AppCodec(), sp, suite.app.AccountKeeper, suite.app.BankKeeper, mockEVMKeeper, authtypes.NewModuleAddress(govtypes.ModuleName).String())

				existingAcc := &statedb.Account{Nonce: uint64(1), Balance: common.Big1}
				balance := make([]uint8, 32)
				balance[31] = uint8(1)
				mockEVMKeeper.On("EstimateGas", mock.Anything, mock.Anything).Return(&evmtypes.EstimateGasResponse{Gas: uint64(200)}, nil)
				mockEVMKeeper.On("ApplyMessage", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&evmtypes.MsgEthereumTxResponse{Ret: balance}, nil).Twice()
				mockEVMKeeper.On("ApplyMessage", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, fmt.Errorf("fail second balance"))
				mockEVMKeeper.On("GetAccountWithoutBalance", mock.Anything, mock.Anything).Return(existingAcc, nil)
			},
			contractMinterBurner,
			false,
		},
		{
			"fail - force fail transfer",
			100,
			10,
			func(common.Address) {},
			func() {
				mockEVMKeeper := &MockEVMKeeper{}
				sp, found := suite.app.ParamsKeeper.GetSubspace(types.ModuleName)
				suite.Require().True(found)
				suite.app.Erc20Keeper = keeper.NewKeeper(runtime.NewKVStoreService(suite.app.GetKey("erc20")), suite.app.AppCodec(), sp, suite.app.AccountKeeper, suite.app.BankKeeper, mockEVMKeeper, authtypes.NewModuleAddress(govtypes.ModuleName).String())

				existingAcc := &statedb.Account{Nonce: uint64(1), Balance: common.Big1}
				balance := make([]uint8, 32)
				mockEVMKeeper.On("EstimateGas", mock.Anything, mock.Anything).Return(&evmtypes.EstimateGasResponse{Gas: uint64(200)}, nil)
				mockEVMKeeper.On("ApplyMessage", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&evmtypes.MsgEthereumTxResponse{Ret: balance}, nil)
				mockEVMKeeper.On("GetAccountWithoutBalance", mock.Anything, mock.Anything).Return(existingAcc, nil)
			},
			contractMinterBurner,
			false,
		},
	}
	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.mintFeeCollector = true
			suite.SetupTest()
			contractAddr = suite.setupRegisterERC20Pair(tc.contractType)
			suite.Require().NotNil(contractAddr)

			id := suite.app.Erc20Keeper.GetTokenPairID(suite.ctx, contractAddr.String())
			pair, _ := suite.app.Erc20Keeper.GetTokenPair(suite.ctx, id)
			coins := sdk.NewCoins(sdk.NewCoin(pair.Denom, sdkmath.NewInt(tc.mint)))
			coinName := types.CreateDenom(contractAddr.String())
			sender := sdk.AccAddress(suite.address.Bytes())

			// Precondition: Mint Coins to convert on sender account
			suite.app.BankKeeper.MintCoins(suite.ctx, types.ModuleName, coins)
			suite.app.BankKeeper.SendCoinsFromModuleToAccount(suite.ctx, types.ModuleName, sender, coins)
			cosmosBalance := suite.app.BankKeeper.GetBalance(suite.ctx, sender, coinName)
			suite.Require().Equal(sdkmath.NewInt(tc.mint), cosmosBalance.Amount)

			// Precondition: Mint escrow tokens on module account
			suite.GrantERC20Token(contractAddr, suite.address, types.ModuleAddress, "MINTER_ROLE")
			suite.MintERC20Token(contractAddr, types.ModuleAddress, types.ModuleAddress, big.NewInt(tc.mint))
			tokenBalance := suite.BalanceOf(contractAddr, types.ModuleAddress)
			suite.Require().Equal(big.NewInt(tc.mint), tokenBalance)

			tc.malleate(contractAddr)
			suite.Commit()

			// Convert Coins back to ERC20s
			receiver := suite.address
			ctx := sdk.WrapSDKContext(suite.ctx)
			msg := types.NewMsgConvertCoin(
				sdk.NewCoin(coinName, sdkmath.NewInt(tc.convert)),
				receiver,
				sender,
			)

			tc.extra()
			res, err := suite.app.Erc20Keeper.ConvertCoin(ctx, msg)

			expRes := &types.MsgConvertCoinResponse{}
			suite.Commit()
			tokenBalance = suite.BalanceOf(contractAddr, suite.address)
			cosmosBalance = suite.app.BankKeeper.GetBalance(suite.ctx, sender, coinName)
			if tc.expPass {
				suite.Require().NoError(err, tc.name)
				suite.Require().Equal(expRes, res)
				suite.Require().Equal(sdkmath.NewInt(tc.mint-tc.convert), cosmosBalance.Amount)
				suite.Require().Equal(big.NewInt(tc.convert), tokenBalance.(*big.Int))
			} else {
				suite.Require().Error(err, tc.name)
			}
		})
	}
	suite.mintFeeCollector = false
}

func (suite *KeeperTestSuite) TestWrongPairOwnerERC20NativeCoin() {
	testCases := []struct {
		name      string
		mint      int64
		burn      int64
		reconvert int64
		expPass   bool
	}{
		{"ok - sufficient funds", 100, 10, 5, true},
	}
	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.mintFeeCollector = true
			suite.SetupTest()
			metadata, pair := suite.setupRegisterCoin()
			suite.Require().NotNil(metadata)
			suite.Require().NotNil(pair)

			// Precondition: Convert Coin to ERC20
			coins := sdk.NewCoins(sdk.NewCoin(cosmosTokenBase, sdkmath.NewInt(tc.mint)))
			sender := sdk.AccAddress(suite.address.Bytes())
			suite.app.BankKeeper.MintCoins(suite.ctx, types.ModuleName, coins)
			suite.app.BankKeeper.SendCoinsFromModuleToAccount(suite.ctx, types.ModuleName, sender, coins)
			msg := types.NewMsgConvertCoin(
				sdk.NewCoin(cosmosTokenBase, sdkmath.NewInt(tc.burn)),
				suite.address,
				sender,
			)

			pair.ContractOwner = types.OWNER_UNSPECIFIED
			suite.app.Erc20Keeper.SetTokenPair(suite.ctx, *pair)

			ctx := sdk.WrapSDKContext(suite.ctx)
			_, err := suite.app.Erc20Keeper.ConvertCoin(ctx, msg)
			suite.Require().Error(err, tc.name)

			// Convert ERC20s back to Coins
			ctx = sdk.WrapSDKContext(suite.ctx)
			contractAddr := common.HexToAddress(pair.Erc20Address)
			msgConvertERC20 := types.NewMsgConvertERC20(
				sdkmath.NewInt(tc.reconvert),
				sender,
				contractAddr,
				suite.address,
			)

			_, err = suite.app.Erc20Keeper.ConvertERC20(ctx, msgConvertERC20)
			suite.Require().Error(err, tc.name)
		})
	}
}

func (suite *KeeperTestSuite) TestConvertCoinNativeIBCVoucher() {
	testCases := []struct {
		name           string
		mint           int64
		burn           int64
		malleate       func(common.Address)
		extra          func()
		expPass        bool
		selfdestructed bool
	}{
		{
			"ok - sufficient funds",
			100,
			10,
			func(common.Address) {},
			func() {},
			true,
			false,
		},
		{
			"ok - equal funds",
			10,
			10,
			func(common.Address) {},
			func() {},
			true,
			false,
		},
		{
			"ok - suicided contract",
			10,
			10,
			func(erc20 common.Address) {
				stateDb := suite.StateDB()
				ok := stateDb.Suicide(erc20)
				suite.Require().True(ok)
				suite.Require().NoError(stateDb.Commit())
			},
			func() {},
			true,
			true,
		},
		{
			"fail - insufficient funds",
			0,
			10,
			func(common.Address) {},
			func() {},
			false,
			false,
		},
		{
			"fail - minting disabled",
			100,
			10,
			func(common.Address) {
				params := types.DefaultParams()
				params.EnableErc20 = false
				suite.app.Erc20Keeper.SetParams(suite.ctx, params)
			},
			func() {},
			false,
			false,
		},
		{
			"fail - deleted module account - force fail", 100, 10, func(common.Address) {},
			func() {
				acc := suite.app.AccountKeeper.GetAccount(suite.ctx, types.ModuleAddress.Bytes())
				suite.app.AccountKeeper.RemoveAccount(suite.ctx, acc)
			}, false, false,
		},
		{
			"fail - force evm fail", 100, 10, func(common.Address) {},
			func() {
				mockEVMKeeper := &MockEVMKeeper{}
				sp, found := suite.app.ParamsKeeper.GetSubspace(types.ModuleName)
				suite.Require().True(found)
				suite.app.Erc20Keeper = keeper.NewKeeper(runtime.NewKVStoreService(suite.app.GetKey("erc20")), suite.app.AppCodec(), sp, suite.app.AccountKeeper, suite.app.BankKeeper, mockEVMKeeper, authtypes.NewModuleAddress(govtypes.ModuleName).String())

				existingAcc := &statedb.Account{Nonce: uint64(1), Balance: common.Big1}
				balance := make([]uint8, 32)
				mockEVMKeeper.On("EstimateGas", mock.Anything, mock.Anything).Return(&evmtypes.EstimateGasResponse{Gas: uint64(200)}, nil)
				mockEVMKeeper.On("ApplyMessage", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&evmtypes.MsgEthereumTxResponse{Ret: balance}, nil).Once()
				mockEVMKeeper.On("ApplyMessage", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, fmt.Errorf("forced ApplyMessage error"))
				mockEVMKeeper.On("GetAccountWithoutBalance", mock.Anything, mock.Anything).Return(existingAcc, nil)
			}, false, false,
		},
		{
			"fail - force evm balance error", 100, 10, func(common.Address) {},
			func() {
				mockEVMKeeper := &MockEVMKeeper{}
				sp, found := suite.app.ParamsKeeper.GetSubspace(types.ModuleName)
				suite.Require().True(found)
				suite.app.Erc20Keeper = keeper.NewKeeper(runtime.NewKVStoreService(suite.app.GetKey("erc20")), suite.app.AppCodec(), sp, suite.app.AccountKeeper, suite.app.BankKeeper, mockEVMKeeper, authtypes.NewModuleAddress(govtypes.ModuleName).String())

				existingAcc := &statedb.Account{Nonce: uint64(1), Balance: common.Big1}
				balance := make([]uint8, 32)
				mockEVMKeeper.On("EstimateGas", mock.Anything, mock.Anything).Return(&evmtypes.EstimateGasResponse{Gas: uint64(200)}, nil)
				// first balance of
				mockEVMKeeper.On("ApplyMessage", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&evmtypes.MsgEthereumTxResponse{Ret: balance}, nil).Once()
				// convert coin
				mockEVMKeeper.On("ApplyMessage", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&evmtypes.MsgEthereumTxResponse{}, nil).Once()
				// second balance of
				mockEVMKeeper.On("ApplyMessage", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&evmtypes.MsgEthereumTxResponse{}, fmt.Errorf("third")).Once()
				// Extra call on test
				mockEVMKeeper.On("ApplyMessage", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&evmtypes.MsgEthereumTxResponse{}, nil)
				mockEVMKeeper.On("GetAccountWithoutBalance", mock.Anything, mock.Anything).Return(existingAcc, nil)
			}, false, false,
		},
		{
			"fail - force balance error", 100, 10, func(common.Address) {},
			func() {
				mockEVMKeeper := &MockEVMKeeper{}
				sp, found := suite.app.ParamsKeeper.GetSubspace(types.ModuleName)
				suite.Require().True(found)
				suite.app.Erc20Keeper = keeper.NewKeeper(runtime.NewKVStoreService(suite.app.GetKey("erc20")), suite.app.AppCodec(), sp, suite.app.AccountKeeper, suite.app.BankKeeper, mockEVMKeeper, authtypes.NewModuleAddress(govtypes.ModuleName).String())

				existingAcc := &statedb.Account{Nonce: uint64(1), Balance: common.Big1}
				balance := make([]uint8, 32)
				mockEVMKeeper.On("EstimateGas", mock.Anything, mock.Anything).Return(&evmtypes.EstimateGasResponse{Gas: uint64(200)}, nil)
				mockEVMKeeper.On("ApplyMessage", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&evmtypes.MsgEthereumTxResponse{Ret: balance}, nil).Times(4)
				mockEVMKeeper.On("GetAccountWithoutBalance", mock.Anything, mock.Anything).Return(existingAcc, nil)
			}, false, false,
		},
	}
	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.mintFeeCollector = true
			suite.SetupTest()
			metadata, pair := suite.setupRegisterIBCVoucher()
			suite.Require().NotNil(metadata)
			erc20 := pair.GetERC20Contract()
			tc.malleate(erc20)
			suite.Commit()

			ctx := sdk.WrapSDKContext(suite.ctx)
			coins := sdk.NewCoins(sdk.NewCoin(ibcBase, sdkmath.NewInt(tc.mint)))
			sender := sdk.AccAddress(suite.address.Bytes())
			msg := types.NewMsgConvertCoin(
				sdk.NewCoin(ibcBase, sdkmath.NewInt(tc.burn)),
				suite.address,
				sender,
			)

			suite.app.BankKeeper.MintCoins(suite.ctx, types.ModuleName, coins)
			suite.app.BankKeeper.SendCoinsFromModuleToAccount(suite.ctx, types.ModuleName, sender, coins)

			tc.extra()
			res, err := suite.app.Erc20Keeper.ConvertCoin(ctx, msg)
			expRes := &types.MsgConvertCoinResponse{}
			suite.Commit()
			balance := suite.BalanceOf(common.HexToAddress(pair.Erc20Address), suite.address)
			cosmosBalance := suite.app.BankKeeper.GetBalance(suite.ctx, sender, metadata.Base)

			if tc.expPass {
				suite.Require().NoError(err, tc.name)

				acc := suite.app.EvmKeeper.GetAccountWithoutBalance(suite.ctx, erc20)
				if tc.selfdestructed {
					suite.Require().Nil(acc, "expected contract to be destroyed")
				} else {
					suite.Require().NotNil(acc)
				}

				if tc.selfdestructed || !acc.IsContract() {
					id := suite.app.Erc20Keeper.GetTokenPairID(suite.ctx, erc20.String())
					_, found := suite.app.Erc20Keeper.GetTokenPair(suite.ctx, id)
					suite.Require().False(found)
				} else {
					suite.Require().Equal(expRes, res)
					suite.Require().Equal(cosmosBalance.Amount.Int64(), sdkmath.NewInt(tc.mint-tc.burn).Int64())
					suite.Require().Equal(balance.(*big.Int).Int64(), big.NewInt(tc.burn).Int64())
				}
			} else {
				suite.Require().Error(err, tc.name)
			}
		})
	}
	suite.mintFeeCollector = false
}

func (suite *KeeperTestSuite) TestConvertERC20NativeIBCVoucher() {
	testCases := []struct {
		name      string
		mint      int64
		burn      int64
		reconvert int64
		malleate  func()
		expPass   bool
	}{
		{"ok - sufficient funds", 100, 10, 5, func() {}, true},
		{"ok - equal funds", 10, 10, 10, func() {}, true},
		{"fail - insufficient funds", 10, 1, 5, func() {}, false},
		{"fail ", 10, 1, -5, func() {}, false},
		{
			"fail - deleted module account - force fail", 100, 10, 5,
			func() {
				acc := suite.app.AccountKeeper.GetAccount(suite.ctx, types.ModuleAddress.Bytes())
				suite.app.AccountKeeper.RemoveAccount(suite.ctx, acc)
			},
			false,
		},
		{
			"fail - force evm fail", 100, 10, 5,
			func() {
				mockEVMKeeper := &MockEVMKeeper{}
				sp, found := suite.app.ParamsKeeper.GetSubspace(types.ModuleName)
				suite.Require().True(found)
				suite.app.Erc20Keeper = keeper.NewKeeper(runtime.NewKVStoreService(suite.app.GetKey("erc20")), suite.app.AppCodec(), sp, suite.app.AccountKeeper, suite.app.BankKeeper, mockEVMKeeper, authtypes.NewModuleAddress(govtypes.ModuleName).String())

				existingAcc := &statedb.Account{Nonce: uint64(1), Balance: common.Big1}
				balance := make([]uint8, 32)
				mockEVMKeeper.On("EstimateGas", mock.Anything, mock.Anything).Return(&evmtypes.EstimateGasResponse{Gas: uint64(200)}, nil)
				mockEVMKeeper.On("ApplyMessage", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&evmtypes.MsgEthereumTxResponse{Ret: balance}, nil).Once()
				mockEVMKeeper.On("ApplyMessage", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, fmt.Errorf("forced ApplyMessage error"))
				mockEVMKeeper.On("GetAccountWithoutBalance", mock.Anything, mock.Anything).Return(existingAcc, nil)
			},
			false,
		},
		{
			"fail - force fail second balance", 100, 10, 5,
			func() {
				mockEVMKeeper := &MockEVMKeeper{}
				sp, found := suite.app.ParamsKeeper.GetSubspace(types.ModuleName)
				suite.Require().True(found)
				suite.app.Erc20Keeper = keeper.NewKeeper(runtime.NewKVStoreService(suite.app.GetKey("erc20")), suite.app.AppCodec(), sp, suite.app.AccountKeeper, suite.app.BankKeeper, mockEVMKeeper, authtypes.NewModuleAddress(govtypes.ModuleName).String())

				existingAcc := &statedb.Account{Nonce: uint64(1), Balance: common.Big1}
				balance := make([]uint8, 32)
				mockEVMKeeper.On("EstimateGas", mock.Anything, mock.Anything).Return(&evmtypes.EstimateGasResponse{Gas: uint64(200)}, nil)
				// first balance of
				mockEVMKeeper.On("ApplyMessage", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&evmtypes.MsgEthereumTxResponse{Ret: balance}, nil).Once()
				// convert coin
				mockEVMKeeper.On("ApplyMessage", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&evmtypes.MsgEthereumTxResponse{}, nil).Once()
				// second balance of
				mockEVMKeeper.On("ApplyMessage", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&evmtypes.MsgEthereumTxResponse{}, fmt.Errorf("third")).Once()
				// Extra call on test
				mockEVMKeeper.On("ApplyMessage", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&evmtypes.MsgEthereumTxResponse{}, nil)
				mockEVMKeeper.On("GetAccountWithoutBalance", mock.Anything, mock.Anything).Return(existingAcc, nil)
			},
			false,
		},
		{
			"fail - force fail second balance", 100, 10, 5,
			func() {
				mockEVMKeeper := &MockEVMKeeper{}
				sp, found := suite.app.ParamsKeeper.GetSubspace(types.ModuleName)
				suite.Require().True(found)
				suite.app.Erc20Keeper = keeper.NewKeeper(runtime.NewKVStoreService(suite.app.GetKey("erc20")), suite.app.AppCodec(), sp, suite.app.AccountKeeper, suite.app.BankKeeper, mockEVMKeeper, authtypes.NewModuleAddress(govtypes.ModuleName).String())

				existingAcc := &statedb.Account{Nonce: uint64(1), Balance: common.Big1}
				balance := make([]uint8, 32)
				mockEVMKeeper.On("EstimateGas", mock.Anything, mock.Anything).Return(&evmtypes.EstimateGasResponse{Gas: uint64(200)}, nil)
				// first balance of
				mockEVMKeeper.On("ApplyMessage", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&evmtypes.MsgEthereumTxResponse{Ret: balance}, nil).Once()
				// convert coin
				mockEVMKeeper.On("ApplyMessage", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&evmtypes.MsgEthereumTxResponse{}, nil).Once()
				// second balance of
				mockEVMKeeper.On("ApplyMessage", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&evmtypes.MsgEthereumTxResponse{Ret: balance}, nil).Once()
				// Extra call on test
				mockEVMKeeper.On("ApplyMessage", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&evmtypes.MsgEthereumTxResponse{}, nil)
				mockEVMKeeper.On("GetAccountWithoutBalance", mock.Anything, mock.Anything).Return(existingAcc, nil)
			},
			false,
		},
		{
			"fail - force fail unescrow", 100, 10, 5,
			func() {
				mockBankKeeper := &MockBankKeeper{}
				sp, found := suite.app.ParamsKeeper.GetSubspace(types.ModuleName)
				suite.Require().True(found)
				suite.app.Erc20Keeper = keeper.NewKeeper(runtime.NewKVStoreService(suite.app.GetKey("erc20")), suite.app.AppCodec(), sp, suite.app.AccountKeeper, mockBankKeeper, suite.app.EvmKeeper, authtypes.NewModuleAddress(govtypes.ModuleName).String())

				mockBankKeeper.On("SendCoinsFromModuleToAccount", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(fmt.Errorf("failed to unescrow"))
				mockBankKeeper.On("BlockedAddr", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(false)
				mockBankKeeper.On("GetBalance", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(sdk.Coin{Denom: "coin", Amount: sdkmath.OneInt()})
			},
			false,
		},
		{
			"fail - force fail balance after transfer", 100, 10, 5,
			func() {
				mockBankKeeper := &MockBankKeeper{}
				sp, found := suite.app.ParamsKeeper.GetSubspace(types.ModuleName)
				suite.Require().True(found)
				suite.app.Erc20Keeper = keeper.NewKeeper(runtime.NewKVStoreService(suite.app.GetKey("erc20")), suite.app.AppCodec(), sp, suite.app.AccountKeeper, mockBankKeeper, suite.app.EvmKeeper, authtypes.NewModuleAddress(govtypes.ModuleName).String())

				mockBankKeeper.On("SendCoinsFromModuleToAccount", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
				mockBankKeeper.On("BlockedAddr", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(false)
				mockBankKeeper.On("GetBalance", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(sdk.Coin{Denom: ibcBase, Amount: sdkmath.OneInt()})
			},
			false,
		},
	}
	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.mintFeeCollector = true
			suite.SetupTest()
			metadata, pair := suite.setupRegisterIBCVoucher()
			suite.Require().NotNil(metadata)
			suite.Require().NotNil(pair)

			// Precondition: Convert Coin to ERC20
			coins := sdk.NewCoins(sdk.NewCoin(ibcBase, sdkmath.NewInt(tc.mint)))
			sender := sdk.AccAddress(suite.address.Bytes())
			suite.app.BankKeeper.MintCoins(suite.ctx, types.ModuleName, coins)
			suite.app.BankKeeper.SendCoinsFromModuleToAccount(suite.ctx, types.ModuleName, sender, coins)
			msg := types.NewMsgConvertCoin(
				sdk.NewCoin(ibcBase, sdkmath.NewInt(tc.burn)),
				suite.address,
				sender,
			)

			ctx := sdk.WrapSDKContext(suite.ctx)
			_, err := suite.app.Erc20Keeper.ConvertCoin(ctx, msg)
			suite.Require().NoError(err, tc.name)
			suite.Commit()
			balance := suite.BalanceOf(common.HexToAddress(pair.Erc20Address), suite.address)
			cosmosBalance := suite.app.BankKeeper.GetBalance(suite.ctx, sender, metadata.Base)
			suite.Require().Equal(cosmosBalance.Amount.Int64(), sdkmath.NewInt(tc.mint-tc.burn).Int64())
			suite.Require().Equal(balance, big.NewInt(tc.burn))

			// Convert ERC20s back to Coins
			ctx = sdk.WrapSDKContext(suite.ctx)
			contractAddr := common.HexToAddress(pair.Erc20Address)
			msgConvertERC20 := types.NewMsgConvertERC20(
				sdkmath.NewInt(tc.reconvert),
				sender,
				contractAddr,
				suite.address,
			)

			tc.malleate()
			res, err := suite.app.Erc20Keeper.ConvertERC20(ctx, msgConvertERC20)
			expRes := &types.MsgConvertERC20Response{}
			suite.Commit()
			balance = suite.BalanceOf(contractAddr, suite.address)
			cosmosBalance = suite.app.BankKeeper.GetBalance(suite.ctx, sender, pair.Denom)
			if tc.expPass {
				suite.Require().NoError(err, tc.name)
				suite.Require().Equal(expRes, res)
				suite.Require().Equal(cosmosBalance.Amount.Int64(), sdkmath.NewInt(tc.mint-tc.burn+tc.reconvert).Int64())
				suite.Require().Equal(balance.(*big.Int).Int64(), big.NewInt(tc.burn-tc.reconvert).Int64())
			} else {
				suite.Require().Error(err, tc.name)
			}
		})
	}
	suite.mintFeeCollector = false
}

func (suite *KeeperTestSuite) TestMsgExecutionByProposal() {
	suite.SetupTest()

	// set denom
	stakingParams, err := suite.app.StakingKeeper.GetParams(suite.ctx)
	suite.Require().NoError(err)
	denom := stakingParams.BondDenom

	// change mindeposit for denom
	govParams, err := suite.app.GovKeeper.Params.Get(suite.ctx)
	suite.Require().NoError(err)
	govParams.MinDeposit = []sdk.Coin{sdk.NewCoin(denom, sdkmath.NewInt(1))}
	err = suite.app.GovKeeper.Params.Set(suite.ctx, govParams)
	suite.Require().NoError(err)

	// create account
	privKey, err := ethsecp256k1.GenerateKey()
	suite.Require().NoError(err)
	proposer := sdk.AccAddress(privKey.PubKey().Address().Bytes())

	// deligate to validator
	initAmount := sdkmath.NewInt(int64(math.Pow10(18)) * 2)
	initBalance := sdk.NewCoins(sdk.NewCoin(denom, initAmount))
	testutil.FundAccount(suite.app.BankKeeper, suite.ctx, proposer, initBalance)
	shares, err := suite.app.StakingKeeper.Delegate(suite.ctx, proposer, sdk.DefaultPowerReduction, stakingtypes.Unbonded, suite.validator, true)
	suite.Require().NoError(err)
	suite.Require().True(shares.GT(sdkmath.LegacyNewDec(0)))

	var (
		erc20Address = suite.DeployContractDirectBalanceManipulation(erc20Name, erc20Symbol).String()
	)

	testCases := []struct {
		name      string
		msg       sdk.Msg
		malleate  func()
		checkFunc func(uint64)
	}{
		{
			"ok - proposal MsgRegisterCoin",
			&types.MsgRegisterCoin{
				Authority:   authtypes.NewModuleAddress(govtypes.ModuleName).String(),
				Title:       "MsgRegisterCoin",
				Description: "MsgRegisterCoin test",
				Metadata: banktypes.Metadata{
					Description: "ATOM IBC voucher (channel 14)",
					Base:        ibcBase,
					// NOTE: Denom units MUST be increasing
					DenomUnits: []*banktypes.DenomUnit{
						{
							Denom:    ibcBase,
							Exponent: 0,
						},
					},
					Name:    "ATOM channel-14",
					Symbol:  "ibcATOM-14",
					Display: ibcBase,
				},
			},
			func() {
				err := suite.app.BankKeeper.MintCoins(suite.ctx, inflationtypes.ModuleName, sdk.Coins{sdk.NewInt64Coin(ibcBase, 1)})
				suite.Require().NoError(err)
			},
			func(proposalId uint64) {
				proposal, err := suite.app.GovKeeper.Proposals.Get(suite.ctx, proposalId)
				suite.Require().NoError(err)
				suite.Require().Equal(govtypesv1.ProposalStatus_PROPOSAL_STATUS_PASSED, proposal.Status)

				id := suite.app.Erc20Keeper.GetTokenPairID(suite.ctx, ibcBase)
				pair, ok := suite.app.Erc20Keeper.GetTokenPair(suite.ctx, id)
				suite.Require().True(ok)
				suite.Require().Equal(suite.app.Erc20Keeper.GetDenomMap(suite.ctx, pair.Denom), id)
				suite.Require().Equal(suite.app.Erc20Keeper.GetERC20Map(suite.ctx, common.HexToAddress(pair.Erc20Address)), id)
			},
		},
		{
			"ok - proposal MsgRegisterERC20",
			&types.MsgRegisterERC20{
				Authority:    authtypes.NewModuleAddress(govtypes.ModuleName).String(),
				Title:        "MsgRegisterERC20",
				Description:  "MsgRegisterERC20 test",
				Erc20Address: erc20Address,
			},
			func() {},
			func(proposalId uint64) {
				proposal, err := suite.app.GovKeeper.Proposals.Get(suite.ctx, proposalId)
				suite.Require().NoError(err)
				suite.Require().Equal(govtypesv1.ProposalStatus_PROPOSAL_STATUS_PASSED, proposal.Status)

				id := suite.app.Erc20Keeper.GetTokenPairID(suite.ctx, erc20Address)
				pair, ok := suite.app.Erc20Keeper.GetTokenPair(suite.ctx, id)
				suite.Require().True(ok)
				suite.Require().Equal(suite.app.Erc20Keeper.GetDenomMap(suite.ctx, pair.Denom), id)
				suite.Require().Equal(suite.app.Erc20Keeper.GetERC20Map(suite.ctx, common.HexToAddress(pair.Erc20Address)), id)
			},
		},
		{
			"ok - proposal MsgToggleTokenConversion",
			&types.MsgToggleTokenConversion{
				Authority:   authtypes.NewModuleAddress(govtypes.ModuleName).String(),
				Title:       "MsgToggleTokenConversion",
				Description: "MsgToggleTokenConversion test",
				Token:       erc20Address,
			},
			func() {},
			func(proposalId uint64) {
				proposal, err := suite.app.GovKeeper.Proposals.Get(suite.ctx, proposalId)
				suite.Require().NoError(err)
				suite.Require().Equal(govtypesv1.ProposalStatus_PROPOSAL_STATUS_PASSED, proposal.Status)

				id := suite.app.Erc20Keeper.GetTokenPairID(suite.ctx, erc20Address)
				pair, ok := suite.app.Erc20Keeper.GetTokenPair(suite.ctx, id)
				suite.Require().True(ok)
				suite.Require().Equal(pair.Enabled, false)
			},
		},
		{
			"ok - proposal MsgUpdateParams",
			&types.MsgUpdateParams{
				Authority: authtypes.NewModuleAddress(govtypes.ModuleName).String(),
				Params:    types.NewParams(false, false),
			},
			func() {},
			func(proposalId uint64) {
				changeParams := types.NewParams(false, false)

				proposal, err := suite.app.GovKeeper.Proposals.Get(suite.ctx, proposalId)
				suite.Require().NoError(err)
				suite.Require().Equal(govtypesv1.ProposalStatus_PROPOSAL_STATUS_PASSED, proposal.Status)
				suite.Require().Equal(suite.app.Erc20Keeper.GetParams(suite.ctx), changeParams)
			},
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			tc.malleate()

			// submit proposal
			proposal, err := suite.app.GovKeeper.SubmitProposal(suite.ctx, []sdk.Msg{tc.msg}, "", "test", "description", proposer, false)
			suite.Require().NoError(err)
			suite.Commit()

			ok, err := suite.app.GovKeeper.AddDeposit(suite.ctx, proposal.Id, proposer, govParams.MinDeposit)
			suite.Require().NoError(err)
			suite.Require().True(ok)
			suite.Commit()

			err = suite.app.GovKeeper.AddVote(suite.ctx, proposal.Id, proposer, govtypesv1.NewNonSplitVoteOption(govtypesv1.OptionYes), "")
			suite.Require().NoError(err)
			suite.CommitAfter(*govParams.VotingPeriod)

			// check proposal result
			tc.checkFunc(proposal.Id)
		})
	}
}
