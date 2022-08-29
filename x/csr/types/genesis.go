package types

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// DefaultIndex is the default capability global index
const DefaultIndex uint64 = 1

// DefaultGenesis returns the default Capability genesis state
func DefaultGenesis() *GenesisState {
	return &GenesisState{
		Params: DefaultParams(),
	}
}

// Validate performs basic genesis state validation returning an error upon any
// failure. Validation will check that there are no duplicate CSRPools registered,
// that there are not the same NFTs registered for different groups, and that the smart contracts
// in a given pool are not repeated elsewhere
func (gs GenesisState) Validate() error {
	if len(gs.Csrs) != 0 {
		return sdkerrors.Wrapf(ErrNonZeroCSRs, "GenesisState::Validate you cannot initialize a genesis state with a set of existing csrs.")
	}

	return gs.Params.Validate()
}
