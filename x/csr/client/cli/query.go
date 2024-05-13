package cli

import (
	"context"
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/version"

	"github.com/Canto-Network/Canto/v7/x/csr/types"
)

// GetQueryCmd returns the cli query commands for the CSR module
func GetQueryCmd(queryRoute string) *cobra.Command {
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("Querying commands for the %s module", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		CmdQueryParams(),
		CmdQueryCSRs(),
		CmdQueryCSRByNFT(),
		CmdQueryCSRByContract(),
		CmdQueryTurnstile(),
	)

	return cmd
}

// CmdQueryParams implements a command that will return the current parameters of the
// CSR module.
func CmdQueryParams() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "params",
		Short: "Query the current parameters of the CSR module",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)

			queryClient := types.NewQueryClient(clientCtx)

			request := &types.QueryParamsRequest{}

			// Query store
			response, err := queryClient.Params(context.Background(), request)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(&response.Params)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// CmdQueryCSRs implements a command that will return the CSRs from the CSR store
func CmdQueryCSRs() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "csrs",
		Short: "Query all registered contracts and NFTs for the CSR module",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)

			queryClient := types.NewQueryClient(clientCtx)

			pageRequest, err := client.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			request := &types.QueryCSRsRequest{
				Pagination: pageRequest,
			}

			// Query store
			response, err := queryClient.CSRs(context.Background(), request)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(response)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// CmdQueryCSRByNFT implements a command that will return a CSR given a NFT ID
func CmdQueryCSRByNFT() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "nft [nftID]",
		Args:    cobra.ExactArgs(1),
		Short:   "Query the CSR associated with a given NFT ID",
		Long:    "Query the CSR associated with a given NFT ID",
		Example: fmt.Sprintf("%s query csr nft <address>", version.AppName),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			// arg must be converted to a uint
			nftID, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return err
			}

			request := &types.QueryCSRByNFTRequest{NftId: nftID}
			// Query store
			response, err := queryClient.CSRByNFT(context.Background(), request)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(response)
		},
	}
	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// CmdQueryCSRByContract implements a cobra command that will return the CSR associated
// given a smart contract address
func CmdQueryCSRByContract() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "contract [address]",
		Args:    cobra.ExactArgs(1),
		Short:   "Query the CSR associated with a given smart contract adddress",
		Long:    "Query the CSR associated with a given smart contract adddress",
		Example: fmt.Sprintf("%s query csr contract <address>", version.AppName),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			request := &types.QueryCSRByContractRequest{Address: args[0]}

			// Query store
			response, err := queryClient.CSRByContract(context.Background(), request)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(response)
		},
	}
	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// CmdQueryTurnstile implements a cobra command that will return the Turnstile address that was deployed by the module account
func CmdQueryTurnstile() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "turnstile",
		Short: "Query the address of the turnstile smart contract deployed by the module account",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)

			queryClient := types.NewQueryClient(clientCtx)

			request := &types.QueryTurnstileRequest{}

			// Query store
			response, err := queryClient.Turnstile(context.Background(), request)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(response)
		},
	}
	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}
