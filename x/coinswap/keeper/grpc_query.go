package keeper

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/store/prefix"
	"github.com/cosmos/cosmos-sdk/runtime"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"

	"github.com/Canto-Network/Canto/v7/x/coinswap/types"
)

var _ types.QueryServer = Keeper{}

// Params queries the parameters of the liquidity module.
func (k Keeper) Params(c context.Context, _ *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	var params types.Params
	k.paramSpace.GetParamSet(ctx, &params)
	return &types.QueryParamsResponse{Params: params}, nil
}

// LiquidityPool returns the liquidity pool information of the denom
func (k Keeper) LiquidityPool(c context.Context, req *types.QueryLiquidityPoolRequest) (*types.QueryLiquidityPoolResponse, error) {
	if req == nil {
		return nil, status.Errorf(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(c)
	pool, exists := k.GetPoolByLptDenom(ctx, req.LptDenom)
	if !exists {
		return nil, errorsmod.Wrapf(types.ErrReservePoolNotExists, "liquidity pool token: %s", req.LptDenom)
	}

	balances, err := k.GetPoolBalancesByLptDenom(ctx, pool.LptDenom)
	if err != nil {
		return nil, err
	}

	standard := sdk.NewCoin(pool.StandardDenom, balances.AmountOf(pool.StandardDenom))
	token := sdk.NewCoin(pool.CounterpartyDenom, balances.AmountOf(pool.CounterpartyDenom))
	liquidity := k.bk.GetSupply(ctx, pool.LptDenom)

	params := k.GetParams(ctx)
	res := types.QueryLiquidityPoolResponse{
		Pool: types.PoolInfo{
			Id:            pool.Id,
			EscrowAddress: pool.EscrowAddress,
			Standard:      standard,
			Token:         token,
			Lpt:           liquidity,
			Fee:           params.Fee.String(),
		},
	}
	return &res, nil
}

func (k Keeper) LiquidityPools(c context.Context, req *types.QueryLiquidityPoolsRequest) (*types.QueryLiquidityPoolsResponse, error) {
	if req == nil {
		return nil, status.Errorf(codes.InvalidArgument, "empty request")
	}
	ctx := sdk.UnwrapSDKContext(c)
	params := k.GetParams(ctx)

	var pools []types.PoolInfo

	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	nftStore := prefix.NewStore(store, []byte(types.KeyPool))
	pageRes, err := query.Paginate(nftStore, req.Pagination, func(key []byte, value []byte) error {
		var pool types.Pool
		k.cdc.MustUnmarshal(value, &pool)

		balances, err := k.GetPoolBalancesByLptDenom(ctx, pool.LptDenom)
		if err != nil {
			return err
		}

		pools = append(pools, types.PoolInfo{
			Id:            pool.Id,
			EscrowAddress: pool.EscrowAddress,
			Standard:      sdk.NewCoin(pool.StandardDenom, balances.AmountOf(pool.StandardDenom)),
			Token:         sdk.NewCoin(pool.CounterpartyDenom, balances.AmountOf(pool.CounterpartyDenom)),
			Lpt:           k.bk.GetSupply(ctx, pool.LptDenom),
			Fee:           params.Fee.String(),
		})
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &types.QueryLiquidityPoolsResponse{
		Pagination: pageRes,
		Pools:      pools,
	}, nil
}
