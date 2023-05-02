package types

import sdk "github.com/cosmos/cosmos-sdk/types"

func NewUnpairingForUnstakeChunkInfo(
	chunkId uint64,
	delegatorAddress string,
	escrowedLsTokens sdk.Coin,
) UnpairingForUnstakeChunkInfo {
	return UnpairingForUnstakeChunkInfo{
		ChunkId:          chunkId,
		DelegatorAddress: delegatorAddress,
		EscrowedLstokens: escrowedLsTokens,
	}
}

func (info *UnpairingForUnstakeChunkInfo) GetDelegator() sdk.AccAddress {
	return sdk.MustAccAddressFromBech32(info.DelegatorAddress)
}
