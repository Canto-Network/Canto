package keeper_test

import (
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/Canto-Network/Canto/v6/x/liquidstaking"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"

	liquidstakingkeeper "github.com/Canto-Network/Canto/v6/x/liquidstaking/keeper"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"

	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	"github.com/cosmos/cosmos-sdk/x/staking"
	ethermint "github.com/evmos/ethermint/types"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/Canto-Network/Canto/v6/app"
	"github.com/Canto-Network/Canto/v6/x/liquidstaking/types"
	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
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

var DefaultInflationAmt = sdk.TokensFromConsensusPower(100, ethermint.PowerReduction)

type KeeperTestSuite struct {
	suite.Suite
	// use keeper for tests
	ctx         sdk.Context
	app         *app.Canto
	queryClient types.QueryClient
	consAddress sdk.ConsAddress
	address     common.Address
	delegator   sdk.AccAddress

	denom string
	// EpochCount counted by epochs module
	rewardEpochCount int64
	// EpochCount counted by liquidstaking module
	lsEpochCount int64
}

// testingEnvOptions is used to configure the testing environment for liquidstaking
type testingEnvOptions struct {
	desc                  string
	numVals               int
	fixedValFeeRate       sdk.Dec
	valFeeRates           []sdk.Dec
	fixedPower            int64
	powers                []int64
	numInsurances         int
	fixedInsuranceFeeRate sdk.Dec
	insuranceFeeRates     []sdk.Dec
	numPairedChunks       int
	// this field influences the total supply of the testing environment
	fundingAccountBalance sdk.Int
}

// testingEnv is used to store the testing environment for liquidstaking
type testingEnv struct {
	delegators      []sdk.AccAddress
	providers       []sdk.AccAddress
	pairedChunks    []types.Chunk
	insurances      []types.Insurance
	valAddrs        []sdk.ValAddress
	pubKeys         []cryptotypes.PubKey
	bondDenom       string
	liquidBondDenom string
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

	suite.delegator = priv.PubKey().Address().Bytes()
	suite.address = common.BytesToAddress(priv.PubKey().Address().Bytes())
	suite.denom = "acanto"

	// consensus key
	privCons, err := ethsecp256k1.GenerateKey()
	require.NoError(t, err)
	suite.consAddress = sdk.ConsAddress(privCons.PubKey().Address())
	initialBlockTime := time.Now().UTC()
	initialHeight := int64(1)
	suite.ctx = suite.app.BaseApp.NewContext(false, tmproto.Header{
		Height:          initialHeight,
		ChainID:         "canto_9001-1",
		Time:            initialBlockTime,
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

	// set current mainnet slahsing params
	downtimeJailDuration, err := time.ParseDuration("1800s")
	require.NoError(t, err)
	suite.app.SlashingKeeper.SetParams(suite.ctx, slashingtypes.NewParams(
		9000,
		sdk.NewDecWithPrec(5, 1), // 0.5
		downtimeJailDuration,
		sdk.NewDecWithPrec(5, 2),  // 0.05
		sdk.NewDecWithPrec(75, 4), // 0.0075
	))

	s.app.LiquidStakingKeeper.SetEpoch(
		suite.ctx,
		types.Epoch{
			CurrentNumber: 0,
			StartTime:     initialBlockTime,
			Duration:      suite.app.StakingKeeper.GetParams(suite.ctx).UnbondingTime,
			StartHeight:   initialHeight,
		},
	)
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

func (suite *KeeperTestSuite) CreateValidators(
	powers []int64,
	fixedFeeRate sdk.Dec,
	feeRates []sdk.Dec,
) (valAddrs []sdk.ValAddress, pubKeys []cryptotypes.PubKey) {
	if feeRates != nil && len(feeRates) > 0 {
		suite.Equal(len(powers), len(feeRates))
	}
	notBondedPool := suite.app.StakingKeeper.GetNotBondedPool(suite.ctx)

	for i, power := range powers {
		priv := ed25519.GenPrivKey()
		pubKeys = append(pubKeys, priv.PubKey())
		valAddr := sdk.ValAddress(priv.PubKey().Address().Bytes())
		validator, err := stakingtypes.NewValidator(valAddr, priv.PubKey(), stakingtypes.Description{})
		suite.NoError(err)

		var feeRate sdk.Dec
		if fixedFeeRate != sdk.ZeroDec() {
			feeRate = fixedFeeRate
		} else {
			feeRate = feeRates[i]
		}
		validator, err = validator.SetInitialCommission(stakingtypes.NewCommission(feeRate, feeRate, feeRate))
		if err != nil {
			return
		}
		// added to avoid invariant check for delegation
		// validator must have self delegation
		suite.app.StakingKeeper.SetValidator(suite.ctx, validator)
		suite.NoError(suite.app.StakingKeeper.SetValidatorByConsAddr(suite.ctx, validator))
		suite.app.StakingKeeper.SetNewValidatorByPowerIndex(suite.ctx, validator)
		suite.app.StakingKeeper.AfterValidatorCreated(suite.ctx, validator.GetOperator())
		valAddrs = append(valAddrs, valAddr)

		tokens := suite.app.StakingKeeper.TokensFromConsensusPower(suite.ctx, power)
		err = suite.app.BankKeeper.MintCoins(suite.ctx, types.ModuleName, sdk.NewCoins(sdk.NewCoin(suite.denom, tokens)))
		suite.NoError(err)
		err = suite.app.BankKeeper.SendCoinsFromModuleToModule(suite.ctx, types.ModuleName, notBondedPool.GetName(), sdk.NewCoins(sdk.NewCoin(suite.denom, tokens)))
		suite.NoError(err)
		suite.app.StakingKeeper.DeleteValidatorByPowerIndex(suite.ctx, validator)
		validator, addedShares := validator.AddTokensFromDel(tokens)
		suite.app.StakingKeeper.SetValidator(suite.ctx, validator)
		suite.app.StakingKeeper.SetValidatorByPowerIndex(suite.ctx, validator)

		del := stakingtypes.NewDelegation(suite.delegator, validator.GetOperator(), addedShares)
		suite.app.StakingKeeper.BeforeDelegationCreated(suite.ctx, suite.delegator, del.GetValidatorAddr())
		suite.app.StakingKeeper.SetDelegation(suite.ctx, del)
		suite.app.StakingKeeper.AfterDelegationModified(suite.ctx, suite.delegator, del.GetValidatorAddr())
	}
	suite.app.EndBlocker(suite.ctx, abci.RequestEndBlock{})
	return
}

// Add test addresses with funds
func (suite *KeeperTestSuite) AddTestAddrsWithFunding(fundingAccount sdk.AccAddress, accNum int, amount sdk.Int) ([]sdk.AccAddress, []sdk.Coin) {
	addrs := make([]sdk.AccAddress, 0, accNum)
	balances := make([]sdk.Coin, 0, accNum)
	for i := 0; i < accNum; i++ {
		addr := sdk.AccAddress(ed25519.GenPrivKey().PubKey().Address())
		addrs = append(addrs, addr)
		balances = append(balances, sdk.NewCoin(suite.denom, amount))

		suite.app.BankKeeper.SendCoins(suite.ctx, fundingAccount, addr, sdk.NewCoins(sdk.NewCoin(suite.denom, amount)))
	}
	return addrs, balances
}

func (suite *KeeperTestSuite) fundAccount(ctx sdk.Context, addr sdk.AccAddress, amount sdk.Int) {
	err := suite.app.BankKeeper.MintCoins(ctx, types.ModuleName, sdk.NewCoins(sdk.NewCoin(suite.denom, amount)))
	suite.NoError(err)
	err = suite.app.BankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, addr, sdk.NewCoins(sdk.NewCoin(suite.denom, amount)))
	suite.NoError(err)
}

func (suite *KeeperTestSuite) advanceHeight(ctx sdk.Context, height int, msg string) sdk.Context {
	fmt.Println("+++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++")
	fmt.Println("advance " + strconv.Itoa(height) + " blocks(= reward epochs)")
	if msg != "" {
		fmt.Println(msg)
	}
	fmt.Println("+++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++")
	feeCollector := suite.app.AccountKeeper.GetModuleAccount(ctx, authtypes.FeeCollectorName)
	for i := 0; i < height; i++ {
		ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 1).WithBlockTime(ctx.BlockTime().Add(time.Second))
		liquidstaking.BeginBlocker(ctx, suite.app.LiquidStakingKeeper)

		// Mimic inflation module AfterEpochEnd Hook
		// - Inflation happened in the end of epoch triggered by AfterEpochEnd hook of epochs module
		mintedCoin := sdk.NewCoin(suite.denom, DefaultInflationAmt)
		_, _, err := suite.app.InflationKeeper.MintAndAllocateInflation(ctx, mintedCoin)
		suite.NoError(err)
		feeCollectorBalances := suite.app.BankKeeper.GetAllBalances(ctx, feeCollector.GetAddress())
		rewardsToBeDistributed := feeCollectorBalances.AmountOf(suite.denom)
		suite.rewardEpochCount += 1

		// Mimic distribution.BeginBlock (AllocateTokens, get rewards from feeCollector, AllocateTokensToValidator, add remaining to feePool)
		suite.NoError(suite.app.BankKeeper.SendCoinsFromModuleToModule(ctx, authtypes.FeeCollectorName, distrtypes.ModuleName, feeCollectorBalances))

		totalPower := int64(0)
		suite.app.StakingKeeper.IterateBondedValidatorsByPower(ctx, func(index int64, validator stakingtypes.ValidatorI) (stop bool) {
			totalPower += validator.GetConsensusPower(suite.app.StakingKeeper.PowerReduction(ctx))
			return false
		})

		totalRewards := sdk.ZeroDec()
		if totalPower != 0 {
			suite.app.StakingKeeper.IterateBondedValidatorsByPower(ctx, func(index int64, validator stakingtypes.ValidatorI) (stop bool) {
				consPower := validator.GetConsensusPower(suite.app.StakingKeeper.PowerReduction(ctx))
				powerFraction := sdk.NewDec(consPower).QuoTruncate(sdk.NewDec(totalPower))
				reward := rewardsToBeDistributed.ToDec().MulTruncate(powerFraction)
				suite.app.DistrKeeper.AllocateTokensToValidator(ctx, validator, sdk.DecCoins{{Denom: suite.denom, Amount: reward}})
				validator = suite.app.StakingKeeper.Validator(ctx, validator.GetOperator())
				totalRewards = totalRewards.Add(reward)
				return false
			})
		}
		remaining := rewardsToBeDistributed.ToDec().Sub(totalRewards)
		suite.False(remaining.GT(sdk.NewDec(1000)), "all rewards should be distributed")
		feePool := suite.app.DistrKeeper.GetFeePool(ctx)
		feePool.CommunityPool = feePool.CommunityPool.Add(
			sdk.NewDecCoin(suite.denom, remaining.TruncateInt()),
		)
		suite.app.DistrKeeper.SetFeePool(ctx, feePool)
		staking.EndBlocker(ctx, suite.app.StakingKeeper)
		liquidstaking.EndBlocker(ctx, suite.app.LiquidStakingKeeper)
		suite.mustPassInvariants()
	}
	return ctx
}

func (suite *KeeperTestSuite) advanceEpoch(ctx sdk.Context) sdk.Context {
	// Set block header time as epochStartTime + duration + 1 second
	epoch := suite.app.LiquidStakingKeeper.GetEpoch(ctx)
	// Lets pass epoch
	ctx = ctx.WithBlockTime(epoch.StartTime.Add(epoch.Duration))
	suite.lsEpochCount += 1

	fmt.Println("===============================================================================")
	fmt.Println("lsEpoch is reached, endblocker will be executed at following block")
	fmt.Println("===============================================================================")

	return ctx
}

func (suite *KeeperTestSuite) resetEpochs() {
	suite.lsEpochCount = 0
	suite.rewardEpochCount = 0
}

func (suite *KeeperTestSuite) mustPassInvariants() {
	res, broken := liquidstakingkeeper.AllInvariants(suite.app.LiquidStakingKeeper)(suite.ctx)
	suite.False(broken)
	suite.Len(res, 0)
}

// unique delegator for each chunks
// - balance of delegator is oneChunk amount of tokens
// unique provider for each insurances
// - balance of provider is oneInsurance amount of tokens
func (suite *KeeperTestSuite) setupLiquidStakeTestingEnv(env testingEnvOptions) testingEnv {
	suite.resetEpochs()
	suite.fundAccount(suite.ctx, fundingAccount, env.fundingAccountBalance)

	if env.fixedPower > 0 {
		env.powers = make([]int64, env.numVals)
		for i := range env.powers {
			env.powers[i] = env.fixedPower
		}
	}
	valAddrs, pubKeys := suite.CreateValidators(env.powers, env.fixedValFeeRate, env.valFeeRates)
	oneChunk, oneInsurance := suite.app.LiquidStakingKeeper.GetMinimumRequirements(suite.ctx)
	providers, providerBalances := suite.AddTestAddrsWithFunding(fundingAccount, env.numInsurances, oneInsurance.Amount)
	insurances := suite.provideInsurances(suite.ctx, providers, valAddrs, providerBalances, env.fixedInsuranceFeeRate, env.insuranceFeeRates)

	// create numPairedChunks delegators
	delegators, delegatorBalances := suite.AddTestAddrsWithFunding(fundingAccount, env.numPairedChunks, oneChunk.Amount)
	nas := suite.app.LiquidStakingKeeper.GetNetAmountState(suite.ctx)
	suite.True(nas.IsZeroState(), "nothing happened yet so it must be zero state")
	pairedChunks := suite.liquidStakes(suite.ctx, delegators, delegatorBalances)

	// update insurance statuses because the status can be changed after liquid staking (pairing -> paired)
	for i, insurance := range insurances {
		insurances[i], _ = suite.app.LiquidStakingKeeper.GetInsurance(suite.ctx, insurance.Id)
	}

	bondDenom := suite.app.StakingKeeper.BondDenom(suite.ctx)
	liquidBondDenom := suite.app.LiquidStakingKeeper.GetLiquidBondDenom(suite.ctx)
	u := suite.app.LiquidStakingKeeper.CalcUtilizationRatio(suite.ctx)
	fmt.Printf(`
===============================================================================
Initial state of %s 
- num of validators: %d
- fixed validator fee rate: %s
- validator fee rates: %s
- num of delegators: %d
- num of insurances: %d
- fixed insurance fee rate: %s
- insurance fee ratesS: %s
- bonded denom: %s
- liquid bond denom: %s
- funding account balance: %s
- total supply: %s
- utilization ratio: %s
===============================================================================
`,
		env.desc,
		len(valAddrs),
		env.fixedValFeeRate.String(),
		env.valFeeRates,
		len(delegators),
		len(providers),
		env.fixedInsuranceFeeRate,
		env.insuranceFeeRates,
		bondDenom,
		liquidBondDenom,
		env.fundingAccountBalance,
		suite.app.BankKeeper.GetSupply(suite.ctx, suite.denom).String(),
		u.String(),
	)
	return testingEnv{
		delegators,
		providers,
		pairedChunks,
		insurances,
		valAddrs,
		pubKeys,
		bondDenom,
		liquidBondDenom,
	}
}

func (suite *KeeperTestSuite) createTestPubKeys(numKeys int) []cryptotypes.PubKey {
	pubKeys := make([]cryptotypes.PubKey, numKeys)
	for i := 0; i < numKeys; i++ {
		pk := ed25519.GenPrivKey()
		pubKeys[i] = pk.PubKey()
	}
	return pubKeys
}
