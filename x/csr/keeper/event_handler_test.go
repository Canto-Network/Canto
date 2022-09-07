package keeper_test

import (
	"errors"
	"math/big"
	"strings"

	"github.com/Canto-Network/Canto/v2/contracts"
	_ "github.com/Canto-Network/Canto/v2/x/csr/keeper"
	"github.com/Canto-Network/Canto/v2/x/csr/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/evmos/ethermint/tests"
	_ "github.com/evmos/ethermint/tests"
	"github.com/evmos/ethermint/x/evm/statedb"
	evmtypes "github.com/evmos/ethermint/x/evm/types"
)

// if smart contract address is not a smart contract - fail
// if smart contract has already been registered - fail
// if the receiver address does not exist - fail
// if smart contract has not yet been registered and is a contract - pass
// check that csr has been set in state
func (suite *KeeperTestSuite) TestRegisterEvent() {
	type testArgs struct {
		SmartContractAddress common.Address
		Receiver             common.Address
	}
	suite.Commit()

	var (
		smartContractAddress = tests.GenerateAddress()
		receiver             = tests.GenerateAddress()
		turnstile, _         = suite.app.CSRKeeper.GetTurnstile(suite.ctx)
	)

	testCases := []struct {
		name       string
		args       testArgs
		expectPass bool
		setup      func()
	}{
		{
			"if smart contract address is not an account in statedb - fail",
			testArgs{
				tests.GenerateAddress(),
				tests.GenerateAddress(),
			},
			false,
			func() {},
		},
		{
			"if the smart contract address is an EOA - fail",
			testArgs{
				SmartContractAddress: smartContractAddress,
				Receiver:             receiver,
			},
			false,
			func() {
				// set smart contract address as an EVM account
				suite.app.EvmKeeper.SetAccount(suite.ctx, smartContractAddress, *statedb.NewEmptyAccount())
			},
		},
		{
			"user is attempting to register a contract that is already registered - fail",
			testArgs{
				SmartContractAddress: smartContractAddress,
				Receiver:             receiver,
			},
			false,
			func() {
				// set the smart contract address to a CSR
				csr := types.CSR{
					Id:        1,
					Contracts: []string{smartContractAddress.Hex()},
				}
				// set the CSR to state
				suite.app.CSRKeeper.SetCSR(suite.ctx, csr)
			},
		},
		{
			"the receiver address is not a valid EVM account",
			testArgs{
				SmartContractAddress: turnstile,
				Receiver:             receiver,
			},
			false,
			func() {
				// receiver is still not a valid account
			},
		},
		{
			"if the smart contract has not been registered yet - pass",
			testArgs{
				SmartContractAddress: turnstile,
				Receiver:             receiver,
			},
			true,
			func() {
				// set receiver to state
				suite.app.EvmKeeper.SetAccount(suite.ctx, receiver, *statedb.NewEmptyAccount())
			},
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			// setup test
			tc.setup()
			data, err := generateRegisterEventData(tc.args.SmartContractAddress, tc.args.Receiver)
			suite.Require().NoError(err)
			// process register CSREvent
			err = suite.app.CSRKeeper.RegisterCSREvent(suite.ctx, data)
			if tc.expectPass {
				suite.Require().NoError(err)
				// check that the CSR exists at nftId 1
				csr, found := suite.app.CSRKeeper.GetCSR(suite.ctx, 1)
				suite.Require().True(found)
				// contract address registered is correct
				suite.Require().Equal(strings.Compare(tc.args.SmartContractAddress.Hex(), csr.Contracts[0]), 0)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

// if smart contract address is not a smart contract - fail
// if smart contract has already been registered - fail
// if the csr appended to does not exist - fail
// if the csr and the smart contract exist - pass
func (suite *KeeperTestSuite) TestUpdateEvent() {
	type testArgs struct {
		smartContractAddress common.Address
		nftId                uint64
	}
	suite.Commit()

	var (
		smartContractAddress = tests.GenerateAddress()
		turnstile, _         = suite.app.CSRKeeper.GetTurnstile(suite.ctx)
	)

	testCases := []struct {
		name       string
		args       testArgs
		expectPass bool
		setup      func()
	}{
		{
			"if the smart contract address is not a smart contract - fails",
			testArgs{
				smartContractAddress: smartContractAddress,
				nftId:                1,
			},
			false,
			func() {
			},
		},
		{
			"if the smart contract has alredy been registered - fail",
			testArgs{
				smartContractAddress: smartContractAddress,
				nftId:                1,
			},
			false,
			func() {
				csr := types.CSR{
					Owner:     sdk.AccAddress(smartContractAddress.Bytes()).String(),
					Account:   sdk.AccAddress(smartContractAddress.Bytes()).String(),
					Id:        1,
					Contracts: []string{smartContractAddress.Hex()},
				}
				// set csr to state
				suite.app.CSRKeeper.SetCSR(suite.ctx, csr)
			},
		},
		{
			"if the csr appended to does not exist - fail",
			testArgs{
				smartContractAddress: turnstile,
				nftId:                2,
			},
			false,
			func() {},
		},
		{
			"if the csr appended to exists, and the contract registered exist - pass",
			testArgs{
				smartContractAddress: turnstile,
				nftId:                1,
			},
			true,
			func() {

			},
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			// setup test
			tc.setup()
			data, err := generateUpdateEventData(tc.args.smartContractAddress, tc.args.nftId)
			suite.Require().NoError(err)
			// process event
			err = suite.app.CSRKeeper.UpdateCSREvent(suite.ctx, data)
			if tc.expectPass {
				suite.Require().NoError(err)
				csr, found := suite.app.CSRKeeper.GetCSR(suite.ctx, 1)
				suite.Require().True(found)
				// contract address registered is correct
				suite.Require().Equal(strings.Compare(tc.args.smartContractAddress.Hex(), csr.Contracts[1]), 0)
			} else {
				suite.Require().Error(err)
			}
		})
	}

}

// test failure in the case that a withdrawal event has been received for a CSR that does not exist
// test failure in the case that a withdrawal event has been received with a recipient that does not exist
// test withdrawing zero return value
// test withdrawing positive rewards value
func (suite *KeeperTestSuite) TestWithdrawalEvent() {
	type testArgs struct {
		withdrawer    common.Address
		receiver      common.Address
		id            uint64
		expectRewards sdk.Coins
	}

	testCases := []struct {
		name       string
		args       testArgs
		expectPass bool
		setup      func(withdrawer, receiver common.Address, id uint64) ([]byte, error)
	}{
		{
			"A CSR that has not yet been committed to state - fail",
			testArgs{
				tests.GenerateAddress(),
				tests.GenerateAddress(),
				uint64(1),
				sdk.Coins{},
			},
			false,
			func(withdrawer, receiver common.Address, id uint64) ([]byte, error) {
				return generateWithdrawEventData(withdrawer, receiver, id)
			},
		},
		{
			"An invalid evm address as receiver - fail",
			testArgs{
				tests.GenerateAddress(),
				tests.GenerateAddress(),
				uint64(1),
				sdk.Coins{},
			},
			false,
			func(withdrawer, receiver common.Address, id uint64) ([]byte, error) {
				// first set CSR to state
				csr := types.CSR{
					Id: id,
				}
				suite.app.CSRKeeper.SetCSR(suite.ctx, csr)
				// generate event
				return generateWithdrawEventData(withdrawer, receiver, id)
			},
		},
		{
			"Withdrawing from a pool w zero rewards - pass",
			testArgs{
				tests.GenerateAddress(),
				tests.GenerateAddress(),
				uint64(1),
				sdk.Coins{
					{
						Denom:  "acanto",
						Amount: sdk.NewIntFromUint64(0),
					},
				},
			},
			true,
			func(withdrawer, receiver common.Address, id uint64) ([]byte, error) {
				// set the receiver account to state
				suite.app.EvmKeeper.SetAccount(suite.ctx, receiver, *statedb.NewEmptyAccount())
				// generate sdk Account
				acct := suite.app.CSRKeeper.CreateNewAccount(suite.ctx)
				// first set CSR to state
				csr := types.CSR{
					Id:      id,
					Account: acct.String(),
				}
				suite.app.CSRKeeper.SetCSR(suite.ctx, csr)
				// generate event
				return generateWithdrawEventData(withdrawer, receiver, id)
			},
		},
		{
			"Withdrawing from a pool w non-zero rewards - pass",
			testArgs{
				tests.GenerateAddress(),
				tests.GenerateAddress(),
				uint64(1),
				sdk.Coins{
					{
						Denom:  "acanto",
						Amount: sdk.NewIntFromUint64(100),
					},
				},
			},
			true,
			func(withdrawer, receiver common.Address, id uint64) ([]byte, error) {
				// set the receiver account to state
				suite.app.EvmKeeper.SetAccount(suite.ctx, receiver, *statedb.NewEmptyAccount())
				// generate sdk Account
				acct := suite.app.CSRKeeper.CreateNewAccount(suite.ctx)
				// send funds to the beneficiary
				coins := sdk.Coins{sdk.Coin{Denom: "acanto", Amount: sdk.NewInt(100)}}
				suite.app.BankKeeper.MintCoins(suite.ctx, types.ModuleName, coins)
				suite.app.BankKeeper.SendCoinsFromModuleToAccount(suite.ctx, types.ModuleName, acct, coins)
				// set CSR to state
				csr := types.CSR{
					Id:      id,
					Account: acct.String(),
				}
				suite.app.CSRKeeper.SetCSR(suite.ctx, csr)
				// generate event
				return generateWithdrawEventData(withdrawer, receiver, id)
			},
		},
	}

	for _, tc := range testCases {
		// setup test
		suite.Run(tc.name, func() {
			data, err := tc.setup(tc.args.withdrawer, tc.args.receiver, tc.args.id)
			// process Withdrawal Event
			suite.Require().NoError(err)
			err = suite.app.CSRKeeper.WithdrawalEvent(suite.ctx, data)
			if tc.expectPass {
				suite.Require().NoError(err)
				// check rewards
				rewards := suite.app.BankKeeper.GetBalance(suite.ctx, sdk.AccAddress(tc.args.receiver.Bytes()), "acanto")
				suite.Require().Equal(rewards.Amount.Uint64(), tc.args.expectRewards[0].Amount.Uint64())
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

func generateUpdateEventData(contract common.Address, nftId uint64) (data []byte, err error) {
	return generateEventData("UpdateCSREvent", contracts.TurnstileContract, contract, nftId)
}

func generateRegisterEventData(contract, receiver common.Address) (data []byte, err error) {
	return generateEventData("RegisterCSREvent", contracts.TurnstileContract, contract, receiver)
}

// generate Withdrawal event log data
func generateWithdrawEventData(withdrawer, receiver common.Address, id uint64) (data []byte, err error) {
	intVal := &big.Int{}
	return generateEventData("Withdrawal", contracts.CSRNFTContract, withdrawer, receiver, intVal.SetUint64(id))
}

// generate event creates data field for arbitrary transaction
// given a set of arguments an a method name, return the abi-encoded bytes
// of the packed event data, withdrawer, receiver, Id (not indexed)
func generateEventData(name string, contract evmtypes.CompiledContract, args ...interface{}) ([]byte, error) {
	//  retrieve arguments from contract
	var event abi.Event
	event, ok := contract.ABI.Events[name]
	if !ok {
		return nil, errors.New("cannot find event")
	}
	// ok now pack arguments
	data, err := event.Inputs.Pack(args...)
	if err != nil {
		return nil, err
	}

	return data, nil
}
