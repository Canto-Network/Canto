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
	if chunk.Status != CHUNK_STATUS_UNPAIRING_FOR_UNSTAKING {
		return ErrInvalidChunkStatus
	}
	return nil
}
