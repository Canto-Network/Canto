package types

// DONTCOVER

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// x/govshuttle module sentinel errors
var (
	Errgovshuttle = sdkerrors.Register(ModuleName, 1100, "govshuttle error")
)
