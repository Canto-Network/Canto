package csr

import (
	"fmt"

	errorsmod "cosmossdk.io/errors"
	"github.com/Canto-Network/Canto/v7/x/csr/keeper"
	"github.com/Canto-Network/Canto/v7/x/csr/types"
	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// NewHandler for the CSR module. This will not handle msg types because all transactions occur directly
// through the CSR Turnstile Smart Contract.
func NewHandler(k keeper.Keeper) baseapp.MsgServiceHandler {
	return func(ctx sdk.Context, msg sdk.Msg) (*sdk.Result, error) {
		errMsg := fmt.Sprintf("unrecognized %s message type: %T", types.ModuleName, msg)
		return nil, errorsmod.Wrap(sdkerrors.ErrUnknownRequest, errMsg)
	}
}
