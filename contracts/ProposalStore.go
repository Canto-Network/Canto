package contracts

import (
	_ "embed" // embed compiled smart contract
	"encoding/json"

	evmtypes "github.com/Canto-Network/ethermint-v2/x/evm/types"
)

var (
	//go:embed compiled_contracts/ProposalStore.json
	ProposalStoreJSON []byte

	// ProposalStore Contract is the EVM object representing the contract object
	ProposalStoreContract evmtypes.CompiledContract
)

func init() {
	err := json.Unmarshal(ProposalStoreJSON, &ProposalStoreContract)
	if err != nil {
		panic(err)
	}
}
