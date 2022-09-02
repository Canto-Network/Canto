package keeper

// DONTCOVER

import (
	"github.com/Canto-Network/Canto/v2/x/csr/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// x/csr module sentinel errors
var (
	ErrPrevRegisteredSmartContract = sdkerrors.Register(types.ModuleName, 2000, "csr::CSR")
	ErrFeeCollectorDistribution    = sdkerrors.Register(types.ModuleName, 2001, "csr::EVMHOOK")
)
