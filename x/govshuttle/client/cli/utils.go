package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/gov/client/cli"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"

	"github.com/Canto-Network/Canto/v8/x/govshuttle/types"
)

// PARSING METADATA ACCORDING TO PROPOSAL STRUCT IN GOVTYPES TYPE IN govshuttle

// ParseRegisterCoinProposal reads and parses a ParseRegisterCoinProposal from a file.
func ParseLendingMarketMetadata(cdc codec.JSONCodec, metadataFile string) (types.LendingMarketMetadata, error) {
	propMetaData := types.LendingMarketMetadata{}

	contents, err := os.ReadFile(filepath.Clean(metadataFile))
	if err != nil {
		return propMetaData, err
	}

	if err = cdc.UnmarshalJSON(contents, &propMetaData); err != nil {
		return propMetaData, err
	}

	propMetaData.PropId = 0

	return propMetaData, nil
}

func ParseTreasuryMetadata(cdc codec.JSONCodec, metadataFile string) (types.TreasuryProposalMetadata, error) {
	propMetaData := types.TreasuryProposalMetadata{}

	contents, err := os.ReadFile(filepath.Clean(metadataFile))
	if err != nil {
		return propMetaData, err
	}

	if err = cdc.UnmarshalJSON(contents, &propMetaData); err != nil {
		return propMetaData, err
	}

	propMetaData.PropID = 0

	return propMetaData, nil
}

// AddGovPropFlagsToCmd adds flags for defining MsgSubmitProposal fields.
//
// See also ReadGovPropFlags.
// ref. github.com/cosmos/cosmos-sdk/x/gov/client/cli/util.go::AddGovPropFlagsToCmd
func AddGovPropFlagsToCmd(cmd *cobra.Command) {
	cmd.Flags().String(cli.FlagTitle, "", "title of proposal")
	cmd.Flags().String(cli.FlagDescription, "", "description of proposal")
	cmd.Flags().String(cli.FlagDeposit, "1acanto", "deposit of proposal")
	cmd.Flags().String(FlagAuthority, "", "The address of the upgrade module authority (defaults to gov)")

	if err := cmd.MarkFlagRequired(cli.FlagTitle); err != nil {
		panic(err)
	}
	if err := cmd.MarkFlagRequired(cli.FlagDescription); err != nil {
		panic(err)
	}
	if err := cmd.MarkFlagRequired(cli.FlagDeposit); err != nil {
		panic(err)
	}
}

// ReadGovPropFlags parses a MsgSubmitProposal from the provided context and flags.
// Setting the messages is up to the caller.
//
// See also AddGovPropFlagsToCmd.
// ref. github.com/cosmos/cosmos-sdk/x/gov/client/cli/util.go::ReadGovPropFlags
func ReadGovPropFlags(clientCtx client.Context, flagSet *pflag.FlagSet) (*govv1.MsgSubmitProposal, error) {
	rv := &govv1.MsgSubmitProposal{}

	deposit, err := flagSet.GetString(cli.FlagDeposit)
	if err != nil {
		return nil, fmt.Errorf("could not read deposit: %w", err)
	}
	if len(deposit) > 0 {
		rv.InitialDeposit, err = sdk.ParseCoinsNormalized(deposit)
		if err != nil {
			return nil, fmt.Errorf("invalid deposit: %w", err)
		}
	}

	rv.Title, err = flagSet.GetString(cli.FlagTitle)
	if err != nil {
		return nil, fmt.Errorf("could not read title: %w", err)
	}

	rv.Summary, err = flagSet.GetString(cli.FlagDescription)
	if err != nil {
		return nil, fmt.Errorf("could not read summary: %w", err)
	}

	rv.Proposer = clientCtx.GetFromAddress().String()

	return rv, nil
}
