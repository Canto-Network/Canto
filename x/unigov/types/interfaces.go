package types

import (
	evmtypes "github.com/Canto-Network/ethermint-v2/x/evm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
)

//Required for deploying Map-Contract/Caling setter methods of Map-Contract
type ERC20Keeper interface {
	CallEVM(ctx sdk.Context, abi abi.ABI, from, contract common.Address, commit bool, method string, args ...interface{}) (*evmtypes.MsgEthereumTxResponse, error)

	CallEVMWithData(
		ctx sdk.Context,
		from common.Address,
		contract *common.Address,
		data []byte,
		commit bool,
	) (*evmtypes.MsgEthereumTxResponse, error)
}

// AccountKeeper defines the expected account keeper used for simulations (noalias)
type AccountKeeper interface {
	//GetAccount(ctx sdk.Context, addr sdk.AccAddress)
	GetModuleAddress(moduleName string) sdk.AccAddress
	GetSequence(sdk.Context, sdk.AccAddress) (uint64, error)
}

type GovKeeper interface {
	GetProposalID(ctx sdk.Context) (uint64, error)
}

// BankKeeper defines the expected interface needed to retrieve account balances.
// type BankKeeper interface {
// 	SpendableCoins(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins
// 	// Methods imported from bank should be defined here
// }
