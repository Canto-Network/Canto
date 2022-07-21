package keeper

import (
	"github.com/Canto-Network/Canto/v1/x/govshuttle/types"
)

var _ types.QueryServer = Keeper{}
