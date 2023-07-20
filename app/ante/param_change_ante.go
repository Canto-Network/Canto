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
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

// ParamChangeLimitDecorator checks that the params change proposals for slashing and staking modules.
// The liquidstaking module works closely with the slashing and staking module's params(e.g. MinimumCollateral constant is calculated based on the slashing params).
// To reduce unexpected risks, it is important to limit the params change proposals for slashing and staking modules.
type ParamChangeLimitDecorator struct {
	slashingKeeper *slashingkeeper.Keeper
	cdc            codec.BinaryCodec
}

// NewParamChangeLimitDecorator creates a new slashing param change limit decorator.
func NewParamChangeLimitDecorator(
	slashingKeeper *slashingkeeper.Keeper,
	cdc codec.BinaryCodec,
) ParamChangeLimitDecorator {
	return ParamChangeLimitDecorator{
		slashingKeeper: slashingKeeper,
		cdc:            cdc,
	}
}

func (s ParamChangeLimitDecorator) AnteHandle(
	ctx sdk.Context,
	tx sdk.Tx,
	simulate bool,
	next sdk.AnteHandler,
) (newCtx sdk.Context, err error) {
	msgs := tx.GetMsgs()
	if err = s.ValidateMsgs(ctx, msgs); err != nil {
		return ctx, err
	}

	return next(ctx, tx, simulate)
}

func (s ParamChangeLimitDecorator) ValidateMsgs(ctx sdk.Context, msgs []sdk.Msg) error {
	var slashingParams slashingtypes.Params
	validMsg := func(m sdk.Msg) error {
		if msg, ok := m.(*govtypes.MsgSubmitProposal); ok {
			switch c := msg.GetContent().(type) {
			case *proposal.ParameterChangeProposal:
				for _, c := range c.Changes {
					switch c.GetSubspace() {
					case slashingtypes.ModuleName:
						slashingParams = s.slashingKeeper.GetParams(ctx)
						switch c.GetKey() {
						// SignedBlocksWindow, MinSignedPerWindow, DowntimeJailDuration are not allowed to be decreased.
						// If we decrease these slashingParams, the slashing penalty can be increased.
						case string(slashingtypes.KeySignedBlocksWindow):
							window, err := strconv.ParseInt(c.GetValue(), 10, 64)
							if err != nil {
								return err
							}
							if window < slashingParams.SignedBlocksWindow {
								return types.ErrInvalidSignedBlocksWindow
							}
						case string(slashingtypes.KeyMinSignedPerWindow):
							minSignedPerWindow, err := sdk.NewDecFromStr(c.GetValue())
							if err != nil {
								return err
							}
							if minSignedPerWindow.LT(slashingParams.MinSignedPerWindow) {
								return types.ErrInvalidMinSignedPerWindow
							}
						case string(slashingtypes.KeyDowntimeJailDuration):
							downtimeJailDuration, err := strconv.ParseInt(c.GetValue(), 10, 64)
							if err != nil {
								return err
							}
							if time.Duration(downtimeJailDuration) < slashingParams.DowntimeJailDuration {
								return types.ErrInvalidDowntimeJailDuration
							}
						// SlashFractionDoubleSign, SlashFractionDowntime are not allowed to be increased.
						// If we increase these slashingParams, the slashing penalty can be increased.
						case string(slashingtypes.KeySlashFractionDoubleSign):
							slashFractionDoubleSign, err := sdk.NewDecFromStr(c.GetValue())
							if err != nil {
								return err
							}
							if slashFractionDoubleSign.GT(slashingParams.SlashFractionDoubleSign) {
								return types.ErrInvalidSlashFractionDoubleSign
							}
						case string(slashingtypes.KeySlashFractionDowntime):
							slashFractionDowntime, err := sdk.NewDecFromStr(c.GetValue())
							if err != nil {
								return err
							}
							if slashFractionDowntime.GT(slashingParams.SlashFractionDowntime) {
								return types.ErrInvalidSlashFractionDowntime
							}
						}
					case stakingtypes.ModuleName:
						switch c.GetKey() {
						case string(stakingtypes.KeyUnbondingTime):
							return types.ErrChangingUnbondingPeriodForbidden
						case string(stakingtypes.KeyBondDenom):
							return types.ErrChangingBondDenomForbidden
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
