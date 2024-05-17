package internal_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/prefab-cloud/prefab-cloud-go/pkg/internal"
)

type EncryptionTestSuite struct {
	decrypter internal.Encryption
	suite.Suite
}

func (suite *EncryptionTestSuite) TestDecryptionWorks() {
	suite.Run("standard case", func() {
		// #nosec G101 -- this is a test key
		secretValue := "b837acfdedb9f6286947fb95f6fb--13490148d8d3ddf0decc3d14--add9b0ed6de775080bec4c5b6025d67e"
		encryptionKey := "e657e0406fc22e17d3145966396b2130d33dcb30ac0edd62a77235cdd01fc49d"
		decryptedValue, err := suite.decrypter.DecryptValue(encryptionKey, secretValue)
		suite.Require().NoError(err)
		suite.Equal("james-was-here", decryptedValue)
	})
}

func (suite *EncryptionTestSuite) TestDecryptionFails() {
	suite.Run("damaged value", func() {
		// #nosec G101 -- this is a test key
		secretValue := "b837acfdedb9f6286947fb95f6fb--13490148d8d3ddf0decc3d14--add9b0ed6de775080bec4c5b6025d67eee"
		encryptionKey := "e657e0406fc22e17d3145966396b2130d33dcb30ac0edd62a77235cdd01fc49d"
		_, err := suite.decrypter.DecryptValue(encryptionKey, secretValue)
		suite.Error(err, "cipher: message authentication failed")
	})

	suite.Run("bad key", func() {
		// #nosec G101 -- this is a test key
		secretValue := "b837acfdedb9f6286947fb95f6fb--13490148d8d3ddf0decc3d14--add9b0ed6de775080bec4c5b6025d67e"
		encryptionKey := "foo"
		_, err := suite.decrypter.DecryptValue(encryptionKey, secretValue)
		suite.Error(err, "cipher: message authentication failed")
	})
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestDecrypterTestSuite(t *testing.T) {
	suite.Run(t, new(EncryptionTestSuite))
}
