package types_test

import (
	"github.com/Canto-Network/Canto/v6/x/liquidstaking/types"
	"github.com/stretchr/testify/suite"
	"testing"
)

type keysTestSuite struct {
	suite.Suite
}

func TestKeysTestSuite(t *testing.T) {
	suite.Run(t, new(keysTestSuite))
}

func (suite *keysTestSuite) TestGetChunkKey() {
	suite.Equal(
		[]byte{0x4, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x1},
		types.GetChunkKey(1),
		"KeyPrefixChunk + 8-bytes represented id as big endian order",
	)
	suite.Equal(
		[]byte{0x4, 0x0, 0x0, 0x0, 0x2, 0x54, 0xb, 0xe3, 0xff},
		types.GetChunkKey(9999999999),
		"KeyPrefixChunk + 8-bytes represented id as big endian order",
	)
}

func (suite *keysTestSuite) TestGetInsuranceKey() {
	suite.Equal(
		[]byte{0x5, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x1},
		types.GetInsuranceKey(1),
		"KeyPrefixInsurance + 8-bytes represented id as big endian order",
	)
	suite.Equal(
		[]byte{0x5, 0x0, 0x0, 0x0, 0x2, 0x54, 0xb, 0xe3, 0xff},
		types.GetInsuranceKey(9999999999),
		"KeyPrefixInsurance + 8-bytes represented id as big endian order",
	)
}

func (suite *keysTestSuite) TestGetWithdrawInsuranceRequestKey() {
	suite.Equal(
		[]byte{0x6, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x1},
		types.GetWithdrawInsuranceRequestKey(1),
		"KeyPrefixWithdrawInsuranceRequest + 8-bytes represented id as big endian order",
	)
	suite.Equal(
		[]byte{0x6, 0x0, 0x0, 0x0, 0x2, 0x54, 0xb, 0xe3, 0xff},
		types.GetWithdrawInsuranceRequestKey(9999999999),
		"KeyPrefixWithdrawInsuranceRequest + 8-bytes represented id as big endian order",
	)
}

func (suite *keysTestSuite) TestGetUnpairingForUnstakingChunkInfoKey() {
	suite.Equal(
		[]byte{0x7, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x1},
		types.GetUnpairingForUnstakingChunkInfoKey(1),
		"KeyPrefixUnpairingForUnstakingChunkInfo + 8-bytes represented id as big endian order",
	)
	suite.Equal(
		[]byte{0x7, 0x0, 0x0, 0x0, 0x2, 0x54, 0xb, 0xe3, 0xff},
		types.GetUnpairingForUnstakingChunkInfoKey(9999999999),
		"KeyPrefixUnpairingForUnstakingChunkInfo + 8-bytes represented id as big endian order",
	)
}

func (suite *keysTestSuite) TestGetRedelegationInfoKey() {
	suite.Equal(
		[]byte{0x8, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x1},
		types.GetRedelegationInfoKey(1),
		"KeyPrefixRedelegationInfo + 8-bytes represented id as big endian order",
	)
	suite.Equal(
		[]byte{0x8, 0x0, 0x0, 0x0, 0x2, 0x54, 0xb, 0xe3, 0xff},
		types.GetRedelegationInfoKey(9999999999),
		"KeyPrefixRedelegationInfo + 8-bytes represented id as big endian order",
	)
}
