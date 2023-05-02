package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

func NewPendingLiquidUnstake(chunkId uint64, delegatorAddress string, escrowedLsTokens sdk.Coin) PendingLiquidUnstake {
	return PendingLiquidUnstake{
		ChunkId:          chunkId,
		DelegatorAddress: delegatorAddress,
		EscrowedLstokens: escrowedLsTokens,
	}
}

func (plu *PendingLiquidUnstake) Delegator() sdk.AccAddress {
	return sdk.MustAccAddressFromBech32(plu.DelegatorAddress)
}

func (plu *PendingLiquidUnstake) Validate(chunkMap map[uint64]Chunk) error {
	chunk, ok := chunkMap[plu.ChunkId]
	if !ok {
		return sdkerrors.Wrapf(
			ErrNotFoundPendingLiquidUnstakeChunkId,
			"chunk id: %d",
			plu.ChunkId,
		)
	}
	if chunk.Status != CHUNK_STATUS_PAIRED {
		return ErrInvalidChunkStatus
	}
	return nil
}
