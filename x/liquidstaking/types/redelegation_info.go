package types

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

func NewRedelegationInfo(id uint64, completionTime time.Time) RedelegationInfo {
	return RedelegationInfo{
		ChunkId:        id,
		CompletionTime: completionTime,
		PenaltyAmt:     sdk.ZeroInt(),
		Deletable:      false,
	}
}

func (rinfo *RedelegationInfo) Matured(currTime time.Time) bool {
	return !rinfo.CompletionTime.Before(currTime)
}

func (rinfo *RedelegationInfo) Validate(chunkMap map[uint64]Chunk) error {
	_, ok := chunkMap[rinfo.ChunkId]
	if !ok {
		return sdkerrors.Wrapf(
			ErrNotFoundRedelegationInfoChunkId,
			"chunk id: %d",
			rinfo.ChunkId,
		)
	}
	return nil

}
