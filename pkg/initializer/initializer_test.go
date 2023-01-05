package initializer

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/aws/amazon-eks-connector/pkg/agent"
	"github.com/aws/amazon-eks-connector/pkg/config"
	"github.com/aws/amazon-eks-connector/pkg/state"
)

const (
	testActivationId          = "310f3d2f-7b9f-4619-9996-f1526bcff3d2"
	testActivationId2         = "3e971d22-c65d-4d89-bac5-ce79b74ff222"
	testActivationCode        = "OJuk0mXbzV"
	testFingerPrint           = "7eb32fa4-5b75-4431-866e-c2af92f5440b"
	testInstanceID            = "9e629669-d5f6-47e6-97e0-3f5c57c01d82"
	testPrivateKey            = "greedisgood"
	testPrivateKeyType        = "war3"
	testPrivateKeyCreatedDate = "2021-07-30 00:00:00.999999999 -0700 PDT"
	testRegion                = "mars-northeast-2"
)

func TestInitializerSuite(t *testing.T) {
	suite.Run(t, new(InitializerSuite))
}

type InitializerSuite struct {
	suite.Suite
	secretPersistence *state.MockPersistence
	fsPersistence     *state.MockPersistence
	registration      *agent.MockRegistration

	initializer Initializer
}

func (suite *InitializerSuite) SetupTest() {
	activationConfig := &config.ActivationConfig{
		Code: testActivationCode,
		ID:   testActivationId,
	}
	suite.secretPersistence = &state.MockPersistence{}
	suite.fsPersistence = &state.MockPersistence{}
	suite.registration = &agent.MockRegistration{}
	suite.initializer = NewInitializer(activationConfig, suite.secretPersistence, suite.fsPersistence, suite.registration)
}

func (suite *InitializerSuite) TestInitializeNoSavedStateHappyCase() {
	// prepare
	state := newTestState()
	serializedState, err := state.Serialize()
	suite.NoError(err)
	suite.secretPersistence.On("Load").Return(nil, nil)
	suite.registration.On("Register").Return(state, nil)
	suite.secretPersistence.On("Save", serializedState).Return(nil)
	suite.fsPersistence.On("Save", serializedState).Return(nil)

	// test
	actualErr := suite.initializer.Initialize()

	// verify
	suite.NoError(actualErr)
	suite.secretPersistence.AssertExpectations(suite.T())
	suite.fsPersistence.AssertExpectations(suite.T())
	suite.registration.AssertExpectations(suite.T())
}

func (suite *InitializerSuite) TestInitializeNoSavedStateFailedRegistration() {
	// prepare
	err := errors.New("failed registration")
	suite.secretPersistence.On("Load").Return(nil, nil)
	suite.registration.On("Register").Return(nil, err)

	// test
	actualErr := suite.initializer.Initialize()

	// verify
	suite.ErrorIs(actualErr, err)
	suite.secretPersistence.AssertExpectations(suite.T())
	suite.fsPersistence.AssertExpectations(suite.T())
	suite.registration.AssertExpectations(suite.T())
}

func (suite *InitializerSuite) TestInitializeNoSavedStateFailedSecretPersistence() {
	// prepare
	state := newTestState()
	serializedState, err := state.Serialize()
	suite.NoError(err)
	err = errors.New("failed persistence")
	suite.secretPersistence.On("Load").Return(nil, nil)
	suite.registration.On("Register").Return(state, nil)
	suite.secretPersistence.On("Save", serializedState).Return(err)

	// test
	actualErr := suite.initializer.Initialize()

	// verify
	suite.ErrorIs(actualErr, err)
	suite.secretPersistence.AssertExpectations(suite.T())
	suite.fsPersistence.AssertExpectations(suite.T())
	suite.registration.AssertExpectations(suite.T())
}

func (suite *InitializerSuite) TestInitializeSavedStateActivationIdMatches() {
	// prepare
	state := newTestState()
	serializedState, err := state.Serialize()
	suite.NoError(err)
	suite.secretPersistence.On("Load").Return(serializedState, nil)
	suite.fsPersistence.On("Save", serializedState).Return(nil)

	// test
	actualErr := suite.initializer.Initialize()

	// verify
	suite.NoError(actualErr)
	suite.secretPersistence.AssertExpectations(suite.T())
	suite.fsPersistence.AssertExpectations(suite.T())
	suite.registration.AssertExpectations(suite.T())
}

func (suite *InitializerSuite) TestInitializeSavedStateCannotDeserialize() {
	// prepare
	testState := newTestState()
	serializedState, err := testState.Serialize()
	suite.NoError(err)
	serializedState[state.EksConnectorConfig] = "{'not a valid json"
	suite.secretPersistence.On("Load").Return(serializedState, nil)

	// test
	err = suite.initializer.Initialize()

	// verify
	suite.Error(err)
	suite.secretPersistence.AssertExpectations(suite.T())
	suite.fsPersistence.AssertExpectations(suite.T())
	suite.registration.AssertExpectations(suite.T())
}

func (suite *InitializerSuite) TestInitializeSavedStateActivationIdEmpty() {
	// prepare
	state := newTestState()
	state.ActivationId = ""
	serializedState, err := state.Serialize()
	suite.NoError(err)
	suite.secretPersistence.On("Load").Return(serializedState, nil)
	suite.fsPersistence.On("Save", serializedState).Return(nil)

	// test
	actualErr := suite.initializer.Initialize()

	// verify
	suite.NoError(actualErr)
	suite.secretPersistence.AssertExpectations(suite.T())
	suite.fsPersistence.AssertExpectations(suite.T())
	suite.registration.AssertExpectations(suite.T())
}

func (suite *InitializerSuite) TestInitializeSavedStateFailedFSPersistence() {
	// prepare
	state := newTestState()
	serializedState, err := state.Serialize()
	suite.NoError(err)
	err = errors.New("failed to persist")
	suite.secretPersistence.On("Load").Return(serializedState, nil)
	suite.fsPersistence.On("Save", serializedState).Return(err)

	// test
	actualErr := suite.initializer.Initialize()

	// verify
	suite.ErrorIs(actualErr, err)
	suite.secretPersistence.AssertExpectations(suite.T())
	suite.fsPersistence.AssertExpectations(suite.T())
	suite.registration.AssertExpectations(suite.T())
}

func (suite *InitializerSuite) TestInitializeSavedStateHappyActivationIdMismatch() {
	// prepare
	state := newTestState()
	state.ActivationId = testActivationId2
	serializedState, err := state.Serialize()
	suite.NoError(err)
	suite.secretPersistence.On("Load").Return(serializedState, nil)
	suite.registration.On("Register").Return(state, nil)
	suite.secretPersistence.On("Save", serializedState).Return(nil)
	suite.fsPersistence.On("Save", serializedState).Return(nil)

	// test
	actualErr := suite.initializer.Initialize()

	// verify
	suite.NoError(actualErr)
	suite.secretPersistence.AssertExpectations(suite.T())
	suite.fsPersistence.AssertExpectations(suite.T())
	suite.registration.AssertExpectations(suite.T())
}

func newTestState() *state.State {
	return &state.State{
		ActivationId:          testActivationId,
		FingerPrint:           testFingerPrint,
		InstanceID:            testInstanceID,
		PrivateKey:            testPrivateKey,
		PrivateKeyType:        testPrivateKeyType,
		PrivateKeyCreatedDate: testPrivateKeyCreatedDate,
		Region:                testRegion,
	}
}
