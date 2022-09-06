package keeper_test

import (
	"errors"
	"math/big"

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
// if smart contract has not yet been registered and is a contract - pass
// check that csr has been set in state
func (suite *KeeperTestSuite) TestRegisterEvent() {

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
				suite.app.BankKeeper.SendCoinsFromModuleToAccount(suite.ctx, types.ModuleName, sdk.AccAddress(receiver.Bytes()), coins)
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
	}
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
