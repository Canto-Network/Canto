package types

// DefaultIndex is the default capability global index
const DefaultIndex uint64 = 1

// DefaultGenesis returns the default Capability genesis state
func DefaultGenesis() *GenesisState {
	return NewGenesisState(DefaultParams(), []CSR{}, "")
}

func NewGenesisState(params Params, csrs []CSR, turnstileAddress string) *GenesisState {
	return &GenesisState{
		Params:           params,
		Csrs:             csrs,
		TurnstileAddress: turnstileAddress,
	}
}

// By default, there should be no CSRs on genesis because the CSR turnstile and NFT smart contracts
// have not been deployed yet. Checks if params are valid.
func (gs GenesisState) Validate() error {
	return gs.Params.Validate()
}
