package serviceaccount

import (
	"io/ioutil"
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/suite"
)

const (
	testSAToken = "0npW4ZyoquYJUtW6b9td"

	// generated with https://golang.org/src/crypto/tls/generate_cert.go
	// go run generate_cert.go --rsa-bits 1024 --host 127.0.0.1,::1,example.com --ca --start-date "Jan 1 00:00:00 1970" --duration=1000000h
	testCACerts = `-----BEGIN CERTIFICATE-----
MIICNTCCAZ6gAwIBAgIRAMusYiYzdGXOLvqV6BcgL5kwDQYJKoZIhvcNAQELBQAw
EjEQMA4GA1UEChMHQWNtZSBDbzAgFw03MDAxMDEwMDAwMDBaGA8yMDg0MDEyOTE2
MDAwMFowEjEQMA4GA1UEChMHQWNtZSBDbzCBnzANBgkqhkiG9w0BAQEFAAOBjQAw
gYkCgYEAu/WV7qoGDjLeFck+oYYdp5xDN3hO8S8d1uzpLNBbMzcsgWsGiN58vIQV
HigulYdz3K9ZOwTyf10AzS6AFvarHnYToYNTkOQ/LLFmIf+M8dE5nHTolEQ0b+y1
JtAU649S67xN2SEYzJ3PsPRaH4K03sZIoxDo1CoxwodHOkvW+psCAwEAAaOBiDCB
hTAOBgNVHQ8BAf8EBAMCAqQwEwYDVR0lBAwwCgYIKwYBBQUHAwEwDwYDVR0TAQH/
BAUwAwEB/zAdBgNVHQ4EFgQUBQd0KF09Xrr7dieSOyyWizX3U8QwLgYDVR0RBCcw
JYILZXhhbXBsZS5jb22HBH8AAAGHEAAAAAAAAAAAAAAAAAAAAAEwDQYJKoZIhvcN
AQELBQADgYEAKX6SSD0+9mPHXxpDhlkYBjMRkS/tyJ4n395vOft3Xc0uVSpzj2mB
zijZYCeUQvuDC+Q1zUZcmuSHmBJP13w/uupioGO0lUUav2H75T0V+h7gTtO/HId0
KWWbADJzw+ZwEtNZoRYip4BmHRXuPkJ3NJBnW1v2++xcLBgdK1r8HJ8=
-----END CERTIFICATE-----
`
)

func TestServiceAccountSuite(t *testing.T) {
	suite.Run(t, new(ServiceAccountSuite))
}

type ServiceAccountSuite struct {
	suite.Suite

	dirName string
}

func (suite *ServiceAccountSuite) TestSecretProvider() {
	// prepare
	secretProvider := &mountedSecretProvider{
		suite.dirName,
	}
	err := suite.writeSAToken(testSAToken)
	suite.NoError(err)
	err = suite.writeCACerts(testCACerts)
	suite.NoError(err)

	// test
	secret, err := secretProvider.Get()

	// verify
	suite.NoError(err)
	suite.Equal(testSAToken, secret.Token)
	suite.Len(secret.RootCAs.Subjects(), 1)
}

func (suite *ServiceAccountSuite) TestSecretProviderMissingCACerts() {
	// prepare
	secretProvider := &mountedSecretProvider{
		suite.dirName,
	}
	err := suite.writeSAToken(testSAToken)
	suite.NoError(err)

	// test
	_, err = secretProvider.Get()

	// verify
	suite.Error(err)
}

func (suite *ServiceAccountSuite) TestSecretProviderMissingSAToken() {
	// prepare
	secretProvider := &mountedSecretProvider{
		suite.dirName,
	}
	err := suite.writeCACerts(testCACerts)
	suite.NoError(err)

	// test
	_, err = secretProvider.Get()

	// verify
	suite.Error(err)
}

func (suite *ServiceAccountSuite) SetupTest() {
	dir, err := ioutil.TempDir("", "eks_connector_sa")
	suite.NoError(err)
	suite.dirName = dir
}

func (suite *ServiceAccountSuite) TearDownTest() {
	err := os.RemoveAll(suite.dirName)
	suite.NoError(err)
}

func (suite *ServiceAccountSuite) writeSAToken(token string) error {
	filePath := path.Join(suite.dirName, FileToken)
	return ioutil.WriteFile(filePath, []byte(token), 0700)
}

func (suite *ServiceAccountSuite) writeCACerts(certs string) error {
	filePath := path.Join(suite.dirName, FileRootCAs)
	return ioutil.WriteFile(filePath, []byte(certs), 0700)
}
