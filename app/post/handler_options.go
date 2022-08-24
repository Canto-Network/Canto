package posthandler

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	authante "github.com/cosmos/cosmos-sdk/x/auth/ante"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	feeskeeper "github.com/evmos/ethermint/x/feemarket/keeper"
)

// define interfaces needed for the postHandler fee burning
// HandlerOptions defines the list of module keepers required to run the canto
// PostHandler decorators.
type HandlerOptions struct {
	// feeMarket Keeper for retrieving baseFeePerGas
	FeesKeeper *feeskeeper.Keeper
	// bank keeper for burning baseFeePerGas
	BankKeeper *bankkeeper.Keeper
}

const (
	// type of EVM messages
	ethTxProto = "/ethermint.evm.v1.ExtensionOptionsEthereumTx"
)

// define validation method for the HandlerOptions struct
func (options HandlerOptions) Validate() error {
	if options.FeesKeeper == nil {
		return sdkerrors.Wrap(sdkerrors.ErrLogic, "fees keeper is required for posthandler")
	}

	if options.BankKeeper == nil {
		return sdkerrors.Wrap(sdkerrors.ErrLogic, "bank keeper is required for posthandler")
	}

	return nil
}

// create postHandler constructor
func NewPostHandler(options HandlerOptions) sdk.AnteHandler {
	return func(
		ctx sdk.Context, tx sdk.Tx, simulate bool,
	) (newCtx sdk.Context, err error) {
		var postHandler sdk.AnteHandler

		//retrieve exension options
		txWithExtensions, ok := tx.(authante.HasExtensionOptionsTx)
		if ok {
			// get Extension Options
			opts := txWithExtensions.GetExtensionOptions()
			if len(opts) > 0 {
				// check if this is indeed an Eth Tx and Add the baseBurnDecorator
				if opts[0].GetTypeUrl() == ethTxProto {
					postHandler = NewEthPostHandler(options)
				}
			}
		} else {
			postHandler = NewCosmosPostHandler(options)
		}

		return postHandler(ctx, tx, simulate)
	}
}

func NewCosmosPostHandler(options HandlerOptions) sdk.AnteHandler {
	return sdk.ChainAnteDecorators()
}

func NewEthPostHandler(options HandlerOptions) sdk.AnteHandler {
	return sdk.ChainAnteDecorators(
		NewBaseBurnDecorator(*options.BankKeeper, *options.FeesKeeper),
	)
}
