package keeper_test

import (
	"github.com/Canto-Network/Canto/v7/x/csr/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/evmos/ethermint/tests"
)

// Special edge case for when the request made is null, cannot be routed to the query client
func (suite *KeeperTestSuite) TestKeeperCSRs() {
	suite.SetupTest()
	ctx := sdk.WrapSDKContext(suite.ctx)
	res, err := suite.app.CSRKeeper.CSRs(ctx, nil)
	suite.Require().Error(err)
	suite.Require().Nil(res)
}

// TestCSRs generates CSRs and makes requests to fetch
func (suite *KeeperTestSuite) TestCSRs() {
	var (
		request          *types.QueryCSRsRequest
		expectedResponse *types.QueryCSRsResponse
	)

	testCases := []struct {
		name    string
		prepare func()
		pass    bool
	}{
		{
			"no csrs registered",
			func() {
				request = &types.QueryCSRsRequest{}
				expectedResponse = &types.QueryCSRsResponse{Pagination: &query.PageResponse{}}
			},
			true,
		},
		{
			"10 csr registered w/o pagination, request 1",
			func() {
				csrs := GenerateCSRs(10)
				for _, csr := range csrs {
					suite.app.CSRKeeper.SetCSR(suite.ctx, csr)
				}

				request = &types.QueryCSRsRequest{
					Pagination: &query.PageRequest{Limit: 1, CountTotal: true},
				}
				expectedResponse = &types.QueryCSRsResponse{
					Pagination: &query.PageResponse{Total: 10},
					Csrs:       csrs[:1],
				}
			},
			true,
		},
		{
			"10 csr registered w/o pagination, request 5",
			func() {
				csrs := GenerateCSRs(10)
				for _, csr := range csrs {
					suite.app.CSRKeeper.SetCSR(suite.ctx, csr)
				}

				request = &types.QueryCSRsRequest{
					Pagination: &query.PageRequest{Limit: 5, CountTotal: true},
				}
				expectedResponse = &types.QueryCSRsResponse{
					Pagination: &query.PageResponse{Total: 10},
					Csrs:       csrs[:5],
				}
			},
			true,
		},
		{
			"10 csr registered w/o pagination, request 10",
			func() {
				csrs := GenerateCSRs(10)
				for _, csr := range csrs {
					suite.app.CSRKeeper.SetCSR(suite.ctx, csr)
				}

				request = &types.QueryCSRsRequest{
					Pagination: &query.PageRequest{Limit: 10, CountTotal: true},
				}
				expectedResponse = &types.QueryCSRsResponse{
					Pagination: &query.PageResponse{Total: 10},
					Csrs:       csrs,
				}
			},
			true,
		},
		{
			"10 csr registered w/o pagination, request 30",
			func() {
				csrs := GenerateCSRs(10)
				for _, csr := range csrs {
					suite.app.CSRKeeper.SetCSR(suite.ctx, csr)
				}

				request = &types.QueryCSRsRequest{
					Pagination: &query.PageRequest{Limit: 30, CountTotal: true},
				}
				expectedResponse = &types.QueryCSRsResponse{
					Pagination: &query.PageResponse{Total: 10},
					Csrs:       csrs,
				}
			},
			true,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			suite.SetupTest()

			ctx := sdk.WrapSDKContext(suite.ctx)
			tc.prepare()

			response, err := suite.queryClient.CSRs(ctx, request)
			if tc.pass {
				suite.Require().NoError(err)
				suite.Require().Equal(expectedResponse.Pagination.Total, response.Pagination.Total)
				suite.Require().ElementsMatch(expectedResponse.Csrs, response.Csrs)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

// Special edge case for when the request made is null, cannot be routed to the query client
func (suite *KeeperTestSuite) TestKeeperCSRsByNFT() {
	suite.SetupTest()
	ctx := sdk.WrapSDKContext(suite.ctx)
	res, err := suite.app.CSRKeeper.CSRByNFT(ctx, nil)
	suite.Require().Error(err)
	suite.Require().Nil(res)
}

// TestCSRs generates CSRs and makes requests to fetch
func (suite *KeeperTestSuite) TestCSRByNFT() {
	var (
		request          *types.QueryCSRByNFTRequest
		expectedResponse *types.QueryCSRByNFTResponse
	)

	testCases := []struct {
		name    string
		prepare func()
		pass    bool
	}{
		{
			"correctly extracting CSR by NFT",
			func() {
				csrs := GenerateCSRs(10)
				for _, csr := range csrs {
					suite.app.CSRKeeper.SetCSR(suite.ctx, csr)
				}

				request = &types.QueryCSRByNFTRequest{NftId: csrs[0].Id}
				expectedResponse = &types.QueryCSRByNFTResponse{Csr: csrs[0]}
			},
			true,
		},
		{
			"invalid request with non-existing NFT",
			func() {
				csrs := GenerateCSRs(10)
				for _, csr := range csrs {
					suite.app.CSRKeeper.SetCSR(suite.ctx, csr)
				}

				request = &types.QueryCSRByNFTRequest{NftId: 34}
				expectedResponse = nil
			},
			false,
		},
		{
			"invalid request with non-existing NFT -> no csrs exist",
			func() {
				request = &types.QueryCSRByNFTRequest{}
				expectedResponse = nil
			},
			false,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			suite.SetupTest()

			ctx := sdk.WrapSDKContext(suite.ctx)
			tc.prepare()

			response, err := suite.queryClient.CSRByNFT(ctx, request)
			if tc.pass {
				suite.Require().NoError(err)
			} else {
				suite.Require().Error(err)
			}
			suite.Require().Equal(expectedResponse, response)
		})
	}
}

// Special edge case for when the request made is null, cannot be routed to the query client
func (suite *KeeperTestSuite) TestKeeperCSRsByContract() {
	suite.SetupTest()
	ctx := sdk.WrapSDKContext(suite.ctx)
	res, err := suite.app.CSRKeeper.CSRByContract(ctx, nil)
	suite.Require().Error(err)
	suite.Require().Nil(res)
}

// TestCSRs generates CSRs and makes requests to fetch
func (suite *KeeperTestSuite) TestCSRByContract() {
	var (
		request          *types.QueryCSRByContractRequest
		expectedResponse *types.QueryCSRByContractResponse
	)

	testCases := []struct {
		name    string
		prepare func()
		pass    bool
	}{
		{
			"correctly extracting CSR by address",
			func() {
				csrs := GenerateCSRs(10)
				for _, csr := range csrs {
					suite.app.CSRKeeper.SetCSR(suite.ctx, csr)
				}

				request = &types.QueryCSRByContractRequest{Address: csrs[0].Contracts[0]}
				expectedResponse = &types.QueryCSRByContractResponse{Csr: csrs[0]}
			},
			true,
		},
		{
			"invalid request empty address",
			func() {
				csrs := GenerateCSRs(10)
				for _, csr := range csrs {
					suite.app.CSRKeeper.SetCSR(suite.ctx, csr)
				}

				request = &types.QueryCSRByContractRequest{Address: ""}
				expectedResponse = nil
			},
			false,
		},
		{
			"invalid request poorly formatted address",
			func() {
				csrs := GenerateCSRs(10)
				for _, csr := range csrs {
					suite.app.CSRKeeper.SetCSR(suite.ctx, csr)
				}

				request = &types.QueryCSRByContractRequest{Address: "920janfoija90we90jfa"}
				expectedResponse = nil
			},
			false,
		},
		{
			"valid request with non csr enabled smart contract address",
			func() {
				csrs := GenerateCSRs(10)
				for _, csr := range csrs {
					suite.app.CSRKeeper.SetCSR(suite.ctx, csr)
				}

				address := tests.GenerateAddress().Hex()
				request = &types.QueryCSRByContractRequest{Address: address}
				expectedResponse = nil
			},
			false,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			suite.SetupTest()

			ctx := sdk.WrapSDKContext(suite.ctx)
			tc.prepare()

			response, err := suite.queryClient.CSRByContract(ctx, request)
			if tc.pass {
				suite.Require().NoError(err)
			} else {
				suite.Require().Error(err)
			}
			suite.Require().Equal(expectedResponse, response)
		})
	}
}

// Test the query service for params
func (suite *KeeperTestSuite) TestQueryParams() {
	ctx := sdk.WrapSDKContext(suite.ctx)
	expectedParams := types.DefaultParams()
	expectedParams.EnableCsr = true

	res, err := suite.queryClient.Params(ctx, &types.QueryParamsRequest{})
	suite.Require().NoError(err)
	suite.Require().Equal(expectedParams, res.Params)
}

// Test the query service that retrieves the turnstile address
func (suite *KeeperTestSuite) TestQueryTurnstile() {
	suite.Commit()
	ctx := sdk.WrapSDKContext(suite.ctx)

	address, found := suite.app.CSRKeeper.GetTurnstile(suite.ctx)
	suite.Require().True(found)

	res, err := suite.queryClient.Turnstile(ctx, &types.QueryTurnstileRequest{})
	suite.Require().NoError(err)
	suite.Require().Equal(address.String(), res.Address)
}
