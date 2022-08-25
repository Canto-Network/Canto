package types

import (
	"testing"

	"github.com/evmos/ethermint/tests"
	"github.com/stretchr/testify/suite"
)

type CSRTestSuite struct {
	suite.Suite
	deployer          string
	contractAddresses []string
	nftSupply         uint64
	nfts              []*CSRNFT
	poolAddress       string
}

func TestCSRSuite(t *testing.T) {
	suite.Run(t, new(CSRTestSuite))
}

func (suite *CSRTestSuite) SetupTest() {
	// deployer is the EVM address of the EOA that deploys everything
	suite.deployer = tests.GenerateAddress().String()

	// contract addresses stores all of the EVM dapps we want to register
	suite.contractAddresses = []string{tests.GenerateAddress().String(), tests.GenerateAddress().String()}

	// NFT supply is the total circulating supply of NFTs
	suite.nftSupply = 4

	// pool address is the address of the csr smart contracted that minted the NFTs
	suite.poolAddress = tests.GenerateAddress().String()

	suite.nfts = []*CSRNFT{
		&CSRNFT{
			Period:  0,
			Id:      0,
			Address: suite.poolAddress,
		},
		&CSRNFT{
			Period:  0,
			Id:      1,
			Address: suite.poolAddress,
		},
		&CSRNFT{
			Period:  0,
			Id:      2,
			Address: suite.poolAddress,
		},
		&CSRNFT{
			Period:  0,
			Id:      3,
			Address: suite.poolAddress,
		},
	}
}

func (suite *CSRTestSuite) TestCSR() {
	testCases := []struct {
		msg        string
		csr        CSR
		expectPass bool
	}{
		{
			"Create CSR object - pass",
			CSR{
				Deployer:  suite.deployer,
				Contracts: suite.contractAddresses,
				CsrPool: &CSRPool{
					CsrNfts:     suite.nfts,
					NftSupply:   suite.nftSupply,
					PoolAddress: suite.poolAddress,
				},
			},
			true,
		},
		{
			"Create CSR object with 0 nft supply - fail",
			CSR{
				Deployer:  suite.deployer,
				Contracts: suite.contractAddresses,
				CsrPool: &CSRPool{
					CsrNfts:     suite.nfts,
					NftSupply:   0,
					PoolAddress: suite.poolAddress,
				},
			},
			false,
		},
		{
			"Create CSR object with invalid deployer address - fail",
			CSR{
				Deployer:  "",
				Contracts: suite.contractAddresses,
				CsrPool: &CSRPool{
					CsrNfts:     suite.nfts,
					NftSupply:   suite.nftSupply,
					PoolAddress: suite.poolAddress,
				},
			},
			false,
		},
		{
			"Create CSR object with invalid pool address - fail",
			CSR{
				Deployer:  suite.deployer,
				Contracts: suite.contractAddresses,
				CsrPool: &CSRPool{
					CsrNfts:     suite.nfts,
					NftSupply:   suite.nftSupply,
					PoolAddress: "",
				},
			},
			false,
		},
		{
			"Create CSR object with no smart contracts (dApps) - fail",
			CSR{
				Deployer:  suite.deployer,
				Contracts: []string{},
				CsrPool: &CSRPool{
					CsrNfts:     suite.nfts,
					NftSupply:   suite.nftSupply,
					PoolAddress: suite.poolAddress,
				},
			},
			false,
		},
		{
			"Create CSR object with mismatched nft supply - fail",
			CSR{
				Deployer:  suite.deployer,
				Contracts: suite.contractAddresses,
				CsrPool: &CSRPool{
					CsrNfts:     suite.nfts[:2],
					NftSupply:   suite.nftSupply,
					PoolAddress: suite.poolAddress,
				},
			},
			true,
		},
	}
	for _, tc := range testCases {
		err := tc.csr.Validate()

		if tc.expectPass {
			suite.Require().NoError(err, tc.msg)
		} else {
			suite.Require().Error(err, tc.msg)
		}
	}
}
