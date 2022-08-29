package types

// this line is used by starport scaffolding # genesis/types/import

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
	// seenCSRPool := make(map[string]bool)
	// seenCSRNFTs := make(map[string]bool)

	// for _, csr := range gs.Csrs {
	// 	// Validate that the csr object was correctly formatted
	// 	if err := csr.Validate(); err != nil {
	// 		return err
	// 	}

	// 	pool := csr.CsrPool
	// 	// Validate that the pool is correctly formatted
	// 	if err := pool.Validate(); err != nil {
	// 		return err
	// 	}

	// 	// Validate that no pool addresses are duplicated
	// 	if seenCSRPool[pool.PoolAddress] {
	// 		return sdkerrors.Wrapf(ErrDuplicatePools, "GensisState::Validate the address of this pool has already been seen before")
	// 	}

	// 	// Validate that no NFTs are duplicated
	// 	for _, nft := range pool.CsrNfts {
	// 		id := pool.PoolAddress + strconv.FormatUint(nft.Id, 10)
	// 		if seenCSRNFTs[id] {
	// 			return sdkerrors.Wrapf(ErrDuplicateNFTs, "GensisState::Validate there are duplicate NFTs across the different pools")
	// 		}
	// 	}

	// 	seenCSRPool[pool.PoolAddress] = true

	// }

	return gs.Params.Validate()
}
