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

// This test suite will run a simulation of sorts where transactions will have
// 1. invalid register events
// 2. valid register events
// 3. invalid update events
// 4. valid update events
// 5. csr enabled smart contracts making transactions
// 6. non-csr enabled smart contracts making transactions
// And every permutation of the above mixed together (csr-disabled/enabled smart contracts with valid/invalid update/register).
// This also tests the factory pattern for smart contracts and will validate that each of the fees are distributed correctly
// amongst the NFTs they belong to.
func (suite *KeeperTestSuite) TestCSRHook() {
	// Set up the test suite
	suite.SetupTest()
	suite.Commit()

	// Deploy test contracts (which are turnstiles for simplicity)
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

	// Set the price of gas
	price := int64(100)
	gasPrice := big.NewInt(price)

	// dynamically set the receipt and msg
	var (
		to      *common.Address
		receipt *ethtypes.Receipt
	)

	turnstileAddress, found := suite.app.CSRKeeper.GetTurnstile(suite.ctx)
	suite.Require().True(found)

	turnstile := contracts.TurnstileContract.ABI
	RegisterCSREvent := turnstile.Events["Register"]
	UpdateCSREvent := turnstile.Events["Assign"]

	// Used to check the expected balance of each NFT
	type nftCheck struct {
		nftID   uint64
		gasUsed uint64
	}

	type result struct {
		shouldReceiveFunds bool
		expectErr          bool
		cumulativeGasUsed  uint64 // cumulative revenue tracking for the turnstile address (expected revenue across all NFTs)
		nft                nftCheck
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
				to = &newAddress

				log := ethtypes.Log{}
				receipt = &ethtypes.Receipt{
					Logs: []*ethtypes.Log{&log},
				}
			},
			result{
				false,
				false,
				0,
				nftCheck{},
			},
		},
		{
			"Unregistered CSR contract (empty logs)", // -> this should effectively do nothing
			func() {
				newAddress := tests.GenerateAddress()
				to = &newAddress

				receipt = &ethtypes.Receipt{
					Logs: []*ethtypes.Log{},
				}
			},
			result{
				false,
				false,
				0,
				nftCheck{},
			},
		},
		{
			"Registered CSR contract (empty log)", // -> this should split the gas fee to the turnstile address
			func() {
				contract := common.HexToAddress(csrs[0].Contracts[0])
				to = &contract

				receipt = &ethtypes.Receipt{
					Logs:    []*ethtypes.Log{},
					GasUsed: 10,
				}
			},
			result{
				true,
				false,
				10,
				nftCheck{nftID: 0, gasUsed: 10},
			},
		},
		{
			"Registered CSR contract (single empty log)", // -> this should split the gas fee to the turnstile address
			func() {
				contract := common.HexToAddress(csrs[0].Contracts[0])
				to = &contract

				log := ethtypes.Log{}
				receipt = &ethtypes.Receipt{
					Logs:    []*ethtypes.Log{&log},
					GasUsed: 13,
				}
			},
			result{
				true,
				false,
				23,
				nftCheck{nftID: 0, gasUsed: 23},
			},
		},
		{
			"Unregistered CSR contract with register event with invalid smart contract", // -> this should through an error bc contract is not deployed
			func() {
				address := tests.GenerateAddress()
				to = &address

				account := tests.GenerateAddress()
				topics := []common.Hash{RegisterCSREvent.ID}
				data, _ := RegisterCSREvent.Inputs.Pack(address, account, big.NewInt(1))
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
				23,
				nftCheck{},
			},
		},
		{
			"Unregistered CSR contract with register event that has an invalid receiver address", // -> this should throw an error because the account sent to is not registered in evm db
			func() {

				address := tests.GenerateAddress()
				to = &address

				account := tests.GenerateAddress()
				topics := []common.Hash{RegisterCSREvent.ID}
				data, _ := RegisterCSREvent.Inputs.Pack(testContract, account, big.NewInt(1))
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
				23,
				nftCheck{},
			},
		},
		{
			"Unregistered CSR contract with valid register event", // -> this should not split fees but will create a new CSR
			func() {
				address := tests.GenerateAddress()
				to = &address

				sdkAccount := suite.CreateNewAccount(suite.ctx)
				account := common.BytesToAddress(sdkAccount.Bytes())
				topics := []common.Hash{RegisterCSREvent.ID}
				data, _ := RegisterCSREvent.Inputs.Pack(testContract, account, big.NewInt(1))
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
				23,
				nftCheck{},
			},
		},
		{
			"Registered smart contract (testContract)", // -> this should split the gas fee to the turnstile address
			func() {
				to = &testContract

				receipt = &ethtypes.Receipt{
					Logs:    []*ethtypes.Log{},
					GasUsed: 10,
				}
			},
			result{
				true,
				false,
				33,
				nftCheck{nftID: 1, gasUsed: 10},
			},
		},
		{
			"Unregistered smart contract with register event that has a duplicated address", // -> this should return an error because the contract is already registered
			func() {
				address := tests.GenerateAddress()
				to = &address

				sdkAccount := suite.CreateNewAccount(suite.ctx)
				account := common.BytesToAddress(sdkAccount.Bytes())
				topics := []common.Hash{RegisterCSREvent.ID}
				data, _ := RegisterCSREvent.Inputs.Pack(testContract, account, big.NewInt(1))
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
				33,
				nftCheck{},
			},
		},
		{
			"Registered smart contract (testContract) test CSR contract with a valid register event nested (testContract2)", // -> might be similar to a factory deployment, should split fees and then register
			func() {
				to = &testContract

				sdkAccount := suite.CreateNewAccount(suite.ctx)
				account := common.BytesToAddress(sdkAccount.Bytes())
				topics := []common.Hash{RegisterCSREvent.ID}
				data, _ := RegisterCSREvent.Inputs.Pack(testContract2, account, big.NewInt(2))
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
				46,
				nftCheck{nftID: 1, gasUsed: 23},
			},
		},
		{
			"Check if smart contract (testContract2) was registered via factory method from above", // -> should split fees to turnstile address
			func() {
				to = &testContract2

				receipt = &ethtypes.Receipt{
					Logs:    []*ethtypes.Log{},
					GasUsed: 10,
				}
			},
			result{
				true,
				false,
				56,
				nftCheck{nftID: 2, gasUsed: 10},
			},
		},
		{
			"Registered Smart contract with an invalid update event (invalid contract)", // -> should return an error because smart contract was not deployed
			func() {
				to = &testContract2

				address := tests.GenerateAddress()
				topics := []common.Hash{UpdateCSREvent.ID}
				data, _ := UpdateCSREvent.Inputs.Pack(address, big.NewInt(1))
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
				56,
				nftCheck{},
			},
		},
		{
			"Registered Smart contract with an invalid update event (invalid nft)", // -> should return an error because the NFT does not exist
			func() {
				to = &testContract2

				topics := []common.Hash{UpdateCSREvent.ID}
				data, _ := UpdateCSREvent.Inputs.Pack(testContract3, big.NewInt(100))
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
				56,
				nftCheck{},
			},
		},
		{
			"Registered Smart contract with an valid update event", //  -> should split fees and update the CSR NFT
			func() {
				to = &testContract2

				topics := []common.Hash{UpdateCSREvent.ID}
				data, _ := UpdateCSREvent.Inputs.Pack(testContract3, big.NewInt(1))
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
				61,
				nftCheck{nftID: 2, gasUsed: 15},
			},
		},
		{
			"Unregistered Smart contract with an invalid update event (invalid contract)", // -> should return an error because smart contract has not been deployed
			func() {
				address := tests.GenerateAddress()
				to = &address

				topics := []common.Hash{UpdateCSREvent.ID}
				data, _ := UpdateCSREvent.Inputs.Pack(address, big.NewInt(1))
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
				61,
				nftCheck{},
			},
		},
		{
			"Unregistered Smart contract with an invalid update event (duplicate contract)", // -> should return an duplicate contract error
			func() {
				address := tests.GenerateAddress()
				to = &address

				topics := []common.Hash{UpdateCSREvent.ID}
				data, _ := UpdateCSREvent.Inputs.Pack(testContract3, big.NewInt(1))
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
				61,
				nftCheck{},
			},
		},
		{
			"Unregistered Smart contract with an valid update event", // -> should update the CSR NFT
			func() {
				address := tests.GenerateAddress()
				to = &address

				topics := []common.Hash{UpdateCSREvent.ID}
				data, _ := UpdateCSREvent.Inputs.Pack(testContract4, big.NewInt(1))
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
				61,
				nftCheck{},
			},
		},
		{
			"Registered Smart Contract (testContract3)", // -> should split fees to the turnstile address
			func() {
				to = &testContract3

				receipt = &ethtypes.Receipt{
					Logs:    []*ethtypes.Log{},
					GasUsed: 7,
				}
			},
			result{
				true,
				false,
				68,
				nftCheck{nftID: 1, gasUsed: 30},
			},
		},
		{
			"Registered Smart Contract (testContract4)", // -> should split fees to the turnstile address
			func() {
				to = &testContract4

				receipt = &ethtypes.Receipt{
					Logs:    []*ethtypes.Log{},
					GasUsed: 7,
				}
			},
			result{
				true,
				false,
				75,
				nftCheck{nftID: 1, gasUsed: 37},
			},
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {

			tc.setUpMsg()

			msg := ethtypes.NewMessage(
				types.ModuleAddress,
				to,
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

			err := suite.app.CSRKeeper.Hooks().PostTxProcessing(suite.ctx, msg, receipt)
			if !tc.test.expectErr {
				suite.Require().NoError(err)
			} else {
				suite.Require().Error(err)
				return
			}

			if tc.test.shouldReceiveFunds {
				// Get the percentage of fees that should be going to the CSR nfts
				csrShare := suite.app.CSRKeeper.GetParams(suite.ctx).CsrShares

				// The contract should be mapped to some NFT
				contract := msg.To()
				nft, found := suite.app.CSRKeeper.GetNFTByContract(suite.ctx, contract.String())
				suite.Require().True(found)

				// The test NFT should match the one found
				testNFT := tc.test.nft.nftID
				gasUsedByNFT := tc.test.nft.gasUsed
				suite.Require().Equal(testNFT, nft)

				// The CSR object should be found
				_, found = suite.app.CSRKeeper.GetCSR(suite.ctx, nft)
				suite.Require().True(found)

				// Checking the turnstile balance
				turnstile := sdk.AccAddress(turnstileAddress.Bytes())
				turnstileBalance := suite.app.BankKeeper.GetAllBalances(suite.ctx, turnstile)

				// Ensuring the turnstile and expected turnstile balances match
				expectedTurnstileBalance := calculateExpectedFee(tc.test.cumulativeGasUsed, gasPrice, csrShare)
				suite.Require().Equal(expectedTurnstileBalance, turnstileBalance.AmountOf(evmDenom))

				// Get the balance of the revenue accumulated at the given NFT
				nftRevenue, err := getNFTRevenue(suite, &turnstileAddress, testNFT)
				suite.Require().NoError(err)

				// Check that the expected NFT balance matches the actual balance
				nftFee := calculateExpectedFee(gasUsedByNFT, gasPrice, csrShare)
				suite.Require().Equal(nftFee.BigInt(), nftRevenue)
			} else {
				contract := msg.To()
				_, found := suite.app.CSRKeeper.GetNFTByContract(suite.ctx, contract.String())
				suite.Require().False(found)
			}
		})
	}

}

// Helper function that will calculate how much revenue a NFT or the Turnstile should accumulate
// Calculation is done by the following: int(gasUsed * gasPrice * csrShares)
func calculateExpectedFee(gasUsed uint64, gasPrice *big.Int, csrShare sdk.Dec) sdk.Int {
	fee := sdk.NewIntFromUint64(gasUsed).Mul(sdk.NewIntFromBigInt(gasPrice))
	expectedTurnstileBalance := sdk.NewDecFromInt(fee).Mul(csrShare).TruncateInt()
	return expectedTurnstileBalance
}

// Helper function that will get the transaction revenue for a given NFT
func getNFTRevenue(suite *KeeperTestSuite, address *common.Address, nft uint64) (*big.Int, error) {
	// Call to retrieve the amount of canto for a given NFT
	resp, err := suite.app.CSRKeeper.CallMethod(suite.ctx, "revenue", contracts.TurnstileContract, types.ModuleAddress, address, big.NewInt(0), new(big.Int).SetUint64(nft))
	if err != nil {
		return nil, err
	}

	// Unpack the results into a big int
	unpackedData, err := turnstileContract.Methods["revenue"].Outputs.Unpack(resp.Ret)
	if err != nil {
		return nil, err
	}
	nftRevenue := unpackedData[0].(*big.Int)

	return nftRevenue, nil
}
