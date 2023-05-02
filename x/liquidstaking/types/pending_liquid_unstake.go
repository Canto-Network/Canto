package types

import sdk "github.com/cosmos/cosmos-sdk/types"

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
