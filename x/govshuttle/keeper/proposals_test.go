package keeper_test

import (
	"fmt"
	"math/big"
	"reflect"
	"testing"

	"github.com/Canto-Network/Canto/v6/x/govshuttle/keeper"
	"github.com/Canto-Network/Canto/v6/x/govshuttle/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	evmtypes "github.com/evmos/ethermint/x/evm/types"
	"github.com/stretchr/testify/mock"
)

func (suite KeeperTestSuite) TestAppendLendingMarketProposal() {
	inputLmProposal := &types.LendingMarketProposal{
		Title:       "title",
		Description: "description",
		Metadata: &types.LendingMarketMetadata{
			Account: []string{
				"0x5E23dC409Fc2F832f83CEc191E245A191a4bCc5C",
			},
			PropId:     0,
			Values:     []uint64{0},
			Calldatas:  []string{""},
			Signatures: []string{"_acceptAdmin"},
		},
	}
	expLmProposal := &types.LendingMarketProposal{
		Title:       "title",
		Description: "description",
		Metadata: &types.LendingMarketMetadata{
			Account: []string{
				"0x5E23dC409Fc2F832f83CEc191E245A191a4bCc5C",
			},
			PropId:     1,
			Values:     []uint64{0},
			Calldatas:  []string{""},
			Signatures: []string{"_acceptAdmin"},
		},
	}
	testCases := []struct {
		name     string
		malleate func()
		expPass  bool
	}{
		{
			name: "Prop Id 0 ",
			malleate: func() {
				inputLmProposal.Metadata.PropId = 0
				expLmProposal.Metadata.PropId = 1
				mockGovKeeper := &MockGovKeeper{}
				mockAccountKeeper := &MockAccountKeeper{}
				mockERC20Keeper := &MockERC20Keeper{}
				subspace, ok := suite.app.ParamsKeeper.GetSubspace(types.ModuleName)
				suite.Require().True(ok)
				suite.app.GovshuttleKeeper = keeper.NewKeeper(suite.app.GetKey("shuttle"), suite.app.AppCodec(), subspace, mockAccountKeeper, mockERC20Keeper, mockGovKeeper)
				mockGovKeeper.On("GetProposalID", mock.Anything).Return(uint64(1), nil)
				mockAccountKeeper.On("GetSequence", mock.Anything, mock.Anything).Return(uint64(0), nil)
				mockERC20Keeper.On("CallEVMWithData", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&evmtypes.MsgEthereumTxResponse{Ret: []uint8{}}, nil)
				mockERC20Keeper.On("CallEVM", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&evmtypes.MsgEthereumTxResponse{Ret: []uint8{}}, nil)
			},
			expPass: true,
		},
		{
			name: "Force fail gov",
			malleate: func() {
				inputLmProposal.Metadata.PropId = 0
				expLmProposal.Metadata.PropId = 1
				mockGovKeeper := &MockGovKeeper{}
				subspace, ok := suite.app.ParamsKeeper.GetSubspace(types.ModuleName)
				suite.Require().True(ok)
				suite.app.GovshuttleKeeper = keeper.NewKeeper(suite.app.GetKey("shuttle"), suite.app.AppCodec(), subspace, suite.app.AccountKeeper, suite.app.Erc20Keeper, mockGovKeeper)
				mockGovKeeper.On("GetProposalID", mock.Anything).Return(uint64(0), fmt.Errorf("forced GetProposalID error"))
			},
			expPass: false,
		},
		{
			name: "Prop Id 1",
			malleate: func() {
				inputLmProposal.Metadata.PropId = 1
				expLmProposal.Metadata.PropId = 1
				mockAccountKeeper := &MockAccountKeeper{}
				mockERC20Keeper := &MockERC20Keeper{}
				subspace, ok := suite.app.ParamsKeeper.GetSubspace(types.ModuleName)
				suite.Require().True(ok)
				suite.app.GovshuttleKeeper = keeper.NewKeeper(suite.app.GetKey("shuttle"), suite.app.AppCodec(), subspace, mockAccountKeeper, mockERC20Keeper, suite.app.GovKeeper)
				mockAccountKeeper.On("GetSequence", mock.Anything, mock.Anything).Return(uint64(0), nil)
				mockERC20Keeper.On("CallEVMWithData", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&evmtypes.MsgEthereumTxResponse{Ret: []uint8{}}, nil)
				mockERC20Keeper.On("CallEVM", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&evmtypes.MsgEthereumTxResponse{Ret: []uint8{}}, nil)
			},
			expPass: true,
		},
		{
			name: "force fail erc20",
			malleate: func() {
				inputLmProposal.Metadata.PropId = 1
				expLmProposal.Metadata.PropId = 1
				mockAccountKeeper := &MockAccountKeeper{}
				mockERC20Keeper := &MockERC20Keeper{}
				subspace, ok := suite.app.ParamsKeeper.GetSubspace(types.ModuleName)
				suite.Require().True(ok)
				suite.app.GovshuttleKeeper = keeper.NewKeeper(suite.app.GetKey("shuttle"), suite.app.AppCodec(), subspace, mockAccountKeeper, mockERC20Keeper, suite.app.GovKeeper)
				mockAccountKeeper.On("GetSequence", mock.Anything, mock.Anything).Return(uint64(0), nil)
				mockERC20Keeper.On("CallEVMWithData", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&evmtypes.MsgEthereumTxResponse{Ret: []uint8{}}, nil)
				mockERC20Keeper.On("CallEVM", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, fmt.Errorf("forced CallEVM error"))
			},
			expPass: false,
		},
		{
			name: "Map contract already deployed",
			malleate: func() {
				inputLmProposal.Metadata.PropId = 1
				expLmProposal.Metadata.PropId = 1
				suite.app.GovshuttleKeeper.SetPort(suite.ctx, common.HexToAddress("0x648a5Aa0C4FbF2C1CF5a3B432c2766EeaF8E402d"))
				mockERC20Keeper := &MockERC20Keeper{}
				subspace, ok := suite.app.ParamsKeeper.GetSubspace(types.ModuleName)
				suite.Require().True(ok)
				suite.app.GovshuttleKeeper = keeper.NewKeeper(suite.app.GetKey("shuttle"), suite.app.AppCodec(), subspace, suite.app.AccountKeeper, mockERC20Keeper, suite.app.GovKeeper)
				mockERC20Keeper.On("CallEVM", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&evmtypes.MsgEthereumTxResponse{Ret: []uint8{}}, nil)
			},
			expPass: true,
		},
	}
	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.SetupTest()

			tc.malleate()

			outputLmProposal, err := suite.app.GovshuttleKeeper.AppendLendingMarketProposal(suite.ctx, inputLmProposal)

			if tc.expPass {
				suite.Require().NoError(err, tc.name)
				suite.Require().Equal(expLmProposal, outputLmProposal)
			} else {
				suite.Require().Error(err, tc.name)
			}
		})
	}
}

func (suite KeeperTestSuite) TestDeployMapContract() {
	inputLmProposal := &types.LendingMarketProposal{
		Title:       "title",
		Description: "description",
		Metadata: &types.LendingMarketMetadata{
			Account: []string{
				"0x5E23dC409Fc2F832f83CEc191E245A191a4bCc5C",
			},
			PropId:     1,
			Values:     []uint64{0},
			Calldatas:  []string{""},
			Signatures: []string{"_acceptAdmin"},
		},
	}
	testCases := []struct {
		name     string
		malleate func()
		expPass  bool
	}{
		{
			name: "ok",
			malleate: func() {
				mockAccountKeeper := &MockAccountKeeper{}
				mockERC20Keeper := &MockERC20Keeper{}
				subspace, ok := suite.app.ParamsKeeper.GetSubspace(types.ModuleName)
				suite.Require().True(ok)
				suite.app.GovshuttleKeeper = keeper.NewKeeper(suite.app.GetKey("shuttle"), suite.app.AppCodec(), subspace, mockAccountKeeper, mockERC20Keeper, suite.app.GovKeeper)
				mockAccountKeeper.On("GetSequence", mock.Anything, mock.Anything).Return(uint64(0), nil)
				mockERC20Keeper.On("CallEVMWithData", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&evmtypes.MsgEthereumTxResponse{Ret: []uint8{}}, nil)
			},
			expPass: true,
		},
		{
			name: "Force fail account",
			malleate: func() {
				mockAccountKeeper := &MockAccountKeeper{}
				mockERC20Keeper := &MockERC20Keeper{}
				subspace, ok := suite.app.ParamsKeeper.GetSubspace(types.ModuleName)
				suite.Require().True(ok)
				suite.app.GovshuttleKeeper = keeper.NewKeeper(suite.app.GetKey("shuttle"), suite.app.AppCodec(), subspace, mockAccountKeeper, mockERC20Keeper, suite.app.GovKeeper)
				mockAccountKeeper.On("GetSequence", mock.Anything, mock.Anything).Return(uint64(0), fmt.Errorf("forced GetSequence error"))
				mockERC20Keeper.On("CallEVMWithData", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&evmtypes.MsgEthereumTxResponse{Ret: []uint8{}}, nil)
			},
			expPass: false,
		},
		{
			name: "Force fail erc20",
			malleate: func() {
				mockAccountKeeper := &MockAccountKeeper{}
				mockERC20Keeper := &MockERC20Keeper{}
				subspace, ok := suite.app.ParamsKeeper.GetSubspace(types.ModuleName)
				suite.Require().True(ok)
				suite.app.GovshuttleKeeper = keeper.NewKeeper(suite.app.GetKey("shuttle"), suite.app.AppCodec(), subspace, mockAccountKeeper, mockERC20Keeper, suite.app.GovKeeper)
				mockAccountKeeper.On("GetSequence", mock.Anything, mock.Anything).Return(uint64(0), nil)
				mockERC20Keeper.On("CallEVMWithData", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, fmt.Errorf("forced CallEVMWithData error"))
			},
			expPass: false,
		},
	}
	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.SetupTest()

			tc.malleate()

			address, err := suite.app.GovshuttleKeeper.DeployMapContract(suite.ctx, inputLmProposal)
			expAddress := crypto.CreateAddress(types.ModuleAddress, 0)
			if tc.expPass {
				suite.Require().NoError(err, tc.name)
				suite.Require().Equal(expAddress, address)
			} else {
				suite.Require().Error(err, tc.name)
			}
		})
	}
}

func TestToAddress(t *testing.T) {
	type args struct {
		addrs []string
	}
	tests := []struct {
		name string
		args args
		want []common.Address
	}{
		{
			name: "empty input",
			args: args{},
			want: []common.Address{},
		},
		{
			name: "valid input",
			args: args{
				addrs: []string{
					"0x1234567890abcdef1234567890abcdef1234561A",
					"0x1234567890AbcdEF1234567890aBcdef1234561B",
				},
			},
			want: []common.Address{
				common.HexToAddress("0x1234567890ABCdeF1234567890abcDEF1234561a"),
				common.HexToAddress("0x1234567890ABCdef1234567890ABCdEf1234561b"),
			},
		},
		{
			name: "invalid input",
			args: args{
				addrs: []string{
					"0x1234567890abcdef1234567890abcdef1234561A",
					"invalid_input",
					"0x1234567890AbcdEF1234567890aBcdef1234561B",
				},
			},
			want: []common.Address{
				common.HexToAddress("0x1234567890ABCdeF1234567890abcDEF1234561a"),
				{},
				common.HexToAddress("0x1234567890ABCdef1234567890ABCdEf1234561b"),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := keeper.ToAddress(tt.args.addrs); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ToAddress() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestToBytes(t *testing.T) {
	type args struct {
		strs []string
	}
	tests := []struct {
		name string
		args args
		want [][]byte
	}{
		{
			name: "empty input",
			args: args{},
			want: [][]byte{},
		},
		{
			name: "valid input",
			args: args{
				strs: []string{
					"1234567890abcdef1234567890abcdef1234561A",
					"1234567890AbcdEF1234567890aBcdef1234561B",
				},
			},
			want: [][]byte{
				common.Hex2Bytes("1234567890abcdef1234567890abcdef1234561A"),
				common.Hex2Bytes("1234567890AbcdEF1234567890aBcdef1234561B"),
			},
		},
		{
			name: "mixed input",
			args: args{
				strs: []string{
					"1234567890abcdef1234567890abcdef1234561A",
					"",
					"1234567890AbcdEF1234567890aBcdef1234561B",
				},
			},
			want: [][]byte{
				common.Hex2Bytes("1234567890abcdef1234567890abcdef1234561A"),
				{},
				common.Hex2Bytes("1234567890AbcdEF1234567890aBcdef1234561B"),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := keeper.ToBytes(tt.args.strs); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ToBytes() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestToBigInt(t *testing.T) {
	type args struct {
		ints []uint64
	}
	tests := []struct {
		name string
		args args
		want []*big.Int
	}{
		{
			name: "empty input",
			args: args{},
			want: []*big.Int{},
		},
		{
			name: "single input",
			args: args{
				ints: []uint64{10},
			},
			want: []*big.Int{
				big.NewInt(10),
			},
		},
		{
			name: "multiple inputs",
			args: args{
				ints: []uint64{10, 20, 30},
			},
			want: []*big.Int{
				big.NewInt(10),
				big.NewInt(20),
				big.NewInt(30),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := keeper.ToBigInt(tt.args.ints); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ToBigInt() = %v, want %v", got, tt.want)
			}
		})
	}
}
