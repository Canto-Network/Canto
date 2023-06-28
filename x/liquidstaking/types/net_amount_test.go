package types_test

import (
	"testing"

	"github.com/Canto-Network/Canto/v6/x/liquidstaking/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"
)

type netAmountTestSuite struct {
	suite.Suite
}

func TestNetAmountTestSuite(t *testing.T) {
	suite.Run(t, new(netAmountTestSuite))
}

func (suite *netAmountTestSuite) TestCalcNetAmount() {
	nas := types.NetAmountState{
		TotalChunksBalance:          sdk.ZeroInt(),
		TotalLiquidTokens:           sdk.MustNewDecFromStr("250000000000000000000000").TruncateInt(),
		TotalUnbondingChunksBalance: sdk.MustNewDecFromStr("250000000000000000000000").TruncateInt(),
		TotalRemainingRewards:       sdk.MustNewDecFromStr("160000000000000000000"),
	}
	suite.Equal(
		"500320000000000000000000.000000000000000000",
		nas.CalcNetAmount(sdk.MustNewDecFromStr("160000000000000000000").TruncateInt()).String(),
	)
}

func (suite *netAmountTestSuite) TestCalcMintRate() {
	nas := types.NetAmountState{
		LsTokensTotalSupply: sdk.MustNewDecFromStr("500000000000000000000000").TruncateInt(),
		NetAmount:           sdk.MustNewDecFromStr("503320000000000000000000"),
	}
	suite.Equal(
		"0.993403798776126519",
		nas.CalcMintRate().String(),
	)

	nas.NetAmount = sdk.ZeroDec()
	suite.Equal(
		"0.000000000000000000",
		nas.CalcMintRate().String(),
	)
}

func (suite *netAmountTestSuite) TestEqual() {
	nas := types.NetAmountState{
		MintRate:                           sdk.ZeroDec(),
		LsTokensTotalSupply:                sdk.ZeroInt(),
		NetAmount:                          sdk.ZeroDec(),
		TotalLiquidTokens:                  sdk.ZeroInt(),
		RewardModuleAccBalance:             sdk.ZeroInt(),
		FeeRate:                            sdk.ZeroDec(),
		UtilizationRatio:                   sdk.ZeroDec(),
		RemainingChunkSlots:                sdk.ZeroInt(),
		DiscountRate:                       sdk.ZeroDec(),
		NumPairedChunks:                    sdk.ZeroInt(),
		ChunkSize:                          sdk.ZeroInt(),
		TotalDelShares:                     sdk.ZeroDec(),
		TotalRemainingRewards:              sdk.ZeroDec(),
		TotalChunksBalance:                 sdk.ZeroInt(),
		TotalUnbondingChunksBalance:        sdk.ZeroInt(),
		TotalInsuranceTokens:               sdk.ZeroInt(),
		TotalPairedInsuranceTokens:         sdk.ZeroInt(),
		TotalUnpairingInsuranceTokens:      sdk.ZeroInt(),
		TotalRemainingInsuranceCommissions: sdk.ZeroDec(),
	}
	cpy := nas
	suite.True(nas.Equal(cpy))

	cpy.ChunkSize = nas.ChunkSize.Add(sdk.OneInt())
	suite.True(
		nas.Equal(cpy),
		"chunk size should not affect equality",
	)

	cpy = nas
	cpy.MintRate = nas.MintRate.Add(sdk.OneDec())
	suite.False(
		nas.Equal(cpy),
		"mint rate should affect equality",
	)

	cpy = nas
	cpy.LsTokensTotalSupply = nas.LsTokensTotalSupply.Add(sdk.OneInt())
	suite.False(
		nas.Equal(cpy),
		"ls tokens total supply should affect equality",
	)

	cpy = nas
	cpy.NetAmount = nas.NetAmount.Add(sdk.OneDec())
	suite.False(
		nas.Equal(cpy),
		"net amount should affect equality",
	)

	cpy = nas
	cpy.TotalLiquidTokens = nas.TotalLiquidTokens.Add(sdk.OneInt())
	suite.False(
		nas.Equal(cpy),
		"total liquid tokens should affect equality",
	)

	cpy = nas
	cpy.RewardModuleAccBalance = nas.RewardModuleAccBalance.Add(sdk.OneInt())
	suite.False(
		nas.Equal(cpy),
		"reward module acc balance should affect equality",
	)

	cpy = nas
	cpy.FeeRate = nas.FeeRate.Add(sdk.OneDec())
	suite.False(
		nas.Equal(cpy),
		"fee rate should affect equality",
	)

	cpy = nas
	cpy.UtilizationRatio = nas.UtilizationRatio.Add(sdk.OneDec())
	suite.False(
		nas.Equal(cpy),
		"utilization ratio should affect equality",
	)

	cpy = nas
	cpy.RemainingChunkSlots = nas.RemainingChunkSlots.Add(sdk.OneInt())
	suite.False(
		nas.Equal(cpy),
		"remaining chunk slots should affect equality",
	)

	cpy = nas
	cpy.DiscountRate = nas.DiscountRate.Add(sdk.OneDec())
	suite.False(
		nas.Equal(cpy),
		"discount rate should affect equality",
	)

	cpy = nas
	cpy.NumPairedChunks = nas.NumPairedChunks.Add(sdk.OneInt())
	suite.False(
		nas.Equal(cpy),
		"num paired chunks should affect equality",
	)

	cpy = nas
	cpy.TotalDelShares = nas.TotalDelShares.Add(sdk.OneDec())
	suite.False(
		nas.Equal(cpy),
		"total del shares should affect equality",
	)

	cpy = nas
	cpy.TotalRemainingRewards = nas.TotalRemainingRewards.Add(sdk.OneDec())
	suite.False(
		nas.Equal(cpy),
		"total remaining rewards should affect equality",
	)

	cpy = nas
	cpy.TotalChunksBalance = nas.TotalChunksBalance.Add(sdk.OneInt())
	suite.False(
		nas.Equal(cpy),
		"total chunks balance should affect equality",
	)

	cpy = nas
	cpy.TotalUnbondingChunksBalance = nas.TotalUnbondingChunksBalance.Add(sdk.OneInt())
	suite.False(
		nas.Equal(cpy),
		"total unbonding chunks balance should affect equality",
	)

	cpy = nas
	cpy.TotalInsuranceTokens = nas.TotalInsuranceTokens.Add(sdk.OneInt())
	suite.False(
		nas.Equal(cpy),
		"total insurance tokens should affect equality",
	)

	cpy = nas
	cpy.TotalPairedInsuranceTokens = nas.TotalPairedInsuranceTokens.Add(sdk.OneInt())
	suite.False(
		nas.Equal(cpy),
		"total paired insurance tokens should affect equality",
	)

	cpy = nas
	cpy.TotalUnpairingInsuranceTokens = nas.TotalUnpairingInsuranceTokens.Add(sdk.OneInt())
	suite.False(
		nas.Equal(cpy),
		"total unpairing insurance tokens should affect equality",
	)

	cpy = nas
	cpy.TotalRemainingInsuranceCommissions = nas.TotalRemainingInsuranceCommissions.Add(sdk.OneDec())
	suite.False(
		nas.Equal(cpy),
		"total remaining insurance commissions should affect equality",
	)

	cpy = nas
	cpy.NumPairedChunks = nas.NumPairedChunks.Add(sdk.OneInt())
	suite.False(
		nas.Equal(cpy),
		"num paired chunks should affect equality",
	)
}

func (suite *netAmountTestSuite) TestIsZeroState() {
	nas := types.NetAmountState{
		MintRate:                           sdk.ZeroDec(),
		LsTokensTotalSupply:                sdk.ZeroInt(),
		NetAmount:                          sdk.ZeroDec(),
		TotalLiquidTokens:                  sdk.ZeroInt(),
		RewardModuleAccBalance:             sdk.ZeroInt(),
		FeeRate:                            sdk.ZeroDec(),
		UtilizationRatio:                   sdk.ZeroDec(),
		RemainingChunkSlots:                sdk.ZeroInt(),
		DiscountRate:                       sdk.ZeroDec(),
		NumPairedChunks:                    sdk.ZeroInt(),
		ChunkSize:                          sdk.ZeroInt(),
		TotalDelShares:                     sdk.ZeroDec(),
		TotalRemainingRewards:              sdk.ZeroDec(),
		TotalChunksBalance:                 sdk.ZeroInt(),
		TotalUnbondingChunksBalance:        sdk.ZeroInt(),
		TotalInsuranceTokens:               sdk.ZeroInt(),
		TotalPairedInsuranceTokens:         sdk.ZeroInt(),
		TotalUnpairingInsuranceTokens:      sdk.ZeroInt(),
		TotalRemainingInsuranceCommissions: sdk.ZeroDec(),
	}
	suite.True(nas.IsZeroState())

	cpy := nas
	cpy.RemainingChunkSlots = nas.RemainingChunkSlots.Add(sdk.OneInt())
	suite.True(
		cpy.IsZeroState(),
		"remaining chunk slots should not affect zero state",
	)

	cpy = nas
	cpy.TotalInsuranceTokens = nas.TotalInsuranceTokens.Add(sdk.OneInt())
	suite.True(
		cpy.IsZeroState(),
		"total insurance tokens should not affect zero state",
	)

	cpy = nas
	cpy.ChunkSize = nas.ChunkSize.Add(sdk.OneInt())
	suite.True(
		cpy.IsZeroState(),
		"chunk size should not affect zero state",
	)
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
