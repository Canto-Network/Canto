package types_test

import (
	"testing"

	"github.com/Canto-Network/Canto/v6/x/liquidstaking/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"
)

type unpairingForUnstakeChunkInfoTestSuite struct {
	suite.Suite
}

func TestUnpairingForUnstakeChunkInfoTestSuite(t *testing.T) {
	suite.Run(t, new(unpairingForUnstakeChunkInfoTestSuite))
}

func (suite *unpairingForUnstakeChunkInfoTestSuite) TestNewUnpairingForUnstakeChunkInfo() {
	c := types.NewChunk(1)
	delegator := sdk.AccAddress("1")
	escrowedLsTokens := sdk.NewCoin(types.DefaultLiquidBondDenom, sdk.NewInt(100))
	info := types.NewUnpairingForUnstakingChunkInfo(
		c.Id,
		delegator.String(),
		escrowedLsTokens,
	)
	suite.Equal(c.Id, info.ChunkId)
	suite.Equal(delegator.String(), info.DelegatorAddress)
	suite.Equal(escrowedLsTokens.String(), info.EscrowedLstokens.String())
}

func (suite *unpairingForUnstakeChunkInfoTestSuite) TestGetDelegator() {
	delegator := sdk.AccAddress("1")
	info := types.UnpairingForUnstakingChunkInfo{
		DelegatorAddress: delegator.String(),
	}
	suite.Equal(delegator, info.GetDelegator())
}

func (suite *unpairingForUnstakeChunkInfoTestSuite) TestEqual() {
	c := types.NewChunk(1)
	delegator := sdk.AccAddress("1")
	escrowedLsTokens := sdk.NewCoin(types.DefaultLiquidBondDenom, sdk.NewInt(100))
	info := types.NewUnpairingForUnstakingChunkInfo(
		c.Id,
		delegator.String(),
		escrowedLsTokens,
	)

	cpy := info
	suite.True(cpy.Equal(info))

	cpy.ChunkId = 2
	suite.False(cpy.Equal(info))

	cpy.ChunkId = info.ChunkId
	cpy.DelegatorAddress = "2"
	suite.False(cpy.Equal(info))

	cpy.DelegatorAddress = info.DelegatorAddress
	cpy.EscrowedLstokens = sdk.NewCoin(types.DefaultLiquidBondDenom, sdk.NewInt(200))
	suite.False(cpy.Equal(info))
}

func (suite *unpairingForUnstakeChunkInfoTestSuite) TestValidate() {
	c := types.NewChunk(1)
	c.Status = types.CHUNK_STATUS_UNPAIRING_FOR_UNSTAKING
	delegator := sdk.AccAddress("1")
	escrowedLsTokens := sdk.NewCoin(types.DefaultLiquidBondDenom, sdk.NewInt(100))
	info := types.NewUnpairingForUnstakingChunkInfo(
		c.Id,
		delegator.String(),
		escrowedLsTokens,
	)
	chunkMap := map[uint64]types.Chunk{
		c.Id: c,
	}
	suite.NoError(info.Validate(chunkMap))

	chunkMap[c.Id] = types.Chunk{
		Id:                   c.Id,
		PairedInsuranceId:    types.Empty,
		UnpairingInsuranceId: types.Empty,
		Status:               types.CHUNK_STATUS_PAIRING,
	}
	suite.Error(info.Validate(chunkMap))

	delete(chunkMap, c.Id)
	suite.Error(info.Validate(chunkMap))
}
