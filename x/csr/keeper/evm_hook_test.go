package keeper_test

import (
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/evmos/ethermint/tests"

	"github.com/Canto-Network/Canto/v2/contracts"

	csrTypes "github.com/Canto-Network/Canto/v2/x/csr/types"

	"github.com/Canto-Network/Canto/v2/x/erc20/types"
)

func (suite *KeeperTestSuite) TestCSRHook() {
	// Set up the test suite
	suite.SetupTest()
	suite.Commit()

	// Deploy a test contract (which is just another turnstile)
	testContract, _ := suite.app.CSRKeeper.DeployTurnstile(suite.ctx)
	suite.Commit()

	// Send some initial funds to the fee module account
	evmDenom := suite.app.EvmKeeper.GetParams(suite.ctx).EvmDenom
	coins := sdk.Coins{{Denom: evmDenom, Amount: sdk.NewIntFromUint64(1000000000)}}
	suite.app.BankKeeper.MintCoins(suite.ctx, csrTypes.ModuleName, coins)
	suite.app.BankKeeper.SendCoinsFromModuleToModule(suite.ctx, csrTypes.ModuleName, suite.app.CSRKeeper.FeeCollectorName, coins)

	// Generate some CSRs that will be used in the store
	numberCSRs := 1
	csrs := GenerateCSRs(numberCSRs)
	for _, csr := range csrs {
		suite.app.CSRKeeper.SetCSR(suite.ctx, csr)
	}

	price := int64(100)
	gasPrice := big.NewInt(price) // gasPrice

	var (
		receipt *ethtypes.Receipt
		msg     ethtypes.Message
	)

	turnstileAddress, found := suite.app.CSRKeeper.GetTurnstile(suite.ctx)
	suite.Require().True(found)
	turnstile := contracts.TurnstileContract.ABI
	// nft := contracts.CSRNFTContract.ABI

	RegisterCSREvent := turnstile.Events["RegisterCSREvent"]
	// UpdateCSREvent := turnstile.Events["UpdateCSREvent"]
	// WithdrawalEvent := nft.Events["Withdrawal"]

	type result struct {
		shouldReceiveFunds bool
		gasUsed            uint64 // cumulative tracking for a particular nft
	}

	testCases := []struct {
		name     string
		setUpMsg func()
		test     result
	}{
		{
			"Unregistered CSR contract (single empty log)",
			func() {
				newAddress := tests.GenerateAddress()
				msg = ethtypes.NewMessage(
					types.ModuleAddress,
					&newAddress,
					0,
					big.NewInt(0), // amount
					uint64(0),     // gasLimit
					gasPrice,      // gasPrice
					big.NewInt(0), // gasFeeCap
					big.NewInt(0), // gasTipCap
					[]byte{},
					ethtypes.AccessList{}, // AccessList
					true,                  // checkNonce
				)

				log := ethtypes.Log{}
				receipt = &ethtypes.Receipt{
					Logs: []*ethtypes.Log{&log},
				}
			},
			result{
				false,
				0,
			},
		},
		{
			"Unregistered CSR contract (empty logs)",
			func() {
				newAddress := tests.GenerateAddress()
				msg = ethtypes.NewMessage(
					types.ModuleAddress,
					&newAddress,
					0,
					big.NewInt(0), // amount
					uint64(0),     // gasLimit
					gasPrice,      // gasPrice
					big.NewInt(0), // gasFeeCap
					big.NewInt(0), // gasTipCap
					[]byte{},
					ethtypes.AccessList{}, // AccessList
					true,                  // checkNonce
				)

				receipt = &ethtypes.Receipt{
					Logs: []*ethtypes.Log{},
				}
			},
			result{
				false,
				0,
			},
		},
		{
			"Registered CSR contract (empty log)",
			func() {
				contract := csrs[0].Contracts[0]
				address := common.HexToAddress(contract)
				msg = ethtypes.NewMessage(
					types.ModuleAddress,
					&address,
					0,
					big.NewInt(0), // amount
					uint64(0),     // gasLimit
					gasPrice,      // gasPrice
					big.NewInt(0), // gasFeeCap
					big.NewInt(0), // gasTipCap
					[]byte{},
					ethtypes.AccessList{}, // AccessList
					true,                  // checkNonce
				)

				receipt = &ethtypes.Receipt{
					Logs:    []*ethtypes.Log{},
					GasUsed: 1,
				}
			},
			result{
				true,
				1,
			},
		},
		{
			"Registered CSR contract (single empty log)",
			func() {
				contract := csrs[0].Contracts[0]
				address := common.HexToAddress(contract)
				msg = ethtypes.NewMessage(
					types.ModuleAddress,
					&address,
					0,
					big.NewInt(0), // amount
					uint64(0),     // gasLimit
					gasPrice,      // gasPrice
					big.NewInt(0), // gasFeeCap
					big.NewInt(0), // gasTipCap
					[]byte{},
					ethtypes.AccessList{}, // AccessList
					true,                  // checkNonce
				)

				log := ethtypes.Log{}
				receipt = &ethtypes.Receipt{
					Logs:    []*ethtypes.Log{&log},
					GasUsed: 10,
				}
			},
			result{
				true,
				11,
			},
		},
		{
			"Unregistered CSR contract (register contract not in state db) event  in logs)",
			func() {
				account := tests.GenerateAddress()

				address := tests.GenerateAddress()
				msg = ethtypes.NewMessage(
					types.ModuleAddress,
					&address,
					0,
					big.NewInt(0), // amount
					uint64(0),     // gasLimit
					gasPrice,      // gasPrice
					big.NewInt(0), // gasFeeCap
					big.NewInt(0), // gasTipCap
					[]byte{},
					ethtypes.AccessList{}, // AccessList
					true,                  // checkNonce
				)

				topics := []common.Hash{RegisterCSREvent.ID, address.Hash(), account.Hash()}
				data, _ := RegisterCSREvent.Inputs.Pack(address, account)
				log := ethtypes.Log{
					Address: turnstileAddress,
					Topics:  topics,
					Data:    data,
				}
				receipt = &ethtypes.Receipt{
					Logs:    []*ethtypes.Log{&log},
					GasUsed: 10,
				}
			},
			result{
				false,
				0,
			},
		},
		{
			"Unregistered CSR contract (register contract, receiver not in state db) event  in logs)",
			func() {
				account := tests.GenerateAddress()

				address := tests.GenerateAddress()
				msg = ethtypes.NewMessage(
					types.ModuleAddress,
					&address,
					0,
					big.NewInt(0), // amount
					uint64(0),     // gasLimit
					gasPrice,      // gasPrice
					big.NewInt(0), // gasFeeCap
					big.NewInt(0), // gasTipCap
					[]byte{},
					ethtypes.AccessList{}, // AccessList
					true,                  // checkNonce
				)

				topics := []common.Hash{RegisterCSREvent.ID, testContract.Hash(), account.Hash()}
				data, _ := RegisterCSREvent.Inputs.Pack(testContract, account)
				log := ethtypes.Log{
					Address: turnstileAddress,
					Topics:  topics,
					Data:    data,
				}
				receipt = &ethtypes.Receipt{
					Logs:    []*ethtypes.Log{&log},
					GasUsed: 10,
				}
			},
			result{
				false,
				0,
			},
		},
		{
			"Unregistered CSR contract (register contract, receiver) event  in logs)",
			func() {
				sdkAccount := suite.app.CSRKeeper.CreateNewAccount(suite.ctx)
				account := common.BytesToAddress(sdkAccount.Bytes())

				address := tests.GenerateAddress()
				msg = ethtypes.NewMessage(
					types.ModuleAddress,
					&address,
					0,
					big.NewInt(0), // amount
					uint64(0),     // gasLimit
					gasPrice,      // gasPrice
					big.NewInt(0), // gasFeeCap
					big.NewInt(0), // gasTipCap
					[]byte{},
					ethtypes.AccessList{}, // AccessList
					true,                  // checkNonce
				)

				topics := []common.Hash{RegisterCSREvent.ID, testContract.Hash(), account.Hash()}
				data, _ := RegisterCSREvent.Inputs.Pack(testContract, account)
				log := ethtypes.Log{
					Address: turnstileAddress,
					Topics:  topics,
					Data:    data,
				}
				receipt = &ethtypes.Receipt{
					Logs:    []*ethtypes.Log{&log},
					GasUsed: 10,
				}
			},
			result{
				false,
				0,
			},
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {

			tc.setUpMsg()

			err := suite.app.CSRKeeper.Hooks().PostTxProcessing(suite.ctx, msg, receipt)
			suite.Require().NoError(err)

			if tc.test.shouldReceiveFunds {
				contract := msg.To()
				nft, found := suite.app.CSRKeeper.GetNFTByContract(suite.ctx, contract.String())
				suite.Require().True(found)

				csr, found := suite.app.CSRKeeper.GetCSR(suite.ctx, nft)
				suite.Require().True(found)

				beneficiary := csr.Account
				cosmosBalance := suite.app.BankKeeper.GetAllBalances(suite.ctx, sdk.AccAddress(beneficiary))

				fee := sdk.NewIntFromUint64(tc.test.gasUsed).Mul(sdk.NewIntFromBigInt(msg.GasPrice()))
				developerFee := sdk.NewDecFromInt(fee).Mul(suite.app.CSRKeeper.GetParams(suite.ctx).CsrShares)
				expected := developerFee.TruncateInt()
				suite.Require().Equal(expected, cosmosBalance.AmountOf(evmDenom))
			} else {
				contract := msg.To()
				_, found := suite.app.CSRKeeper.GetNFTByContract(suite.ctx, contract.String())
				suite.Require().False(found)
			}
		})
	}

}
