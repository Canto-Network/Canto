package keeper_test

import (
	"encoding/json"
	"math/big"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/Canto-Network/Canto-Testnet-v2/v1/app"
	"github.com/Canto-Network/ethermint-v2/server/config"
	evm "github.com/Canto-Network/ethermint-v2/x/evm/types"
	feemarkettypes "github.com/Canto-Network/ethermint-v2/x/feemarket/types"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/require"

	//used for deploying contracts
	"github.com/Canto-Network/Canto-Testnet-v2/v1/contracts"
	"github.com/Canto-Network/Canto-Testnet-v2/v1/x/erc20/types"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
)

type KeeperTestSuite struct {
	suite.Suite //top level testing suite

	ctx     sdk.Context
	app     *app.Canto
	address common.Address

	queryClient    types.QueryClient
	queryClientEvm evm.QueryClient
	signer         keyring.Signer
}

var s *KeeperTestSuite

func TestKeeperTestSuite(t *testing.T) {
	s = new(KeeperTestSuite)
	suite.Run(t, s)
}

//Test Helpers
func (suite *KeeperTestSuite) DoSetupTest(t require.TestingT) {
	checkTx := false

	feemarketGenesis := feemarkettypes.DefaultGenesisState()
	feemarketGenesis.Params.EnableHeight = 1	
	feemarketGenesis.Params.NoBaseFee = false

	//init app
	suite.app = app.Setup(checkTx, feemarketGenesis)
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
