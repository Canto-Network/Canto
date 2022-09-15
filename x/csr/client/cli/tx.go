package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/Canto-Network/Canto/v2/x/csr/types"
	"github.com/cosmos/cosmos-sdk/client"
)

// GetTxCmd returns the transaction commands for this module
func GetTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("%s has no transaction commands (everything is done on the EVM)", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}
	return cmd
}
