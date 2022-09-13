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

	turnstile := contracts.TurnstileContract.ABI

	RegisterCSREvent := turnstile.Events["Register"]
	UpdateCSREvent := turnstile.Events["Attach"]

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
			"Unregistered CSR contract (single empty log)", //  -> this should effectively do nothing
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
			"Unregistered CSR contract (empty logs)", // -> this should effectively do nothing
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
			"Registered CSR contract (empty log)", // -> this should split the gas fee to the turnstile address
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
			"Registered CSR contract (single empty log)", // -> this should split the gas fee to the turnstile address
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
			"Unregistered CSR contract with register event with invalid smart contract", // -> this should through an error bc contract is not deployed
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

				topics := []common.Hash{RegisterCSREvent.ID}
				data, _ := RegisterCSREvent.Inputs.Pack(address, account, uint64(1))
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
			"Unregistered CSR contract with register event that has an invalid receiver address", // -> this should throw an error because the account sent to is not registered in evm db
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

				topics := []common.Hash{RegisterCSREvent.ID}
				data, _ := RegisterCSREvent.Inputs.Pack(testContract, account, uint64(1))
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
			"Unregistered CSR contract with valid register event", // -> this should not split fees but will create a new CSR
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

				topics := []common.Hash{RegisterCSREvent.ID}
				data, _ := RegisterCSREvent.Inputs.Pack(testContract, account, uint64(1))
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
			"Registered smart contract (testContract)", // -> this should split the gas fee to the turnstile address
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
			"Unregistered smart contract with register event that has a duplicated address", // -> this should return an error because the contract is already registered
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

				topics := []common.Hash{RegisterCSREvent.ID}
				data, _ := RegisterCSREvent.Inputs.Pack(testContract, account, uint64(1))
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
			"Registered smart contract (testContract) test CSR contract with a valid register event nested (testContract2)", // -> might be similar to a factory deployment, should split fees and then register
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

				topics := []common.Hash{RegisterCSREvent.ID}
				data, _ := RegisterCSREvent.Inputs.Pack(testContract2, account, uint64(2))
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
			"Check if smart contract (testContract2) was registered via factory method from above", // -> should split fees to turnstile address
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
			"Registered Smart contract with an invalid update event (invalid contract)", // -> should return an error because smart contract was not deployed
			func() {
				address := tests.GenerateAddress()

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
				true,
				true,
				10,
			},
		},
		{
			"Registered Smart contract with an invalid update event (invalid nft)", // -> should return an error because the NFT does not exist
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
			"Registered Smart contract with an valid update event", //  -> should split fees and update the CSR NFT
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
			"Unregistered Smart contract with an invalid update event (invalid contract)", // -> should return an error because smart contract has not been deployed
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
			"Unregistered Smart contract with an invalid update event (duplicate contract)", // -> should return an duplicate contract error
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
			"Unregistered Smart contract with an valid update event", // -> should update the CSR NFT
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
			"Registered Smart Contract (testContract3)", // -> should split fees to the turnstile address
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
