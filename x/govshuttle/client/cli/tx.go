package cli

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"

	addresscodec "cosmossdk.io/core/address"
	errorsmod "cosmossdk.io/errors"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
	"github.com/cosmos/cosmos-sdk/version"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	"github.com/Canto-Network/Canto/v8/x/govshuttle/types"
)

var (
	DefaultRelativePacketTimeoutTimestamp = uint64((time.Duration(10) * time.Minute).Nanoseconds())

	FlagAuthority = "authority"
)

// NewTxCmd returns a root CLI command handler for certain modules/govshuttle transaction commands.
func NewTxCmd(ac addresscodec.Codec) *cobra.Command {
	txCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "govshuttle subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	txCmd.AddCommand(
		NewLendingMarketProposalCmd(ac),
		NewTreasuryProposalCmd(ac),
	)
	return txCmd
}

// NewRegisterCoinProposalCmd implements the command to submit a community-pool-spend proposal
func NewLendingMarketProposalCmd(ac addresscodec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "lending-market [metadata]",
		Args:  cobra.ExactArgs(1),
		Short: "Submit a lending market proposal",
		Long: `Submit a proposal for the Canto Lending Market along with an initial deposit.
Upon passing, the
The proposal details must be supplied via a JSON file.`,
		Example: fmt.Sprintf(`$ %s tx gov submit-proposal lending-market <path/to/metadata.json> --from=<key_or_address> --title=<title> --description=<description>

Where metadata.json contains (example):

{
	"Account": ["address_1", "address_2"],
	"PropId":  1,
	"values": ["canto", "osmo"],
	"calldatas": ["calldata1", "calldata2"],
	"signatures": ["func1", "func2"]
}`, version.AppName,
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			proposal, err := ReadGovPropFlags(clientCtx, cmd.Flags())
			if err != nil {
				return err
			}

			propMetaData, err := ParseLendingMarketMetadata(clientCtx.Codec, args[0])
			if err != nil {
				return errorsmod.Wrap(err, "Failure to parse JSON object")
			}

			// validate basic logic
			cd, vals, sigs := len(propMetaData.GetCalldatas()), len(propMetaData.GetValues()), len(propMetaData.GetSignatures())
			if cd != vals {
				return errorsmod.Wrap(govtypes.ErrInvalidProposalContent, "proposal array arguments must be same length")
			}
			if vals != sigs {
				return errorsmod.Wrap(govtypes.ErrInvalidProposalContent, "proposal array arguments must be same length")
			}

			authority, _ := cmd.Flags().GetString(FlagAuthority)
			if authority != "" {
				if _, err = ac.StringToBytes(authority); err != nil {
					return fmt.Errorf("invalid authority address: %w", err)
				}
			} else {
				authority = sdk.AccAddress(address.Module("gov")).String()
			}

			if err := proposal.SetMsgs([]sdk.Msg{
				&types.MsgLendingMarketProposal{
					Authority:   authority,
					Title:       proposal.Title,
					Description: proposal.Summary,
					Metadata:    &propMetaData,
				},
			}); err != nil {
				return fmt.Errorf("failed to create submit lending market proposal message: %w", err)
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), proposal)
		},
	}
	flags.AddTxFlagsToCmd(cmd)
	AddGovPropFlagsToCmd(cmd)

	return cmd
}

// Register TreasuryProposal submit cmd
func NewTreasuryProposalCmd(ac addresscodec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "treasury-proposal [metadata]",
		Args:  cobra.ExactArgs(1),
		Short: "Submit a proposal to the Canto Treasury",
		Long: `Submit a proposal for the Canto Treasury along with an initial deposit.
Upon passing, the
The proposal details must be supplied via a JSON file.`,
		Example: fmt.Sprintf(`$ %s tx gov submit-proposal treasury-proposal <path/to/metadata.json> --from=<key_or_address> --title=<title> --description=<description>

Where metadata.json contains (example):

{
	"recipient": "0xfffffff...",
	"PropID":  1,
	"amount": 1,
	"denom": "canto/note"
}`, version.AppName,
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			proposal, err := ReadGovPropFlags(clientCtx, cmd.Flags())
			if err != nil {
				return err
			}

			propMetaData, err := ParseTreasuryMetadata(clientCtx.Codec, args[0])
			if err != nil {
				return errorsmod.Wrap(err, "Failure to parse JSON object")
			}

			// validate basic logic
			s := strings.ToLower(propMetaData.GetDenom())
			if s != "canto" && s != "note" {
				return errorsmod.Wrapf(govtypes.ErrInvalidProposalContent, "%s is not a valid denom string", propMetaData.GetDenom())
			}

			authority, _ := cmd.Flags().GetString(FlagAuthority)
			if authority != "" {
				if _, err = ac.StringToBytes(authority); err != nil {
					return fmt.Errorf("invalid authority address: %w", err)
				}
			} else {
				authority = sdk.AccAddress(address.Module("gov")).String()
			}

			if err := proposal.SetMsgs([]sdk.Msg{
				&types.MsgTreasuryProposal{
					Authority:   authority,
					Title:       proposal.Title,
					Description: proposal.Summary,
					Metadata:    &propMetaData,
				},
			}); err != nil {
				return fmt.Errorf("failed to create submit treasury proposal message: %w", err)
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), proposal)
		},
	}
	flags.AddTxFlagsToCmd(cmd)
	AddGovPropFlagsToCmd(cmd)

	return cmd
}
