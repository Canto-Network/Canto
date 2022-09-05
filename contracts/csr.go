package contracts

import (
	_ "embed" // embed compiled smart contract
	"encoding/json"

	evmtypes "github.com/evmos/ethermint/x/evm/types"
)

var (
	//go:embed compiled_contracts/Turnstile.json
	TurnstileJSON     []byte
	TurnstileContract evmtypes.CompiledContract
	//go:embed compiled_contracts/CSRNFT.json
	CSRNFTJSON     []byte
	CSRNFTContract evmtypes.CompiledContract
)

func init() {

	err := json.Unmarshal(TurnstileJSON, &TurnstileContract)
	if err != nil {
		panic(err)
	}

	err = json.Unmarshal(CSRNFTJSON, &CSRNFTContract)
	if err != nil {
		panic(err)
	}

	if len(TurnstileContract.Bin) == 0 {
		panic("The turnstile contract was not loaded")
	}

	if len(CSRNFTContract.Bin) == 0 {
		panic("The CSRNFT contract was not loaded")
	}
}
