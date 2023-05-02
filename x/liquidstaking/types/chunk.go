package types

import (
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	ethermint "github.com/evmos/ethermint/types"
)

const (
	MaxPairedChunks = 10
)

var ChunkSize = sdk.TokensFromConsensusPower(5000000, ethermint.PowerReduction)

func NewChunk(id uint64) Chunk {
	return Chunk{
		Id:          id,
		InsuranceId: 0, // Not yet assigned
		Status:      CHUNK_STATUS_PAIRING,
	}
}

func (c *Chunk) DerivedAddress() sdk.AccAddress {
	return DeriveAddress(ModuleName, fmt.Sprintf("chunk%d", c.Id))
}

func (c *Chunk) Equal(other Chunk) bool {
	return c.Id == other.Id &&
		c.InsuranceId == other.InsuranceId &&
		c.Status == other.Status
}

func (c *Chunk) SetStatus(status ChunkStatus) {
	c.Status = status
}
