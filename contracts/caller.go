package contracts

import (
	_ "embed" // embed compiled smart contract

	evmtypes "github.com/evmos/ethermint/x/evm/types"
)

//go:embed compiled_contracts/caller.json

// CallerContract is the compiled ERC20Burnable contract
var CallerContract evmtypes.CompiledContract

func init() {
	// err := json.Unmarshal(callerJSON, &CallerContract)
	// if err != nil {
	// 	panic(err)
	// }
}
