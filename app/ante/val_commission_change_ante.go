package ante

import (
	"fmt"
	"github.com/Canto-Network/Canto/v7/types"
	liquidstakingkeeper "github.com/Canto-Network/Canto/v7/x/liquidstaking/keeper"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/authz"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"time"

	authzante "github.com/Canto-Network/Canto/v7/app/ante/cosmos"
)

// ValCommissionChangeLimitDecorator checks that if MsgEditValidator tries to change the commission rate.
// If so, it allows msg only within 23 hours and 50 minutes of the next epoch.
// Because liquidstaking module rank insurances based on insurance fee rate + validator's commission rate at every Epoch and
// validators can change their commission rate by MsgEditValidator at any time, we need to limit the commission rate change.
// If not, validator can change their commission rate as high to get more delegation rewards while epoch duration and as low right before the epoch, so it can still be ranked in.
type ValCommissionChangeLimitDecorator struct {
	liquidstakingKeeper *liquidstakingkeeper.Keeper
	stakingKeeper       *stakingkeeper.Keeper
	cdc                 codec.BinaryCodec
}

// NewValCommissionChangeLimitDecorator creates a new ValCommissionChangeLimitDecorator
func NewValCommissionChangeLimitDecorator(
	liquidstakingKeeper *liquidstakingkeeper.Keeper,
	stakingKeeper *stakingkeeper.Keeper,
	cdc codec.BinaryCodec,
) ValCommissionChangeLimitDecorator {
	return ValCommissionChangeLimitDecorator{
		liquidstakingKeeper: liquidstakingKeeper,
		stakingKeeper:       stakingKeeper,
		cdc:                 cdc,
	}
}

func (vccld ValCommissionChangeLimitDecorator) AnteHandle(
	ctx sdk.Context,
	tx sdk.Tx,
	simulate bool,
	next sdk.AnteHandler,
) (newCtx sdk.Context, err error) {
	msgs := tx.GetMsgs()
	if err = vccld.ValidateMsgs(ctx, msgs); err != nil {
		return ctx, err
	}

	return next(ctx, tx, simulate)
}

func (vccld ValCommissionChangeLimitDecorator) ValidateMsgs(ctx sdk.Context, msgs []sdk.Msg) error {
	var validMsg func(m sdk.Msg, nestedCnt int) error
	validMsg = func(m sdk.Msg, nestedCnt int) error {
		if nestedCnt >= authzante.MaxNestedMsgs {
			return fmt.Errorf("found more nested msgs than permited. Limit is : %d", authzante.MaxNestedMsgs)
		}
		switch msg := m.(type) {
		case *authz.MsgExec:
			for _, v := range msg.Msgs {
				var innerMsg sdk.Msg
				if err := vccld.cdc.UnpackAny(v, &innerMsg); err != nil {
					return sdkerrors.Wrapf(sdkerrors.ErrUnauthorized, "cannot unmarshal authz exec msgs")
				}
				nestedCnt++
				if err := validMsg(innerMsg, nestedCnt); err != nil {
					return err
				}
			}

		case *stakingtypes.MsgEditValidator:
			valAddr, err := sdk.ValAddressFromBech32(msg.ValidatorAddress)
			if err != nil {
				return err
			}
			val, found := vccld.stakingKeeper.GetValidator(ctx, valAddr)
			if !found {
				return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "validator does not exist")
			}
			if val.Commission.Rate.Equal(*msg.CommissionRate) {
				// This is not a commission rate change.
				return nil
			}
			// Check if the commission rate change is within 23 hours and 50 minutes of the epoch.
			epoch := vccld.liquidstakingKeeper.GetEpoch(ctx)
			nextEpochTime := epoch.StartTime.Add(epoch.Duration)
			timeLimit := nextEpochTime.Add(-23 * time.Hour).Add(-50 * time.Minute)
			if ctx.BlockTime().After(timeLimit) {
				return nil
			}
			timeLeft := timeLimit.Sub(ctx.BlockTime())
			return sdkerrors.Wrap(
				types.ErrChangingValCommissionForbidden,
				fmt.Sprintf("%s left", timeLeft.String()),
			)
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
