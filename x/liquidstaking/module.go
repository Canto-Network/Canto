package liquidstaking

import (
	"context"
	"encoding/json"
	"fmt"
	inflationtypes "github.com/Canto-Network/Canto/v7/x/inflation/types"
	"github.com/armon/go-metrics"
	"github.com/cosmos/cosmos-sdk/telemetry"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	"github.com/cosmos/cosmos-sdk/x/staking"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	ethermint "github.com/evmos/ethermint/types"
	"math/rand"

	inflationkeeper "github.com/Canto-Network/Canto/v7/x/inflation/keeper"
	"github.com/Canto-Network/Canto/v7/x/liquidstaking/simulation"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	distrkeeper "github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	"github.com/gorilla/mux"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/spf13/cobra"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/Canto-Network/Canto/v7/x/liquidstaking/client/cli"
	"github.com/Canto-Network/Canto/v7/x/liquidstaking/keeper"
	"github.com/Canto-Network/Canto/v7/x/liquidstaking/types"
)

// type check to ensure the interface is properly implemented
var (
	_ module.AppModule           = AppModule{}
	_ module.AppModuleBasic      = AppModuleBasic{}
	_ module.AppModuleSimulation = AppModule{}
)

// AppModuleBasic type for the liquidstaking module
type AppModuleBasic struct {
	cdc codec.Codec
}

// Name returns the liquidstaking module's name.
func (AppModuleBasic) Name() string {
	return types.ModuleName
}

// RegisterLegacyAminoCodec performs a no-op as the liquidstaking do not support Amino
// encoding.
func (AppModuleBasic) RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {}

// ConsensusVersion returns the consensus state-breaking version for the module.
func (AppModuleBasic) ConsensusVersion() uint64 {
	return 1
}

// RegisterInterfaces registers interfaces and implementations of the liquidstaking
// module.
func (AppModuleBasic) RegisterInterfaces(interfaceRegistry codectypes.InterfaceRegistry) {
	types.RegisterInterfaces(interfaceRegistry)
}

// DefaultGenesis returns default genesis state as raw bytes for the liquidstaking
// module.
func (AppModuleBasic) DefaultGenesis(cdc codec.JSONCodec) json.RawMessage {
	return cdc.MustMarshalJSON(types.DefaultGenesisState())
}

// ValidateGenesis performs genesis state validation for the liquidstaking module.
func (b AppModuleBasic) ValidateGenesis(cdc codec.JSONCodec, config client.TxEncodingConfig, bz json.RawMessage) error {
	var genesisState types.GenesisState
	if err := cdc.UnmarshalJSON(bz, &genesisState); err != nil {
		return fmt.Errorf("failed to unmarshal %s genesis state: %w", types.ModuleName, err)
	}

	return genesisState.Validate()
}

// RegisterRESTRoutes performs a no-op as the liquidstaking module doesn't expose REST
// endpoints
func (AppModuleBasic) RegisterRESTRoutes(clientCtx client.Context, rtr *mux.Router) {}

// RegisterGRPCGatewayRoutes registers the gRPC Gateway routes for the liquidstaking
// module.
func (b AppModuleBasic) RegisterGRPCGatewayRoutes(c client.Context, serveMux *runtime.ServeMux) {
	if err := types.RegisterQueryHandlerClient(context.Background(), serveMux, types.NewQueryClient(c)); err != nil {
		panic(err)
	}
}

// GetTxCmd returns the root tx command for the liquidstaking module.
func (AppModuleBasic) GetTxCmd() *cobra.Command {
	return cli.GetTxCmd()
}

// GetQueryCmd returns the liquidstaking module's root query command.
func (AppModuleBasic) GetQueryCmd() *cobra.Command {
	return cli.GetQueryCmd(types.StoreKey)
}

// ___________________________________________________________________________

// AppModule implements the AppModule interface for the liquidstaking module.
type AppModule struct {
	AppModuleBasic
	keeper keeper.Keeper
	ak     authkeeper.AccountKeeper
	bk     bankkeeper.Keeper
	sk     stakingkeeper.Keeper
	dk     distrkeeper.Keeper
	ik     inflationkeeper.Keeper
}

// NewAppModule creates a new AppModule Object
func NewAppModule(
	cdc codec.Codec,
	k keeper.Keeper,
	ak authkeeper.AccountKeeper,
	bk bankkeeper.Keeper,
	sk stakingkeeper.Keeper,
	dk distrkeeper.Keeper,
	ik inflationkeeper.Keeper,
) AppModule {
	return AppModule{
		AppModuleBasic: AppModuleBasic{cdc: cdc},
		keeper:         k,
		ak:             ak,
		bk:             bk,
		sk:             sk,
		dk:             dk,
		ik:             ik,
	}
}

// Name returns the liquidstaking module's name.
func (AppModule) Name() string {
	return types.ModuleName
}

// RegisterInvariants registers the liquidstaking module's invariants.
func (am AppModule) RegisterInvariants(ir sdk.InvariantRegistry) {
	keeper.RegisterInvariants(ir, am.keeper)
}

// NewHandler returns nil - liquidstaking module doesn't expose tx gRPC endpoints
func (am AppModule) NewHandler() sdk.Handler {
	return nil
}

// Route returns the liquidstaking module's message routing key.
func (am AppModule) Route() sdk.Route {
	return sdk.NewRoute(types.RouterKey, am.NewHandler())
}

// QuerierRoute returns the claim module's query routing key.
func (am AppModule) QuerierRoute() string {
	return types.RouterKey
}

// LegacyQuerierHandler returns the claim module's Querier.
func (am AppModule) LegacyQuerierHandler(amino *codec.LegacyAmino) sdk.Querier {
	return nil
}

// RegisterServices registers a GRPC query service to respond to the
// module-specific GRPC queries.
func (am AppModule) RegisterServices(cfg module.Configurator) {
	types.RegisterMsgServer(cfg.MsgServer(), am.keeper)
	types.RegisterQueryServer(cfg.QueryServer(), am.keeper)
}

// BeginBlock executes all ABCI BeginBlock logic respective to the liquidstaking module.
func (am AppModule) BeginBlock(ctx sdk.Context, _ abci.RequestBeginBlock) {
	BeginBlocker(ctx, am.keeper)
}

// EndBlock executes all ABCI EndBlock logic respective to the liquidstaking module. It
// returns no validator updates.
func (am AppModule) EndBlock(ctx sdk.Context, _ abci.RequestEndBlock) []abci.ValidatorUpdate {
	EndBlocker(ctx, am.keeper)
	return nil
}

// InitGenesis performs the liquidstaking module's genesis initialization. It returns
// no validator updates.
func (am AppModule) InitGenesis(ctx sdk.Context, cdc codec.JSONCodec, data json.RawMessage) []abci.ValidatorUpdate {
	var genesisState types.GenesisState

	cdc.MustUnmarshalJSON(data, &genesisState)
	InitGenesis(ctx, am.keeper, genesisState)
	return []abci.ValidatorUpdate{}
}

// ExportGenesis returns the liquidstaking module's exported genesis state as raw JSON bytes.
func (am AppModule) ExportGenesis(ctx sdk.Context, cdc codec.JSONCodec) json.RawMessage {
	gs := ExportGenesis(ctx, am.keeper)
	return cdc.MustMarshalJSON(gs)
}

// ___________________________________________________________________________

// AppModuleSimulation functions

// GenerateGenesisState creates a randomized GenState of the liquidstaking module.
func (am AppModule) GenerateGenesisState(simState *module.SimulationState) {
	simulation.RandomizedGenState(simState)
}

// ProposalContents returns content functions for governance proposals.
func (am AppModule) ProposalContents(simState module.SimulationState) []simtypes.WeightedProposalContent {
	return []simtypes.WeightedProposalContent{}
}

// RandomizedParams creates randomized liquidstaking param changes for the simulator.
func (am AppModule) RandomizedParams(r *rand.Rand) []simtypes.ParamChange {
	return []simtypes.ParamChange{}
}

// RegisterStoreDecoder registers a decoder for liquidstaking module's types.
func (am AppModule) RegisterStoreDecoder(sdr sdk.StoreDecoderRegistry) {
	sdr[types.StoreKey] = simulation.NewDecodeStore(am.cdc)
}

// WeightedOperations returns liquidstaking module weighted operations
func (am AppModule) WeightedOperations(simState module.SimulationState) []simtypes.WeightedOperation {
	return simulation.WeightedOperations(
		simState.AppParams, simState.Cdc, am.ak, am.bk, am.sk, am.keeper,
	)
}

func (am AppModule) AdvanceEpochBeginBlock(ctx sdk.Context) {
	bondDenom := am.sk.BondDenom(ctx)
	lsmEpoch := am.keeper.GetEpoch(ctx)
	ctx = ctx.WithBlockTime(lsmEpoch.StartTime.Add(lsmEpoch.Duration))
	staking.BeginBlocker(ctx, am.sk)

	// mimic the begin block logic of epoch module
	// currently epoch module use hooks when begin block and inflation module
	// implemented that hook, so actual logic is in inflation module.
	{
		epochMintProvision, found := am.ik.GetEpochMintProvision(ctx)
		if !found {
			panic("epoch mint provision not found")
		}
		inflationParams := am.ik.GetParams(ctx)
		// mintedCoin := sdk.NewCoin(inflationParams.MintDenom, epochMintProvision.TruncateInt())
		mintedCoin := sdk.NewCoin(inflationParams.MintDenom, sdk.TokensFromConsensusPower(100, ethermint.PowerReduction))
		staking, communityPool, err := am.ik.MintAndAllocateInflation(ctx, mintedCoin)
		if err != nil {
			panic(err)
		}
		defer func() {
			if mintedCoin.Amount.IsInt64() {
				telemetry.IncrCounterWithLabels(
					[]string{"inflation", "allocate", "total"},
					float32(mintedCoin.Amount.Int64()),
					[]metrics.Label{telemetry.NewLabel("denom", mintedCoin.Denom)},
				)
			}
			if staking.AmountOf(mintedCoin.Denom).IsInt64() {
				telemetry.IncrCounterWithLabels(
					[]string{"inflation", "allocate", "staking", "total"},
					float32(staking.AmountOf(mintedCoin.Denom).Int64()),
					[]metrics.Label{telemetry.NewLabel("denom", mintedCoin.Denom)},
				)
			}
			if communityPool.AmountOf(mintedCoin.Denom).IsInt64() {
				telemetry.IncrCounterWithLabels(
					[]string{"inflation", "allocate", "community_pool", "total"},
					float32(communityPool.AmountOf(mintedCoin.Denom).Int64()),
					[]metrics.Label{telemetry.NewLabel("denom", mintedCoin.Denom)},
				)
			}
		}()

		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				inflationtypes.EventTypeMint,
				sdk.NewAttribute(inflationtypes.AttributeEpochNumber, fmt.Sprintf("%d", -1)),
				sdk.NewAttribute(inflationtypes.AttributeKeyEpochProvisions, epochMintProvision.String()),
				sdk.NewAttribute(sdk.AttributeKeyAmount, mintedCoin.Amount.String()),
			),
		)
	}

	feeCollector := am.ak.GetModuleAccount(ctx, authtypes.FeeCollectorName)
	// mimic the begin block logic of distribution module
	{
		feeCollectorBalance := am.bk.SpendableCoins(ctx, feeCollector.GetAddress())
		rewardsToBeDistributed := feeCollectorBalance.AmountOf(bondDenom)

		// mimic distribution.BeginBlock (AllocateTokens, get rewards from feeCollector, AllocateTokensToValidator, add remaining to feePool)
		err := am.bk.SendCoinsFromModuleToModule(ctx, authtypes.FeeCollectorName, distrtypes.ModuleName, feeCollectorBalance)
		if err != nil {
			panic(err)
		}
		totalRewards := sdk.ZeroDec()
		totalPower := int64(0)
		am.sk.IterateBondedValidatorsByPower(ctx, func(index int64, validator stakingtypes.ValidatorI) (stop bool) {
			consPower := validator.GetConsensusPower(am.sk.PowerReduction(ctx))
			totalPower = totalPower + consPower
			return false
		})
		if totalPower != 0 {
			am.sk.IterateBondedValidatorsByPower(ctx, func(index int64, validator stakingtypes.ValidatorI) (stop bool) {
				consPower := validator.GetConsensusPower(am.sk.PowerReduction(ctx))
				powerFraction := sdk.NewDec(consPower).QuoTruncate(sdk.NewDec(totalPower))
				reward := rewardsToBeDistributed.ToDec().MulTruncate(powerFraction)
				am.dk.AllocateTokensToValidator(ctx, validator, sdk.DecCoins{{Denom: bondDenom, Amount: reward}})
				totalRewards = totalRewards.Add(reward)
				return false
			})
		}
		remaining := rewardsToBeDistributed.ToDec().Sub(totalRewards)
		feePool := am.dk.GetFeePool(ctx)
		feePool.CommunityPool = feePool.CommunityPool.Add(sdk.DecCoins{
			{Denom: bondDenom, Amount: remaining}}...)
		am.dk.SetFeePool(ctx, feePool)
	}
	am.keeper.CoverRedelegationPenalty(ctx)
}

func (am AppModule) AdvanceEpochEndBlock(ctx sdk.Context) {
	lsmEpoch := am.keeper.GetEpoch(ctx)
	ctx = ctx.WithBlockTime(lsmEpoch.StartTime.Add(lsmEpoch.Duration))

	staking.EndBlocker(ctx, am.sk)
	// mimic liquidstaking endblocker except increasing epoch
	{
		am.keeper.DistributeReward(ctx)
		am.keeper.CoverSlashingAndHandleMatureUnbondings(ctx)
		am.keeper.RemoveDeletableRedelegationInfos(ctx)
		am.keeper.HandleQueuedLiquidUnstakes(ctx)
		am.keeper.HandleUnprocessedQueuedLiquidUnstakes(ctx)
		am.keeper.HandleQueuedWithdrawInsuranceRequests(ctx)
		newlyRankedInInsurances, rankOutInsurances := am.keeper.RankInsurances(ctx)
		am.keeper.RePairRankedInsurances(ctx, newlyRankedInInsurances, rankOutInsurances)
		am.keeper.IncrementEpoch(ctx)
	}
}
