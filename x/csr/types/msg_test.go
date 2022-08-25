package types_test

import (
	"testing"

	"github.com/Canto-Network/Canto/v2/x/csr/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/evmos/ethermint/tests"
	"github.com/stretchr/testify/suite"
)

type UIntArray = types.UIntArray

type MsgTestSuite struct {
	suite.Suite

	nftsupply   uint64
	deployer    sdk.AccAddress
	allocations map[string]uint64 // map between bech32 address and
	contracts   []string
	nonces      []*UIntArray
}

func TestMsgTestSuite(t *testing.T) {
	suite.Run(t, new(MsgTestSuite))
}

func (suite *MsgTestSuite) SetupTest() {
	suite.nftsupply = uint64(100)
	deployer := tests.GenerateAddress()
	suite.deployer = sdk.AccAddress(deployer.Bytes())
	suite.contracts, suite.nonces = generateAddresses(deployer, 5)
	suite.allocations = generateAllocations(3, []int{33, 33, 34})
}

func generateAddresses(deployer common.Address, len int) ([]string, []*UIntArray) {
	contracts := make([]string, len)
	nonces := make([]*UIntArray, len)
	for i := 0; i < len; i++ {
		// generate nonces
		nonces[i] = &UIntArray{Value: []uint64{uint64(i + 1)}}
		// generate contract addresses
		contracts[i] = crypto.CreateAddress(deployer, uint64(i+1)).String()
	}
	return contracts, nonces
}

func generateAllocations(numShares int, allocations []int) map[string]uint64 {
	alloc := make(map[string]uint64)
	for i := 0; i < numShares; i++ {
		alloc[sdk.AccAddress(tests.GenerateAddress().Bytes()).String()] = uint64(allocations[i])
	}
	return alloc
}

//Test Msg Instantiation
func (suite *MsgTestSuite) TestMsgRegisterCSR() {
	msg := types.NewMsgRegisterCSR(
		suite.deployer,
		suite.nftsupply,
		suite.allocations,
		suite.contracts,
		suite.nonces,
	)
	//check basic methods and validation
	suite.Require().Equal(types.RouterKey, msg.Route())
	suite.Require().Equal(types.TypeMsgRegisterCSR, msg.Type())
	suite.Require().NotNil(msg.GetSignBytes())
	suite.Require().Equal(suite.deployer.String(), msg.GetSigners()[0].String())
	suite.Require().NoError(msg.ValidateBasic())
}

//Test CheckAllocations
func (suite *MsgTestSuite) TestCheckAllocations() {
	type testArgs struct {
		allocations map[string]uint64
		nftsupply   uint64
	}

	acct1 := sdk.AccAddress(tests.GenerateAddress().Bytes()).String()
	acct2 := sdk.AccAddress(tests.GenerateAddress().Bytes()).String()
	acct3 := sdk.AccAddress(tests.GenerateAddress().Bytes()).String()

	testCases := []struct {
		name       string
		args       testArgs
		expectPass bool
	}{
		{
			"when allocations is less than nftsupply - fail",
			testArgs{
				map[string]uint64{acct1: uint64(12), acct2: uint64(7), acct3: uint64(2)},
				uint64(23),
			},
			false,
		},
		{
			"when allocations is greater than nftsupply - fail",
			testArgs{
				map[string]uint64{acct1: uint64(12), acct2: uint64(7), acct3: uint64(2)},
				uint64(20),
			},
			false,
		},
		{
			"when allocations is equal to nftsupply but an invalid address - fail",
			testArgs{
				map[string]uint64{acct1: uint64(12), "": uint64(7), acct3: uint64(1)},
				uint64(20),
			},
			false,
		},
		{
			"when allocations is equal to  nftsupply and all addresses are valid - pass ",
			testArgs{
				map[string]uint64{acct1: uint64(12), acct2: uint64(7), acct3: uint64(1)},
				uint64(20),
			},
			true,
		},
	}

	// now process testCases
	for _, tc := range testCases {
		// construct Msg
		msg := types.NewMsgRegisterCSR(
			suite.deployer,
			suite.nftsupply,
			suite.allocations,
			suite.contracts,
			suite.nonces,
		)
		msg.Allocations = tc.args.allocations
		msg.NftSupply = tc.args.nftsupply

		if tc.expectPass {
			suite.Require().NoError(msg.CheckAllocations())
			// test that validate basic also passes
			suite.Require().NoError(msg.ValidateBasic())
		} else {
			suite.Require().Error(msg.CheckAllocations())
			// test that validate Basic also fails
			suite.Require().Error(msg.ValidateBasic())
		}
	}
}

//Test checkContracts
func (suite *MsgTestSuite) TestCheckContracts() {
	type testArgs struct {
		contracts []string
	}

	addr1 := tests.GenerateAddress().String()
	addr2 := tests.GenerateAddress().String()

	testCases := []struct {
		name       string
		args       testArgs
		expectPass bool
	}{
		{
			"if there is a non-existent address - fail",
			testArgs{
				[]string{addr1, ""},
			},
			false,
		},
		{
			"if there is an incorrectly formatted address - fail",
			testArgs{
				[]string{addr1, addr2[4:]},
			},
			false,
		},
		{
			"if there is a zero-address - fail",
			testArgs{
				[]string{addr1, "0x0000000000000000000000000000000000000000"},
			},
			false,
		},
		{
			"if all addresses are correctly formatted - pass",
			testArgs{
				[]string{addr1, addr2},
			},
			true,
		},
	}

	for _, tc := range testCases {
		// construct Msg
		msg := types.NewMsgRegisterCSR(
			suite.deployer,
			suite.nftsupply,
			suite.allocations,
			suite.contracts,
			suite.nonces,
		)
		// overwrite contracts
		msg.Contracts = tc.args.contracts
		//adjust length of nonces to correct length so msg.ValidateBasic does not fail
		msg.Nonces = msg.Nonces[:len(msg.Contracts)]
		if tc.expectPass {
			suite.Require().NoError(msg.CheckContracts())
			// test that ValidateBasic also Passes
			suite.Require().NoError(msg.ValidateBasic())
		} else {
			suite.Require().Error(msg.CheckContracts())
			// test that ValidateBasic also fails
			suite.Require().Error(msg.ValidateBasic())
		}
	}

}

//Test checkNonces
func (suite *MsgTestSuite) TestCheckNonces() {
	type testArgs struct {
		nonces []*UIntArray
	}

	testCases := []struct {
		name       string
		args       testArgs
		expectPass bool
	}{
		{
			"if any of the nonces are less than 1 - fail",
			testArgs{
				[]*UIntArray{
					{
						Value: []uint64{
							0, 1, 2, 3,
						},
					},
					{
						Value: []uint64{
							1, 2, 3, 4,
						},
					},
				},
			},
			false,
		},
		{
			"if all nonces are greater than 1 - pass",
			testArgs{
				[]*UIntArray{
					{
						Value: []uint64{
							1, 2, 3, 4,
						},
					},
					{
						Value: []uint64{
							1, 2, 3, 4,
						},
					},
				},
			},
			true,
		},
	}

	for _, tc := range testCases {
		// instantiate msg
		// construct Msg
		msg := types.NewMsgRegisterCSR(
			suite.deployer,
			suite.nftsupply,
			suite.allocations,
			suite.contracts,
			suite.nonces,
		)
		// overwrite Nonces in message object
		msg.Nonces = tc.args.nonces
		msg.Contracts = msg.Contracts[:len(msg.Nonces)]
		if tc.expectPass {
			suite.Require().NoError(msg.CheckNonces())
			// test that ValidateBasic also passes
			suite.Require().NoError(msg.ValidateBasic())
		} else {
			suite.Require().Error(msg.CheckNonces())
			// test that ValidateBasic also fails
			suite.Require().Error(msg.ValidateBasic())
		}
	}
}

func (suite *MsgTestSuite) TestValidateBasic() {
	type testArgs struct {
		deployer  string
		NFTSupply uint64
		noncesLen uint64
	}

	testCases := []struct {
		name       string
		args       testArgs
		expectPass bool
	}{
		{
			"if contracts / nonces are not same length - fail",
			testArgs{
				deployer:  "",
				NFTSupply: uint64(100),
				noncesLen: 3,
			},
			false,
		},
		{
			"if deployer address is invalid - fail",
			testArgs{
				deployer:  "x",
				NFTSupply: uint64(100),
				noncesLen: 5,
			},
			false,
		},
		{
			"if NFT supply is 0 - fail",
			testArgs{
				deployer:  "",
				NFTSupply: uint64(0),
				noncesLen: 5,
			},
			false,
		},
		{
			"if none if msg setupTest msg - pass",
			testArgs{
				deployer:  "",
				NFTSupply: uint64(100),
				noncesLen: 5,
			},
			true,
		},
	}

	for _, tc := range testCases {
		// construct Msg
		msg := types.NewMsgRegisterCSR(
			suite.deployer,
			suite.nftsupply,
			suite.allocations,
			suite.contracts,
			suite.nonces,
		)
		if tc.args.deployer != "" {
			msg.Deployer = tc.args.deployer
		}

		msg.NftSupply = tc.args.NFTSupply
		msg.Contracts = msg.Contracts[:tc.args.noncesLen]

		if tc.expectPass {
			suite.Require().NoError(msg.ValidateBasic())
		} else {
			suite.Require().Error(msg.ValidateBasic())
		}
	}
}
