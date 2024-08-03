package keeper

import (
	"github.com/Canto-Network/Canto/v8/x/govshuttle/types"
)

var _ types.QueryServer = Keeper{}
