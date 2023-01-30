package contracts

import (
	_ "embed" // embed compiled smart contract

	evmtypes "github.com/evmos/ethermint/x/evm/types"
)

//go:embed compiled_contracts/callee.json

// CalleeContract is the compiled ERC20Burnable contract
var CalleeContract evmtypes.CompiledContract

func init() {
	// err := json.Unmarshal(calleeJSON, &CalleeContract)
	// if err != nil {
	// 	// panic(err)
	// 	fmt.Println("ERROR HERE")
	// }
}
