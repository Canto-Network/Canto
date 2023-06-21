package types

import (
	"time"

	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

// NewGenesisState creates a new GenesisState instance.
func NewGenesisState(
	params Params,
	epoch Epoch,
	lastChunkId, lastInsuranceId uint64,
	chunks []Chunk,
	insurances []Insurance,
	infos []UnpairingForUnstakingChunkInfo,
	reqs []WithdrawInsuranceRequest,
) GenesisState {
	return GenesisState{
		LiquidBondDenom:                 DefaultLiquidBondDenom,
		Params:                          params,
		Epoch:                           epoch,
		LastChunkId:                     lastChunkId,
		LastInsuranceId:                 lastInsuranceId,
		Chunks:                          chunks,
		Insurances:                      insurances,
		UnpairingForUnstakingChunkInfos: infos,
		WithdrawInsuranceRequests:       reqs,
	}
}

func DefaultGenesisState() *GenesisState {
	return &GenesisState{
		LiquidBondDenom: DefaultLiquidBondDenom,
		Params:          DefaultParams(),
		Epoch: Epoch{
			CurrentNumber: 0,
			StartTime:     time.Time{},
			Duration:      types.DefaultUnbondingTime,
			StartHeight:   0,
		},
		LastChunkId:                     0,
		LastInsuranceId:                 0,
		Chunks:                          []Chunk{},
		Insurances:                      []Insurance{},
		UnpairingForUnstakingChunkInfos: []UnpairingForUnstakingChunkInfo{},
		WithdrawInsuranceRequests:       []WithdrawInsuranceRequest{},
	}
}

// Validate performs basic genesis state validation returning an error upon any
func (gs GenesisState) Validate() error {
	if err := gs.Params.Validate(); err != nil {
		return err
	}
	if err := gs.Epoch.Validate(); err != nil {
		return err
	}
	chunkMap := make(map[uint64]Chunk)
	for _, chunk := range gs.Chunks {
		if err := chunk.Validate(gs.LastChunkId); err != nil {
			return err
		}
		chunkMap[chunk.Id] = chunk
	}
	insuranceMap := make(map[uint64]Insurance)
	for _, insurance := range gs.Insurances {
		if err := insurance.Validate(gs.LastInsuranceId); err != nil {
			return err
		}
		insuranceMap[insurance.Id] = insurance
	}
	for _, info := range gs.UnpairingForUnstakingChunkInfos {
		if err := info.Validate(chunkMap); err != nil {
			return err
		}
	}
	for _, req := range gs.WithdrawInsuranceRequests {
		if err := req.Validate(insuranceMap); err != nil {
			return err
		}
	}
	return nil
}
