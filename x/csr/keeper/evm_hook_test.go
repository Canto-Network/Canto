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

	// Deploy test contracts (which are turnstiles)
	testContract, _ := suite.app.CSRKeeper.DeployTurnstile(suite.ctx)
	testContract2, _ := suite.app.CSRKeeper.DeployTurnstile(suite.ctx)
	testContract3, _ := suite.app.CSRKeeper.DeployTurnstile(suite.ctx)
	testContract4, _ := suite.app.CSRKeeper.DeployTurnstile(suite.ctx)
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
	csrNFTAddress, found := suite.app.CSRKeeper.GetCSRNFT(suite.ctx)
	suite.Require().True(found)

	turnstile := contracts.TurnstileContract.ABI
	nft := contracts.CSRNFTContract.ABI

	RegisterCSREvent := turnstile.Events["RegisterCSREvent"]
	UpdateCSREvent := turnstile.Events["UpdateCSREvent"]
	WithdrawalEvent := nft.Events["Withdrawal"]

	type result struct {
		shouldReceiveFunds bool
		expectErr          bool
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
				false,
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
				false,
				11,
			},
		},
		{
			"Unregistered CSR contract (register contract not in state db) event in logs)",
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
				true,
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
				true,
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
				false,
				0,
			},
		},
		{
			"Register test CSR contract (no events)",
			func() {
				msg = ethtypes.NewMessage(
					types.ModuleAddress,
					&testContract,
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
					GasUsed: 10,
				}
			},
			result{
				true,
				false,
				10,
			},
		},
		{
			"Register test CSR contract (register duplicate smart contract event) -> might be similar to a factory deployment",
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
				true,
				10,
			},
		},
		{
			"Register test CSR contract (register smart contract event) -> might be similar to a factory deployment",
			func() {
				sdkAccount := suite.app.CSRKeeper.CreateNewAccount(suite.ctx)
				account := common.BytesToAddress(sdkAccount.Bytes())

				msg = ethtypes.NewMessage(
					types.ModuleAddress,
					&testContract,
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

				topics := []common.Hash{RegisterCSREvent.ID, testContract2.Hash(), account.Hash()}
				data, _ := RegisterCSREvent.Inputs.Pack(testContract2, account)
				log := ethtypes.Log{
					Address: turnstileAddress,
					Topics:  topics,
					Data:    data,
				}
				receipt = &ethtypes.Receipt{
					Logs:    []*ethtypes.Log{&log},
					GasUsed: 13,
				}
			},
			result{
				true,
				false,
				23,
			},
		},
		{
			"Check if smart contract was registered via factory method from above",
			func() {
				msg = ethtypes.NewMessage(
					types.ModuleAddress,
					&testContract2,
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
					GasUsed: 10,
				}
			},
			result{
				true,
				false,
				10,
			},
		},
		{
			"Registered Smart contract with an invalid update event (invalid contract)",
			func() {
				addr := tests.GenerateAddress()

				msg = ethtypes.NewMessage(
					types.ModuleAddress,
					&testContract2,
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

				topics := []common.Hash{UpdateCSREvent.ID}
				data, _ := UpdateCSREvent.Inputs.Pack(addr, uint64(100))
				log := ethtypes.Log{
					Address: turnstileAddress,
					Topics:  topics,
					Data:    data,
				}
				receipt = &ethtypes.Receipt{
					Logs:    []*ethtypes.Log{&log},
					GasUsed: 0,
				}
			},
			result{
				true,
				true,
				10,
			},
		},
		{
			"Registered Smart contract with an invalid update event (invalid nft)",
			func() {
				msg = ethtypes.NewMessage(
					types.ModuleAddress,
					&testContract2,
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

				topics := []common.Hash{UpdateCSREvent.ID}
				data, _ := UpdateCSREvent.Inputs.Pack(testContract3, uint64(100))
				log := ethtypes.Log{
					Address: turnstileAddress,
					Topics:  topics,
					Data:    data,
				}
				receipt = &ethtypes.Receipt{
					Logs:    []*ethtypes.Log{&log},
					GasUsed: 0,
				}
			},
			result{
				true,
				true,
				10,
			},
		},
		{
			"Registered Smart contract with an valid update event",
			func() {
				msg = ethtypes.NewMessage(
					types.ModuleAddress,
					&testContract2,
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

				topics := []common.Hash{UpdateCSREvent.ID}
				data, _ := UpdateCSREvent.Inputs.Pack(testContract3, uint64(1))
				log := ethtypes.Log{
					Address: turnstileAddress,
					Topics:  topics,
					Data:    data,
				}
				receipt = &ethtypes.Receipt{
					Logs:    []*ethtypes.Log{&log},
					GasUsed: 5,
				}
			},
			result{
				true,
				false,
				15,
			},
		},
		{
			"Unregistered Smart contract with an invalid update event (invalid contract)",
			func() {
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

				topics := []common.Hash{UpdateCSREvent.ID}
				data, _ := UpdateCSREvent.Inputs.Pack(address, uint64(1))
				log := ethtypes.Log{
					Address: turnstileAddress,
					Topics:  topics,
					Data:    data,
				}
				receipt = &ethtypes.Receipt{
					Logs:    []*ethtypes.Log{&log},
					GasUsed: 0,
				}
			},
			result{
				false,
				true,
				0,
			},
		},
		{
			"Unregistered Smart contract with an invalid update event (duplicate contract)",
			func() {
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

				topics := []common.Hash{UpdateCSREvent.ID}
				data, _ := UpdateCSREvent.Inputs.Pack(testContract3, uint64(1))
				log := ethtypes.Log{
					Address: turnstileAddress,
					Topics:  topics,
					Data:    data,
				}
				receipt = &ethtypes.Receipt{
					Logs:    []*ethtypes.Log{&log},
					GasUsed: 0,
				}
			},
			result{
				false,
				true,
				0,
			},
		},
		{
			"Unregistered Smart contract with an valid update event",
			func() {
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

				topics := []common.Hash{UpdateCSREvent.ID}
				data, _ := UpdateCSREvent.Inputs.Pack(testContract4, uint64(1))
				log := ethtypes.Log{
					Address: turnstileAddress,
					Topics:  topics,
					Data:    data,
				}
				receipt = &ethtypes.Receipt{
					Logs:    []*ethtypes.Log{&log},
					GasUsed: 0,
				}
			},
			result{
				false,
				false,
				0,
			},
		},
		{
			"Registered Smart Contract test3",
			func() {
				msg = ethtypes.NewMessage(
					types.ModuleAddress,
					&testContract3,
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
					GasUsed: 7,
				}
			},
			result{
				true,
				false,
				30,
			},
		},
		{
			"Unregistered Smart Contract with valid Withdraw event",
			func() {
				withdrawer := common.BytesToAddress(suite.app.CSRKeeper.CreateNewAccount(suite.ctx).Bytes())
				receiver := common.BytesToAddress(suite.app.CSRKeeper.CreateNewAccount(suite.ctx).Bytes())

				msg = ethtypes.NewMessage(
					types.ModuleAddress,
					&testContract,
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

				topics := []common.Hash{WithdrawalEvent.ID}
				data, _ := WithdrawalEvent.Inputs.Pack(withdrawer, receiver, big.NewInt(1))
				log := ethtypes.Log{
					Address: csrNFTAddress,
					Topics:  topics,
					Data:    data,
				}
				receipt = &ethtypes.Receipt{
					Logs:    []*ethtypes.Log{&log},
					GasUsed: 1,
				}
			},
			result{
				true,
				false,
				1,
			},
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {

			tc.setUpMsg()

			err := suite.app.CSRKeeper.Hooks().PostTxProcessing(suite.ctx, msg, receipt)
			if !tc.test.expectErr {
				suite.Require().NoError(err)
			} else {
				suite.Require().Error(err)
				return
			}

			if tc.test.shouldReceiveFunds {
				contract := msg.To()
				nft, found := suite.app.CSRKeeper.GetNFTByContract(suite.ctx, contract.String())
				suite.Require().True(found)

				csr, found := suite.app.CSRKeeper.GetCSR(suite.ctx, nft)
				suite.Require().True(found)

				beneficiary := csr.Beneficiary
				cosmosBalance := suite.app.BankKeeper.GetAllBalances(suite.ctx, sdk.MustAccAddressFromBech32(beneficiary))

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
