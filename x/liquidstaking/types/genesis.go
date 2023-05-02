package types

// NewGenesisState creates a new GenesisState instance.
func NewGenesisState(params Params, epoch Epoch, lastChunkId, lastInsuranceId uint64, chunks []Chunk, insurances []Insurance, withdrawingInsurances []WithdrawingInsurance, liquidUnstakeUnbondingDelegationInfos []LiquidUnstakeUnbondingDelegationInfo) GenesisState {
	return GenesisState{
		LiquidBondDenom:                       DefaultLiquidBondDenom,
		Params:                                params,
		Epoch:                                 epoch,
		LastChunkId:                           lastChunkId,
		LastInsuranceId:                       lastInsuranceId,
		Chunks:                                chunks,
		Insurances:                            insurances,
		WithdrawingInsurances:                 withdrawingInsurances,
		LiquidUnstakeUnbondingDelegationInfos: liquidUnstakeUnbondingDelegationInfos,
	}
}

func DefaultGenesisState() *GenesisState {
	return &GenesisState{
		LiquidBondDenom:                       DefaultLiquidBondDenom,
		Params:                                DefaultParams(),
		Epoch:                                 Epoch{},
		LastChunkId:                           0,
		LastInsuranceId:                       0,
		Chunks:                                []Chunk{},
		Insurances:                            []Insurance{},
		WithdrawingInsurances:                 []WithdrawingInsurance{},
		LiquidUnstakeUnbondingDelegationInfos: []LiquidUnstakeUnbondingDelegationInfo{},
	}
}

func (gs GenesisState) Validate() error {
	return gs.Params.Validate()
}
