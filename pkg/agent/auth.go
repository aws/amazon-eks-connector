package agent

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"

	"github.com/google/uuid"
)

// https://github.com/aws/amazon-ssm-agent/blob/mainline/agent/managedInstances/auth/rsa_key.go

const (
	keySize int = 2048

	// KeyType returns the RSA Key Type
	KeyType = "Rsa"
)

type rsaKey struct {
	privateKey *rsa.PrivateKey
}

// createFingerPrint creates a new random fingerprint
func createFingerPrint() (string, error) {
	uid, err := uuid.NewRandom()
	if err != nil {
		return "", nil
	}
	return uid.String(), nil
}

// createKeypair creates a new RSA keypair
func createKeypair() (*rsaKey, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, keySize)
	if err != nil {
		return nil, err
	}

	return &rsaKey{
		privateKey: privateKey,
	}, nil
}

// encodePublicKey encodes a public key to a base 64 DER encoded string
func (rsaKey *rsaKey) encodePublicKey() (string, error) {
	publicKeyBytes, err := x509.MarshalPKIXPublicKey(&rsaKey.privateKey.PublicKey)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(publicKeyBytes), nil
}

// encodePrivateKey encodes a private key to a base 64 DER encoded string
func (rsaKey *rsaKey) encodePrivateKey() string {
	privateKeyBytes := x509.MarshalPKCS1PrivateKey(rsaKey.privateKey)

	return base64.StdEncoding.EncodeToString(privateKeyBytes)
}
