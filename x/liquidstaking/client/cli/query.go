package cli

import (
	"context"
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/version"

	"github.com/Canto-Network/Canto/v6/x/liquidstaking/types"
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
		CmdQueryEpoch(),
		CmdQueryChunk(),
		CmdQueryInsurances(),
		CmdQueryInsurance(),
	)

	return cmd
}

// CmdQueryParams implements a command that will return the current parameters of the
// liquidstaking module.
func CmdQueryParams() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "params",
		Short: fmt.Sprintf("Query the current parameters of %s module", types.ModuleName),
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

// CmdQueryEpoch implements a command that will return the Epoch from the Epoch store
func CmdQueryEpoch() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "epoch",
		Short: fmt.Sprintf("Query the epoch of %s module", types.ModuleName),
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)

			queryClient := types.NewQueryClient(clientCtx)

			request := &types.QueryEpochRequest{}

			// Query store
			response, err := queryClient.Epoch(context.Background(), request)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(response)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// CmdQueryChunk implements a command that will return a Chunk given a chunk id
func CmdQueryChunks() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "chunks",
		Args:    cobra.ExactArgs(1),
		Short:   "Query Chunks",
		Example: fmt.Sprintf("query %s chunks --status", version.AppName),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			pageRequest, err := client.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			chunkStatusStr, _ := cmd.Flags().GetString(FlagChunkStatus)
			request := &types.QueryChunksRequest{
				Status:     types.ChunkStatus(types.ChunkStatus_value[chunkStatusStr]),
				Pagination: pageRequest,
			}

			queryClient := types.NewQueryClient(clientCtx)

			// Query store
			response, err := queryClient.Chunks(context.Background(), request)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(response)
		},
	}
	cmd.Flags().AddFlagSet(flagSetChunks())
	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// CmdQueryChunk implements a command that will return a Chunk given a chunk id
func CmdQueryChunk() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "chunk [chunkId]",
		Args:    cobra.ExactArgs(1),
		Short:   "Query the Chunk associated with a given chunk id",
		Example: fmt.Sprintf("%s query liquidstaking chunk 1", version.AppName),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			// arg must be converted to a uint
			chunkId, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return err
			}

			request := &types.QueryChunkRequest{Id: chunkId}
			// Query store
			response, err := queryClient.Chunk(context.Background(), request)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(response)
		},
	}
	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// CmdQueryInsurances implements a command that will return insurances in liquidstaking module
func CmdQueryInsurances() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "insurances",
		Args:    cobra.ExactArgs(1),
		Short:   "Query Insurances",
		Example: fmt.Sprintf("query %s insurances --status <InsuranceStatus> --validator-address <validatorAddress> --provider-address <providerAddress>", version.AppName),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			pageRequest, err := client.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			insuranceStatusStr, _ := cmd.Flags().GetString(FlagInsuranceStatus)
			validatorAddress, _ := cmd.Flags().GetString(FlagValidatorAddress)
			providerAddress, _ := cmd.Flags().GetString(FlagProviderAddress)

			request := &types.QueryInsurancesRequest{
				Status:           types.InsuranceStatus(types.InsuranceStatus_value[insuranceStatusStr]),
				ValidatorAddress: validatorAddress,
				ProviderAddress:  providerAddress,
				Pagination:       pageRequest,
			}

			queryClient := types.NewQueryClient(clientCtx)

			// Query store
			response, err := queryClient.Insurances(context.Background(), request)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(response)
		},
	}
	cmd.Flags().AddFlagSet(flagSetChunks())
	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// CmdQueryInsurance implements a command that will return a Chunk given an insurance id
func CmdQueryInsurance() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "insurance [insuranceId]",
		Args:    cobra.ExactArgs(1),
		Short:   "Query the Insurance associated with a given insurance id",
		Example: fmt.Sprintf("%s query liquidstaking insurance 1", version.AppName),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			// arg must be converted to a uint
			insuranceId, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return err
			}

			request := &types.QueryInsuranceRequest{Id: insuranceId}
			// Query store
			response, err := queryClient.Insurance(context.Background(), request)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(response)
		},
	}
	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}
