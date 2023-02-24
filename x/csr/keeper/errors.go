package keeper

// DONTCOVER

import (
	"github.com/Canto-Network/Canto/v6/x/csr/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// x/csr module sentinel errors
var (
	ErrPrevRegisteredSmartContract = sdkerrors.Register(types.ModuleName, 2000, "You cannot register a smart contract that was previously registered")
	ErrFeeDistribution             = sdkerrors.Register(types.ModuleName, 2001, "There was an error sending fees from the fee collector to NFT")
	ErrContractDeployments         = sdkerrors.Register(types.ModuleName, 2004, "There was an error deploying a contract from the CSR module")
	ErrMethodCall                  = sdkerrors.Register(types.ModuleName, 2005, "There was an error calling a method on a smart contract from the module account")
	ErrRegisterInvalidContract     = sdkerrors.Register(types.ModuleName, 2007, "Register/update must have a valid smart contract address entered")
	ErrNonexistentAcct             = sdkerrors.Register(types.ModuleName, 2008, "The recipient of a register event must be an EOA that exists")
	ErrNonexistentCSR              = sdkerrors.Register(types.ModuleName, 2009, "The CSR that was queried does not currently exist")
	ErrNFTNotFound                 = sdkerrors.Register(types.ModuleName, 2010, "The NFT that was queried does not currently exist")
	ErrDuplicateNFTID              = sdkerrors.Register(types.ModuleName, 2011, "There cannot be duplicate NFT IDs passed into a register event")
)
