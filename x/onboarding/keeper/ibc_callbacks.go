package keeper

import (
	"strconv"

	errorsmod "cosmossdk.io/errors"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	transfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v8/modules/core/04-channel/types"
	"github.com/cosmos/ibc-go/v8/modules/core/exported"

	"github.com/ethereum/go-ethereum/common"

	"github.com/Canto-Network/Canto/v7/ibc"
	coinswaptypes "github.com/Canto-Network/Canto/v7/x/coinswap/types"
	erc20types "github.com/Canto-Network/Canto/v7/x/erc20/types"
	"github.com/Canto-Network/Canto/v7/x/onboarding/types"
)

// OnRecvPacket performs an IBC receive callback.
// It swaps the transferred IBC denom to acanto and
// convert the remaining balance to ERC20 tokens.
// If the balance of acanto is greater than the predefined value,
// the swap is omitted and the entire transferred amount is converted to ERC20.
func (k Keeper) OnRecvPacket(
	ctx sdk.Context,
	packet channeltypes.Packet,
	ack exported.Acknowledgement,
) exported.Acknowledgement {
	logger := k.Logger(ctx)

	// It always returns original ACK
	// meaning that even if the swap or conversion fails, it does not revert IBC transfer
	// and the asset transferred to the Canto network will still remain in the Canto network

	params := k.GetParams(ctx)
	if !params.EnableOnboarding {
		return ack
	}

	// check source channel is in the whitelist channels
	var found bool
	for _, s := range params.WhitelistedChannels {
		if s == packet.DestinationChannel {
			found = true
		}
	}

	if !found {
		return ack
	}

	// Get recipient addresses in `canto1` and the original bech32 format
	_, recipient, senderBech32, recipientBech32, err := ibc.GetTransferSenderRecipient(packet)
	if err != nil {
		return channeltypes.NewErrorAcknowledgement(err)
	}

	//get the recipient account
	account := k.accountKeeper.GetAccount(ctx, recipient)

	// onboarding is not supported for module accounts
	if _, isModuleAccount := account.(sdk.ModuleAccountI); isModuleAccount {
		return ack
	}

	standardDenom, err := k.coinswapKeeper.GetStandardDenom(ctx)
	if err != nil {
		return ack
	}

	var data transfertypes.FungibleTokenPacketData
	if err = transfertypes.ModuleCdc.UnmarshalJSON(packet.GetData(), &data); err != nil {
		// NOTE: shouldn't happen as the packet has already
		// been decoded on ICS20 transfer logic
		err = errorsmod.Wrapf(types.ErrInvalidType, "cannot unmarshal ICS-20 transfer packet data")
		return channeltypes.NewErrorAcknowledgement(err)
	}

	// parse the transferred denom
	transferredCoin := ibc.GetReceivedCoin(
		packet.SourcePort, packet.SourceChannel,
		packet.DestinationPort, packet.DestinationChannel,
		data.Denom, data.Amount,
	)

	autoSwapThreshold := k.GetParams(ctx).AutoSwapThreshold
	swapCoins := sdk.NewCoin(standardDenom, autoSwapThreshold)
	standardCoinBalance := k.bankKeeper.SpendableCoins(ctx, recipient).AmountOf(standardDenom)
	swappedAmount := sdkmath.ZeroInt()

	if standardCoinBalance.LT(autoSwapThreshold) {
		swappedAmount, err = k.coinswapKeeper.TradeInputForExactOutput(ctx, coinswaptypes.Input{Coin: transferredCoin, Address: recipient.String()}, coinswaptypes.Output{Coin: swapCoins, Address: recipient.String()})
		if err != nil {
			swappedAmount = sdkmath.ZeroInt()
			logger.Error("failed to swap coins", "error", err)
		} else {
			ctx.EventManager().EmitEvent(
				sdk.NewEvent(
					coinswaptypes.EventTypeSwap,
					sdk.NewAttribute(coinswaptypes.AttributeValueAmount, swappedAmount.String()),
					sdk.NewAttribute(coinswaptypes.AttributeValueSender, recipient.String()),
					sdk.NewAttribute(coinswaptypes.AttributeValueRecipient, recipient.String()),
					sdk.NewAttribute(coinswaptypes.AttributeValueIsBuyOrder, strconv.FormatBool(true)),
					sdk.NewAttribute(coinswaptypes.AttributeValueTokenPair, coinswaptypes.GetTokenPairByDenom(transferredCoin.Denom, swapCoins.Denom)),
				),
			)
		}
	}

	//convert coins to ERC20 token
	pairID := k.erc20Keeper.GetTokenPairID(ctx, transferredCoin.Denom)
	if len(pairID) == 0 {
		// short-circuit: if the denom is not registered, conversion will fail
		// so we can continue with the rest of the stack
		return ack
	}

	pair, _ := k.erc20Keeper.GetTokenPair(ctx, pairID)
	if !pair.Enabled {
		// no-op: continue with the rest of the stack without conversion
		return ack
	}

	convertCoin := sdk.NewCoin(transferredCoin.Denom, transferredCoin.Amount.Sub(swappedAmount))

	// Build MsgConvertCoin, from recipient to recipient since IBC transfer already occurred
	convertMsg := erc20types.NewMsgConvertCoin(convertCoin, common.BytesToAddress(recipient.Bytes()), recipient)

	// NOTE: we don't use ValidateBasic the msg since we've already validated
	// the ICS20 packet data

	// Use MsgConvertCoin to convert the Cosmos Coin to an ERC20
	if _, err = k.erc20Keeper.ConvertCoin(sdk.WrapSDKContext(ctx), convertMsg); err != nil {
		logger.Error("failed to convert coins", "error", err)
		return ack
	}

	logger.Info(
		"coinswap and erc20 conversion completed",
		"sender", senderBech32,
		"receiver", recipientBech32,
		"source-port", packet.SourcePort,
		"source-channel", packet.SourceChannel,
		"dest-port", packet.DestinationPort,
		"dest-channel", packet.DestinationChannel,
		"swap amount", swappedAmount,
		"convert amount", convertCoin.Amount,
	)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeOnboarding,
			sdk.NewAttribute(sdk.AttributeKeySender, senderBech32),
			sdk.NewAttribute(transfertypes.AttributeKeyReceiver, recipientBech32),
			sdk.NewAttribute(channeltypes.AttributeKeySrcChannel, packet.SourceChannel),
			sdk.NewAttribute(channeltypes.AttributeKeySrcPort, packet.SourcePort),
			sdk.NewAttribute(channeltypes.AttributeKeyDstPort, packet.DestinationPort),
			sdk.NewAttribute(channeltypes.AttributeKeyDstChannel, packet.DestinationChannel),
			sdk.NewAttribute(types.AttributeKeySwapAmount, swappedAmount.String()),
			sdk.NewAttribute(types.AttributeKeyConvertAmount, convertCoin.Amount.String()),
		),
	)

	// return original acknowledgement
	return ack
}
