package keeper

import (
	"github.com/Canto-Network/Canto/v5/x/govshuttle/types"
)

var _ types.QueryServer = Keeper{}
