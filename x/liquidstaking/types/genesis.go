package types

// NewGenesisState creates a new GenesisState instance.
func NewGenesisState(
	params Params,
	epoch Epoch,
	lastChunkId, lastInsuranceId uint64,
	chunks []Chunk,
	insurances []Insurance,
	pendingUnstakes []PendingLiquidUnstake,
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
		PendingLiquidUnstakes:           pendingUnstakes,
		UnpairingForUnstakingChunkInfos: infos,
		WithdrawInsuranceRequests:       reqs,
	}
}

func DefaultGenesisState() *GenesisState {
	return &GenesisState{
		LiquidBondDenom:                 DefaultLiquidBondDenom,
		Params:                          DefaultParams(),
		Epoch:                           Epoch{},
		LastChunkId:                     0,
		LastInsuranceId:                 0,
		Chunks:                          []Chunk{},
		Insurances:                      []Insurance{},
		PendingLiquidUnstakes:           []PendingLiquidUnstake{},
		UnpairingForUnstakingChunkInfos: []UnpairingForUnstakingChunkInfo{},
		WithdrawInsuranceRequests:       []WithdrawInsuranceRequest{},
	}
}

func (gs GenesisState) Validate() error {
	if err := gs.Params.Validate(); err != nil {
		return err
	}
	if err := gs.Epoch.Validate(); err != nil {
	}
	if gs.LastChunkId < 0 {
		return ErrInvalidLastChunkId
	}
	if gs.LastInsuranceId < 0 {
		return ErrInvalidLastInsuranceId
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
	for _, pendingUnstake := range gs.PendingLiquidUnstakes {
		if err := pendingUnstake.Validate(chunkMap); err != nil {
			return err
		}
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
