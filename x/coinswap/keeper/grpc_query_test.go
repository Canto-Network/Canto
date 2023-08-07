package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Canto-Network/Canto/v7/x/coinswap/types"
)

func (s *TestSuite) TestGRPCParams() {
	resp, err := s.queryClient.Params(sdk.WrapSDKContext(s.ctx), &types.QueryParamsRequest{})
	s.Require().NoError(err)
	s.Require().Equal(s.keeper.GetParams(s.ctx), resp.Params)
}

func (s *TestSuite) TestGRPCPool() {
	_, _ = createReservePool(s, denomBTC)
	poolId := types.GetPoolId(denomBTC)
	pool, _ := s.app.CoinswapKeeper.GetPool(s.ctx, poolId)

	resp, err := s.queryClient.LiquidityPool(sdk.WrapSDKContext(s.ctx), &types.QueryLiquidityPoolRequest{LptDenom: pool.LptDenom})
	s.Require().NoError(err)
	s.Require().Equal(pool.Id, resp.Pool.Id)
	s.Require().Equal(pool.EscrowAddress, resp.Pool.EscrowAddress)

	balances, _ := s.app.CoinswapKeeper.GetPoolBalancesByLptDenom(s.ctx, pool.LptDenom)
	liquidity := s.app.BankKeeper.GetSupply(s.ctx, pool.LptDenom)
	params := s.app.CoinswapKeeper.GetParams(s.ctx)
	actCoins := sdk.NewCoins(
		resp.Pool.Standard,
		resp.Pool.Token,
	)
	s.Equal(actCoins.Sort().String(), balances.Sort().String())
	s.Equal(liquidity, resp.Pool.Lpt)
	s.Equal(params.Fee.String(), resp.Pool.Fee)
}

func (s *TestSuite) TestGRPCPools() {
	_, _ = createReservePool(s, denomBTC)
	resp, err := s.queryClient.LiquidityPools(sdk.WrapSDKContext(s.ctx), &types.QueryLiquidityPoolsRequest{})
	s.Require().NoError(err)
	s.Require().Len(resp.Pools, 1)

	_, _ = createReservePool(s, denomETH)
	resp, err = s.queryClient.LiquidityPools(sdk.WrapSDKContext(s.ctx), &types.QueryLiquidityPoolsRequest{})
	s.Require().NoError(err)
	s.Require().Len(resp.Pools, 2)
}
