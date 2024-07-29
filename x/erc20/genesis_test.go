package erc20_test

import (
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/cometbft/cometbft/crypto/tmhash"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	tmversion "github.com/cometbft/cometbft/proto/tendermint/version"
	"github.com/cometbft/cometbft/version"

	"github.com/evmos/ethermint/tests"
	feemarkettypes "github.com/evmos/ethermint/x/feemarket/types"

	"github.com/Canto-Network/Canto/v8/app"
	"github.com/Canto-Network/Canto/v8/x/erc20"
	"github.com/Canto-Network/Canto/v8/x/erc20/types"
)

var (
	uqstars = "ibc/13B6057538B93225F6EBACCB64574C49B2C1568C5AE6CCFE0A039D7DAC02BF29"
	// uqstars1 and uqstars2 have same denom
	// uqstars1 is deployed first.
	uqstars1 = types.TokenPair{
		Erc20Address:  "0x2C68D1d6aB986Ff4640b51e1F14C716a076E44C4",
		Denom:         uqstars,
		Enabled:       true,
		ContractOwner: types.OWNER_MODULE,
	}
	// uqstars2 is deployed later than uqstars1.
	uqstars2 = types.TokenPair{
		Erc20Address:  "0xD32eB974468ed767338533842D2D4Cc90B9BAb46",
		Denom:         uqstars,
		Enabled:       true,
		ContractOwner: types.OWNER_MODULE,
	}
	customERC20 = types.TokenPair{
		Erc20Address:  "0xC5e00D3b04563950941f7137B5AfA3a534F0D6d6",
		Denom:         "custom",
		Enabled:       true,
		ContractOwner: types.OWNER_EXTERNAL,
	}

	tokenPairs = []types.TokenPair{
		uqstars2,
		// even if we put uqstars1 later, it should be disabled because
		// uqstars2 is the deployed later than uqstars1
		uqstars1,
		customERC20,
	}
	denomIdxs = []types.TokenPairDenomIndex{
		{
			Denom:       customERC20.Denom,
			TokenPairId: customERC20.GetID(),
		},
		{
			Denom: uqstars,
			// denomIdx must have the latest token pair id assigned
			// if there are multiple token pairs with the same denom
			TokenPairId: uqstars2.GetID(),
		},
	}
	erc20AddrIdxs = []types.TokenPairERC20AddressIndex{
		{
			Erc20Address: common.HexToAddress(uqstars1.Erc20Address).Bytes(),
			TokenPairId:  uqstars2.GetID(),
		},
		{
			Erc20Address: customERC20.GetERC20Contract().Bytes(),
			TokenPairId:  customERC20.GetID(),
		},

		{
			Erc20Address: common.HexToAddress(uqstars2.Erc20Address).Bytes(),
			TokenPairId:  uqstars2.GetID(),
		},
	}
)

type GenesisTestSuite struct {
	suite.Suite
	ctx     sdk.Context
	app     *app.Canto
	genesis types.GenesisState
}

func TestGenesisTestSuite(t *testing.T) {
	suite.Run(t, new(GenesisTestSuite))
}

func (suite *GenesisTestSuite) SetupTest() {
	// consensus key
	consAddress := sdk.ConsAddress(tests.GenerateAddress().Bytes())

	suite.app = app.Setup(false, feemarkettypes.DefaultGenesisState())
	suite.ctx = suite.app.BaseApp.NewContextLegacy(false, tmproto.Header{
		Height:          1,
		ChainID:         "canto_9000-1",
		Time:            time.Now().UTC(),
		ProposerAddress: consAddress.Bytes(),

		Version: tmversion.Consensus{
			Block: version.BlockProtocol,
		},
		LastBlockId: tmproto.BlockID{
			Hash: tmhash.Sum([]byte("block_id")),
			PartSetHeader: tmproto.PartSetHeader{
				Total: 11,
				Hash:  tmhash.Sum([]byte("partset_header")),
			},
		},
		AppHash:            tmhash.Sum([]byte("app")),
		DataHash:           tmhash.Sum([]byte("data")),
		EvidenceHash:       tmhash.Sum([]byte("evidence")),
		ValidatorsHash:     tmhash.Sum([]byte("validators")),
		NextValidatorsHash: tmhash.Sum([]byte("next_validators")),
		ConsensusHash:      tmhash.Sum([]byte("consensus")),
		LastResultsHash:    tmhash.Sum([]byte("last_result")),
	})

	suite.genesis = *types.DefaultGenesisState()
}

func (suite *GenesisTestSuite) TestERC20InitGenesis() {
	testCases := []struct {
		name         string
		genesisState types.GenesisState
	}{
		{
			"empty genesis",
			types.GenesisState{},
		},
		{
			"default genesis",
			*types.DefaultGenesisState(),
		},
		{
			"custom genesis",
			types.NewGenesisState(
				types.DefaultParams(),
				tokenPairs,
				denomIdxs,
				erc20AddrIdxs,
			),
		},
	}

	for _, tc := range testCases {
		suite.Require().NotPanics(func() {
			erc20.InitGenesis(suite.ctx, suite.app.Erc20Keeper, suite.app.AccountKeeper, tc.genesisState)
		})
		params := suite.app.Erc20Keeper.GetParams(suite.ctx)
		suite.Require().Equal(tc.genesisState.Params, params)

		tokenPairs := suite.app.Erc20Keeper.GetTokenPairs(suite.ctx)
		if len(tokenPairs) > 0 {
			suite.Require().Equal(tc.genesisState.TokenPairs, tokenPairs)
			suite.Equal(denomIdxs, suite.app.Erc20Keeper.GetAllTokenPairDenomIndexes(suite.ctx))
			suite.Equal(erc20AddrIdxs, suite.app.Erc20Keeper.GetAllTokenPairERC20AddressIndexes(suite.ctx))
			suite.Equal(
				uqstars2.GetID(), suite.app.Erc20Keeper.GetTokenPairIdByDenom(suite.ctx, uqstars),
				"denom index must have latest token pair id",
			)
		} else {
			suite.Len(tc.genesisState.TokenPairs, 0)
			suite.Len(suite.app.Erc20Keeper.GetAllTokenPairDenomIndexes(suite.ctx), 0)
			suite.Len(suite.app.Erc20Keeper.GetAllTokenPairERC20AddressIndexes(suite.ctx), 0)
		}
	}
}

func (suite *GenesisTestSuite) TestErc20ExportGenesis() {
	testGenCases := []struct {
		name         string
		genesisState types.GenesisState
	}{
		{
			"empty genesis",
			types.GenesisState{},
		},
		{
			"default genesis",
			*types.DefaultGenesisState(),
		},
		{
			"custom genesis",
			types.NewGenesisState(
				types.DefaultParams(),
				tokenPairs,
				denomIdxs,
				erc20AddrIdxs,
			),
		},
	}

	for _, tc := range testGenCases {
		erc20.InitGenesis(suite.ctx, suite.app.Erc20Keeper, suite.app.AccountKeeper, tc.genesisState)
		suite.Require().NotPanics(func() {
			genesisExported := erc20.ExportGenesis(suite.ctx, suite.app.Erc20Keeper)
			params := suite.app.Erc20Keeper.GetParams(suite.ctx)
			suite.Require().Equal(genesisExported.Params, params)

			tokenPairs := suite.app.Erc20Keeper.GetTokenPairs(suite.ctx)
			if len(tokenPairs) > 0 {
				suite.Require().Equal(tc.genesisState.TokenPairs, tokenPairs)
				suite.Equal(denomIdxs, suite.app.Erc20Keeper.GetAllTokenPairDenomIndexes(suite.ctx))
				suite.Equal(erc20AddrIdxs, suite.app.Erc20Keeper.GetAllTokenPairERC20AddressIndexes(suite.ctx))
				suite.Equal(
					uqstars2.GetID(), suite.app.Erc20Keeper.GetTokenPairIdByDenom(suite.ctx, uqstars),
					"denom index must have latest token pair id",
				)
			} else {
				suite.Len(tc.genesisState.TokenPairs, 0)
				suite.Len(suite.app.Erc20Keeper.GetAllTokenPairDenomIndexes(suite.ctx), 0)
				suite.Len(suite.app.Erc20Keeper.GetAllTokenPairERC20AddressIndexes(suite.ctx), 0)
			}
		})
	}
}
