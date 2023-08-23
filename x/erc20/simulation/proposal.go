package simulation

import (
	"math/rand"
	"strings"

	"github.com/Canto-Network/Canto/v7/app/params"
	"github.com/Canto-Network/Canto/v7/contracts"
	"github.com/Canto-Network/Canto/v7/x/erc20/keeper"
	"github.com/Canto-Network/Canto/v7/x/erc20/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/simulation"
	"github.com/ethereum/go-ethereum/common"
	"github.com/evmos/ethermint/crypto/ethsecp256k1"
	"github.com/evmos/ethermint/tests"
	evmtypes "github.com/evmos/ethermint/x/evm/types"
)

// Simulation operation weights constants.
const (
	OpWeightSimulateRegisterCoinProposal          = "op_weight_register_coin_proposal"
	OpWeightSimulateRegisterERC20Proposal         = "op_weight_register_erc20_proposal"
	OpWeightSimulateToggleTokenConversionProposal = "op_weight_toggle_token_conversion_proposal"

	erc20Decimals = uint8(18)
)

// ProposalContents defines the module weighted proposals' contents
func ProposalContents(k keeper.Keeper, ak types.AccountKeeper, bk types.BankKeeper, ek types.EVMKeeper, fk types.FeeMarketKeeper) []simtypes.WeightedProposalContent {
	return []simtypes.WeightedProposalContent{
		simulation.NewWeightedProposalContent(
			OpWeightSimulateRegisterCoinProposal,
			params.DefaultWeightRegisterCoinProposal,
			SimulateRegisterCoinProposal(k, bk),
		),
		simulation.NewWeightedProposalContent(
			OpWeightSimulateRegisterERC20Proposal,
			params.DefaultWeightRegisterERC20Proposal,
			SimulateRegisterERC20Proposal(k, ak, bk, ek, fk),
		),
		simulation.NewWeightedProposalContent(
			OpWeightSimulateToggleTokenConversionProposal,
			params.DefaultWeightToggleTokenConversionProposal,
			SimulateToggleTokenConversionProposal(k, bk, ek, fk),
		),
	}
}

func SimulateRegisterCoinProposal(k keeper.Keeper, bk types.BankKeeper) simtypes.ContentSimulatorFn {
	return func(r *rand.Rand, ctx sdk.Context, accs []simtypes.Account) simtypes.Content {
		coinMetadata := types.GenRandomCoinMetadata(r)
		if err := bk.MintCoins(ctx, types.ModuleName, sdk.NewCoins(sdk.NewCoin(coinMetadata.Base, sdk.NewInt(1)))); err != nil {
			panic(err)
		}

		params := k.GetParams(ctx)
		params.EnableErc20 = true
		k.SetParams(ctx, params)

		// mint cosmos coin to random accounts
		randomIteration := r.Intn(10)
		for i := 0; i < randomIteration; i++ {
			simAccount, _ := simtypes.RandomAcc(r, accs)

			if err := bk.MintCoins(ctx, types.ModuleName, sdk.NewCoins(sdk.NewCoin(coinMetadata.Base, sdk.NewInt(100000000)))); err != nil {
				panic(err)
			}
			if err := bk.SendCoinsFromModuleToAccount(ctx, types.ModuleName, simAccount.Address, sdk.NewCoins(sdk.NewCoin(coinMetadata.Base, sdk.NewInt(100000000)))); err != nil {
				panic(err)
			}
		}

		proposal := types.RegisterCoinProposal{
			Title:       simtypes.RandStringOfLength(r, 10),
			Description: simtypes.RandStringOfLength(r, 100),
			Metadata:    coinMetadata,
		}

		if err := keeper.NewErc20ProposalHandler(&k)(ctx, &proposal); err != nil {
			panic(err)
		}

		return nil
	}
}

func SimulateRegisterERC20Proposal(k keeper.Keeper, accountKeeper types.AccountKeeper, bankKeeper types.BankKeeper, evmKeeper types.EVMKeeper, feemarketKeeper types.FeeMarketKeeper) simtypes.ContentSimulatorFn {
	return func(r *rand.Rand, ctx sdk.Context, accs []simtypes.Account) simtypes.Content {
		params := k.GetParams(ctx)
		params.EnableErc20 = true
		k.SetParams(ctx, params)

		evmParams := evmtypes.DefaultParams()
		evmParams.EvmDenom = "stake"
		evmKeeper.SetParams(ctx, evmParams)

		isNativeErc20 := r.Intn(2) == 1
		// account key
		priv, err := ethsecp256k1.GenerateKey()
		if err != nil {
			panic(err)
		}
		address := common.BytesToAddress(priv.PubKey().Address().Bytes())
		signer := tests.NewSigner(priv)

		erc20ABI := contracts.ERC20MinterBurnerDecimalsContract.ABI

		var deployer common.Address
		var contractAddr common.Address
		coinMetadata := types.GenRandomCoinMetadata(r)

		coins := sdk.NewCoins(sdk.NewCoin(evmParams.EvmDenom, sdk.NewInt(10000000000000000)))
		if err = bankKeeper.MintCoins(ctx, types.ModuleName, coins); err != nil {
			panic(err)
		}

		if err = bankKeeper.SendCoinsFromModuleToModule(ctx, types.ModuleName, authtypes.FeeCollectorName, coins); err != nil {
			panic(err)
		}
		if isNativeErc20 {
			deployer = types.ModuleAddress
			contractAddr, err = keeper.DeployERC20Contract(ctx, k, accountKeeper, coinMetadata)
		} else {
			deployer = address
			erc20Name := coinMetadata.Name
			erc20Symbol := coinMetadata.Symbol
			contractAddr, err = keeper.DeployContract(ctx, evmKeeper, feemarketKeeper, deployer, signer, erc20Name, erc20Symbol, erc20Decimals)
		}

		if err != nil {
			panic(err)
		}

		// mint cosmos coin to random accounts
		randomIteration := r.Intn(10)
		for i := 0; i < randomIteration; i++ {
			simAccount, _ := simtypes.RandomAcc(r, accs)

			mintAmt := sdk.NewInt(100000000)
			receiver := common.BytesToAddress(simAccount.Address.Bytes())
			before := k.BalanceOf(ctx, erc20ABI, contractAddr, receiver)
			_, err = k.CallEVM(ctx, erc20ABI, deployer, contractAddr, true, "mint", receiver, mintAmt.BigInt())
			if err != nil {
				panic(err)
			}
			after := k.BalanceOf(ctx, erc20ABI, contractAddr, receiver)
			if after.Cmp(before.Add(before, mintAmt.BigInt())) != 0 {
				panic("mint failed")
			}
		}

		if err != nil {
			panic(err)
		}

		proposal := types.RegisterERC20Proposal{
			Title:        simtypes.RandStringOfLength(r, 10),
			Description:  simtypes.RandStringOfLength(r, 100),
			Erc20Address: contractAddr.String(),
		}

		if err := keeper.NewErc20ProposalHandler(&k)(ctx, &proposal); err != nil {
			panic(err)
		}

		return nil
	}
}

func SimulateToggleTokenConversionProposal(k keeper.Keeper, bankKeeper types.BankKeeper, evmKeeper types.EVMKeeper, feemarketKeeper types.FeeMarketKeeper) simtypes.ContentSimulatorFn {
	return func(r *rand.Rand, ctx sdk.Context, accs []simtypes.Account) simtypes.Content {
		params := k.GetParams(ctx)
		params.EnableErc20 = true
		k.SetParams(ctx, params)

		evmParams := evmtypes.DefaultParams()
		evmParams.EvmDenom = "stake"
		evmKeeper.SetParams(ctx, evmParams)

		// account key
		priv, err := ethsecp256k1.GenerateKey()
		if err != nil {
			panic(err)
		}
		address := common.BytesToAddress(priv.PubKey().Address().Bytes())
		signer := tests.NewSigner(priv)

		erc20Name := simtypes.RandStringOfLength(r, 10)
		erc20Symbol := strings.ToUpper(simtypes.RandStringOfLength(r, 4))

		coins := sdk.NewCoins(sdk.NewCoin(evmParams.EvmDenom, sdk.NewInt(10000000000000000)))
		if err = bankKeeper.MintCoins(ctx, types.ModuleName, coins); err != nil {
			panic(err)
		}

		if err = bankKeeper.SendCoinsFromModuleToModule(ctx, types.ModuleName, authtypes.FeeCollectorName, coins); err != nil {
			panic(err)
		}

		contractAddr, err := keeper.DeployContract(ctx, evmKeeper, feemarketKeeper, address, signer, erc20Name, erc20Symbol, erc20Decimals)
		if err != nil {
			panic(err)
		}

		_, err = k.RegisterERC20(ctx, contractAddr)
		if err != nil {
			panic(err)
		}

		proposal := types.ToggleTokenConversionProposal{
			Title:       simtypes.RandStringOfLength(r, 10),
			Description: simtypes.RandStringOfLength(r, 100),
			Token:       contractAddr.String(),
		}

		if err := keeper.NewErc20ProposalHandler(&k)(ctx, &proposal); err != nil {
			panic(err)
		}

		return nil
	}
}
