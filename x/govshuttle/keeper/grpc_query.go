package keeper

import (
	"github.com/Canto-Network/Canto/v6/x/govshuttle/types"
)

var _ types.QueryServer = Keeper{}
