package keeper_test

import (
	"errors"
	"math/big"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	sdkmath "cosmossdk.io/math"
	"github.com/Canto-Network/Canto/v7/app"
	"github.com/Canto-Network/Canto/v7/contracts"
	"github.com/Canto-Network/Canto/v7/x/csr/types"
	"github.com/cosmos/cosmos-sdk/baseapp"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cosmosed25519 "github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/evmos/ethermint/crypto/ethsecp256k1"
	"github.com/evmos/ethermint/tests"
	evmtypes "github.com/evmos/ethermint/x/evm/types"
	feemarkettypes "github.com/evmos/ethermint/x/feemarket/types"

	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cometbft/cometbft/crypto/ed25519"
	"github.com/cometbft/cometbft/crypto/tmhash"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	tmversion "github.com/cometbft/cometbft/proto/tendermint/version"
	"github.com/cometbft/cometbft/version"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type KeeperTestSuite struct {
	suite.Suite
	// use keeper for tests
	ctx            sdk.Context
	app            *app.Canto
	queryClient    types.QueryClient
	queryClientEvm evmtypes.QueryClient
	consAddress    sdk.ConsAddress
	ethSigner      ethtypes.Signer
	address        common.Address
	validator      stakingtypes.Validator

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
	feemarketGenesis := feemarkettypes.DefaultGenesisState()
	feemarketGenesis.Params.EnableHeight = 0
	feemarketGenesis.Params.NoBaseFee = false

	// instantiate app
	suite.app = app.Setup(false, feemarketGenesis)
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
	pubKey := cosmosed25519.GenPrivKey().PubKey()
	require.NoError(t, err)
	suite.consAddress = sdk.ConsAddress(pubKey.Address())
	suite.ctx = suite.app.BaseApp.NewContextLegacy(false, tmproto.Header{
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

	queryHelperEvm := baseapp.NewQueryServerTestHelper(suite.ctx, suite.app.InterfaceRegistry())
	evmtypes.RegisterQueryServer(queryHelperEvm, suite.app.EvmKeeper)
	suite.queryClientEvm = evmtypes.NewQueryClient(queryHelperEvm)

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

	stakingParams, err := suite.app.StakingKeeper.GetParams(suite.ctx)
	require.NoError(t, err)
	stakingParams.BondDenom = suite.denom
	suite.app.StakingKeeper.SetParams(suite.ctx, stakingParams)

	// Set Validator
	valAddr := sdk.ValAddress(suite.address.Bytes())
	validator, err := stakingtypes.NewValidator(valAddr.String(), pubKey, stakingtypes.Description{})
	require.NoError(t, err)

	validator = stakingkeeper.TestingUpdateValidator(suite.app.StakingKeeper, suite.ctx, validator, true)
	valbz, err := s.app.StakingKeeper.ValidatorAddressCodec().StringToBytes(validator.GetOperator())
	s.NoError(err)
	suite.app.StakingKeeper.Hooks().AfterValidatorCreated(suite.ctx, valbz)
	err = suite.app.StakingKeeper.SetValidatorByConsAddr(suite.ctx, validator)
	require.NoError(t, err)
	suite.validator = validator

	suite.ethSigner = ethtypes.LatestSignerForChainID(s.app.EvmKeeper.ChainID())
}

// Commit commits and starts a new block with an updated context.
func (suite *KeeperTestSuite) Commit() {
	suite.CommitAfter(time.Second * 0)
}

// Commit commits a block at a given time.
func (suite *KeeperTestSuite) CommitAfter(t time.Duration) {
	header := suite.ctx.BlockHeader()
	suite.app.FinalizeBlock(&abci.RequestFinalizeBlock{
		Height:          header.Height,
		Time:            header.Time,
		ProposerAddress: header.ProposerAddress,
	})

	suite.app.Commit()

	header.Height += 1
	header.Time = header.Time.Add(t)
	suite.app.FinalizeBlock(&abci.RequestFinalizeBlock{
		Height:          header.Height,
		Time:            header.Time,
		ProposerAddress: header.ProposerAddress,
	})

	// update ctx
	suite.ctx = suite.app.BaseApp.NewContextLegacy(false, header)

	queryHelper := baseapp.NewQueryServerTestHelper(suite.ctx, suite.app.InterfaceRegistry())
	evmtypes.RegisterQueryServer(queryHelper, suite.app.EvmKeeper)
	suite.queryClientEvm = evmtypes.NewQueryClient(queryHelper)

	queryHelper = baseapp.NewQueryServerTestHelper(suite.ctx, suite.app.InterfaceRegistry())
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

// Helper function that will calculate how much revenue a NFT or the Turnstile should accumulate
// Calculation is done by the following: int(gasUsed * gasPrice * csrShares)
func CalculateExpectedFee(gasUsed uint64, gasPrice *big.Int, csrShare sdkmath.LegacyDec) sdkmath.Int {
	fee := sdkmath.NewIntFromUint64(gasUsed).Mul(sdkmath.NewIntFromBigInt(gasPrice))
	expectedTurnstileBalance := sdkmath.LegacyNewDecFromInt(fee).Mul(csrShare).TruncateInt()
	return expectedTurnstileBalance
}

// Helper function that checks the state of the CSR objects
func CheckCSRValues(csr types.CSR, expectedID uint64, expectedContracts []string, expectedTxs uint64, expectedRevenue *big.Int) {
	s.Require().Equal(expectedID, csr.Id)
	s.Require().Equal(expectedContracts, csr.Contracts)
	s.Require().Equal(expectedTxs, csr.Txs)
	s.Require().Equal(expectedRevenue, csr.Revenue.BigInt())
}

// Generates a new private private key and corresponding SDK Account Address
func GenerateKey() (*ethsecp256k1.PrivKey, sdk.AccAddress) {
	address, priv := tests.NewAddrKey()
	return priv.(*ethsecp256k1.PrivKey), sdk.AccAddress(address.Bytes())
}

// Helper function to create and make a ethereum transaction
func EVMTX(
	priv *ethsecp256k1.PrivKey,
	to *common.Address,
	amount *big.Int,
	gasLimit uint64,
	gasPrice *big.Int,
	gasFeeCap *big.Int,
	gasTipCap *big.Int,
	data []byte,
	accesses *ethtypes.AccessList,
) (*abci.ResponseFinalizeBlock, error) {
	msgEthereumTx := BuildEthTx(priv, to, amount, gasLimit, gasPrice, gasFeeCap, gasTipCap, data, accesses)
	result, err := FinalizeEthBlock(priv, msgEthereumTx)
	return result, err
}

// Helper function that creates an ethereum transaction
func BuildEthTx(
	priv *ethsecp256k1.PrivKey,
	to *common.Address,
	amount *big.Int,
	gasLimit uint64,
	gasPrice *big.Int,
	gasFeeCap *big.Int,
	gasTipCap *big.Int,
	data []byte,
	accesses *ethtypes.AccessList,
) *evmtypes.MsgEthereumTx {
	chainID := s.app.EvmKeeper.ChainID()
	from := common.BytesToAddress(priv.PubKey().Address().Bytes())
	nonce, err := s.app.AccountKeeper.GetSequence(s.ctx, sdk.AccAddress(from.Bytes()))
	s.Require().NoError(err)
	msgEthereumTx := evmtypes.NewTx(
		chainID,
		nonce,
		to,
		amount,
		gasLimit,
		gasPrice,
		gasFeeCap,
		gasTipCap,
		data,
		accesses,
	)
	msgEthereumTx.From = from.String()
	return msgEthereumTx
}

func FinalizeEthBlock(priv *ethsecp256k1.PrivKey, msgEthereumTx *evmtypes.MsgEthereumTx) (*abci.ResponseFinalizeBlock, error) {
	bz := PrepareEthTx(priv, msgEthereumTx)
	res, err := s.app.BaseApp.FinalizeBlock(&abci.RequestFinalizeBlock{
		Height:          s.app.LastBlockHeight() + 1,
		Txs:             [][]byte{bz},
		ProposerAddress: s.ctx.BlockHeader().ProposerAddress,
	})
	return res, err
}

func PrepareEthTx(priv *ethsecp256k1.PrivKey, msgEthereumTx *evmtypes.MsgEthereumTx) []byte {
	txConfig := s.app.TxConfig()
	option, err := codectypes.NewAnyWithValue(&evmtypes.ExtensionOptionsEthereumTx{})
	s.Require().NoError(err)

	txBuilder := txConfig.NewTxBuilder()
	builder, ok := txBuilder.(authtx.ExtensionOptionsTxBuilder)
	s.Require().True(ok)
	builder.SetExtensionOptions(option)

	err = msgEthereumTx.Sign(s.ethSigner, tests.NewSigner(priv))
	s.Require().NoError(err)

	msgEthereumTx.From = ""
	err = txBuilder.SetMsgs(msgEthereumTx)
	s.Require().NoError(err)

	txData, err := evmtypes.UnpackTxData(msgEthereumTx.Data)
	s.Require().NoError(err)

	evmDenom := s.app.EvmKeeper.GetParams(s.ctx).EvmDenom
	fees := sdk.Coins{{Denom: evmDenom, Amount: sdkmath.NewIntFromBigInt(txData.Fee())}}
	builder.SetFeeAmount(fees)
	builder.SetGasLimit(msgEthereumTx.GetGas())

	// bz are bytes to be broadcasted over the network
	bz, err := txConfig.TxEncoder()(txBuilder.GetTx())
	s.Require().NoError(err)

	return bz
}
