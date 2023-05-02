package types

import (
	"time"
)

func NewWithdrawingInsurance(insuranceId, chunkId uint64, completionTime time.Time) WithdrawingInsurance {
	return WithdrawingInsurance{
		InsuranceId:    insuranceId,
		ChunkId:        chunkId,
		CompletionTime: completionTime,
	}
}
