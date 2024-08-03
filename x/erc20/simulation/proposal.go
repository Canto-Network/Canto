package simulation

import (
	"math/rand"
	"strings"

	sdkmath "cosmossdk.io/math"
	"github.com/Canto-Network/Canto/v8/app/params"
	"github.com/Canto-Network/Canto/v8/contracts"
	"github.com/Canto-Network/Canto/v8/x/erc20/keeper"
	"github.com/Canto-Network/Canto/v8/x/erc20/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/simulation"
	"github.com/ethereum/go-ethereum/common"
	"github.com/evmos/ethermint/crypto/ethsecp256k1"
	"github.com/evmos/ethermint/tests"
	evmtypes "github.com/evmos/ethermint/x/evm/types"
)

// Simulation operation weights constants
const (
	DefaultWeightMsgUpdateParams int = 100

	OpWeightMsgUpdateParams                       = "op_weight_msg_update_params"
	OpWeightSimulateRegisterCoinProposal          = "op_weight_register_coin_proposal"
	OpWeightSimulateRegisterERC20Proposal         = "op_weight_register_erc20_proposal"
	OpWeightSimulateToggleTokenConversionProposal = "op_weight_toggle_token_conversion_proposal"

	erc20Decimals = uint8(18)
)

// ProposalMsgs defines the module weighted proposals' contents
func ProposalMsgs(k keeper.Keeper, ak types.AccountKeeper, bk types.BankKeeper, ek types.EVMKeeper, fk types.FeeMarketKeeper) []simtypes.WeightedProposalMsg {
	return []simtypes.WeightedProposalMsg{
		simulation.NewWeightedProposalMsg(
			OpWeightMsgUpdateParams,
			DefaultWeightMsgUpdateParams,
			SimulateMsgUpdateParams,
		),
		simulation.NewWeightedProposalMsg(
			OpWeightSimulateRegisterCoinProposal,
			params.DefaultWeightRegisterCoinProposal,
			SimulateMsgRegisterCoin(k, bk),
		),
		simulation.NewWeightedProposalMsg(
			OpWeightSimulateRegisterERC20Proposal,
			params.DefaultWeightRegisterERC20Proposal,
			SimulateMsgRegisterERC20(k, ak, bk, ek, fk),
		),
		simulation.NewWeightedProposalMsg(
			OpWeightSimulateToggleTokenConversionProposal,
			params.DefaultWeightToggleTokenConversionProposal,
			SimulateMsgToggleTokenConversion(k, bk, ek, fk),
		),
	}
}

// SimulateMsgUpdateParams returns a random MsgUpdateParams
func SimulateMsgUpdateParams(r *rand.Rand, _ sdk.Context, _ []simtypes.Account) sdk.Msg {
	// use the default gov module account address as authority
	var authority sdk.AccAddress = address.Module("gov")

	params := types.DefaultParams()

	params.EnableErc20 = generateRandomBool(r)
	params.EnableEVMHook = generateRandomBool(r)

	return &types.MsgUpdateParams{
		Authority: authority.String(),
		Params:    params,
	}
}

func SimulateRegisterCoin(r *rand.Rand, ctx sdk.Context, accs []simtypes.Account, k keeper.Keeper, bk types.BankKeeper) (sdk.Msg, error) {
	coinMetadata := types.GenRandomCoinMetadata(r)
	if err := bk.MintCoins(ctx, types.ModuleName, sdk.NewCoins(sdk.NewCoin(coinMetadata.Base, sdkmath.NewInt(1000000000000)))); err != nil {
		panic(err)
	}
	bankparams := bk.GetParams(ctx)
	bankparams.DefaultSendEnabled = true
	bk.SetParams(ctx, bankparams)

	params := k.GetParams(ctx)
	params.EnableErc20 = true
	k.SetParams(ctx, params)

	// mint cosmos coin to random accounts
	randomIteration := r.Intn(10)
	mintAmt := sdkmath.NewInt(100000000)
	for i := 0; i < randomIteration; i++ {
		simAccount, _ := simtypes.RandomAcc(r, accs)

		if err := bk.MintCoins(ctx, types.ModuleName, sdk.NewCoins(sdk.NewCoin(coinMetadata.Base, mintAmt))); err != nil {
			return &types.MsgRegisterCoin{}, err
		}
		if err := bk.SendCoinsFromModuleToAccount(ctx, types.ModuleName, simAccount.Address, sdk.NewCoins(sdk.NewCoin(coinMetadata.Base, sdkmath.NewInt(1000000000)))); err != nil {
			return &types.MsgRegisterCoin{}, err
		}

	}

	// use the default gov module account address as authority
	var authority sdk.AccAddress = address.Module("gov")

	msg := &types.MsgRegisterCoin{
		Authority:   authority.String(),
		Title:       simtypes.RandStringOfLength(r, 10),
		Description: simtypes.RandStringOfLength(r, 100),
		Metadata:    coinMetadata,
	}

	if _, err := k.RegisterCoinProposal(ctx, msg); err != nil {
		return &types.MsgRegisterCoin{}, err
	}

	return msg, nil
}

func SimulateMsgRegisterCoin(k keeper.Keeper, bk types.BankKeeper) simtypes.MsgSimulatorFn {
	return func(r *rand.Rand, ctx sdk.Context, accs []simtypes.Account) sdk.Msg {
		msg, err := SimulateRegisterCoin(r, ctx, accs, k, bk)
		if err != nil {
			panic(err)
		}
		return msg
	}
}

func SimulateRegisterERC20(r *rand.Rand, ctx sdk.Context, accs []simtypes.Account, k keeper.Keeper, ak types.AccountKeeper, bk types.BankKeeper, ek types.EVMKeeper, fk types.FeeMarketKeeper) (sdk.Msg, error) {
	params := k.GetParams(ctx)
	params.EnableErc20 = true
	k.SetParams(ctx, params)

	evmParams := evmtypes.DefaultParams()
	evmParams.EvmDenom = "stake"
	ek.SetParams(ctx, evmParams)

	// account key
	priv, err := ethsecp256k1.GenerateKey()
	if err != nil {
		panic(err)
	}
	addr := common.BytesToAddress(priv.PubKey().Address().Bytes())
	signer := tests.NewSigner(priv)

	erc20ABI := contracts.ERC20MinterBurnerDecimalsContract.ABI

	var deployer common.Address
	var contractAddr common.Address
	coinMetadata := types.GenRandomCoinMetadata(r)

	deployer = addr
	erc20Name := coinMetadata.Name
	erc20Symbol := coinMetadata.Symbol
	contractAddr, err = keeper.DeployContract(ctx, ek, fk, deployer, signer, erc20Name, erc20Symbol, erc20Decimals)

	// mint cosmos coin to random accounts
	randomIteration := r.Intn(10)
	for i := 0; i < randomIteration; i++ {
		simAccount, _ := simtypes.RandomAcc(r, accs)

		mintAmt := sdkmath.NewInt(100000000)
		receiver := common.BytesToAddress(simAccount.Address.Bytes())
		before := k.BalanceOf(ctx, erc20ABI, contractAddr, receiver)
		_, err = k.CallEVM(ctx, erc20ABI, deployer, contractAddr, true, "mint", receiver, mintAmt.BigInt())
		if err != nil {
			return &types.MsgRegisterERC20{}, err
		}
		after := k.BalanceOf(ctx, erc20ABI, contractAddr, receiver)
		if after.Cmp(before.Add(before, mintAmt.BigInt())) != 0 {
			return &types.MsgRegisterERC20{}, err
		}
	}

	// use the default gov module account address as authority
	var authority sdk.AccAddress = address.Module("gov")

	msg := &types.MsgRegisterERC20{
		Authority:    authority.String(),
		Title:        simtypes.RandStringOfLength(r, 10),
		Description:  simtypes.RandStringOfLength(r, 100),
		Erc20Address: contractAddr.String(),
	}

	if _, err := k.RegisterERC20Proposal(ctx, msg); err != nil {
		return &types.MsgRegisterERC20{}, err
	}

	return msg, nil
}

func SimulateMsgRegisterERC20(k keeper.Keeper, accountKeeper types.AccountKeeper, bankKeeper types.BankKeeper, evmKeeper types.EVMKeeper, feemarketKeeper types.FeeMarketKeeper) simtypes.MsgSimulatorFn {
	return func(r *rand.Rand, ctx sdk.Context, accs []simtypes.Account) sdk.Msg { // use the default gov module account address as authority
		msg, err := SimulateRegisterERC20(r, ctx, accs, k, accountKeeper, bankKeeper, evmKeeper, feemarketKeeper)
		if err != nil {
			panic(err)
		}
		return msg
	}
}

func SimulateMsgToggleTokenConversion(k keeper.Keeper, bankKeeper types.BankKeeper, evmKeeper types.EVMKeeper, feemarketKeeper types.FeeMarketKeeper) simtypes.MsgSimulatorFn {
	return func(r *rand.Rand, ctx sdk.Context, accs []simtypes.Account) sdk.Msg {
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
		addr := common.BytesToAddress(priv.PubKey().Address().Bytes())
		signer := tests.NewSigner(priv)

		erc20Name := simtypes.RandStringOfLength(r, 10)
		erc20Symbol := strings.ToUpper(simtypes.RandStringOfLength(r, 4))

		coins := sdk.NewCoins(sdk.NewCoin(evmParams.EvmDenom, sdkmath.NewInt(10000000000000000)))
		if err = bankKeeper.MintCoins(ctx, types.ModuleName, coins); err != nil {
			panic(err)
		}

		if err = bankKeeper.SendCoinsFromModuleToModule(ctx, types.ModuleName, authtypes.FeeCollectorName, coins); err != nil {
			panic(err)
		}

		contractAddr, err := keeper.DeployContract(ctx, evmKeeper, feemarketKeeper, addr, signer, erc20Name, erc20Symbol, erc20Decimals)
		if err != nil {
			panic(err)
		}

		_, err = k.RegisterERC20(ctx, contractAddr)
		if err != nil {
			panic(err)
		}

		var authority sdk.AccAddress = address.Module("gov")

		msg := &types.MsgToggleTokenConversion{
			Authority:   authority.String(),
			Title:       simtypes.RandStringOfLength(r, 10),
			Description: simtypes.RandStringOfLength(r, 100),
			Token:       contractAddr.String(),
		}

		if _, err := k.ToggleTokenConversionProposal(ctx, msg); err != nil {
			panic(err)
		}

		return msg
	}
}
