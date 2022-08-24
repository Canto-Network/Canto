package posthandler

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	feeskeeper "github.com/evmos/ethermint/x/feemarket/keeper"
)

// the baseFeeBurn PostHandler decorator defines the necessary logic for burning the base fee *gasUsed per tx
// notice that this logic must be handled after RunTx as the gasUsed value will have been populated,
type BaseBurnDecorator struct {
	// fees keeper required for retrieving the baseFee for the current block
	FeesKeeper feeskeeper.Keeper
	// bank keeper required for burning tx fees
	BankKeeper bankkeeper.Keeper
}

const (
	// denom of gas fees
	denomMint = "acanto"
	// account for which all tx fees will be sent after runTx
	feeCollector = authtypes.FeeCollectorName
)

// NewBaseBurnDecorator returns the decorator used in post handler for burning of the base fee per Tx
func NewBaseBurnDecorator(bankKeeper bankkeeper.Keeper, feesKeeper feeskeeper.Keeper) sdk.AnteDecorator {
	return BaseBurnDecorator{
		FeesKeeper: feesKeeper,
		BankKeeper: bankKeeper,
	}
}

// baseBurnDecorator implements the sdk.AnteDecorator interface, this method is executed once the PostHandler AnteHandle method is executed,
func (bd BaseBurnDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (sdk.Context, error) {
	if err := bd.burnBaseFee(ctx, tx); err != nil {
		return ctx, err
	}

	return next(ctx, tx, simulate)
}

// burnBaseFee burns tx.gasUsed * blockBaseFee as per the feeMarket Keeper
// No need for stateful validation of the msg as it has already been validated and checked against state (post RunTx)
func (bd BaseBurnDecorator) burnBaseFee(ctx sdk.Context, tx sdk.Tx) error {
	feeTx, ok := tx.(sdk.FeeTx)
	// this is not a feeTx so proceed without error
	if !ok || feeTx.GetGas() == 0 {
		return nil
	}

	// retrieve baseFee from feemarket param-space, if nil, return 
	baseFee :=  bd.FeesKeeper.GetBaseFee(ctx)
	if baseFee == nil {
		// EIP1559 has not been enabled for this block
		return nil
	}
	// identify the amt of Gas used for this tx
	gasUsed := ctx.BlockGasMeter().GasConsumed()
	amtToBurn := sdk.NewDecFromBigInt(baseFee).MulInt64(int64(gasUsed))
	
	// now burn the amt to burn
	err := bd.BankKeeper.BurnCoins(ctx, feeCollector, sdk.NewCoins(sdk.NewCoin(denomMint, sdk.Int(amtToBurn))))
	if err != nil { 
		return err
	}

	// burn coins from the feeCollector account after they have been sent
	return nil
}
