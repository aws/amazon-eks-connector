package agent

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/aws/amazon-eks-connector/pkg/config"
	"github.com/aws/amazon-eks-connector/pkg/ssm"
)

const (
	testActivationID   = "1897a40d-ce57-42f3-8228-25aeaf9dc4f1"
	testActivationCode = "testActivationCode"
	testInstanceID     = "7a68a33a-7a1e-4db3-bd0c-581e5e252319"
	testRegion         = "mars-northeast-2"
)

func TestRunRegistrationSuite(t *testing.T) {
	suite.Run(t, new(RegistrationSuite))
}

type RegistrationSuite struct {
	suite.Suite
	ssm          *ssm.MockClient
	registration Registration
}

func (suite *RegistrationSuite) SetupTest() {
	suite.ssm = &ssm.MockClient{}
	activationConfig := &config.ActivationConfig{
		ID:   testActivationID,
		Code: testActivationCode,
	}
	suite.registration = NewRegistration(suite.ssm, activationConfig)
}

func (suite *RegistrationSuite) TestRegisterHappyCase() {
	// prepare
	suite.ssm.On("RegisterManagedInstance",
		testActivationID,
		testActivationCode,
		mock.Anything,
		KeyType,
		mock.Anything,
	).Return(testInstanceID, nil)
	suite.ssm.On("Region").Return(testRegion)

	// test
	state, err := suite.registration.Register()

	// verify
	suite.NoError(err)
	suite.Equal(testInstanceID, state.InstanceID)
	suite.Equal(testRegion, state.Region)
	suite.Equal(testActivationID, state.ActivationId)
	suite.ssm.AssertExpectations(suite.T())
}

func (suite *RegistrationSuite) TestRegisterSSMCallFailed() {
	// prepare
	err := errors.New("SSMServiceException")
	suite.ssm.On("RegisterManagedInstance",
		testActivationID,
		testActivationCode,
		mock.Anything,
		KeyType,
		mock.Anything,
	).Return("", err)

	// test
	state, err := suite.registration.Register()

	// verify
	suite.Error(err)
	suite.Nil(state)
	suite.ssm.AssertExpectations(suite.T())
}
