package types

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// errors
var (
	ErrBlockedAddress = sdkerrors.Register(ModuleName, 2, "blocked address")
	ErrInvalidType    = sdkerrors.Register(ModuleName, 3, "invalid type")
)
