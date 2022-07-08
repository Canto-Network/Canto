package ante

import (
	"math/big"

	evmtypes "github.com/Canto-Network/ethermint-v2/x/evm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/params"
)

// EvmKeeper defines the expected keeper interface used on the AnteHandler
type EvmKeeper interface {
	GetParams(ctx sdk.Context) (params evmtypes.Params)
	ChainID() *big.Int
	GetBaseFee(ctx sdk.Context, ethCfg *params.ChainConfig) *big.Int
}
