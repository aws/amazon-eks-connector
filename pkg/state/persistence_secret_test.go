package state

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/aws/amazon-eks-connector/pkg/k8s"
)

const (
	testSecretStateFileManifest        = "manifest file content"
	testSecretStateFileFingerPrint     = "finger print file content"
	testSecretStateFileRegistrationKey = "registration key file content"
	testEksConnectorConfig             = "EKS Connector config content"
)

func TestSecretPersistenceSuite(t *testing.T) {
	suite.Run(t, new(SecretPersistenceSuite))
}

type SecretPersistenceSuite struct {
	suite.Suite

	secret      *k8s.MockSecret
	persistence Persistence
}

func (suite *SecretPersistenceSuite) TestSave() {
	// prepare
	state := SerializedState{
		FileManifest:            testSecretStateFileManifest,
		FileRegistrationKey:     testSecretStateFileRegistrationKey,
		FileInstanceFingerprint: testSecretStateFileFingerPrint,
		EksConnectorConfig:      testEksConnectorConfig,
	}
	secretMap := map[string][]byte{
		SecretKeyManifest:        []byte(testSecretStateFileManifest),
		SecretKeyRegistrationKey: []byte(testSecretStateFileRegistrationKey),
		SecretKeyFingerprint:     []byte(testSecretStateFileFingerPrint),
		SecretKeyConnectorConfig: []byte(testEksConnectorConfig),
	}
	suite.secret.On("Put", secretMap).Return(nil)

	// test
	err := suite.persistence.Save(state)

	// verify
	suite.NoError(err)
	suite.secret.AssertExpectations(suite.T())
}

func (suite *SecretPersistenceSuite) TestLoad() {
	// prepare
	state := SerializedState{
		FileManifest:            testSecretStateFileManifest,
		FileRegistrationKey:     testSecretStateFileRegistrationKey,
		FileInstanceFingerprint: testSecretStateFileFingerPrint,
		EksConnectorConfig:      testEksConnectorConfig,
	}
	secretMap := map[string][]byte{
		SecretKeyManifest:        []byte(testSecretStateFileManifest),
		SecretKeyRegistrationKey: []byte(testSecretStateFileRegistrationKey),
		SecretKeyFingerprint:     []byte(testSecretStateFileFingerPrint),
		SecretKeyConnectorConfig: []byte(testEksConnectorConfig),
	}
	suite.secret.On("Get").Return(secretMap, nil)

	// test
	actualState, err := suite.persistence.Load()

	// verify
	suite.NoError(err)
	suite.Equal(state, actualState)
	suite.secret.AssertExpectations(suite.T())
}

func (suite *SecretPersistenceSuite) SetupTest() {
	suite.secret = &k8s.MockSecret{}
	suite.persistence = NewSecretPersistence(suite.secret)
}
