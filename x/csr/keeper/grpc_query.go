package keeper

import (
	"github.com/Canto-Network/Canto/v2/x/csr/types"
)

var _ types.QueryServer = Keeper{}
