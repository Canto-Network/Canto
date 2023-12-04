package types

// DONTCOVER

import (
	errorsmod "cosmossdk.io/errors"
)

// x/govshuttle module sentinel errors
var (
	Errgovshuttle = errorsmod.Register(ModuleName, 1100, "govshuttle error")
)
