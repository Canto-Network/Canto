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

	var chunkResponses []types.ResponseChunk
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
			chunkResponses = append(chunkResponses, chunkToResponseChunk(ctx, k, chunk))
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
	return &types.QueryChunkResponse{Chunk: chunkToResponseChunk(ctx, k, chunk)}, nil
}

func (k Keeper) Insurances(c context.Context, req *types.QueryInsurancesRequest) (*types.QueryInsurancesResponse, error) {
	if req == nil {
		return nil, status.Errorf(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(c)
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixInsurance)

	var insuranceResponses []types.ResponseInsurance
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
			insuranceResponses = append(insuranceResponses, insuranceToResponseInsurance(ctx, k, insurance))
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
	return &types.QueryInsuranceResponse{Insurance: insuranceToResponseInsurance(ctx, k, insurance)}, nil
}

func (k Keeper) States(c context.Context, _ *types.QueryStatesRequest) (*types.QueryStatesResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	return &types.QueryStatesResponse{NetAmountState: k.GetNetAmountState(ctx)}, nil
}

func (k Keeper) WithdrawInsuranceRequests(c context.Context, req *types.QueryWithdrawInsuranceRequestsRequest) (*types.QueryWithdrawInsuranceRequestsResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	var reqs []types.ResponseWithdrawInsuranceRequest
	k.IterateAllWithdrawInsuranceRequests(ctx, func(request types.WithdrawInsuranceRequest) (bool, error) {
		insurance, found := k.GetInsurance(ctx, request.InsuranceId)
		if !found {
			return false, fmt.Errorf("no insurance is associated with Insurance Id %d", request.InsuranceId)
		}
		if req.ProviderAddress != "" {
			if insurance.ProviderAddress != req.ProviderAddress {
				return false, nil
			}
		}
		reqs = append(reqs, types.ResponseWithdrawInsuranceRequest{
			Insurance: insuranceToResponseInsurance(ctx, k, insurance),
		})
		return false, nil

	})
	return &types.QueryWithdrawInsuranceRequestsResponse{Requests: reqs}, nil
}

func (k Keeper) WithdrawInsuranceRequest(c context.Context, req *types.QueryWithdrawInsuranceRequestRequest) (*types.QueryWithdrawInsuranceRequestResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	request, found := k.GetWithdrawInsuranceRequest(ctx, req.Id)
	if !found {
		return nil, fmt.Errorf("no withdraw insurance request is associated with Insurance Id %d", req.Id)
	}
	insurance, found := k.GetInsurance(ctx, request.InsuranceId)
	if !found {
		return nil, fmt.Errorf("no insurance is associated with Insurance Id %d", request.InsuranceId)
	}
	return &types.QueryWithdrawInsuranceRequestResponse{
		Request: types.ResponseWithdrawInsuranceRequest{
			Insurance: insuranceToResponseInsurance(ctx, k, insurance),
		},
	}, nil
}

func (k Keeper) UnpairingForUnstakingChunkInfos(c context.Context, req *types.QueryUnpairingForUnstakingChunkInfosRequest) (*types.QueryUnpairingForUnstakingChunkInfosResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	var infos []types.ResponseUnpairingForUnstakingChunkInfo
	k.IterateAllUnpairingForUnstakingChunkInfos(ctx, func(info types.UnpairingForUnstakingChunkInfo) (bool, error) {
		chunk, found := k.GetChunk(ctx, info.ChunkId)
		if !found {
			return false, fmt.Errorf("no chunk is associated with Chunk Id %d", info.ChunkId)
		}
		// TODO: Optional field but it handled like required, check other queries also
		if req.DelegatorAddress != info.DelegatorAddress {
			return false, nil
		}
		infos = append(infos, types.ResponseUnpairingForUnstakingChunkInfo{
			Chunk: chunkToResponseChunk(ctx, k, chunk),
		})
		return false, nil

	})
	return &types.QueryUnpairingForUnstakingChunkInfosResponse{Infos: infos}, nil
}

func (k Keeper) UnpairingForUnstakingChunkInfo(c context.Context, req *types.QueryUnpairingForUnstakingChunkInfoRequest) (*types.QueryUnpairingForUnstakingChunkInfoResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	info, found := k.GetUnpairingForUnstakingChunkInfo(ctx, req.Id)
	if !found {
		return nil, fmt.Errorf("no unpairing for unstaking chunk info is associated with Id %d", req.Id)
	}
	chunk, found := k.GetChunk(ctx, info.ChunkId)
	if !found {
		return nil, fmt.Errorf("no chunk is associated with Chunk Id %d", info.ChunkId)
	}
	return &types.QueryUnpairingForUnstakingChunkInfoResponse{
		Info: types.ResponseUnpairingForUnstakingChunkInfo{
			Chunk: chunkToResponseChunk(ctx, k, chunk),
		},
	}, nil
}

func (k Keeper) MaxPairedChunks(_ context.Context, _ *types.QueryMaxPairedChunksRequest) (*types.QueryMaxPairedChunksResponse, error) {
	return &types.QueryMaxPairedChunksResponse{MaxPairedChunks: types.MaxPairedChunks}, nil
}

func (k Keeper) ChunkSize(_ context.Context, _ *types.QueryChunkSizeRequest) (*types.QueryChunkSizeResponse, error) {
	return &types.QueryChunkSizeResponse{ChunkSize: types.ChunkSize.Uint64()}, nil
}

func chunkToResponseChunk(ctx sdk.Context, k Keeper, chunk types.Chunk) types.ResponseChunk {
	pairedInsurance, _ := k.GetInsurance(ctx, chunk.PairedInsuranceId)
	unpairingInsurance, _ := k.GetInsurance(ctx, chunk.UnpairingInsuranceId)
	// TODO: Add validation for nil insurances
	// TODO: Handle chunks which have no delegation obj
	del, _ := k.stakingKeeper.GetDelegation(ctx, chunk.DerivedAddress(), sdk.ValAddress(pairedInsurance.ValidatorAddress))
	val, _ := k.stakingKeeper.GetValidator(ctx, sdk.ValAddress(pairedInsurance.ValidatorAddress))

	return types.ResponseChunk{
		Id: chunk.Id,
		// TODO: Need following fields?
		Tokens: val.TokensFromShares(del.Shares),
		Shares: del.Shares,
		// TODO: Meaningless field and it is temporary value because reward goes to module account, so need to re-name it or remove it
		// TODO: or Balance + Unclaimed reward (delegation reward)
		// TODO: It will be ok to use native state to service
		AccumulatedReward:  k.bankKeeper.GetBalance(ctx, chunk.DerivedAddress(), k.stakingKeeper.BondDenom(ctx)),
		PairedInsurance:    pairedInsurance,
		UnpairingInsurance: unpairingInsurance,
		Status:             chunk.Status,
	}
}

func insuranceToResponseInsurance(ctx sdk.Context, k Keeper, insurance types.Insurance) types.ResponseInsurance {
	chunk, _ := k.GetChunk(ctx, insurance.ChunkId)
	return types.ResponseInsurance{
		Id:               insurance.Id,
		ValidatorAddress: insurance.ValidatorAddress,
		ProviderAddress:  insurance.ProviderAddress,
		Amount:           k.bankKeeper.GetBalance(ctx, insurance.DerivedAddress(), k.stakingKeeper.BondDenom(ctx)),
		FeeRate:          insurance.FeeRate,
		Chunk:            chunk,
		Status:           insurance.Status,
	}
}
