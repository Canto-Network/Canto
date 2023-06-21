package types

import (
	"fmt"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	ethermint "github.com/evmos/ethermint/types"
)

var ChunkSize = sdk.TokensFromConsensusPower(250_000, ethermint.PowerReduction)

func NewChunk(id uint64) Chunk {
	return Chunk{
		Id:                   id,
		PairedInsuranceId:    0, // Not yet assigned
		UnpairingInsuranceId: 0, // Not yet assigned
		Status:               CHUNK_STATUS_PAIRING,
	}
}

func (c *Chunk) DerivedAddress() sdk.AccAddress {
	return DeriveAddress(ModuleName, fmt.Sprintf("chunk%d", c.Id))
}

func (c *Chunk) Equal(other Chunk) bool {
	return c.Id == other.Id &&
		c.PairedInsuranceId == other.PairedInsuranceId &&
		c.UnpairingInsuranceId == other.UnpairingInsuranceId &&
		c.Status == other.Status
}

func (c *Chunk) SetStatus(status ChunkStatus) {
	c.Status = status
}

func (c *Chunk) Validate(lastChunkId uint64) error {
	if c.Id > lastChunkId {
		return sdkerrors.Wrapf(
			ErrInvalidChunkId,
			"chunk id must be %d or less",
			lastChunkId,
		)
	}
	if c.Status == CHUNK_STATUS_UNSPECIFIED {
		return ErrInvalidChunkStatus
	}
	return nil
}

func (c *Chunk) HasPairedInsurance() bool {
	return c.PairedInsuranceId != Empty
}
