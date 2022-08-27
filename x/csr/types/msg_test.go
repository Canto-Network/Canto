package types_test

import (
	"testing"

	"github.com/Canto-Network/Canto/v2/x/csr/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
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
	poolAddress sdk.AccAddress
	deployer    sdk.AccAddress
	allocations map[string]uint64 // map between bech32 address and amount
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
	suite.poolAddress = sdk.AccAddress((tests.GenerateAddress().Bytes()))
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

func (suite *MsgTestSuite) TestMsgUpdateCSR() {
	msg := types.NewMsgUpdateCSR(
		suite.deployer,
		suite.poolAddress,
		suite.contracts,
		suite.nonces,
	)

	// check basic methods and validation
	suite.Require().Equal(types.RouterKey, msg.Route())
	suite.Require().Equal(types.TypeMsgUpdateCSR, msg.Type())
	suite.Require().NotNil(msg.GetSignBytes())
	suite.Require().Equal(msg.Deployer, msg.GetSigners()[0].String())
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
			tc.args.nftsupply,
			tc.args.allocations,
			suite.contracts,
			suite.nonces,
		)

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
		// construct MsgRegisterCSR
		msgRegister := types.NewMsgRegisterCSR(
			suite.deployer,
			suite.nftsupply,
			suite.allocations,
			tc.args.contracts,
			suite.nonces[:len(tc.args.contracts)],
		)

		// construct MsgUpdateCSR
		msgUpdate := types.NewMsgUpdateCSR(
			suite.deployer,
			suite.poolAddress,
			tc.args.contracts,
			msgRegister.ContractData.Nonces,
		)

		if tc.expectPass {
			suite.Require().NoError(msgRegister.ContractData.CheckContracts())
			// test that MsgRegisterCSR ValidateBasic Passes
			suite.Require().NoError(msgRegister.ValidateBasic())
			// test taht MsgUpdateCSR ValidateBasic also passes
			suite.Require().NoError(msgUpdate.ValidateBasic())
		} else {
			suite.Require().Error(msgRegister.ContractData.CheckContracts())
			// test that MsgRegisterCSR ValidateBasic fails
			suite.Require().Error(msgRegister.ValidateBasic())
			// test taht MsgUpdateCSR ValidateBsic also fails
			suite.Require().Error(msgUpdate.ValidateBasic())
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
		// construct MsgRegisterCSR
		msgRegister := types.NewMsgRegisterCSR(
			suite.deployer,
			suite.nftsupply,
			suite.allocations,
			suite.contracts[:len(tc.args.nonces)],
			tc.args.nonces,
		)
		// construct MsgUpdateCSR
		msgUpdate := types.NewMsgUpdateCSR(
			suite.deployer,
			suite.poolAddress,
			suite.contracts[:len(tc.args.nonces)],
			tc.args.nonces,
		)

		if tc.expectPass {
			suite.Require().NoError(msgRegister.ContractData.CheckNonces())
			// test that ValidateBasic also passes for MsgRegisterCSR
			suite.Require().NoError(msgRegister.ValidateBasic())
			// test that ValidateBasic also passes for MsgUpdateCSR
			suite.Require().NoError(msgUpdate.ValidateBasic())
		} else {
			suite.Require().Error(msgRegister.ContractData.CheckNonces())
			// test that ValidateBasic for MsgRegisterCSR fails
			suite.Require().Error(msgRegister.ValidateBasic())
			// test that Validate Basic for MsgUpdateCSR also fails
			suite.Require().Error(msgUpdate.ValidateBasic())
		}
	}
}

func (suite *MsgTestSuite) TestRegisterValidateBasic() {
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
		msg.ContractData.Contracts = msg.ContractData.Contracts[:tc.args.noncesLen]

		if tc.expectPass {
			suite.Require().NoError(msg.ValidateBasic())
		} else {
			suite.Require().Error(msg.ValidateBasic())
		}
	}
}

// test validate basic for MsgRegisterCSR
func (suite *MsgTestSuite) TestUpdateValidateBasic() {
	type testArgs struct {
		deployer  string
		poolAddr  string
		noncesLen int
	}

	testCases := []struct {
		name       string
		args       testArgs
		expectPass bool
	}{
		{
			"if deployer address is invalid - fail",
			testArgs{
				"x",
				sdk.AccAddress(tests.GenerateAddress().Bytes()).String(),
				5,
			},
			false,
		},
		{
			"if pool address is invalid - fail",
			testArgs{
				sdk.AccAddress(tests.GenerateAddress().Bytes()).String(),
				"x",
				5,
			},
			false,
		},
		{
			"if nonces/ contracts len is not equal - fail",
			testArgs{
				sdk.AccAddress(tests.GenerateAddress().Bytes()).String(),
				sdk.AccAddress(tests.GenerateAddress().Bytes()).String(),
				3,
			},
			false,
		},
		{
			"if deployer / poolAddress is valid - pass",
			testArgs{
				sdk.AccAddress(tests.GenerateAddress().Bytes()).String(),
				sdk.AccAddress(tests.GenerateAddress().Bytes()).String(),
				5,
			},
			true,
		},
	}

	for _, tc := range testCases {
		// construct MsgUpdateCSR
		msg := types.NewMsgUpdateCSR(
			suite.deployer,
			suite.poolAddress,
			suite.contracts,
			suite.nonces[:tc.args.noncesLen],
		)

		msg.Deployer = tc.args.deployer
		msg.PoolAddress = tc.args.poolAddr

		if tc.expectPass {
			// ValidateBasic should pass
			suite.Require().NoError(msg.ValidateBasic())
		} else {
			// ValidateBasic should fail
			suite.Require().Error(msg.ValidateBasic())
		}
	}
}

type WithdrawTestSuite struct {
	suite.Suite

	withdrawer sdk.AccAddress
	receiver   sdk.AccAddress
	csrPools   []string
	nftIds     []*UIntArray
}

// run the Withdrawal Test Suite
func TestWithdrawTestSuite(t *testing.T) {
	suite.Run(t, new(WithdrawTestSuite))
}

func (suite *WithdrawTestSuite) SetupTest() {
	// generate both deployer and receiver addresses
	suite.withdrawer = sdk.MustAccAddressFromBech32(generatePools(1)[0])
	suite.receiver = sdk.MustAccAddressFromBech32(generatePools(1)[0])
	// generate pools
	suite.csrPools = generatePools(5)
	// generate nftIds
	_, suite.nftIds = generateAddresses(tests.GenerateAddress(), 5)
}

// check that upon construction, the setters/getter methods are well-defined and return expected values
func (suite *WithdrawTestSuite) TestMsgWithdraw() {
	// construct a msgWithdraw
	msg := types.NewMsgWithdrawCSR(
		suite.withdrawer,
		suite.receiver,
		suite.csrPools,
		suite.nftIds,
	)
	// check that the type / routes are correct
	suite.Require().Equal(msg.Type(), types.TypeMsgWithdrawCSR)
	suite.Require().Equal(msg.Route(), types.RouterKey)
	// check that getSigners is the withdrawer address
	suite.Require().Equal(msg.GetSigners()[0].String(), msg.Withdrawer)
	// check that GetSignBytes is non-nil (not many better ways to test this :()
	suite.Require().NotNil(msg.GetSignBytes())
	// check that ValidateBasic Passes
	suite.Require().NoError(msg.ValidateBasic())
}

// check that if receiver is "", the receiver is made to the deployer's address
func (suite *WithdrawTestSuite) TestEmptyReceiverAdddr() {
	// construct msg Withdraw with empty receiver
	msg := types.NewMsgWithdrawCSR(
		suite.withdrawer,
		nil,
		suite.csrPools,
		suite.nftIds,
	)
	// check that the receiver address and withdrawer are the same
	suite.Require().Equal(msg.Withdrawer, msg.Receiver)
	// check that ValidateBasic passes
	suite.Require().NoError(msg.ValidateBasic())
}

// test checkAllocations, first,
// fail when there is invalid addr, fail when there is a repeated address
func (suite *WithdrawTestSuite) TestCheckPools() {
	type testArgs struct {
		poolAddrs []string
		nftIds    []*UIntArray
		setup     func(*types.MsgWithdrawCSR, []*UIntArray, []string)
	}

	testCases := []struct {
		name       string
		args       testArgs
		expectPass bool
	}{
		{
			"if repeated nftId in UintArray - fail",
			testArgs{
				suite.csrPools,
				[]*UIntArray{
					{
						[]uint64{uint64(1), uint64(1)},
					},
				},
				func(msg *types.MsgWithdrawCSR, nftIds []*UIntArray, poolAddrs []string) {
					// make the csrPools len shorter than the len(nftIds)
					msg.CsrPools = poolAddrs[:2]
					msg.Nfts = nftIds
					return
				},
			},
			false,
		},
		{
			"if there is an invalid sdk address in pool Addrs - fail",
			testArgs{
				// take all but last sdk address, last one will be empty (invalid address)
				append(suite.csrPools[:4], "x"),
				suite.nftIds,
				func(msg *types.MsgWithdrawCSR, nftIds []*UIntArray, poolAddrs []string) {
					// set pool addrs to what is passed to this func
					msg.CsrPools = poolAddrs
				},
			},
			false,
		},
	}

	for _, tc := range testCases {
		// instantiate msg
		msg := types.NewMsgWithdrawCSR(
			suite.withdrawer,
			suite.receiver,
			suite.csrPools,
			suite.nftIds,
		)
		// run setup for test
		tc.args.setup(msg, tc.args.nftIds, tc.args.poolAddrs)

		if tc.expectPass {
			// CheckPools should pass
			suite.Require().NoError(msg.CheckPools())
			// ValidateBasic should Pass
			suite.Require().NoError(msg.CheckPools())
		} else {
			// CheckPools should fail
			suite.Require().Error(msg.CheckPools())
			// validate basic should fail
			suite.Require().Error(msg.ValidateBasic())
		}
	}

}

// Test Validate Basic, fail if either receiver or withdrawer is an invalid accound address
// fail if the length of the pools / nftIds is not equal
func (suite *WithdrawTestSuite) TestValidateBasic() {
	type testArgs struct {
		withdrawer string
		receiver   string
		setup      func(msg *types.MsgWithdrawCSR, withdrawer, receiver string)
	}

	testCases := []struct {
		name       string
		args       testArgs
		expectPass bool
	}{
		{
			"if withdrawer address is invalid - fail",
			testArgs{
				"x",
				suite.receiver.String(),
				func(msg *types.MsgWithdrawCSR, withdrawer, receiver string) {
					msg.Withdrawer = withdrawer
				},
			},
			false,
		},
		{
			"if the receiver address is invalid - fail",
			testArgs{
				suite.withdrawer.String(),
				"x",
				func(msg *types.MsgWithdrawCSR, withdrawer, receiver string) {
					msg.Receiver = receiver
				},
			},
			false,
		},
		{
			"if the len of poolAddrs and nftIds is not equal fail",
			testArgs{
				"x",
				"x",
				func(msg *types.MsgWithdrawCSR, withdrawer, receiver string) {
					// cut off last 3 entries of pool
					msg.CsrPools = msg.CsrPools[:2]
				},
			},
			false,
		},
	}

	for _, tc := range testCases {
		// instantiate msg
		msg := types.NewMsgWithdrawCSR(
			suite.withdrawer,
			suite.receiver,
			suite.csrPools,
			suite.nftIds,
		)

		// setup test case
		tc.args.setup(msg, tc.args.withdrawer, "")

		if tc.expectPass {
			// expect ValidateBasic to pass
			suite.Require().NoError(msg.ValidateBasic())
		} else {
			// expect ValidateBasic to fail
			suite.Require().Error(msg.ValidateBasic())
		}
	}
}

// helper function to generate test addresses for CheckPools tests
func generatePools(numAccts int) []string {
	// generate pks
	accts := make([]string, numAccts)
	for i := 0; i < numAccts; i++ {
		// fill PrivKeyField
		pk := ed25519.GenPrivKey().PubKey()
		// generate account
		accts[i] = sdk.AccAddress(pk.Bytes()).String()
	}

	return accts
}
