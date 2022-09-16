package keeper_test

import (
	"errors"
	"math/big"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/Canto-Network/Canto/v2/app"
	"github.com/Canto-Network/Canto/v2/contracts"
	"github.com/Canto-Network/Canto/v2/x/csr/types"
	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/evmos/ethermint/crypto/ethsecp256k1"
	evmtypes "github.com/evmos/ethermint/x/evm/types"
	feemarkettypes "github.com/evmos/ethermint/x/feemarket/types"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto/ed25519"
	"github.com/tendermint/tendermint/crypto/tmhash"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	tmversion "github.com/tendermint/tendermint/proto/tendermint/version"
	"github.com/tendermint/tendermint/version"
)

type KeeperTestSuite struct {
	suite.Suite
	// use keeper for tests
	ctx         sdk.Context
	app         *app.Canto
	queryClient types.QueryClient
	consAddress sdk.ConsAddress
	ethSigner   ethtypes.Signer
	address     common.Address
	validator   stakingtypes.Validator

	denom string
}

var s *KeeperTestSuite

func TestKeeperTestSuite(t *testing.T) {
	s = new(KeeperTestSuite)

	suite.Run(t, s)

	// Run Ginkgo integration tests
	RegisterFailHandler(Fail)
	RunSpecs(t, "Keeper Suite")
}

func (suite *KeeperTestSuite) SetupTest() {
	// instantiate app
	suite.app = app.Setup(false, feemarkettypes.DefaultGenesisState())
	// initialize ctx for tests
	suite.SetupApp()
}

func (suite *KeeperTestSuite) SetupApp() {
	t := suite.T()

	priv, err := ethsecp256k1.GenerateKey()
	require.NoError(t, err)

	suite.address = common.BytesToAddress(priv.PubKey().Address().Bytes())
	suite.denom = "acanto"

	// consensus key
	privCons, err := ethsecp256k1.GenerateKey()
	require.NoError(t, err)
	suite.consAddress = sdk.ConsAddress(privCons.PubKey().Address())
	suite.ctx = suite.app.BaseApp.NewContext(false, tmproto.Header{
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

	queryHelper := baseapp.NewQueryServerTestHelper(suite.ctx, suite.app.InterfaceRegistry())
	types.RegisterQueryServer(queryHelper, suite.app.CSRKeeper)
	suite.queryClient = types.NewQueryClient(queryHelper)

	bigInt := &big.Int{}
	bigInt.SetUint64(100)
	s.app.FeeMarketKeeper.SetBaseFee(suite.ctx, bigInt)

	params := types.DefaultParams()
	params.EnableCsr = true
	suite.app.CSRKeeper.SetParams(suite.ctx, params)

	evmParams := suite.app.EvmKeeper.GetParams(suite.ctx)
	evmParams.EvmDenom = suite.denom
	suite.app.EvmKeeper.SetParams(suite.ctx, evmParams)

	stakingParams := suite.app.StakingKeeper.GetParams(suite.ctx)
	stakingParams.BondDenom = suite.denom
	suite.app.StakingKeeper.SetParams(suite.ctx, stakingParams)

	// Set Validator
	valAddr := sdk.ValAddress(suite.address.Bytes())
	validator, err := stakingtypes.NewValidator(valAddr, privCons.PubKey(), stakingtypes.Description{})
	require.NoError(t, err)

	validator = stakingkeeper.TestingUpdateValidator(suite.app.StakingKeeper, suite.ctx, validator, true)
	suite.app.StakingKeeper.AfterValidatorCreated(suite.ctx, validator.GetOperator())
	err = suite.app.StakingKeeper.SetValidatorByConsAddr(suite.ctx, validator)
	require.NoError(t, err)

	validators := s.app.StakingKeeper.GetValidators(s.ctx, 1)
	suite.validator = validators[0]

	suite.ethSigner = ethtypes.LatestSignerForChainID(s.app.EvmKeeper.ChainID())
}

// Commit commits and starts a new block with an updated context.
func (suite *KeeperTestSuite) Commit() {
	suite.CommitAfter(time.Second * 0)
}

// Commit commits a block at a given time.
func (suite *KeeperTestSuite) CommitAfter(t time.Duration) {
	header := suite.ctx.BlockHeader()
	suite.app.EndBlock(abci.RequestEndBlock{Height: header.Height})
	_ = suite.app.Commit()

	header.Height += 1
	header.Time = header.Time.Add(t)
	suite.app.BeginBlock(abci.RequestBeginBlock{
		Header: header,
	})

	// update ctx
	suite.ctx = suite.app.BaseApp.NewContext(false, header)

	queryHelper := baseapp.NewQueryServerTestHelper(suite.ctx, suite.app.InterfaceRegistry())
	types.RegisterQueryServer(queryHelper, suite.app.CSRKeeper)
	suite.queryClient = types.NewQueryClient(queryHelper)
}

// CreateNewAccount creates a new account and sets it in the account keeper.
func (suite *KeeperTestSuite) CreateNewAccount(ctx sdk.Context) sdk.AccAddress {
	pubKey := ed25519.GenPrivKey().PubKey()
	address := sdk.AccAddress(pubKey.Address())
	beneficiary := suite.app.AccountKeeper.NewAccountWithAddress(ctx, address)
	suite.app.AccountKeeper.SetAccount(ctx, beneficiary)
	return address
}

// GenerateUpdateEventData is a helper function that will generate the update event data given a smart contract address and nft id.
func GenerateUpdateEventData(contract common.Address, nftID uint64) (data []byte, err error) {
	bigInt := &big.Int{}
	bigInt.SetUint64(nftID)
	return GenerateEventData("Assign", contracts.TurnstileContract, contract, bigInt)
}

// GenerateRegisterEventData is a helper function that will generate the register event data given a smart contract address, receiver address and nft id.
func GenerateRegisterEventData(contract, receiver common.Address, nftID uint64) (data []byte, err error) {
	bigInt := &big.Int{}
	bigInt.SetUint64(nftID)
	return GenerateEventData("Register", contracts.TurnstileContract, contract, receiver, bigInt)
}

// GenerateEventData creates data field for an arbitrary transaction given a set of arguments an a method name. Returns the byte data
// associated with the the inputed event data.
func GenerateEventData(name string, contract evmtypes.CompiledContract, args ...interface{}) ([]byte, error) {
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
