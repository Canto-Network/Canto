package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

func NewUnpairingForUnstakingChunkInfo(
	chunkId uint64,
	delegatorAddress string,
	escrowedLsTokens sdk.Coin,
) UnpairingForUnstakingChunkInfo {
	return UnpairingForUnstakingChunkInfo{
		ChunkId:          chunkId,
		DelegatorAddress: delegatorAddress,
		EscrowedLstokens: escrowedLsTokens,
	}
}

func (info *UnpairingForUnstakingChunkInfo) GetDelegator() sdk.AccAddress {
	return sdk.MustAccAddressFromBech32(info.DelegatorAddress)
}

func (info *UnpairingForUnstakingChunkInfo) Validate(chunkMap map[uint64]Chunk) error {
	chunk, ok := chunkMap[info.ChunkId]
	if !ok {
		return sdkerrors.Wrapf(
			ErrNotFoundUnpairingForUnstakingChunkInfoChunkId,
			"chunk id: %d",
			info.ChunkId,
		)
	}
	// Chunk related with this info must be in PAIRED or UNPAIRING_FOR_UNSTAKING statuses.
	// PAIRED: unstaking request is just queued, not yet started.
	// UNPAIRING_FOR_UNSTAKING: unstaking request is already started at latest epoch.
	if chunk.Status != CHUNK_STATUS_PAIRED &&
		chunk.Status != CHUNK_STATUS_UNPAIRING_FOR_UNSTAKING {
		return ErrInvalidChunkStatus
	}
	return nil
}

func (info *UnpairingForUnstakingChunkInfo) Equal(other UnpairingForUnstakingChunkInfo) bool {
	return info.ChunkId == other.ChunkId &&
		info.DelegatorAddress == other.DelegatorAddress &&
		info.EscrowedLstokens.IsEqual(other.EscrowedLstokens)
}
