package keeper

import (
	"context"
	"github.com/Canto-Network/Canto/v6/x/liquidstaking/types"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _ types.QueryServer = Keeper{}

func (k Keeper) Params(c context.Context, _ *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	params := k.GetParams(ctx)
	return &types.QueryParamsResponse{Params: params}, nil
}

func (k Keeper) Epoch(c context.Context, _ *types.QueryEpochRequest) (*types.QueryEpochResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	epoch := k.GetEpoch(ctx)
	return &types.QueryEpochResponse{Epoch: epoch}, nil
}

func (k Keeper) Chunks(c context.Context, req *types.QueryChunksRequest) (*types.QueryChunksResponse, error) {
	if req == nil {
		return nil, status.Errorf(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(c)
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixChunk)

	var chunkResponses []types.ChunkResponse
	pageRes, err := query.FilteredPaginate(store, req.Pagination, func(key []byte, value []byte, accumulate bool) (bool, error) {
		var chunk types.Chunk
		if err := k.cdc.Unmarshal(value, &chunk); err != nil {
			return false, err
		}

		if req.Status != 0 && chunk.Status != req.Status {
			return false, nil
		}

		if accumulate {
			// for all chunks, get the insurance and convert to chunk response
			chunkResponses = append(chunkResponses, chunkToChunkResponse(ctx, k, chunk))
		}

		return true, nil
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryChunksResponse{Chunks: chunkResponses, Pagination: pageRes}, nil
}

func (k Keeper) Chunk(c context.Context, req *types.QueryChunkRequest) (*types.QueryChunkResponse, error) {
	if req == nil {
		return nil, status.Errorf(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(c)
	chunk, found := k.GetChunk(ctx, req.Id)
	if !found {
		return nil, status.Errorf(codes.NotFound, "no chunk is associated with Chunk Id %d", req.Id)
	}
	return &types.QueryChunkResponse{Chunk: chunkToChunkResponse(ctx, k, chunk)}, nil
}

func (k Keeper) Insurances(c context.Context, req *types.QueryInsurancesRequest) (*types.QueryInsurancesResponse, error) {
	if req == nil {
		return nil, status.Errorf(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(c)
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixInsurance)

	var insuranceResponses []types.InsuranceResponse
	pageRes, err := query.FilteredPaginate(store, req.Pagination, func(key []byte, value []byte, accumulate bool) (bool, error) {
		var insurance types.Insurance
		if err := k.cdc.Unmarshal(value, &insurance); err != nil {
			return false, err
		}

		if req.Status != 0 && insurance.Status != req.Status {
			return false, nil
		}

		if req.ValidatorAddress != "" && insurance.ValidatorAddress != req.ValidatorAddress {
			return false, nil
		}

		if req.ProviderAddress != "" && insurance.ProviderAddress != req.ProviderAddress {
			return false, nil
		}

		if accumulate {
			// for all insurances, get the chunks and convert to insurance response
			insuranceResponses = append(insuranceResponses, insuranceToInsuranceResponse(ctx, k, insurance))
		}

		return true, nil
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryInsurancesResponse{Insurances: insuranceResponses, Pagination: pageRes}, nil
}

func (k Keeper) Insurance(c context.Context, req *types.QueryInsuranceRequest) (*types.QueryInsuranceResponse, error) {
	if req == nil {
		return nil, status.Errorf(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(c)
	insurance, found := k.GetInsurance(ctx, req.Id)
	if !found {
		return nil, status.Errorf(codes.NotFound, "no insurance is associated with Insurance Id %d", req.Id)
	}
	return &types.QueryInsuranceResponse{Insurance: insuranceToInsuranceResponse(ctx, k, insurance)}, nil
}

func (k Keeper) States(c context.Context, _ *types.QueryStatesRequest) (*types.QueryStatesResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	return &types.QueryStatesResponse{NetAmountState: k.GetNetAmountState(ctx)}, nil
}

func chunkToChunkResponse(ctx sdk.Context, k Keeper, chunk types.Chunk) types.ChunkResponse {
	insurance, _ := k.GetInsurance(ctx, chunk.InsuranceId)
	del, _ := k.stakingKeeper.GetDelegation(ctx, chunk.DerivedAddress(), sdk.ValAddress(insurance.ValidatorAddress))
	val, _ := k.stakingKeeper.GetValidator(ctx, sdk.ValAddress(insurance.ValidatorAddress))

	return types.ChunkResponse{
		Id:                chunk.Id,
		Tokens:            val.TokensFromShares(del.Shares),
		Shares:            del.Shares,
		AccumulatedReward: k.bankKeeper.GetBalance(ctx, chunk.DerivedAddress(), k.stakingKeeper.BondDenom(ctx)),
		Insurance:         insurance,
		Status:            chunk.Status,
	}
}

func insuranceToInsuranceResponse(ctx sdk.Context, k Keeper, insurance types.Insurance) types.InsuranceResponse {
	chunk, _ := k.GetChunk(ctx, insurance.ChunkId)
	return types.InsuranceResponse{
		Id:               insurance.Id,
		ValidatorAddress: insurance.ValidatorAddress,
		ProviderAddress:  insurance.ProviderAddress,
		Amount:           k.bankKeeper.GetBalance(ctx, insurance.DerivedAddress(), k.stakingKeeper.BondDenom(ctx)),
		FeeRate:          insurance.FeeRate,
		Chunk:            chunk,
		Status:           insurance.Status,
	}
}
