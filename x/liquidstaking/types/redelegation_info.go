package types

import (
	"time"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

func NewRedelegationInfo(id uint64, completionTime time.Time) RedelegationInfo {
	return RedelegationInfo{
		ChunkId:        id,
		CompletionTime: completionTime,
	}
}

// Matured returns true if currTime pass the completion time.
func (rinfo *RedelegationInfo) Matured(currTime time.Time) bool {
	return !currTime.Before(rinfo.CompletionTime)
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
