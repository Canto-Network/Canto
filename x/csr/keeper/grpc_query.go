package keeper

import (
	"context"
	"math/big"
	"strings"

	"github.com/Canto-Network/Canto/v2/x/csr/types"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	ethermint "github.com/evmos/ethermint/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _ types.QueryServer = Keeper{}

// Params returns the CSR module parameters
func (k Keeper) Params(c context.Context, _ *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	params := k.GetParams(ctx)
	return &types.QueryParamsResponse{Params: params}, nil
}

// CSRs returns all of the CSRs in the CSR module with optional pagination
func (k Keeper) CSRs(c context.Context, request *types.QueryCSRsRequest) (*types.QueryCSRsResponse, error) {
	if request == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(c)
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixCSR)

	wrappedCSRs := make([]types.WrappedCSR, 0)
	pageRes, err := query.Paginate(
		store,
		request.Pagination,
		func(key, value []byte) error {
			nft := BytesToUInt64(key)
			csr, _ := k.GetCSR(ctx, nft)
			revenue := big.NewInt(0).SetBytes(csr.Revenue)
			wrappedCSR := types.WrappedCSR{
				Csr:           *csr,
				RevenueString: revenue.String(),
			}
			wrappedCSRs = append(wrappedCSRs, wrappedCSR)
			return nil
		},
	)

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryCSRsResponse{
		Csrs:       wrappedCSRs,
		Pagination: pageRes,
	}, nil
}

// CSRByNFT returns the CSR associated with a given NFT ID passed into the request. This will return nil if the NFT does not
// match up to any CSR
func (k Keeper) CSRByNFT(c context.Context, request *types.QueryCSRByNFTRequest) (*types.QueryCSRByNFTResponse, error) {
	if request == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(c)

	csr, found := k.GetCSR(ctx, request.NftId)
	if !found {
		return nil, status.Errorf(codes.NotFound, "no csr is associated with NFT ID %d", request.NftId)
	}

	revenue := big.NewInt(0).SetBytes(csr.Revenue)
	wrappedCSR := types.WrappedCSR{
		Csr:           *csr,
		RevenueString: revenue.String(),
	}

	return &types.QueryCSRByNFTResponse{Csr: wrappedCSR}, nil
}

// CSRByContract returns the CSR associated with a given smart contracted address passed into the request. This will return nil if the smart contract
// address does not match up to any CSR
func (k Keeper) CSRByContract(c context.Context, request *types.QueryCSRByContractRequest) (*types.QueryCSRByContractResponse, error) {
	if request == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(c)

	if strings.TrimSpace(request.Address) == "" {
		return nil, status.Error(
			codes.InvalidArgument,
			"smart contract address is empty",
		)
	}

	if err := ethermint.ValidateNonZeroAddress(request.Address); err != nil {
		return nil, status.Errorf(
			codes.InvalidArgument,
			"invalid format for contract %s, should be non-zero hex ('0x...')", request.Address,
		)
	}

	nft, found := k.GetNFTByContract(ctx, request.Address)
	if !found {
		return nil, status.Errorf(codes.NotFound, "no csr contains an smart contract with address %s", request.Address)
	}
	csr, _ := k.GetCSR(ctx, nft)

	revenue := big.NewInt(0).SetBytes(csr.Revenue)
	wrappedCSR := types.WrappedCSR{
		Csr:           *csr,
		RevenueString: revenue.String(),
	}

	return &types.QueryCSRByContractResponse{Csr: wrappedCSR}, nil
}

// Turnstile returns the turnstile address that was deployed by the module account. This function does not take in any request params.
func (k Keeper) Turnstile(c context.Context, _ *types.QueryTurnstileRequest) (*types.QueryTurnstileResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	address, found := k.GetTurnstile(ctx)
	if !found {
		return nil, status.Errorf(codes.NotFound, "the turnstile address has not been found.")
	}
	return &types.QueryTurnstileResponse{Address: address.String()}, nil
}
