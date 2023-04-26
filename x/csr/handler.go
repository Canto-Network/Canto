package csr

import (
	"fmt"

	"github.com/Canto-Network/Canto/v6/x/csr/keeper"
	"github.com/Canto-Network/Canto/v6/x/csr/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// NewHandler for the CSR module. This will not handle msg types because all transactions occur directly
// through the CSR Turnstile Smart Contract.
func NewHandler(k keeper.Keeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) (*sdk.Result, error) {
		errMsg := fmt.Sprintf("unrecognized %s message type: %T", types.ModuleName, msg)
		return nil, sdkerrors.Wrap(sdkerrors.ErrUnknownRequest, errMsg)
	}
}
