package keeper

// DONTCOVER

import (
	"github.com/Canto-Network/Canto/v2/x/csr/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// x/csr module sentinel errors
var (
	ErrPrevRegisteredSmartContract = sdkerrors.Register(types.ModuleName, 2000, "csr::CSR")
	ErrFeeDistribution             = sdkerrors.Register(types.ModuleName, 2001, "csr::EVMHOOK")
	ErrCSRNFTNotDeployed           = sdkerrors.Register(types.ModuleName, 2002, "csr::Keeper")
	ErrAddressDerivation           = sdkerrors.Register(types.ModuleName, 2003, "csr::Keeper")
	ErrContractDeployments         = sdkerrors.Register(types.ModuleName, 2004, "csr::Keeper")
	ErrMethodCall                  = sdkerrors.Register(types.ModuleName, 2005, "csr::Keeper")
	ErrUnpackData                  = sdkerrors.Register(types.ModuleName, 2006, "csr:Keeper")
	ErrRegisterEOA                 = sdkerrors.Register(types.ModuleName, 2007, "csr::EventHandler")
	ErrNonexistentAcct             = sdkerrors.Register(types.ModuleName, 2008, "csr::EventHandler")
	ErrNonexistentCSR              = sdkerrors.Register(types.ModuleName, 2009, "csr::EventHandler")
	ErrNFTNotFound                 = sdkerrors.Register(types.ModuleName, 2010, "csr::EventHandler")
	ErrDuplicateNFTID              = sdkerrors.Register(types.ModuleName, 2011, "csr::EventHandler")
)
