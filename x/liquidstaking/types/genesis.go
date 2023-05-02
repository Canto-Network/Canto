package types

// NewGenesisState creates a new GenesisState instance.
func NewGenesisState(
	params Params,
	epoch Epoch,
	lastChunkId, lastInsuranceId uint64,
	chunks []Chunk,
	insurances []Insurance,
	pendingUnstakes []PendingLiquidUnstake,
	infos []UnpairingForUnstakeChunkInfo,
	reqs []WithdrawInsuranceRequest,
) GenesisState {
	return GenesisState{
		LiquidBondDenom:               DefaultLiquidBondDenom,
		Params:                        params,
		Epoch:                         epoch,
		LastChunkId:                   lastChunkId,
		LastInsuranceId:               lastInsuranceId,
		Chunks:                        chunks,
		Insurances:                    insurances,
		PendingLiquidUnstakes:         pendingUnstakes,
		UnpairingForUnstakeChunkInfos: infos,
		WithdrawInsuranceRequests:     reqs,
	}
}

func DefaultGenesisState() *GenesisState {
	return &GenesisState{
		LiquidBondDenom:               DefaultLiquidBondDenom,
		Params:                        DefaultParams(),
		Epoch:                         Epoch{},
		LastChunkId:                   0,
		LastInsuranceId:               0,
		Chunks:                        []Chunk{},
		Insurances:                    []Insurance{},
		PendingLiquidUnstakes:         []PendingLiquidUnstake{},
		UnpairingForUnstakeChunkInfos: []UnpairingForUnstakeChunkInfo{},
		WithdrawInsuranceRequests:     []WithdrawInsuranceRequest{},
	}
}

func (gs GenesisState) Validate() error {
	return gs.Params.Validate()
}
