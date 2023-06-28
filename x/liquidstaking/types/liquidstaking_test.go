package types_test

import (
	"github.com/Canto-Network/Canto/v6/x/liquidstaking/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"
	"testing"
)

type liquistakingTestSuite struct {
	suite.Suite
}

func TestLiquidstakingTestSuite(t *testing.T) {
	suite.Run(t, new(liquistakingTestSuite))
}

func (suite *liquistakingTestSuite) TestNativeTokenToLiquidStakeToken() {
	tcs := []struct {
		name                     string
		nativeToken              sdk.Int
		lsTokenTotalSupplyAmount sdk.Int
		netAmount                sdk.Dec
		expected                 string
	}{
		{
			"test1",
			types.ChunkSize,
			sdk.MustNewDecFromStr("750000000000000000000000").TruncateInt(),
			sdk.MustNewDecFromStr("750161999352002591325000"),
			"249946011877386975000000",
		},
	}
	for _, tc := range tcs {
		suite.Run(tc.name, func() {
			result := types.NativeTokenToLiquidStakeToken(tc.nativeToken, tc.lsTokenTotalSupplyAmount, tc.netAmount)
			suite.Equal(tc.expected, result.String())
		})
	}
}
