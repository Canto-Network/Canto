package keeper

import (
	"context"
	"fmt"
	"math/big"
	"strings"

	errorsmod "cosmossdk.io/errors"
	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/hashicorp/go-metrics"

	"github.com/Canto-Network/Canto/v7/contracts"
	"github.com/Canto-Network/Canto/v7/x/erc20/types"
)

var _ types.MsgServer = &Keeper{}

// ConvertCoin converts native Cosmos coins into ERC20 tokens for both
// Cosmos-native and ERC20 TokenPair Owners
func (k Keeper) ConvertCoin(
	goCtx context.Context,
	msg *types.MsgConvertCoin,
) (*types.MsgConvertCoinResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Error checked during msg validation
	receiver := common.HexToAddress(msg.Receiver)
	sender := sdk.MustAccAddressFromBech32(msg.Sender)

	pair, err := k.MintingEnabled(ctx, sender, receiver.Bytes(), msg.Coin.Denom)
	if err != nil {
		return nil, err
	}

	// Remove token pair if contract is suicided
	erc20 := common.HexToAddress(pair.Erc20Address)
	acc := k.evmKeeper.GetAccountWithoutBalance(ctx, erc20)

	if acc == nil || !acc.IsContract() {
		k.DeleteTokenPair(ctx, pair)
		k.Logger(ctx).Debug(
			"deleting selfdestructed token pair from state",
			"contract", pair.Erc20Address,
		)
		// NOTE: return nil error to persist the changes from the deletion
		return nil, nil
	}

	// Check ownership and execute conversion
	switch {
	case pair.IsNativeCoin():
		return k.convertCoinNativeCoin(ctx, pair, msg, receiver, sender) // case 1.1
	case pair.IsNativeERC20():
		return k.convertCoinNativeERC20(ctx, pair, msg, receiver, sender) // case 2.2
	default:
		return nil, types.ErrUndefinedOwner
	}
}

// ConvertERC20 converts ERC20 tokens into native Cosmos coins for both
// Cosmos-native and ERC20 TokenPair Owners
func (k Keeper) ConvertERC20(
	goCtx context.Context,
	msg *types.MsgConvertERC20,
) (*types.MsgConvertERC20Response, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Error checked during msg validation
	receiver := sdk.MustAccAddressFromBech32(msg.Receiver)
	sender := common.HexToAddress(msg.Sender)

	pair, err := k.MintingEnabled(ctx, sender.Bytes(), receiver, msg.ContractAddress)
	if err != nil {
		return nil, err
	}

	// Remove token pair if contract is suicided
	erc20 := common.HexToAddress(pair.Erc20Address)
	acc := k.evmKeeper.GetAccountWithoutBalance(ctx, erc20)

	if acc == nil || !acc.IsContract() {
		k.DeleteTokenPair(ctx, pair)
		k.Logger(ctx).Debug(
			"deleting selfdestructed token pair from state",
			"contract", pair.Erc20Address,
		)
		// NOTE: return nil error to persist the changes from the deletion
		return nil, nil
	}

	// Check ownership and execute conversion
	switch {
	case pair.IsNativeCoin():
		return k.convertERC20NativeCoin(ctx, pair, msg, receiver, sender) // case 1.2
	case pair.IsNativeERC20():
		return k.convertERC20NativeToken(ctx, pair, msg, receiver, sender) // case 2.1
	default:
		return nil, types.ErrUndefinedOwner
	}
}

// convertCoinNativeCoin handles the coin conversion for a native Cosmos coin
// token pair:
//   - escrow coins on module account
//   - mint tokens and send to receiver
//   - check if token balance increased by amount
func (k Keeper) convertCoinNativeCoin(
	ctx sdk.Context,
	pair types.TokenPair,
	msg *types.MsgConvertCoin,
	receiver common.Address,
	sender sdk.AccAddress,
) (*types.MsgConvertCoinResponse, error) {
	// NOTE: ignore validation from NewCoin constructor
	coins := sdk.Coins{msg.Coin}
	erc20 := contracts.ERC20MinterBurnerDecimalsContract.ABI
	contract := pair.GetERC20Contract()
	balanceToken := k.BalanceOf(ctx, erc20, contract, receiver)
	if balanceToken == nil {
		return nil, errorsmod.Wrap(types.ErrEVMCall, "failed to retrieve balance")
	}

	// Escrow coins on module account
	err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, sender, types.ModuleName, coins)
	if err != nil {
		return nil, errorsmod.Wrap(err, "failed to escrow coins")
	}

	// Mint tokens and send to receiver
	_, err = k.CallEVM(ctx, erc20, types.ModuleAddress, contract, true, "mint", receiver, msg.Coin.Amount.BigInt())
	if err != nil {
		return nil, err
	}

	// Check expected receiver balance after transfer
	tokens := msg.Coin.Amount.BigInt()
	balanceTokenAfter := k.BalanceOf(ctx, erc20, contract, receiver)
	if balanceTokenAfter == nil {
		return nil, errorsmod.Wrap(types.ErrEVMCall, "failed to retrieve balance")
	}
	expToken := big.NewInt(0).Add(balanceToken, tokens)

	if r := balanceTokenAfter.Cmp(expToken); r != 0 {
		return nil, errorsmod.Wrapf(
			types.ErrBalanceInvariance,
			"invalid token balance - expected: %v, actual: %v", expToken, balanceTokenAfter,
		)
	}

	defer func() {
		telemetry.IncrCounterWithLabels(
			[]string{"tx", "msg", "convert", "coin", "total"},
			1,
			[]metrics.Label{
				telemetry.NewLabel("denom", pair.Denom),
				telemetry.NewLabel("erc20", pair.Erc20Address),
			},
		)

		if msg.Coin.Amount.IsInt64() {
			telemetry.IncrCounterWithLabels(
				[]string{"tx", "msg", "convert", "coin", "amount", "total"},
				float32(msg.Coin.Amount.Int64()),
				[]metrics.Label{
					telemetry.NewLabel("denom", pair.Denom),
					telemetry.NewLabel("erc20", pair.Erc20Address),
				},
			)
		}
	}()

	ctx.EventManager().EmitEvents(
		sdk.Events{
			sdk.NewEvent(
				types.EventTypeConvertCoin,
				sdk.NewAttribute(sdk.AttributeKeySender, msg.Sender),
				sdk.NewAttribute(types.AttributeKeyReceiver, msg.Receiver),
				sdk.NewAttribute(sdk.AttributeKeyAmount, msg.Coin.Amount.String()),
				sdk.NewAttribute(types.AttributeKeyCosmosCoin, msg.Coin.Denom),
				sdk.NewAttribute(types.AttributeKeyERC20Token, pair.Erc20Address),
			),
		},
	)

	return &types.MsgConvertCoinResponse{}, nil
}

// convertERC20NativeCoin handles the erc20 conversion for a native Cosmos coin
// token pair:
//   - burn escrowed tokens
//   - unescrow coins that have been previously escrowed with ConvertCoin
//   - check if coin balance increased by amount
//   - check if token balance decreased by amount
func (k Keeper) convertERC20NativeCoin(
	ctx sdk.Context,
	pair types.TokenPair,
	msg *types.MsgConvertERC20,
	receiver sdk.AccAddress,
	sender common.Address,
) (*types.MsgConvertERC20Response, error) {
	// NOTE: coin fields already validated
	coins := sdk.Coins{sdk.Coin{Denom: pair.Denom, Amount: msg.Amount}}
	erc20 := contracts.ERC20MinterBurnerDecimalsContract.ABI
	contract := pair.GetERC20Contract()
	balanceCoin := k.bankKeeper.GetBalance(ctx, receiver, pair.Denom)
	balanceToken := k.BalanceOf(ctx, erc20, contract, sender)
	if balanceToken == nil {
		return nil, errorsmod.Wrap(types.ErrEVMCall, "failed to retrieve balance")
	}

	// Burn escrowed tokens
	_, err := k.CallEVM(ctx, erc20, types.ModuleAddress, contract, true, "burnCoins", sender, msg.Amount.BigInt())
	if err != nil {
		return nil, err
	}

	// Unescrow coins and send to receiver
	err = k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, receiver, coins)
	if err != nil {
		return nil, err
	}

	// Check expected receiver balance after transfer
	balanceCoinAfter := k.bankKeeper.GetBalance(ctx, receiver, pair.Denom)
	expCoin := balanceCoin.Add(coins[0])
	if ok := balanceCoinAfter.IsEqual(expCoin); !ok {
		return nil, errorsmod.Wrapf(
			types.ErrBalanceInvariance,
			"invalid coin balance - expected: %v, actual: %v",
			expCoin, balanceCoinAfter,
		)
	}

	// Check expected Sender balance after transfer
	tokens := coins[0].Amount.BigInt()
	balanceTokenAfter := k.BalanceOf(ctx, erc20, contract, sender)
	if balanceTokenAfter == nil {
		return nil, errorsmod.Wrap(types.ErrEVMCall, "failed to retrieve balance")
	}

	expToken := big.NewInt(0).Sub(balanceToken, tokens)
	if r := balanceTokenAfter.Cmp(expToken); r != 0 {
		return nil, errorsmod.Wrapf(
			types.ErrBalanceInvariance,
			"invalid token balance - expected: %v, actual: %v",
			expToken, balanceTokenAfter,
		)
	}

	defer func() {
		telemetry.IncrCounterWithLabels(
			[]string{"tx", "msg", "convert", "erc20", "total"},
			1,
			[]metrics.Label{
				telemetry.NewLabel("denom", pair.Denom),
				telemetry.NewLabel("erc20", pair.Erc20Address),
			},
		)

		if msg.Amount.IsInt64() {
			telemetry.IncrCounterWithLabels(
				[]string{"tx", "msg", "convert", "erc20", "amount", "total"},
				float32(msg.Amount.Int64()),
				[]metrics.Label{
					telemetry.NewLabel("denom", pair.Denom),
					telemetry.NewLabel("erc20", pair.Erc20Address),
				},
			)
		}
	}()

	ctx.EventManager().EmitEvents(
		sdk.Events{
			sdk.NewEvent(
				types.EventTypeConvertERC20,
				sdk.NewAttribute(sdk.AttributeKeySender, msg.Sender),
				sdk.NewAttribute(types.AttributeKeyReceiver, msg.Receiver),
				sdk.NewAttribute(sdk.AttributeKeyAmount, msg.Amount.String()),
				sdk.NewAttribute(types.AttributeKeyCosmosCoin, pair.Denom),
				sdk.NewAttribute(types.AttributeKeyERC20Token, msg.ContractAddress),
			),
		},
	)

	return &types.MsgConvertERC20Response{}, nil
}

// convertERC20NativeToken handles the erc20 conversion for a native erc20 token
// pair:
//   - escrow tokens on module account
//   - mint coins on bank module
//   - send minted coins to the receiver
//   - check if coin balance increased by amount
//   - check if token balance decreased by amount
//   - check for unexpected `Approval` event in logs
func (k Keeper) convertERC20NativeToken(
	ctx sdk.Context,
	pair types.TokenPair,
	msg *types.MsgConvertERC20,
	receiver sdk.AccAddress,
	sender common.Address,
) (*types.MsgConvertERC20Response, error) {
	// NOTE: coin fields already validated
	coins := sdk.Coins{sdk.Coin{Denom: pair.Denom, Amount: msg.Amount}}
	erc20 := contracts.ERC20MinterBurnerDecimalsContract.ABI
	contract := pair.GetERC20Contract()
	balanceCoin := k.bankKeeper.GetBalance(ctx, receiver, pair.Denom)
	balanceToken := k.BalanceOf(ctx, erc20, contract, types.ModuleAddress)
	if balanceToken == nil {
		return nil, errorsmod.Wrap(types.ErrEVMCall, "failed to retrieve balance")
	}

	// Escrow tokens on module account
	transferData, err := erc20.Pack("transfer", types.ModuleAddress, msg.Amount.BigInt())
	if err != nil {
		return nil, err
	}

	res, err := k.CallEVMWithData(ctx, sender, &contract, transferData, true)
	if err != nil {
		return nil, err
	}

	// Check evm call response
	var unpackedRet types.ERC20BoolResponse
	if err := erc20.UnpackIntoInterface(&unpackedRet, "transfer", res.Ret); err != nil {
		return nil, err
	}

	if !unpackedRet.Value {
		return nil, errorsmod.Wrap(sdkerrors.ErrLogic, "failed to execute transfer")
	}

	// Check expected escrow balance after transfer execution
	tokens := coins[0].Amount.BigInt()
	balanceTokenAfter := k.BalanceOf(ctx, erc20, contract, types.ModuleAddress)
	if balanceTokenAfter == nil {
		return nil, errorsmod.Wrap(types.ErrEVMCall, "failed to retrieve balance")
	}

	expToken := big.NewInt(0).Add(balanceToken, tokens)

	if r := balanceTokenAfter.Cmp(expToken); r != 0 {
		return nil, errorsmod.Wrapf(
			types.ErrBalanceInvariance,
			"invalid token balance - expected: %v, actual: %v",
			expToken, balanceTokenAfter,
		)
	}

	// Mint coins
	if err := k.bankKeeper.MintCoins(ctx, types.ModuleName, coins); err != nil {
		return nil, err
	}

	// Send minted coins to the receiver
	if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, receiver, coins); err != nil {
		return nil, err
	}

	// Check expected receiver balance after transfer
	balanceCoinAfter := k.bankKeeper.GetBalance(ctx, receiver, pair.Denom)
	expCoin := balanceCoin.Add(coins[0])

	if ok := balanceCoinAfter.IsEqual(expCoin); !ok {
		return nil, errorsmod.Wrapf(
			types.ErrBalanceInvariance,
			"invalid coin balance - expected: %v, actual: %v",
			expCoin, balanceCoinAfter,
		)
	}

	// Check for unexpected `Approval` event in logs
	if err := k.monitorApprovalEvent(res); err != nil {
		return nil, err
	}

	defer func() {
		telemetry.IncrCounterWithLabels(
			[]string{"tx", "msg", "convert", "erc20", "total"},
			1,
			[]metrics.Label{
				telemetry.NewLabel("coin", pair.Denom),
				telemetry.NewLabel("erc20", pair.Erc20Address),
			},
		)

		if msg.Amount.IsInt64() {
			telemetry.IncrCounterWithLabels(
				[]string{"tx", "msg", "convert", "erc20", "amount", "total"},
				float32(msg.Amount.Int64()),
				[]metrics.Label{
					telemetry.NewLabel("denom", pair.Denom),
					telemetry.NewLabel("erc20", pair.Erc20Address),
				},
			)
		}
	}()

	ctx.EventManager().EmitEvents(
		sdk.Events{
			sdk.NewEvent(
				types.EventTypeConvertERC20,
				sdk.NewAttribute(sdk.AttributeKeySender, msg.Sender),
				sdk.NewAttribute(types.AttributeKeyReceiver, msg.Receiver),
				sdk.NewAttribute(sdk.AttributeKeyAmount, msg.Amount.String()),
				sdk.NewAttribute(types.AttributeKeyCosmosCoin, pair.Denom),
				sdk.NewAttribute(types.AttributeKeyERC20Token, msg.ContractAddress),
			),
		},
	)

	return &types.MsgConvertERC20Response{}, nil
}

// convertCoinNativeERC20 handles the coin conversion for a native ERC20 token
// pair:
//   - escrow Coins on module account
//   - unescrow Tokens that have been previously escrowed with ConvertERC20 and send to receiver
//   - burn escrowed Coins
//   - check if token balance increased by amount
//   - check for unexpected `Approval` event in logs
func (k Keeper) convertCoinNativeERC20(
	ctx sdk.Context,
	pair types.TokenPair,
	msg *types.MsgConvertCoin,
	receiver common.Address,
	sender sdk.AccAddress,
) (*types.MsgConvertCoinResponse, error) {
	// NOTE: ignore validation from NewCoin constructor
	coins := sdk.Coins{msg.Coin}

	erc20 := contracts.ERC20MinterBurnerDecimalsContract.ABI
	contract := pair.GetERC20Contract()
	balanceToken := k.BalanceOf(ctx, erc20, contract, receiver)
	if balanceToken == nil {
		return nil, errorsmod.Wrap(types.ErrEVMCall, "failed to retrieve balance")
	}

	// Escrow Coins on module account
	if err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, sender, types.ModuleName, coins); err != nil {
		return nil, errorsmod.Wrap(err, "failed to escrow coins")
	}

	// Unescrow Tokens and send to receiver
	res, err := k.CallEVM(ctx, erc20, types.ModuleAddress, contract, true, "transfer", receiver, msg.Coin.Amount.BigInt())
	if err != nil {
		return nil, err
	}

	// Check unpackedRet execution
	var unpackedRet types.ERC20BoolResponse
	if err := erc20.UnpackIntoInterface(&unpackedRet, "transfer", res.Ret); err != nil {
		return nil, err
	}

	if !unpackedRet.Value {
		return nil, errorsmod.Wrap(sdkerrors.ErrLogic, "failed to execute unescrow tokens from user")
	}

	// Check expected Receiver balance after transfer execution
	tokens := msg.Coin.Amount.BigInt()
	balanceTokenAfter := k.BalanceOf(ctx, erc20, contract, receiver)
	if balanceTokenAfter == nil {
		return nil, errorsmod.Wrap(types.ErrEVMCall, "failed to retrieve balance")
	}

	exp := big.NewInt(0).Add(balanceToken, tokens)

	if r := balanceTokenAfter.Cmp(exp); r != 0 {
		return nil, errorsmod.Wrapf(
			types.ErrBalanceInvariance,
			"invalid token balance - expected: %v, actual: %v", exp, balanceTokenAfter,
		)
	}

	// Burn escrowed Coins
	err = k.bankKeeper.BurnCoins(ctx, types.ModuleName, coins)
	if err != nil {
		return nil, errorsmod.Wrap(err, "failed to burn coins")
	}

	// Check for unexpected `Approval` event in logs
	if err := k.monitorApprovalEvent(res); err != nil {
		return nil, err
	}

	defer func() {
		telemetry.IncrCounterWithLabels(
			[]string{"tx", "msg", "convert", "coin", "total"},
			1,
			[]metrics.Label{
				telemetry.NewLabel("denom", pair.Denom),
				telemetry.NewLabel("erc20", pair.Erc20Address),
			},
		)

		if msg.Coin.Amount.IsInt64() {
			telemetry.IncrCounterWithLabels(
				[]string{"tx", "msg", "convert", "coin", "amount", "total"},
				float32(msg.Coin.Amount.Int64()),
				[]metrics.Label{
					telemetry.NewLabel("denom", pair.Denom),
					telemetry.NewLabel("erc20", pair.Erc20Address),
				},
			)
		}
	}()

	ctx.EventManager().EmitEvents(
		sdk.Events{
			sdk.NewEvent(
				types.EventTypeConvertCoin,
				sdk.NewAttribute(sdk.AttributeKeySender, msg.Sender),
				sdk.NewAttribute(types.AttributeKeyReceiver, msg.Receiver),
				sdk.NewAttribute(sdk.AttributeKeyAmount, msg.Coin.Amount.String()),
				sdk.NewAttribute(types.AttributeKeyCosmosCoin, msg.Coin.Denom),
				sdk.NewAttribute(types.AttributeKeyERC20Token, pair.Erc20Address),
			),
		},
	)

	return &types.MsgConvertCoinResponse{}, nil
}

func (k Keeper) UpdateParams(goCtx context.Context, req *types.MsgUpdateParams) (*types.MsgUpdateParamsResponse, error) {
	if k.GetAuthority() != req.Authority {
		return nil, errorsmod.Wrapf(govtypes.ErrInvalidSigner, "invalid authority; expected %s, got %s", k.GetAuthority(), req.Authority)
	}

	if err := req.Params.Validate(); err != nil {
		return nil, err
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	k.SetParams(ctx, req.Params)
	//if err := k.SetParams(ctx, req.Params); err != nil {
	//	return nil, err
	//}

	return &types.MsgUpdateParamsResponse{}, nil
}

func (k Keeper) RegisterCoinProposal(goCtx context.Context, req *types.MsgRegisterCoinProposal) (*types.MsgRegisterCoinProposalResponse, error) {
	if k.GetAuthority() != req.Authority {
		return nil, errorsmod.Wrapf(govtypes.ErrInvalidSigner, "invalid authority; expected %s, got %s", k.GetAuthority(), req.Authority)
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	coinMetadata := req.Metadata

	_, err := k.RegisterCoin(ctx, coinMetadata)
	if err != nil {
		return nil, err
	}

	return &types.MsgRegisterCoinProposalResponse{}, nil
}

func (k Keeper) RegisterERC20Proposal(goCtx context.Context, req *types.MsgRegisterERC20Proposal) (*types.MsgRegisterERC20ProposalResponse, error) {
	if k.GetAuthority() != req.Authority {
		return nil, errorsmod.Wrapf(govtypes.ErrInvalidSigner, "invalid authority; expected %s, got %s", k.GetAuthority(), req.Authority)
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	contract := common.HexToAddress(req.Erc20Address)

	_, err := k.RegisterERC20(ctx, contract)
	if err != nil {
		return nil, err
	}

	return &types.MsgRegisterERC20ProposalResponse{}, nil
}

func (k Keeper) ToggleTokenConversionProposal(goCtx context.Context, req *types.MsgToggleTokenConversionProposal) (*types.MsgToggleTokenConversionProposalResponse, error) {
	if k.GetAuthority() != req.Authority {
		return nil, errorsmod.Wrapf(govtypes.ErrInvalidSigner, "invalid authority; expected %s, got %s", k.GetAuthority(), req.Authority)
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	token := req.Token

	_, err := k.ToggleConversion(ctx, token)
	if err != nil {
		return nil, err
	}

	return &types.MsgToggleTokenConversionProposalResponse{}, nil
}

// RegisterCoin deploys an erc20 contract and creates the token pair for the
// existing cosmos coin
func (k Keeper) RegisterCoin(
	ctx sdk.Context,
	coinMetadata banktypes.Metadata,
) (*types.TokenPair, error) {
	// Check if the conversion is globally enabled
	params := k.GetParams(ctx)
	if !params.EnableErc20 {
		return nil, errorsmod.Wrap(
			types.ErrERC20Disabled, "registration is currently disabled by governance",
		)
	}

	// Prohibit denominations that contain the evm denom
	if strings.Contains(coinMetadata.Base, "CANTO") {
		return nil, errorsmod.Wrapf(
			types.ErrEVMDenom, "cannot register the EVM denomination %s", coinMetadata.Base,
		)
	}

	// Check if denomination is already registered
	if k.IsDenomRegistered(ctx, coinMetadata.Name) {
		return nil, errorsmod.Wrapf(
			types.ErrTokenPairAlreadyExists, "coin denomination already registered: %s", coinMetadata.Name,
		)
	}

	// Check if the coin exists by ensuring the supply is set
	if !k.bankKeeper.HasSupply(ctx, coinMetadata.Base) {
		return nil, errorsmod.Wrapf(
			sdkerrors.ErrInvalidCoins, "base denomination '%s' cannot have a supply of 0", coinMetadata.Base,
		)
	}

	if err := k.verifyMetadata(ctx, coinMetadata); err != nil {
		return nil, errorsmod.Wrapf(
			types.ErrInternalTokenPair, "coin metadata is invalid %s", coinMetadata.Name,
		)
	}

	addr, err := k.DeployERC20Contract(ctx, coinMetadata)
	if err != nil {
		return nil, errorsmod.Wrap(
			err, "failed to create wrapped coin denom metadata for ERC20",
		)
	}

	pair := types.NewTokenPair(addr, coinMetadata.Base, true, types.OWNER_MODULE)
	k.SetTokenPair(ctx, pair)
	k.SetDenomMap(ctx, pair.Denom, pair.GetID())
	k.SetERC20Map(ctx, common.HexToAddress(pair.Erc20Address), pair.GetID())

	return &pair, nil
}

// RegisterERC20 creates a Cosmos coin and registers the token pair between the
// coin and the ERC20
func (k Keeper) RegisterERC20(
	ctx sdk.Context,
	contract common.Address,
) (*types.TokenPair, error) {
	// Check if the conversion is globally enabled
	params := k.GetParams(ctx)
	if !params.EnableErc20 {
		return nil, errorsmod.Wrap(
			types.ErrERC20Disabled, "registration is currently disabled by governance",
		)
	}

	// Check if ERC20 is already registered
	if k.IsERC20Registered(ctx, contract) {
		return nil, errorsmod.Wrapf(
			types.ErrTokenPairAlreadyExists, "token ERC20 contract already registered: %s", contract.String(),
		)
	}

	metadata, err := k.CreateCoinMetadata(ctx, contract)
	if err != nil {
		return nil, errorsmod.Wrap(
			err, "failed to create wrapped coin denom metadata for ERC20",
		)
	}

	pair := types.NewTokenPair(contract, metadata.Name, true, types.OWNER_EXTERNAL)
	k.SetTokenPair(ctx, pair)
	k.SetDenomMap(ctx, pair.Denom, pair.GetID())
	k.SetERC20Map(ctx, common.HexToAddress(pair.Erc20Address), pair.GetID())
	return &pair, nil
}

// ToggleConversion toggles conversion for a given token pair
func (k Keeper) ToggleConversion(
	ctx sdk.Context,
	token string,
) (types.TokenPair, error) {
	id := k.GetTokenPairID(ctx, token)
	if len(id) == 0 {
		return types.TokenPair{}, errorsmod.Wrapf(
			types.ErrTokenPairNotFound, "token '%s' not registered by id", token,
		)
	}

	pair, found := k.GetTokenPair(ctx, id)
	if !found {
		return types.TokenPair{}, errorsmod.Wrapf(
			types.ErrTokenPairNotFound, "token '%s' not registered", token,
		)
	}

	pair.Enabled = !pair.Enabled

	k.SetTokenPair(ctx, pair)
	return pair, nil
}

// CreateCoinMetadata generates the metadata to represent the ERC20 token on
// canto.
func (k Keeper) CreateCoinMetadata(
	ctx sdk.Context,
	contract common.Address,
) (*banktypes.Metadata, error) {
	strContract := contract.String()

	erc20Data, err := k.QueryERC20(ctx, contract)
	if err != nil {
		return nil, err
	}

	// Check if metadata already exists
	_, found := k.bankKeeper.GetDenomMetaData(ctx, CreateDenom(strContract))
	if found {
		return nil, errorsmod.Wrap(
			types.ErrInternalTokenPair, "denom metadata already registered",
		)
	}

	if k.IsDenomRegistered(ctx, CreateDenom(strContract)) {
		return nil, errorsmod.Wrapf(
			types.ErrInternalTokenPair, "coin denomination already registered: %s", erc20Data.Name,
		)
	}

	// base denomination
	base := CreateDenom(strContract)

	// create a bank denom metadata based on the ERC20 token ABI details
	// metadata name is should always be the contract since it's the key
	// to the bank store
	metadata := banktypes.Metadata{
		Description: CreateDenomDescription(strContract),
		Base:        base,
		// NOTE: Denom units MUST be increasing
		DenomUnits: []*banktypes.DenomUnit{
			{
				Denom:    base,
				Exponent: 0,
			},
		},
		Name:    CreateDenom(strContract),
		Symbol:  erc20Data.Symbol,
		Display: base,
	}

	// only append metadata if decimals > 0, otherwise validation fails
	if erc20Data.Decimals > 0 {
		nameSanitized := types.SanitizeERC20Name(erc20Data.Name)
		metadata.DenomUnits = append(
			metadata.DenomUnits,
			&banktypes.DenomUnit{
				Denom:    nameSanitized,
				Exponent: uint32(erc20Data.Decimals),
			},
		)
		metadata.Display = nameSanitized
	}

	if err := metadata.Validate(); err != nil {
		return nil, errorsmod.Wrapf(
			err, "ERC20 token data is invalid for contract %s", strContract,
		)
	}

	k.bankKeeper.SetDenomMetaData(ctx, metadata)

	return &metadata, nil
}

// verifyMetadata verifies if the metadata matches the existing one, if not it
// sets it to the store
func (k Keeper) verifyMetadata(
	ctx sdk.Context,
	coinMetadata banktypes.Metadata,
) error {
	meta, found := k.bankKeeper.GetDenomMetaData(ctx, coinMetadata.Base)
	if !found {
		k.bankKeeper.SetDenomMetaData(ctx, coinMetadata)
		return nil
	}

	// If it already existed, check that is equal to what is stored
	return types.EqualMetadata(meta, coinMetadata)
}

// CreateDenomDescription generates a string with the coin description
func CreateDenomDescription(address string) string {
	return fmt.Sprintf("Cosmos coin token representation of %s", address)
}

// CreateDenom generates a string the module name plus the address to avoid conflicts with names staring with a number
func CreateDenom(address string) string {
	return fmt.Sprintf("%s/%s", types.ModuleName, address)
}
