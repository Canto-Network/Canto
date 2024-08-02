package keeper

// DONTCOVER

import (
	errorsmod "cosmossdk.io/errors"
	"github.com/Canto-Network/Canto/v7/x/csr/types"
)

// x/csr module sentinel errors
var (
	ErrPrevRegisteredSmartContract = errorsmod.Register(types.ModuleName, 2000, "You cannot register a smart contract that was previously registered")
	ErrFeeDistribution             = errorsmod.Register(types.ModuleName, 2001, "There was an error sending fees from the fee collector to NFT")
	ErrContractDeployments         = errorsmod.Register(types.ModuleName, 2004, "There was an error deploying a contract from the CSR module")
	ErrMethodCall                  = errorsmod.Register(types.ModuleName, 2005, "There was an error calling a method on a smart contract from the module account")
	ErrRegisterInvalidContract     = errorsmod.Register(types.ModuleName, 2007, "Register/update must have a valid smart contract address entered")
	ErrNonexistentAcct             = errorsmod.Register(types.ModuleName, 2008, "The recipient of a register event must be an EOA that exists")
	ErrNonexistentCSR              = errorsmod.Register(types.ModuleName, 2009, "The CSR that was queried does not currently exist")
	ErrNFTNotFound                 = errorsmod.Register(types.ModuleName, 2010, "The NFT that was queried does not currently exist")
	ErrDuplicateNFTID              = errorsmod.Register(types.ModuleName, 2011, "There cannot be duplicate NFT IDs passed into a register event")
)
