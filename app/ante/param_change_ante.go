package ante

import (
	"fmt"
	"strconv"
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/authz"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/cosmos/cosmos-sdk/x/params/types/proposal"
	slashingkeeper "github.com/cosmos/cosmos-sdk/x/slashing/keeper"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	authzante "github.com/Canto-Network/Canto/v7/app/ante/cosmos"
	"github.com/Canto-Network/Canto/v7/types"
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

func (pcld ParamChangeLimitDecorator) AnteHandle(
	ctx sdk.Context,
	tx sdk.Tx,
	simulate bool,
	next sdk.AnteHandler,
) (newCtx sdk.Context, err error) {
	msgs := tx.GetMsgs()
	if err = pcld.ValidateMsgs(ctx, msgs); err != nil {
		return ctx, err
	}

	return next(ctx, tx, simulate)
}

func (pcld ParamChangeLimitDecorator) ValidateMsgs(ctx sdk.Context, msgs []sdk.Msg) error {
	var slashingParams slashingtypes.Params
	var validMsg func(m sdk.Msg, nestedCnt int) error
	validMsg = func(m sdk.Msg, nestedCnt int) error {
		if nestedCnt >= authzante.MaxNestedMsgs {
			return fmt.Errorf("found more nested msgs than permited. Limit is : %d", authzante.MaxNestedMsgs)
		}
		switch msg := m.(type) {
		case *authz.MsgExec:
			for _, v := range msg.Msgs {
				var innerMsg sdk.Msg
				if err := pcld.cdc.UnpackAny(v, &innerMsg); err != nil {
					return sdkerrors.Wrapf(sdkerrors.ErrUnauthorized, "cannot unmarshal authz exec msgs")
				}
				nestedCnt++
				if err := validMsg(innerMsg, nestedCnt); err != nil {
					return err
				}
			}

		case *govtypes.MsgSubmitProposal:
			switch c := msg.GetContent().(type) {
			case *proposal.ParameterChangeProposal:
				for _, c := range c.Changes {
					switch c.GetSubspace() {
					case slashingtypes.ModuleName:
						slashingParams = pcld.slashingKeeper.GetParams(ctx)
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
		default:
			return nil
		}
		return nil
	}

	for _, m := range msgs {
		if err := validMsg(m, 1); err != nil {
			return err
		}
	}
	return nil
}
