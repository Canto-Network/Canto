package erc20

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/spf13/cobra"

	"cosmossdk.io/core/address"
	"cosmossdk.io/core/appmodule"
	"github.com/Canto-Network/Canto/v8/x/erc20/simulation"
	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"

	"github.com/Canto-Network/Canto/v8/x/erc20/client/cli"
	"github.com/Canto-Network/Canto/v8/x/erc20/keeper"
	"github.com/Canto-Network/Canto/v8/x/erc20/types"
)

// type check to ensure the interface is properly implemented
var (
	_ module.AppModuleBasic      = AppModuleBasic{}
	_ module.AppModuleBasic      = AppModule{}
	_ module.HasServices         = AppModule{}
	_ module.HasABCIGenesis      = AppModule{}
	_ module.AppModuleSimulation = AppModule{}

	_ appmodule.AppModule = AppModule{}
)

// app module Basics object
type AppModuleBasic struct {
	ac  address.Codec
	cdc codec.Codec
}

func NewAppModuleBasic(ac address.Codec, cdc codec.Codec) AppModuleBasic {
	return AppModuleBasic{ac: ac, cdc: cdc}
}

func (AppModuleBasic) Name() string {
	return types.ModuleName
}

// RegisterLegacyAminoCodec performs a no-op as the erc20 doesn't support Amino encoding
func (AppModuleBasic) RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	types.RegisterLegacyAminoCodec(cdc)
}

// ConsensusVersion returns the consensus state-breaking version for the module.
func (AppModuleBasic) ConsensusVersion() uint64 {
	return 2
}

// RegisterInterfaces registers interfaces and implementations of the erc20 module.
func (AppModuleBasic) RegisterInterfaces(interfaceRegistry codectypes.InterfaceRegistry) {
	types.RegisterInterfaces(interfaceRegistry)
}

// DefaultGenesis returns default genesis state as raw bytes for the erc20
// module.
func (AppModuleBasic) DefaultGenesis(cdc codec.JSONCodec) json.RawMessage {
	return cdc.MustMarshalJSON(types.DefaultGenesisState())
}

func (b AppModuleBasic) ValidateGenesis(cdc codec.JSONCodec, config client.TxEncodingConfig, bz json.RawMessage) error {
	var genesisState types.GenesisState
	if err := cdc.UnmarshalJSON(bz, &genesisState); err != nil {
		return fmt.Errorf("failed to unmarshal %s genesis state: %w", types.ModuleName, err)
	}

	return genesisState.Validate()
}

func (b AppModuleBasic) RegisterGRPCGatewayRoutes(c client.Context, serveMux *runtime.ServeMux) {
	if err := types.RegisterQueryHandlerClient(context.Background(), serveMux, types.NewQueryClient(c)); err != nil {
		panic(err)
	}
}

// GetTxCmd returns the root tx command for the erc20 module.
func (b AppModuleBasic) GetTxCmd() *cobra.Command {
	return cli.NewTxCmd(b.ac)
}

// GetQueryCmd returns no root query command for the erc20 module.
func (AppModuleBasic) GetQueryCmd() *cobra.Command {
	return cli.GetQueryCmd()
}

type AppModule struct {
	AppModuleBasic
	keeper keeper.Keeper
	// TODO: Add keepers that your application module simulation requires
	ak authkeeper.AccountKeeper
	bk types.BankKeeper
	ek types.EVMKeeper
	fk types.FeeMarketKeeper
}

// NewAppModule creates a new AppModule Object
func NewAppModule(
	cdc codec.Codec,
	k keeper.Keeper,
	ak authkeeper.AccountKeeper,
	bk types.BankKeeper,
	ek types.EVMKeeper,
	fk types.FeeMarketKeeper,
	ac address.Codec,
) AppModule {
	return AppModule{
		AppModuleBasic: AppModuleBasic{ac: ac, cdc: cdc},
		keeper:         k,
		ak:             ak,
		bk:             bk,
		ek:             ek,
		fk:             fk,
	}
}

// IsOnePerModuleType implements the depinject.OnePerModuleType interface.
func (am AppModule) IsOnePerModuleType() {}

// IsAppModule implements the appmodule.AppModule interface.
func (am AppModule) IsAppModule() {}

func (AppModule) Name() string {
	return types.ModuleName
}

func (am AppModule) RegisterInvariants(ir sdk.InvariantRegistry) {}

func (am AppModule) RegisterServices(cfg module.Configurator) {
	types.RegisterMsgServer(cfg.MsgServer(), am.keeper)
	types.RegisterQueryServer(cfg.QueryServer(), am.keeper)

	migrator := keeper.NewMigrator(am.keeper)

	// NOTE: the migrations below will only run if the consensus version has changed
	// since the last release

	// register v1 -> v2 migration
	if err := cfg.RegisterMigration(types.ModuleName, 1, migrator.Migrate1to2); err != nil {
		panic(fmt.Errorf("failed to migrate %s to v2: %w", types.ModuleName, err))
	}
}

// ProposalMsgs returns msgs used for governance proposals for simulations.
func (am AppModule) ProposalMsgs(simState module.SimulationState) []simtypes.WeightedProposalMsg {
	return simulation.ProposalMsgs(am.keeper, am.ak, am.bk, am.ek, am.fk)
}

func (am AppModule) InitGenesis(ctx sdk.Context, cdc codec.JSONCodec, data json.RawMessage) []abci.ValidatorUpdate {
	var genesisState types.GenesisState

	cdc.MustUnmarshalJSON(data, &genesisState)
	InitGenesis(ctx, am.keeper, am.ak, genesisState)
	return []abci.ValidatorUpdate{}
}

func (am AppModule) ExportGenesis(ctx sdk.Context, cdc codec.JSONCodec) json.RawMessage {
	gs := ExportGenesis(ctx, am.keeper)
	return cdc.MustMarshalJSON(gs)
}

func (am AppModule) GenerateGenesisState(input *module.SimulationState) {
}

func (am AppModule) ProposalContents(simState module.SimulationState) []simtypes.WeightedProposalContent {
	return []simtypes.WeightedProposalContent{}
}

func (am AppModule) RegisterStoreDecoder(decoderRegistry simtypes.StoreDecoderRegistry) {
	decoderRegistry[types.ModuleName] = simulation.NewDecodeStore(am.cdc)
}

func (am AppModule) WeightedOperations(simState module.SimulationState) []simtypes.WeightedOperation {
	return simulation.WeightedOperations(
		simState.AppParams, simState.Cdc, am.keeper, am.ak, am.bk, am.ek, am.fk,
	)
}
