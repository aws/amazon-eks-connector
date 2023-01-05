package agent

import (
	"crypto"
	"crypto/x509"
	"encoding/base64"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

func TestCreateFingerPrint(t *testing.T) {
	fingerprint, err := createFingerPrint()
	assert.NoError(t, err)
	_, err = uuid.Parse(fingerprint)
	assert.NoError(t, err, "Fingerprint must be UUID")
}

func TestRunRsaKeySuite(t *testing.T) {
	suite.Run(t, new(RsaKeySuite))
}

type RsaKeySuite struct {
	suite.Suite
}

func (suite *RsaKeySuite) TestEncodePublicKey() {
	// prepare
	keypair, err := createKeypair()
	suite.NoError(err)
	suite.NotNil(keypair.privateKey)

	// test
	encoded, err := keypair.encodePublicKey()

	// verify
	suite.NoError(err)
	publicKey, err := decodePublicKey(encoded)
	suite.NoError(err)
	suite.Equal(keypair.privateKey.Public(), publicKey)
}

func (suite *RsaKeySuite) TestEncodePrivateKey() {
	// prepare
	keypair, err := createKeypair()
	suite.NoError(err)
	suite.NotNil(keypair.privateKey)

	// test
	encoded := keypair.encodePrivateKey()

	// verify
	decodedKeypair, err := decodePrivateKey(encoded)
	suite.NoError(err)
	suite.Equal(keypair, decodedKeypair)
}

func decodePublicKey(encodedKey string) (crypto.PublicKey, error) {
	publicKeyBytes, err := base64.StdEncoding.DecodeString(encodedKey)
	if err != nil {
		return nil, err
	}
	return x509.ParsePKIXPublicKey(publicKeyBytes)
}

func decodePrivateKey(privateKey string) (*rsaKey, error) {
	privateKeyBytes, err := base64.StdEncoding.DecodeString(privateKey)
	if err != nil {
		return nil, err
	}
	rsaPrivateKey, err := x509.ParsePKCS1PrivateKey(privateKeyBytes)
	if err != nil {
		return nil, err
	}

	return &rsaKey{
		privateKey: rsaPrivateKey,
	}, nil
}
