package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/Canto-Network/Canto/v7/x/csr/types"
	"github.com/cosmos/cosmos-sdk/client"
)

// GetTxCmd returns the transaction methods allowed for the CLI. However, currently all transaction or state transition
// functionality is triggered through the Turnstile Smart Contract.
func GetTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("%s has no transaction commands (everything transaction is triggered via the Turnstile Smart Contract)", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}
	return cmd
}
