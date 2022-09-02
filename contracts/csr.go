package contracts

import (
	_ "embed" // embed compiled smart contract
	"encoding/json"

	"github.com/ethereum/go-ethereum/common"

	"github.com/Canto-Network/Canto/v2/x/csr/types"
	evmtypes "github.com/evmos/ethermint/x/evm/types"
)

var (
	//go:embed compiled_contracts/Turnstile.json
	TurnstileJSON            []byte
	TurnstileContract        evmtypes.CompiledContract
	TurnstileContractAddress common.Address

	// TODO: Need to add the CSR NFT code here once written
)

func init() {
	TurnstileContractAddress = types.ModuleAddress

	err := json.Unmarshal(TurnstileJSON, &TurnstileContract)
	if err != nil {
		panic(err)
	}

	if len(TurnstileContract.Bin) == 0 {
		panic("The turnstile contract was not loaded")
	}
}
