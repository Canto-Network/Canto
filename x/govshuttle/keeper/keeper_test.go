package keeper_test

import (
	"encoding/json"
	"math/big"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/Canto-Network/Canto/v8/app"
	"github.com/Canto-Network/Canto/v8/x/govshuttle/types"
	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cometbft/cometbft/crypto/tmhash"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	tmversion "github.com/cometbft/cometbft/proto/tendermint/version"
	"github.com/cometbft/cometbft/version"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/evmos/ethermint/server/config"
	evm "github.com/evmos/ethermint/x/evm/types"
	feemarkettypes "github.com/evmos/ethermint/x/feemarket/types"

	//used for deploying contracts
	"github.com/Canto-Network/Canto/v8/contracts"
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
	validator      stakingtypes.Validator
}

var s *KeeperTestSuite

func TestKeeperTestSuite(t *testing.T) {
	s = new(KeeperTestSuite)
	suite.Run(t, s)
}

// Test Helpers
func (suite *KeeperTestSuite) DoSetupTest(t require.TestingT) {
	feemarketGenesis := feemarkettypes.DefaultGenesisState()
	feemarketGenesis.Params.EnableHeight = 1
	feemarketGenesis.Params.NoBaseFee = false

	// init app
	suite.app = app.Setup(false, feemarketGenesis)

	// consensus key
	pubKey := ed25519.GenPrivKey().PubKey()
	consAddress := sdk.ConsAddress(pubKey.Address())

	suite.ctx = suite.app.BaseApp.NewContextLegacy(false, tmproto.Header{
		Height:          1,
		ChainID:         "canto_9001-1",
		Time:            time.Now().UTC(),
		ProposerAddress: consAddress.Bytes(),

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

	queryHelper := baseapp.NewQueryServerTestHelper(suite.ctx, suite.app.InterfaceRegistry())
	types.RegisterQueryServer(queryHelper, suite.app.GovshuttleKeeper)
	suite.queryClient = types.NewQueryClient(queryHelper)

	// Set Validator
	valAddr := sdk.ValAddress(pubKey.Address().Bytes())
	validator, err := stakingtypes.NewValidator(valAddr.String(), pubKey, stakingtypes.Description{})
	suite.NoError(err)

	validator = stakingkeeper.TestingUpdateValidator(suite.app.StakingKeeper, suite.ctx, validator, true)
	valbz, err := suite.app.StakingKeeper.ValidatorAddressCodec().StringToBytes(validator.GetOperator())
	suite.NoError(err)
	suite.app.StakingKeeper.Hooks().AfterValidatorCreated(suite.ctx, valbz)
	err = suite.app.StakingKeeper.SetValidatorByConsAddr(suite.ctx, validator)
	suite.NoError(err)
	suite.validator = validator

	stakingParams, err := suite.app.StakingKeeper.GetParams(suite.ctx)
	suite.NoError(err)
	stakingParams.BondDenom = "acanto"
	suite.app.StakingKeeper.SetParams(suite.ctx, stakingParams)
}

func (suite *KeeperTestSuite) SetupTest() {
	suite.DoSetupTest(suite.T())
}

func (suite *KeeperTestSuite) DeployCaller() (common.Address, error) {
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

	res, err := suite.queryClientEvm.EstimateGas(suite.ctx, &evm.EthCallRequest{
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

	rsp, err := suite.app.EvmKeeper.EthereumTx(suite.ctx, erc20DeployTx)
	if err != nil {
		return common.Address{}, err
	}

	suite.Require().Empty(rsp.VmError)
	return crypto.CreateAddress(suite.address, nonce), nil
}

func (suite *KeeperTestSuite) DeployCallee() {

}

func (suite *KeeperTestSuite) Commit() {
	suite.CommitAfter(time.Nanosecond)
}

func (suite *KeeperTestSuite) CommitAfter(t time.Duration) {
	header := suite.ctx.BlockHeader()
	header.Time = header.Time.Add(t)
	suite.app.FinalizeBlock(&abci.RequestFinalizeBlock{
		Height:          header.Height,
		Time:            header.Time,
		ProposerAddress: suite.ctx.BlockHeader().ProposerAddress,
	})
	suite.app.Commit()

	// update ctx
	header.Height += 1
	suite.ctx = suite.app.BaseApp.NewUncachedContext(false, header)

	queryHelper := baseapp.NewQueryServerTestHelper(suite.ctx, suite.app.InterfaceRegistry())
	evm.RegisterQueryServer(queryHelper, suite.app.EvmKeeper)
	suite.queryClientEvm = evm.NewQueryClient(queryHelper)
}
