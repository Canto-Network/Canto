package ante

import (
	"strconv"
	"time"

	"github.com/Canto-Network/Canto/v6/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/authz"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/cosmos/cosmos-sdk/x/params/types/proposal"
	slashingkeeper "github.com/cosmos/cosmos-sdk/x/slashing/keeper"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
)

type SlashingParamChangeLimitDecorator struct {
	cdc            codec.BinaryCodec
	slashingKeeper *slashingkeeper.Keeper
}

func NewSlashingParamChangeLimitDecorator(
	cdc codec.BinaryCodec,
	slashingKeeper *slashingkeeper.Keeper,
) SlashingParamChangeLimitDecorator {
	return SlashingParamChangeLimitDecorator{
		cdc:            cdc,
		slashingKeeper: slashingKeeper,
	}
}

func (s SlashingParamChangeLimitDecorator) AnteHandle(
	ctx sdk.Context,
	tx sdk.Tx,
	simulate bool,
	next sdk.AnteHandler,
) (newCtx sdk.Context, err error) {
	if !ctx.IsCheckTx() || simulate {
		return next(ctx, tx, simulate)
	}

	msgs := tx.GetMsgs()
	if err = s.ValidateMsgs(ctx, msgs); err != nil {
		return ctx, err
	}

	return next(ctx, tx, simulate)
}

func (s SlashingParamChangeLimitDecorator) ValidateMsgs(ctx sdk.Context, msgs []sdk.Msg) error {
	validMsg := func(m sdk.Msg) error {
		params := s.slashingKeeper.GetParams(ctx)
		if msg, ok := m.(*govtypes.MsgSubmitProposal); ok {
			switch c := msg.GetContent().(type) {
			case *proposal.ParameterChangeProposal:
				for _, c := range c.Changes {
					if c.GetSubspace() != slashingtypes.ModuleName {
						return nil
					}
					switch c.GetKey() {
					case string(slashingtypes.KeySignedBlocksWindow):
						window, err := strconv.ParseInt(c.GetValue(), 10, 64)
						if err != nil {
							return err
						}
						if window < params.SignedBlocksWindow {
							return sdkerrors.Wrapf(types.ErrInvalidSignedBlocksWindow, "given: %d, current: %d", window, params.SignedBlocksWindow)
						}
					case string(slashingtypes.KeyMinSignedPerWindow):
						minSignedPerWindow, err := sdk.NewDecFromStr(c.GetValue())
						if err != nil {
							return err
						}
						if minSignedPerWindow.LT(params.MinSignedPerWindow) {
							return sdkerrors.Wrapf(types.ErrInvalidMinSignedPerWindow, "given: %s, current: %s", minSignedPerWindow, params.MinSignedPerWindow)
						}
					case string(slashingtypes.KeyDowntimeJailDuration):
						downtimeJailDuration, err := strconv.ParseInt(c.GetValue(), 10, 64)
						if err != nil {
							return err
						}
						if time.Duration(downtimeJailDuration) < params.DowntimeJailDuration {
							return sdkerrors.Wrapf(types.ErrInvalidDowntimeJailDuration, "given: %d, current: %d", downtimeJailDuration, params.DowntimeJailDuration)
						}
					case string(slashingtypes.KeySlashFractionDoubleSign):
						slashFractionDoubleSign, err := sdk.NewDecFromStr(c.GetValue())
						if err != nil {
							return err
						}
						if slashFractionDoubleSign.GT(params.SlashFractionDoubleSign) {
							return sdkerrors.Wrapf(types.ErrInvalidSlashFractionDoubleSign, "given: %s, current: %s", slashFractionDoubleSign, params.SlashFractionDoubleSign)
						}
					case string(slashingtypes.KeySlashFractionDowntime):
						slashFractionDowntime, err := sdk.NewDecFromStr(c.GetValue())
						if err != nil {
							return err
						}
						if slashFractionDowntime.GT(params.SlashFractionDowntime) {
							return sdkerrors.Wrapf(types.ErrInvalidSlashFractionDowntime, "given: %s, current: %s", slashFractionDowntime, params.SlashFractionDowntime)
						}
					}
				}
			default:
				return nil
			}
		}
		return nil
	}
	validAuthz := func(execMsg *authz.MsgExec) error {
		for _, v := range execMsg.Msgs {
			var innerMsg sdk.Msg
			if err := s.cdc.UnpackAny(v, &innerMsg); err != nil {
				return sdkerrors.Wrapf(sdkerrors.ErrUnauthorized, "cannot unmarshal authz exec msgs")
			}
			if err := validMsg(innerMsg); err != nil {
				return err
			}
		}

		return nil
	}
	for _, m := range msgs {
		if msg, ok := m.(*authz.MsgExec); ok {
			if err := validAuthz(msg); err != nil {
				return err
			}
			continue
		}

		// validate normal msgs
		if err := validMsg(m); err != nil {
			return err
		}
	}
	return nil
}
