package simulation

import (
	"math/rand"
	"strings"

	"github.com/Canto-Network/Canto/v7/contracts"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	"github.com/cosmos/cosmos-sdk/x/simulation"
	"github.com/ethereum/go-ethereum/common"
	"github.com/evmos/ethermint/crypto/ethsecp256k1"
	"github.com/evmos/ethermint/tests"
	evmkeeper "github.com/evmos/ethermint/x/evm/keeper"
	feemarketkeeper "github.com/evmos/ethermint/x/feemarket/keeper"

	"github.com/Canto-Network/Canto/v7/app/params"
	"github.com/Canto-Network/Canto/v7/x/erc20"
	"github.com/Canto-Network/Canto/v7/x/erc20/keeper"
	"github.com/Canto-Network/Canto/v7/x/erc20/types"
)

// Simulation operation weights constants.
const (
	OpWeightSimulateRegisterCoinProposal          = "op_weight_register_coin_proposal"
	OpWeightSimulateRegisterERC20Proposal         = "op_weight_register_erc20_proposal"
	OpWeightSimulateToggleTokenConversionProposal = "op_weight_toggle_token_conversion_proposal"

	erc20Decimals = uint8(18)
)

// ProposalContents defines the module weighted proposals' contents
func ProposalContents(k keeper.Keeper, bk bankkeeper.Keeper, ek evmkeeper.Keeper, fk feemarketkeeper.Keeper) []simtypes.WeightedProposalContent {
	return []simtypes.WeightedProposalContent{
		simulation.NewWeightedProposalContent(
			OpWeightSimulateRegisterCoinProposal,
			params.DefaultWeightRegisterCoinProposal,
			SimulateRegisterCoinProposal(k, bk),
		),
		simulation.NewWeightedProposalContent(
			OpWeightSimulateRegisterERC20Proposal,
			params.DefaultWeightRegisterERC20Proposal,
			SimulateRegisterERC20Proposal(k, ek, fk),
		),
		simulation.NewWeightedProposalContent(
			OpWeightSimulateToggleTokenConversionProposal,
			params.DefaultWeightToggleTokenConversionProposal,
			SimulateToggleTokenConversionProposal(k, ek, fk),
		),
	}
}

func SimulateRegisterCoinProposal(k keeper.Keeper, bk bankkeeper.Keeper) simtypes.ContentSimulatorFn {
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

		if err := erc20.NewErc20ProposalHandler(&k)(ctx, &proposal); err != nil {
			panic(err)
		}

		return nil
	}
}

func SimulateRegisterERC20Proposal(k keeper.Keeper, evmKeepr evmkeeper.Keeper, feemarketKeeper feemarketkeeper.Keeper) simtypes.ContentSimulatorFn {
	return func(r *rand.Rand, ctx sdk.Context, accs []simtypes.Account) simtypes.Content {
		params := k.GetParams(ctx)
		params.EnableErc20 = true
		k.SetParams(ctx, params)

		// account key
		priv, err := ethsecp256k1.GenerateKey()
		if err != nil {
			panic(err)
		}
		address := common.BytesToAddress(priv.PubKey().Address().Bytes())
		signer := tests.NewSigner(priv)

		erc20Name := simtypes.RandStringOfLength(r, 10)
		erc20Symbol := strings.ToUpper(simtypes.RandStringOfLength(r, 4))
		contractAddr, err := erc20.DeployContract(ctx, evmKeepr, feemarketKeeper, address, signer, erc20Name, erc20Symbol, erc20Decimals)
		if err != nil {
			panic(err)
		}

		erc20ABI := contracts.ERC20MinterBurnerDecimalsContract.ABI

		// mint cosmos coin to random accounts
		randomIteration := r.Intn(10)
		for i := 0; i < randomIteration; i++ {
			simAccount, _ := simtypes.RandomAcc(r, accs)

			mintAmt := sdk.NewInt(100000000)
			receiver := common.BytesToAddress(simAccount.Address.Bytes())
			before := k.BalanceOf(ctx, erc20ABI, contractAddr, receiver)
			_, err = k.CallEVM(ctx, erc20ABI, types.ModuleAddress, contractAddr, true, "mint", receiver, mintAmt.BigInt())
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

		if err := erc20.NewErc20ProposalHandler(&k)(ctx, &proposal); err != nil {
			panic(err)
		}

		return nil
	}
}

func SimulateToggleTokenConversionProposal(k keeper.Keeper, evmKeeper evmkeeper.Keeper, feemarketKeeper feemarketkeeper.Keeper) simtypes.ContentSimulatorFn {
	return func(r *rand.Rand, ctx sdk.Context, accs []simtypes.Account) simtypes.Content {
		params := k.GetParams(ctx)
		params.EnableErc20 = true
		k.SetParams(ctx, params)

		// account key
		priv, err := ethsecp256k1.GenerateKey()
		if err != nil {
			panic(err)
		}
		address := common.BytesToAddress(priv.PubKey().Address().Bytes())
		signer := tests.NewSigner(priv)

		erc20Name := simtypes.RandStringOfLength(r, 10)
		erc20Symbol := strings.ToUpper(simtypes.RandStringOfLength(r, 4))

		contractAddr, err := erc20.DeployContract(ctx, evmKeeper, feemarketKeeper, address, signer, erc20Name, erc20Symbol, erc20Decimals)
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

		if err := erc20.NewErc20ProposalHandler(&k)(ctx, &proposal); err != nil {
			panic(err)
		}

		return nil
	}
}
