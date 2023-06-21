package keeper

import (
	"context"
	"fmt"

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

	var chunks []types.QueryChunkResponse
	pageRes, err := query.FilteredPaginate(store, req.Pagination, func(key []byte, value []byte, accumulate bool) (bool, error) {
		var chunk types.Chunk
		if err := k.cdc.Unmarshal(value, &chunk); err != nil {
			return false, err
		}

		if req.Status != 0 && chunk.Status != req.Status {
			return false, nil
		}

		if accumulate {

			chunks = append(chunks, types.QueryChunkResponse{
				Chunk:          chunk,
				DerivedAddress: chunk.DerivedAddress().String(),
			})
		}

		return true, nil
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryChunksResponse{Chunks: chunks, Pagination: pageRes}, nil
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
	return &types.QueryChunkResponse{Chunk: chunk, DerivedAddress: chunk.DerivedAddress().String()}, nil
}

func (k Keeper) Insurances(c context.Context, req *types.QueryInsurancesRequest) (*types.QueryInsurancesResponse, error) {
	if req == nil {
		return nil, status.Errorf(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(c)
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixInsurance)

	var insurances []types.QueryInsuranceResponse
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
			insurances = append(insurances, types.QueryInsuranceResponse{
				Insurance:      insurance,
				DerivedAddress: insurance.DerivedAddress().String(),
				FeePoolAddress: insurance.FeePoolAddress().String(),
			})
		}

		return true, nil
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryInsurancesResponse{Insurances: insurances, Pagination: pageRes}, nil
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
	return &types.QueryInsuranceResponse{
		Insurance:      insurance,
		DerivedAddress: insurance.DerivedAddress().String(),
		FeePoolAddress: insurance.FeePoolAddress().String(),
	}, nil
}

func (k Keeper) WithdrawInsuranceRequests(c context.Context, req *types.QueryWithdrawInsuranceRequestsRequest) (*types.QueryWithdrawInsuranceRequestsResponse, error) {
	if req == nil {
		return nil, status.Errorf(codes.InvalidArgument, "empty request")
	}
	ctx := sdk.UnwrapSDKContext(c)
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixWithdrawInsuranceRequest)

	var reqs []types.WithdrawInsuranceRequest
	pageRes, err := query.FilteredPaginate(store, req.Pagination, func(key []byte, value []byte, accumulate bool) (bool, error) {
		var withdrawInsuranceRequest types.WithdrawInsuranceRequest
		if err := k.cdc.Unmarshal(value, &withdrawInsuranceRequest); err != nil {
			return false, err
		}

		insurance, found := k.GetInsurance(ctx, withdrawInsuranceRequest.InsuranceId)
		if !found {
			return false, fmt.Errorf("no insurance is associated with Insurance Id %d", withdrawInsuranceRequest.InsuranceId)
		}

		if req.ProviderAddress != "" && insurance.ProviderAddress != req.ProviderAddress {
			return false, nil
		}

		if accumulate {
			reqs = append(reqs, withdrawInsuranceRequest)
		}
		return true, nil
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryWithdrawInsuranceRequestsResponse{reqs, pageRes}, nil
}

func (k Keeper) WithdrawInsuranceRequest(c context.Context, req *types.QueryWithdrawInsuranceRequestRequest) (*types.QueryWithdrawInsuranceRequestResponse, error) {
	if req == nil {
		return nil, status.Errorf(codes.InvalidArgument, "empty request")
	}
	ctx := sdk.UnwrapSDKContext(c)
	request, found := k.GetWithdrawInsuranceRequest(ctx, req.Id)
	if !found {
		return nil, fmt.Errorf("no withdraw insurance request is associated with Insurance Id %d", req.Id)
	}
	_, found = k.GetInsurance(ctx, request.InsuranceId)
	if !found {
		return nil, fmt.Errorf("no insurance is associated with Insurance Id %d", request.InsuranceId)
	}
	return &types.QueryWithdrawInsuranceRequestResponse{
		request,
	}, nil
}

func (k Keeper) UnpairingForUnstakingChunkInfos(c context.Context, req *types.QueryUnpairingForUnstakingChunkInfosRequest) (*types.QueryUnpairingForUnstakingChunkInfosResponse, error) {
	if req == nil {
		return nil, status.Errorf(codes.InvalidArgument, "empty request")
	}
	ctx := sdk.UnwrapSDKContext(c)
	var infos []types.UnpairingForUnstakingChunkInfo
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixUnpairingForUnstakingChunkInfo)

	pageRes, err := query.FilteredPaginate(store, req.Pagination, func(key []byte, value []byte, accumulate bool) (bool, error) {
		var info types.UnpairingForUnstakingChunkInfo
		if err := k.cdc.Unmarshal(value, &info); err != nil {
			return false, err
		}

		chunk, found := k.GetChunk(ctx, info.ChunkId)
		if !found {
			return false, fmt.Errorf("no chunk is associated with Chunk Id %d", info.ChunkId)
		}
		if req.Queued {
			// Only return queued(=not yet started) liquid unstake.
			if chunk.Status == types.CHUNK_STATUS_UNPAIRING_FOR_UNSTAKING {
				return false, nil
			}
		}

		if req.DelegatorAddress != "" && req.DelegatorAddress != info.DelegatorAddress {
			return false, nil
		}

		if accumulate {
			infos = append(infos, info)
		}
		return true, nil
	})

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryUnpairingForUnstakingChunkInfosResponse{infos, pageRes}, nil
}

func (k Keeper) UnpairingForUnstakingChunkInfo(c context.Context, req *types.QueryUnpairingForUnstakingChunkInfoRequest) (*types.QueryUnpairingForUnstakingChunkInfoResponse, error) {
	if req == nil {
		return nil, status.Errorf(codes.InvalidArgument, "empty request")
	}
	ctx := sdk.UnwrapSDKContext(c)
	info, found := k.GetUnpairingForUnstakingChunkInfo(ctx, req.Id)
	if !found {
		return nil, fmt.Errorf("no unpairing for unstaking chunk info is associated with Id %d", req.Id)
	}
	return &types.QueryUnpairingForUnstakingChunkInfoResponse{
		info,
	}, nil
}

func (k Keeper) ChunkSize(c context.Context, req *types.QueryChunkSizeRequest) (*types.QueryChunkSizeResponse, error) {
	if req == nil {
		return nil, status.Errorf(codes.InvalidArgument, "empty request")
	}
	ctx := sdk.UnwrapSDKContext(c)
	return &types.QueryChunkSizeResponse{
		ChunkSize: sdk.NewCoin(
			k.stakingKeeper.BondDenom(ctx),
			types.ChunkSize,
		),
	}, nil
}

func (k Keeper) MinimumCollateral(c context.Context, req *types.QueryMinimumCollateralRequest) (*types.QueryMinimumCollateralResponse, error) {
	if req == nil {
		return nil, status.Errorf(codes.InvalidArgument, "empty request")
	}
	ctx := sdk.UnwrapSDKContext(c)
	minimumCollateral, err := sdk.NewDecFromStr(types.MinimumCollateral)
	if err != nil {
		return nil, err
	}
	return &types.QueryMinimumCollateralResponse{
		MinimumCollateral: sdk.NewDecCoinFromDec(
			k.stakingKeeper.BondDenom(ctx),
			types.ChunkSize.ToDec().Mul(minimumCollateral),
		),
	}, nil
}

func (k Keeper) States(c context.Context, req *types.QueryStatesRequest) (*types.QueryStatesResponse, error) {
	if req == nil {
		return nil, status.Errorf(codes.InvalidArgument, "empty request")
	}
	ctx := sdk.UnwrapSDKContext(c)
	return &types.QueryStatesResponse{NetAmountState: k.GetNetAmountState(ctx)}, nil
}
