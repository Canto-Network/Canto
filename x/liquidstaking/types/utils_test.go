package types_test

import (
	math_rand "math/rand"
	"testing"
	"time"

	"github.com/Canto-Network/Canto/v7/x/liquidstaking/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"
)

type utilsTestSuite struct {
	suite.Suite
}

func TestUtilsTestSuite(t *testing.T) {
	suite.Run(t, new(utilsTestSuite))
}

func (suite *utilsTestSuite) TestDeriveAddress() {
}

func (suite *utilsTestSuite) TestRandomInt() {
	r := math_rand.New(math_rand.NewSource(time.Now().UnixNano()))
	v := types.RandomInt(r, sdk.ZeroInt(), sdk.NewInt(100))
	suite.True(v.GTE(sdk.ZeroInt()))
	suite.True(v.LT(sdk.NewInt(100)))
}

func (suite *utilsTestSuite) TestRandomDec() {
	r := math_rand.New(math_rand.NewSource(time.Now().UnixNano()))
	v := types.RandomDec(r, sdk.ZeroDec(), sdk.NewDec(100))
	suite.True(v.GTE(sdk.ZeroDec()))
	suite.True(v.LT(sdk.NewDec(100)))
}
