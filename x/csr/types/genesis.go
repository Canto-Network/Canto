package types

// DefaultIndex is the default capability global index
const DefaultIndex uint64 = 1

// DefaultGenesis returns the default Capability genesis state
func DefaultGenesis() *GenesisState {
	return &GenesisState{
		Params: DefaultParams(),
	}
}

// By default, there should be no CSRs on genesis because the CSR turnstile and NFT smart contracts
// have not been deployed yet. Checks if params are valid.
func (gs GenesisState) Validate() error {
	return gs.Params.Validate()
}
