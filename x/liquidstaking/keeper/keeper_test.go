package keeper_test

import (
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	"github.com/cosmos/cosmos-sdk/x/staking"
	ethermint "github.com/evmos/ethermint/types"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/Canto-Network/Canto/v6/app"
	"github.com/Canto-Network/Canto/v6/x/liquidstaking/types"
	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/evmos/ethermint/crypto/ethsecp256k1"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	abci "github.com/tendermint/tendermint/abci/types"
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
	suite.app = app.Setup(false, nil)
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
	types.RegisterQueryServer(queryHelper, suite.app.LiquidStakingKeeper)
	suite.queryClient = types.NewQueryClient(queryHelper)

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

	validators := s.app.StakingKeeper.GetValidators(suite.ctx, 1)
	suite.validator = validators[0]
}

// Commit commits and starts a new block with an updated context.
func (suite *KeeperTestSuite) Commit() {
	suite.CommitAfter(time.Second * 0)
}

// Commit commits a block at a given time.
func (suite *KeeperTestSuite) CommitAfter(t time.Duration) {
	_ = suite.app.Commit()
	header := suite.ctx.BlockHeader()

	header.Height += 1
	header.Time = header.Time.Add(t)
	suite.app.BeginBlock(abci.RequestBeginBlock{
		Header: header,
	})

	// update ctx
	suite.ctx = suite.app.BaseApp.NewContext(false, header)
}

func (suite *KeeperTestSuite) CreateValidators(powers []int64) (valAddrs []sdk.ValAddress) {
	notBondedPool := suite.app.StakingKeeper.GetNotBondedPool(suite.ctx)

	for _, power := range powers {
		priv := ed25519.GenPrivKey()
		valAddr := sdk.ValAddress(priv.PubKey().Address().Bytes())
		validator, err := stakingtypes.NewValidator(valAddr, priv.PubKey(), stakingtypes.Description{})
		suite.NoError(err)

		tokens := suite.app.StakingKeeper.TokensFromConsensusPower(suite.ctx, power)

		// Mint tokens for not bonded pool
		err = suite.app.BankKeeper.MintCoins(suite.ctx, types.ModuleName, sdk.NewCoins(sdk.NewCoin(suite.denom, tokens)))
		suite.NoError(err)
		err = suite.app.BankKeeper.SendCoinsFromModuleToModule(suite.ctx, types.ModuleName, notBondedPool.GetName(), sdk.NewCoins(sdk.NewCoin(suite.denom, tokens)))
		suite.NoError(err)

		validator, err = validator.SetInitialCommission(stakingtypes.NewCommission(sdk.NewDecWithPrec(10, 2), sdk.NewDecWithPrec(10, 2), sdk.NewDecWithPrec(10, 2)))
		if err != nil {
			return
		}
		validator, _ = validator.AddTokensFromDel(tokens)
		validator = stakingkeeper.TestingUpdateValidator(suite.app.StakingKeeper, suite.ctx, validator, true)
		suite.app.StakingKeeper.AfterValidatorCreated(suite.ctx, validator.GetOperator())
		suite.app.StakingKeeper.SetValidator(suite.ctx, validator)
		suite.NoError(suite.app.StakingKeeper.SetValidatorByConsAddr(suite.ctx, validator))
		valAddrs = append(valAddrs, valAddr)
	}
	suite.app.EndBlocker(suite.ctx, abci.RequestEndBlock{})
	return
}

// Add test addresses with funds
func (suite *KeeperTestSuite) AddTestAddrs(accNum int, amount sdk.Int) ([]sdk.AccAddress, []sdk.Coin) {
	addrs := make([]sdk.AccAddress, 0, accNum)
	balances := make([]sdk.Coin, 0, accNum)
	for i := 0; i < accNum; i++ {
		addr := sdk.AccAddress(ed25519.GenPrivKey().PubKey().Address())
		addrs = append(addrs, addr)
		balances = append(balances, sdk.NewCoin(suite.denom, amount))

		// fund each account
		err := suite.app.BankKeeper.MintCoins(suite.ctx, types.ModuleName, sdk.NewCoins(sdk.NewCoin(suite.denom, amount)))
		suite.NoError(err)
		err = suite.app.BankKeeper.SendCoinsFromModuleToAccount(suite.ctx, types.ModuleName, addr, sdk.NewCoins(sdk.NewCoin(suite.denom, amount)))
		suite.NoError(err)
	}
	return addrs, balances
}

func (suite *KeeperTestSuite) advanceHeight(height int) {

	feeCollector := suite.app.AccountKeeper.GetModuleAccount(suite.ctx, authtypes.FeeCollectorName)
	for i := 0; i < height; i++ {
		suite.ctx = suite.ctx.WithBlockHeight(suite.ctx.BlockHeight() + 1).WithBlockTime(suite.ctx.BlockTime().Add(time.Second))

		// Mimic inflation module AfterEpochEnd Hook
		// - Inflation happened in the end of epoch triggered by AfterEpochEnd hook of epochs module
		mintedCoin := sdk.NewCoin(suite.denom, sdk.TokensFromConsensusPower(100, ethermint.PowerReduction)) // 100 Canto
		_, _, err := suite.app.InflationKeeper.MintAndAllocateInflation(suite.ctx, mintedCoin)
		suite.NoError(err)
		feeCollectorBalances := suite.app.BankKeeper.GetAllBalances(suite.ctx, feeCollector.GetAddress())
		rewardsToBeDistributed := feeCollectorBalances.AmountOf(suite.denom)

		// Mimic distribution.BeginBlock (AllocateTokens, get rewards from feeCollector, AllocateTokensToValidator, add remaining to feePool)
		suite.NoError(suite.app.BankKeeper.SendCoinsFromModuleToModule(suite.ctx, authtypes.FeeCollectorName, distrtypes.ModuleName, feeCollectorBalances))

		totalPower := int64(0)
		suite.app.StakingKeeper.IterateBondedValidatorsByPower(suite.ctx, func(index int64, validator stakingtypes.ValidatorI) (stop bool) {
			totalPower += validator.GetConsensusPower(suite.app.StakingKeeper.PowerReduction(suite.ctx))
			return false
		})

		totalRewards := sdk.ZeroDec()
		if totalPower != 0 {
			suite.app.StakingKeeper.IterateBondedValidatorsByPower(suite.ctx, func(index int64, validator stakingtypes.ValidatorI) (stop bool) {
				consPower := validator.GetConsensusPower(suite.app.StakingKeeper.PowerReduction(suite.ctx))
				powerFraction := sdk.NewDec(consPower).QuoTruncate(sdk.NewDec(totalPower))
				reward := rewardsToBeDistributed.ToDec().MulTruncate(powerFraction)
				suite.app.DistrKeeper.AllocateTokensToValidator(suite.ctx, validator, sdk.DecCoins{{Denom: suite.denom, Amount: reward}})
				totalRewards = totalRewards.Add(reward)
				return false
			})
		}
		remaining := rewardsToBeDistributed.ToDec().Sub(totalRewards)
		suite.False(remaining.GT(sdk.NewDec(100)), "all rewards should be distributed")
		feePool := suite.app.DistrKeeper.GetFeePool(suite.ctx)
		feePool.CommunityPool = feePool.CommunityPool.Add(
			sdk.NewDecCoin(suite.denom, remaining.TruncateInt()),
		)
		suite.app.DistrKeeper.SetFeePool(suite.ctx, feePool)
		staking.EndBlocker(suite.ctx, suite.app.StakingKeeper)
	}
}
