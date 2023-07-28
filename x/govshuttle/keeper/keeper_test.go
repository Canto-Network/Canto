package keeper_test

import (
	"encoding/json"
	"math/big"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"github.com/tendermint/tendermint/crypto/tmhash"
	"github.com/tendermint/tendermint/version"

	"github.com/Canto-Network/Canto/v6/app"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/evmos/ethermint/crypto/ethsecp256k1"
	"github.com/evmos/ethermint/server/config"
	evm "github.com/evmos/ethermint/x/evm/types"
	feemarkettypes "github.com/evmos/ethermint/x/feemarket/types"
	"github.com/stretchr/testify/require"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	tmversion "github.com/tendermint/tendermint/proto/tendermint/version"

	//used for deploying contracts
	"github.com/Canto-Network/Canto/v6/contracts"
	"github.com/Canto-Network/Canto/v6/x/govshuttle/types"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	evmtypes "github.com/evmos/ethermint/x/evm/types"
)

type KeeperTestSuite struct {
	suite.Suite //top level testing suite

	ctx            sdk.Context
	app            *app.Canto
	address        common.Address
	consAddress    sdk.ConsAddress
	queryClientEvm evm.QueryClient
	signer         keyring.Signer
}

var s *KeeperTestSuite

func TestKeeperTestSuite(t *testing.T) {
	s = new(KeeperTestSuite)
	suite.Run(t, s)
}

// Test Helpers
func (suite *KeeperTestSuite) DoSetupTest(t require.TestingT) {

	// consensus key
	priv, err := ethsecp256k1.GenerateKey()
	require.NoError(t, err)
	suite.consAddress = sdk.ConsAddress(priv.PubKey().Address())
	checkTx := false

	feemarketGenesis := feemarkettypes.DefaultGenesisState()
	feemarketGenesis.Params.EnableHeight = 1
	feemarketGenesis.Params.NoBaseFee = false

	//init app
	suite.app = app.Setup(checkTx, feemarketGenesis)
	suite.ctx = suite.app.BaseApp.NewContext(checkTx, tmproto.Header{
		Height:          1,
		ChainID:         "canto_9001-1",
		Time:            time.Now().UTC(),
		ProposerAddress: suite.consAddress.Bytes(),

		Version: tmversion.Consensus{
			Block: version.BlockProtocol,
		},
		LastBlockId: tmproto.BlockID{
			Hash: tmhash.Sum([]byte("block_id")),
			PartSetHeader: tmproto.PartSetHeader{
				Total: 11,
				Hash:  tmhash.Sum([]byte("partset_header")),
			},
		},
		AppHash:            tmhash.Sum([]byte("app")),
		DataHash:           tmhash.Sum([]byte("data")),
		EvidenceHash:       tmhash.Sum([]byte("evidence")),
		ValidatorsHash:     tmhash.Sum([]byte("validators")),
		NextValidatorsHash: tmhash.Sum([]byte("next_validators")),
		ConsensusHash:      tmhash.Sum([]byte("consensus")),
		LastResultsHash:    tmhash.Sum([]byte("last_result")),
	})
}

func (suite *KeeperTestSuite) SetupTest() {
	suite.DoSetupTest(suite.T())
}

func (suite *KeeperTestSuite) DeployCaller() (common.Address, error) {
	ctx := sdk.WrapSDKContext(suite.ctx)
	chainID := suite.app.EvmKeeper.ChainID()

	ctorArgs, err := contracts.CallerContract.ABI.Pack("")

	if err != nil {
		return common.Address{}, err
	}

	data := append(contracts.CallerContract.Bin, ctorArgs...)
	args, err := json.Marshal(&evm.TransactionArgs{
		From: &suite.address,
		Data: (*hexutil.Bytes)(&data),
	})

	if err != nil {
		return common.Address{}, err
	}

	res, err := suite.queryClientEvm.EstimateGas(ctx, &evm.EthCallRequest{
		Args:   args,
		GasCap: uint64(config.DefaultGasCap),
	})
	if err != nil {
		return common.Address{}, err
	}

	nonce := suite.app.EvmKeeper.GetNonce(suite.ctx, suite.address)

	erc20DeployTx := evm.NewTxContract(
		chainID,
		nonce,
		nil,     // amount
		res.Gas, // gasLimit
		nil,     // gasPrice
		suite.app.FeeMarketKeeper.GetBaseFee(suite.ctx),
		big.NewInt(1),
		data,                   // input
		&ethtypes.AccessList{}, // accesses
	)

	erc20DeployTx.From = suite.address.Hex()
	err = erc20DeployTx.Sign(ethtypes.LatestSignerForChainID(chainID), suite.signer)
	if err != nil {
		return common.Address{}, err
	}

	rsp, err := suite.app.EvmKeeper.EthereumTx(ctx, erc20DeployTx)
	if err != nil {
		return common.Address{}, err
	}

	suite.Require().Empty(rsp.VmError)
	return crypto.CreateAddress(suite.address, nonce), nil
}

func (suite *KeeperTestSuite) DeployCallee() {

}

var _ types.GovKeeper = &MockGovKeeper{}

type MockGovKeeper struct {
	mock.Mock
}

func (m *MockGovKeeper) GetProposalID(ctx sdk.Context) (uint64, error) {
	args := m.Called(mock.Anything)
	return args.Get(0).(uint64), args.Error(1)
}

var _ types.AccountKeeper = &MockAccountKeeper{}

type MockAccountKeeper struct {
	mock.Mock
}

func (m *MockAccountKeeper) GetSequence(ctx sdk.Context, addr sdk.AccAddress) (uint64, error) {
	args := m.Called(mock.Anything, mock.Anything)
	return args.Get(0).(uint64), args.Error(1)
}

func (m *MockAccountKeeper) GetModuleAddress(moduleName string) sdk.AccAddress {
	args := m.Called(mock.Anything)
	return args.Get(0).(sdk.AccAddress)
}

var _ types.ERC20Keeper = &MockERC20Keeper{}

type MockERC20Keeper struct {
	mock.Mock
}

func (m *MockERC20Keeper) CallEVM(ctx sdk.Context, abi abi.ABI, from common.Address, contract common.Address, commit bool, method string, args ...interface{}) (*evmtypes.MsgEthereumTxResponse, error) {
	resArgs := m.Called(mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
	if resArgs.Get(0) == nil {
		return nil, resArgs.Error(1)
	}
	return resArgs.Get(0).(*evmtypes.MsgEthereumTxResponse), resArgs.Error(1)
}

func (m *MockERC20Keeper) CallEVMWithData(ctx sdk.Context, from common.Address, contract *common.Address, data []byte, commit bool) (*evmtypes.MsgEthereumTxResponse, error) {
	args := m.Called(mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*evmtypes.MsgEthereumTxResponse), args.Error(1)
}

func (suite *KeeperTestSuite) TestKeeper() {
	address, found := suite.app.GovshuttleKeeper.GetPort(suite.ctx)
	suite.Require().Equal(false, found)
	suite.Require().Equal(common.Address{}, address)
	testAddress := common.HexToAddress("0x648a5Aa0C4FbF2C1CF5a3B432c2766EeaF8E402d")
	suite.app.GovshuttleKeeper.SetPort(suite.ctx, testAddress)
	address, found = suite.app.GovshuttleKeeper.GetPort(suite.ctx)
	suite.Require().Equal(true, found)
	suite.Require().Equal(testAddress, address)
}
