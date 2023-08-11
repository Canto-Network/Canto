package types_test

import (
	"testing"

	"github.com/Canto-Network/Canto/v7/x/liquidstaking/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"
)

type netAmountTestSuite struct {
	suite.Suite
}

func TestNetAmountTestSuite(t *testing.T) {
	suite.Run(t, new(netAmountTestSuite))
}

func (suite *netAmountTestSuite) TestString() {
	nas := types.NetAmountState{
		MintRate:                           sdk.NewDec(1),
		LsTokensTotalSupply:                sdk.NewInt(1),
		NetAmount:                          sdk.NewDec(1),
		TotalLiquidTokens:                  sdk.NewInt(1),
		RewardModuleAccBalance:             sdk.NewInt(1),
		FeeRate:                            sdk.NewDec(1),
		UtilizationRatio:                   sdk.NewDec(1),
		RemainingChunkSlots:                sdk.NewInt(1),
		DiscountRate:                       sdk.NewDec(1),
		NumPairedChunks:                    sdk.NewInt(1),
		ChunkSize:                          sdk.NewInt(1),
		TotalDelShares:                     sdk.NewDec(1),
		TotalRemainingRewards:              sdk.NewDec(1),
		TotalChunksBalance:                 sdk.NewInt(1),
		TotalUnbondingChunksBalance:        sdk.NewInt(1),
		TotalInsuranceTokens:               sdk.NewInt(1),
		TotalPairedInsuranceTokens:         sdk.NewInt(1),
		TotalUnpairingInsuranceTokens:      sdk.NewInt(1),
		TotalRemainingInsuranceCommissions: sdk.NewDec(1),
	}
	suite.Equal(
		`NetAmountState:
	MintRate:                   1.000000000000000000
	LsTokensTotalSupply:        1
	NetAmount: 	                1.000000000000000000	
	TotalLiquidTokens:          1	
	RewardModuleAccountBalance: 1
	FeeRate:                    1.000000000000000000
	UtilizationRatio:           1.000000000000000000
	RemainingChunkSlots:        1
	DiscountRate:               1.000000000000000000
	NumPairedChunks:            1
	ChunkSize:                  1
	TotalDelShares:             1.000000000000000000
	TotalRemainingRewards:      1.000000000000000000	
	TotalChunksBalance:         1	
	TotalUnbondingBalance:      1
	TotalInsuranceTokens:       1
	TotalPairedInsuranceTokens: 1
    TotalUnpairingInsuranceTokens: 1
    TotalRemainingInsuranceCommissions: 1.000000000000000000
`,
		nas.String(),
	)
}
