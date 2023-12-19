package types

// DefaultIndex is the default capability global index
const DefaultIndex uint64 = 1

// DefaultGenesis returns the default Capability genesis state
func DefaultGenesis() *GenesisState {
	return NewGenesisState(DefaultParams(), nil)
}

func NewGenesisState(params Params, portAddress []byte) *GenesisState {
	return &GenesisState{
		Params:      params,
		PortAddress: portAddress,
		// this line is used by starport scaffolding # genesis/types/init
	}
}

// Validate performs basic genesis state validation returning an error upon any
// failure.
func (gs GenesisState) Validate() error {
	// this line is used by starport scaffolding # genesis/types/validate

	return gs.Params.Validate()
}
