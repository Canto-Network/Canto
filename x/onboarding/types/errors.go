package types

import (
	errorsmod "cosmossdk.io/errors"
)

// errors
var (
	ErrBlockedAddress = errorsmod.Register(ModuleName, 2, "blocked address")
	ErrInvalidType    = errorsmod.Register(ModuleName, 3, "invalid type")
)
