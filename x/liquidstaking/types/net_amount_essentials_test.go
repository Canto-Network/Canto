package types_test

import (
	"testing"

	"github.com/Canto-Network/Canto/v7/x/liquidstaking/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"
)

type netAmountEssentialsTestSuite struct {
	suite.Suite
}

func TestNetAmountEssentialsTestSuite(t *testing.T) {
	suite.Run(t, new(netAmountEssentialsTestSuite))
}

func (suite *netAmountEssentialsTestSuite) TestCalcNetAmount() {
	nase := types.NetAmountStateEssentials{
		TotalChunksBalance:          sdk.ZeroInt(),
		TotalLiquidTokens:           sdk.MustNewDecFromStr("250000000000000000000000").TruncateInt(),
		TotalUnbondingChunksBalance: sdk.MustNewDecFromStr("250000000000000000000000").TruncateInt(),
		TotalRemainingRewards:       sdk.MustNewDecFromStr("160000000000000000000"),
		RewardModuleAccBalance:      sdk.MustNewDecFromStr("160000000000000000000").TruncateInt(),
	}
	suite.Equal(
		"500320000000000000000000.000000000000000000",
		nase.CalcNetAmount().String(),
	)
}

func (suite *netAmountEssentialsTestSuite) TestCalcMintRate() {
	nase := types.NetAmountStateEssentials{
		LsTokensTotalSupply: sdk.MustNewDecFromStr("500000000000000000000000").TruncateInt(),
		NetAmount:           sdk.MustNewDecFromStr("503320000000000000000000"),
	}
	suite.Equal(
		"0.993403798776126519",
		nase.CalcMintRate().String(),
	)

	nase.NetAmount = sdk.ZeroDec()
	suite.Equal(
		"0.000000000000000000",
		nase.CalcMintRate().String(),
	)
}

func (suite *netAmountEssentialsTestSuite) TestEqual() {
	nase := types.NetAmountStateEssentials{
		MintRate:                    sdk.ZeroDec(),
		LsTokensTotalSupply:         sdk.ZeroInt(),
		NetAmount:                   sdk.ZeroDec(),
		TotalLiquidTokens:           sdk.ZeroInt(),
		RewardModuleAccBalance:      sdk.ZeroInt(),
		FeeRate:                     sdk.ZeroDec(),
		UtilizationRatio:            sdk.ZeroDec(),
		RemainingChunkSlots:         sdk.ZeroInt(),
		DiscountRate:                sdk.ZeroDec(),
		NumPairedChunks:             sdk.ZeroInt(),
		ChunkSize:                   sdk.ZeroInt(),
		TotalDelShares:              sdk.ZeroDec(),
		TotalRemainingRewards:       sdk.ZeroDec(),
		TotalChunksBalance:          sdk.ZeroInt(),
		TotalUnbondingChunksBalance: sdk.ZeroInt(),
	}
	cpy := nase
	suite.True(nase.Equal(cpy))

	cpy.ChunkSize = nase.ChunkSize.Add(sdk.OneInt())
	suite.True(
		nase.Equal(cpy),
		"chunk size should not affect equality",
	)

	cpy = nase
	cpy.MintRate = nase.MintRate.Add(sdk.OneDec())
	suite.False(
		nase.Equal(cpy),
		"mint rate should affect equality",
	)

	cpy = nase
	cpy.LsTokensTotalSupply = nase.LsTokensTotalSupply.Add(sdk.OneInt())
	suite.False(
		nase.Equal(cpy),
		"ls tokens total supply should affect equality",
	)

	cpy = nase
	cpy.NetAmount = nase.NetAmount.Add(sdk.OneDec())
	suite.False(
		nase.Equal(cpy),
		"net amount should affect equality",
	)

	cpy = nase
	cpy.TotalLiquidTokens = nase.TotalLiquidTokens.Add(sdk.OneInt())
	suite.False(
		nase.Equal(cpy),
		"total liquid tokens should affect equality",
	)

	cpy = nase
	cpy.RewardModuleAccBalance = nase.RewardModuleAccBalance.Add(sdk.OneInt())
	suite.False(
		nase.Equal(cpy),
		"reward module acc balance should affect equality",
	)

	cpy = nase
	cpy.FeeRate = nase.FeeRate.Add(sdk.OneDec())
	suite.False(
		nase.Equal(cpy),
		"fee rate should affect equality",
	)

	cpy = nase
	cpy.UtilizationRatio = nase.UtilizationRatio.Add(sdk.OneDec())
	suite.False(
		nase.Equal(cpy),
		"utilization ratio should affect equality",
	)

	cpy = nase
	cpy.RemainingChunkSlots = nase.RemainingChunkSlots.Add(sdk.OneInt())
	suite.False(
		nase.Equal(cpy),
		"remaining chunk slots should affect equality",
	)

	cpy = nase
	cpy.DiscountRate = nase.DiscountRate.Add(sdk.OneDec())
	suite.False(
		nase.Equal(cpy),
		"discount rate should affect equality",
	)

	cpy = nase
	cpy.NumPairedChunks = nase.NumPairedChunks.Add(sdk.OneInt())
	suite.False(
		nase.Equal(cpy),
		"num paired chunks should affect equality",
	)

	cpy = nase
	cpy.TotalDelShares = nase.TotalDelShares.Add(sdk.OneDec())
	suite.False(
		nase.Equal(cpy),
		"total del shares should affect equality",
	)

	cpy = nase
	cpy.TotalRemainingRewards = nase.TotalRemainingRewards.Add(sdk.OneDec())
	suite.False(
		nase.Equal(cpy),
		"total remaining rewards should affect equality",
	)

	cpy = nase
	cpy.TotalChunksBalance = nase.TotalChunksBalance.Add(sdk.OneInt())
	suite.False(
		nase.Equal(cpy),
		"total chunks balance should affect equality",
	)

	cpy = nase
	cpy.TotalUnbondingChunksBalance = nase.TotalUnbondingChunksBalance.Add(sdk.OneInt())
	suite.False(
		nase.Equal(cpy),
		"total unbonding chunks balance should affect equality",
	)

	cpy = nase
	cpy.NumPairedChunks = nase.NumPairedChunks.Add(sdk.OneInt())
	suite.False(
		nase.Equal(cpy),
		"num paired chunks should affect equality",
	)
}

func (suite *netAmountEssentialsTestSuite) TestIsZeroState() {
	nas := types.NetAmountStateEssentials{
		MintRate:                    sdk.ZeroDec(),
		LsTokensTotalSupply:         sdk.ZeroInt(),
		NetAmount:                   sdk.ZeroDec(),
		TotalLiquidTokens:           sdk.ZeroInt(),
		RewardModuleAccBalance:      sdk.ZeroInt(),
		FeeRate:                     sdk.ZeroDec(),
		UtilizationRatio:            sdk.ZeroDec(),
		RemainingChunkSlots:         sdk.ZeroInt(),
		DiscountRate:                sdk.ZeroDec(),
		NumPairedChunks:             sdk.ZeroInt(),
		ChunkSize:                   sdk.ZeroInt(),
		TotalDelShares:              sdk.ZeroDec(),
		TotalRemainingRewards:       sdk.ZeroDec(),
		TotalChunksBalance:          sdk.ZeroInt(),
		TotalUnbondingChunksBalance: sdk.ZeroInt(),
	}
	suite.True(nas.IsZeroState())

	cpy := nas
	cpy.RemainingChunkSlots = nas.RemainingChunkSlots.Add(sdk.OneInt())
	suite.True(
		cpy.IsZeroState(),
		"remaining chunk slots should not affect zero state",
	)

	cpy = nas
	cpy.ChunkSize = nas.ChunkSize.Add(sdk.OneInt())
	suite.True(
		cpy.IsZeroState(),
		"chunk size should not affect zero state",
	)
}

func (suite *netAmountEssentialsTestSuite) TestString() {
	nase := types.NetAmountStateEssentials{
		MintRate:                    sdk.NewDec(1),
		LsTokensTotalSupply:         sdk.NewInt(1),
		NetAmount:                   sdk.NewDec(1),
		TotalLiquidTokens:           sdk.NewInt(1),
		RewardModuleAccBalance:      sdk.NewInt(1),
		FeeRate:                     sdk.NewDec(1),
		UtilizationRatio:            sdk.NewDec(1),
		RemainingChunkSlots:         sdk.NewInt(1),
		DiscountRate:                sdk.NewDec(1),
		NumPairedChunks:             sdk.NewInt(1),
		ChunkSize:                   sdk.NewInt(1),
		TotalDelShares:              sdk.NewDec(1),
		TotalRemainingRewards:       sdk.NewDec(1),
		TotalChunksBalance:          sdk.NewInt(1),
		TotalUnbondingChunksBalance: sdk.NewInt(1),
	}
	suite.Equal(
		`NetAmountStateEssentials:
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
`,
		nase.String(),
	)
}
