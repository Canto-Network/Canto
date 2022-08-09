package keeper

import (
	"github.com/Canto-Network/Canto/v2/x/govshuttle/types"
)

var _ types.QueryServer = Keeper{}
