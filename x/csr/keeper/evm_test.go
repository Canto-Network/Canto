package keeper_test

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"log"

	_ "github.com/Canto-Network/Canto/v2/x/csr/keeper"
	"github.com/Canto-Network/Canto/v2/x/csr/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	evmtypes "github.com/evmos/ethermint/x/evm/types"
	_ "github.com/Canto-Network/Canto/v2/contracts"
)

// embed test contracts in test suite
var (
	//go:embed test_contracts/test_deploy.json
	testDeployJSON []byte
	// used for contract deployments
	bytecode = common.Hex2Bytes("0x608060405234801561001057600080fd5b5061001f61002460201b60201c565b610129565b7f142f41d272585cc7a6eae3dbcac228c0151c4c458c743eddab11b2c2fbac73553360405161005391906100fb565b60405180910390a1565b600082825260208201905092915050565b7f75706461746564206576656e7400000000000000000000000000000000000000600082015250565b60006100a4600d8361005d565b91506100af8261006e565b602082019050919050565b600073ffffffffffffffffffffffffffffffffffffffff82169050919050565b60006100e5826100ba565b9050919050565b6100f5816100da565b82525050565b6000604082019050818103600083015261011481610097565b905061012360208301846100ec565b92915050565b610175806101386000396000f3fe608060405234801561001057600080fd5b506004361061002b5760003560e01c80637b0cb83914610030575b600080fd5b61003861003a565b005b7f142f41d272585cc7a6eae3dbcac228c0151c4c458c743eddab11b2c2fbac7355336040516100699190610111565b60405180910390a1565b600082825260208201905092915050565b7f75706461746564206576656e7400000000000000000000000000000000000000600082015250565b60006100ba600d83610073565b91506100c582610084565b602082019050919050565b600073ffffffffffffffffffffffffffffffffffffffff82169050919050565b60006100fb826100d0565b9050919050565b61010b816100f0565b82525050565b6000604082019050818103600083015261012a816100ad565b90506101396020830184610102565b9291505056fea26469706673582212203907ed7d0b543881f2292494961d2548a3b5e14fac4b6823dbb85069899ea63364736f6c63430008100033")
	contract evmtypes.CompiledContract
)

// first test contract deployments

func (suite *KeeperTestSuite) TestContractDeployment() {
	type testArgs struct {
		methodName   string
		setup        func() (error, common.Address)
		expectReturn func(contract common.Address) bool
	}

	testCases := []struct {
		name       string
		args       testArgs
		expectPass bool
	}{
		{
			"module calls deploy1 on testDeploy contract",
			testArgs{
				"deploy1",
				func() (error, common.Address) {
					//  deploy test Contract
					addr := suite.DeployContract()
					acc := s.app.EvmKeeper.GetAccountWithoutBalance(suite.ctx, addr)
					suite.Require().True(acc.IsContract())
					// now call deploy1 on testContract and receive address
					_, err := suite.app.CSRKeeper.CallMethod(suite.ctx, "deploy1", contract, &addr)
					// now return ret and expect it to be an address
					if err != nil {
						return err, common.Address{}
					}

					return nil, addr
				},
				func(contract common.Address) bool {
					// retrieve nonce of the contract
					nonce := getNonce(contract.Bytes())
					// expected return, is CREATE address
					expectAddr := crypto.CreateAddress(contract, nonce-1)
					acc := suite.app.EvmKeeper.GetAccountWithoutBalance(suite.ctx, expectAddr)
					return acc != nil
				},
			},
			true,
		},
		{
			"module calls deploy2 with salt [32]byte{\"\"} on testDeploy contract",
			testArgs{
				"module calls deploy2 on testDeploy contract verify that contracts exist",
				func() (error, common.Address) {
					// deploy test Contract
					addr := suite.DeployContract()
					acc := suite.app.EvmKeeper.GetAccountWithoutBalance(suite.ctx, addr)
					suite.Require().True(acc.IsContract())
					//  now return ret and expect it to be an address
					salt := [32]byte{0x01}
					_, err := suite.app.CSRKeeper.CallMethod(suite.ctx, "deploy2", contract, &addr, salt)
					if err != nil {
						return err, common.Address{}
					}

					return nil, addr
				},
				func(contract common.Address) bool {
					// generate createAddress2 with correct data
					expectedAddr := crypto.CreateAddress2(contract, [32]byte{0x01}, crypto.Keccak256(bytecode))
					// retrieve account at address
					acc := s.app.EvmKeeper.GetAccountWithoutBalance(suite.ctx, expectedAddr)
					return acc != nil
				},
			},
			true,
		},
	}

	for _, tc := range testCases {
		// setup test
		err, addr := tc.args.setup()
		suite.Require().NoError(err)
		if tc.expectPass {
			suite.Require().True(tc.args.expectReturn(addr))
		} else {
			suite.Require().False(tc.args.expectReturn(addr))
		}
	}
}

// Test Address derivation with CREATE / CREATE2
func (suite *KeeperTestSuite) TestAddressDerivation() {
	type testArgs struct {
		setup  func() common.Address
		nonces []uint64
		salts  [][32]byte
	}

	testCases := []struct {
		name       string
		args       testArgs
		expectPass bool
	}{
		{
			"len of nonces / salts incorrect - fail",
			testArgs{
				func() common.Address {
					return common.Address{}
				},
				[]uint64{
					1,
				},
				[][32]byte{
					{},
					{},
				},
			},
			false,
		},
		{
			"contract deployed through create address - pass",
			testArgs{
				func() common.Address {
					addr := suite.DeployContract()
					// retrieve account at this address
					acc := suite.app.EvmKeeper.GetAccount(s.ctx, addr)
					// contract exists
					suite.Require().NotNil(acc.IsContract())
					//  now deploy contract from this address
					_, err := suite.app.CSRKeeper.CallMethod(suite.ctx, "deploy1", contract, &addr)
					suite.Require().NoError(err)
					suite.Commit()
					// return address of contract and 1, nonce when second contract was deployed
					return crypto.CreateAddress(addr, 1)
				},
				[]uint64{
					0,
					1,
				},
				[][32]byte{
					{},
					{},
				},
			},
			true,
		},
	}

	for _, tc := range testCases {
		//  run setup test
		fmt.Printf("Running Test: %s \n", tc.name)
		expectAddr := tc.args.setup()

		if tc.expectPass {
			err, addr := suite.app.CSRKeeper.DeriveAddress(suite.ctx, types.ModuleAddress, tc.args.nonces, tc.args.salts)
			suite.Require().NoError(err)
			suite.Require().True(addr == expectAddr)
		} else {
			err, _ := suite.app.CSRKeeper.DeriveAddress(suite.ctx, types.ModuleAddress, tc.args.nonces, tc.args.salts)
			suite.Require().Error(err)
		}
	}
}

// Test deployment of the Turnstile Contract
func (suite *KeeperTestSuite) TestDeployTurnstile() {
	// first deploy Turnstile
	addr, err := suite.app.CSRKeeper.DeployTurnstile(suite.ctx) 
	suite.Require().NoError(err)
	// now find the account in state indexed by addr
	acc := suite.app.EvmKeeper.GetAccountWithoutBalance(suite.ctx, addr)
	// code hash must not be nil
	suite.Require().NotEqual(acc.CodeHash, crypto.Keccak256(nil))
	// now index into state with code hash, 
	code := suite.app.EvmKeeper.GetCode(suite.ctx, common.BytesToHash(acc.CodeHash))
	// Turnstile code exists at address
	suite.Require().NotNil(code)
}

func init() {
	// unmarshal json object into contract object
	err := json.Unmarshal(testDeployJSON, &contract)
	if err != nil {
		// log error and quit
		log.Fatalf("ERROR:: %s", err.Error())
	}
}

// Deploy Contract, check that derived address is correct
func (suite *KeeperTestSuite) DeployContract() common.Address {
	// deploy compiled contract object and return address
	addr, err := suite.app.CSRKeeper.DeployContract(suite.ctx, contract, bytecode)
	if err != nil {
		// log this failure
		log.Fatalf("KeeperTestSuite::DeployContract:: Error: %s", err.Error())
		return common.Address{}
	}
	return addr
}
