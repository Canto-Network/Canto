package types

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// RootCodespace is the codespace for all errors defined in this package
const RootCodespace = "canto"

// root error codes for canto
const (
	codeKeyTypeNotSupported = iota + 2
	codeInvalidSignedBlocksWindow
	codeInvalidMinSignedPerWindow
	codeInvalidDowntimeJailDuration
	codeInvalidSlashFractionDoubleSign
	codeInvalidSlashFractionDowntime
	codeChangingUnbondingPeriodForbidden
	codeChangingBondDenomForbidden
)

// errors
var (
	ErrKeyTypeNotSupported              = sdkerrors.Register(RootCodespace, codeKeyTypeNotSupported, "key type 'secp256k1' not supported")
	ErrInvalidSignedBlocksWindow        = sdkerrors.Register(RootCodespace, codeInvalidSignedBlocksWindow, "cannot decrease signed blocks window")
	ErrInvalidMinSignedPerWindow        = sdkerrors.Register(RootCodespace, codeInvalidMinSignedPerWindow, "cannot decrease minimum signed per window")
	ErrInvalidDowntimeJailDuration      = sdkerrors.Register(RootCodespace, codeInvalidDowntimeJailDuration, "cannot decrease downtime jail duration")
	ErrInvalidSlashFractionDoubleSign   = sdkerrors.Register(RootCodespace, codeInvalidSlashFractionDoubleSign, "cannot increase slash fraction double sign")
	ErrInvalidSlashFractionDowntime     = sdkerrors.Register(RootCodespace, codeInvalidSlashFractionDowntime, "cannot increase slash fraction downtime")
	ErrChangingUnbondingPeriodForbidden = sdkerrors.Register(RootCodespace, codeChangingUnbondingPeriodForbidden, "changing unbonding period not allowed")
	ErrChangingBondDenomForbidden       = sdkerrors.Register(RootCodespace, codeChangingBondDenomForbidden, "changing bond denom not allowed")
)
