package testutil

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

func TestIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}