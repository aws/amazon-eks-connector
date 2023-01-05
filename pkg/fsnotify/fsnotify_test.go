package fsnotify

import (
	"errors"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/suite"

	"github.com/aws/amazon-eks-connector/pkg/state"
)

func TestFSNotifySuite(t *testing.T) {
	suite.Run(t, new(FSNotifySuite))
}

type FSNotifySuite struct {
	suite.Suite
	secretPersistence *state.MockPersistence
	fsPersistence     *state.MockPersistence

	fsNotify fsWatchProvider
}

func (suite *FSNotifySuite) SetupTest() {
	suite.secretPersistence = &state.MockPersistence{}
	suite.fsPersistence = &state.MockPersistence{}

	suite.fsNotify = fsWatchProvider{
		secretPersistence: suite.secretPersistence,
		fsPersistence:     suite.fsPersistence,
		viper:             viper.New(),
	}
}

func (suite *FSNotifySuite) TestFSNoUpdateHappyCase() {
	// prepare
	serializedState := getSerializedState("testPrivateKey", "")
	suite.secretPersistence.On("Load").Return(serializedState, nil)
	suite.fsPersistence.On("Load").Return(serializedState, nil)

	// test
	isSuccess, actualErr := suite.fsNotify.SyncSecrets()

	// verify
	suite.True(isSuccess)
	suite.NoError(actualErr)
	suite.fsPersistence.AssertExpectations(suite.T())
	suite.secretPersistence.AssertExpectations(suite.T())
}

func (suite *FSNotifySuite) TestFSSavedStateHappyCase() {
	// prepare
	newSerializedState := getSerializedState("newtestPrivateKey", "")
	existingState := getSerializedState("oldtestPrivateKey", "oldActivationId")
	// expected state is a merge of new state plus the old activationId
	expectedState := getSerializedState("newtestPrivateKey", "oldActivationId")

	suite.secretPersistence.On("Load").Return(existingState, nil)
	suite.fsPersistence.On("Load").Return(newSerializedState, nil)
	suite.secretPersistence.On("Save", expectedState).Return(nil)

	// test
	isSuccess, actualErr := suite.fsNotify.SyncSecrets()

	// verify
	suite.True(isSuccess)
	suite.NoError(actualErr)
	suite.fsPersistence.AssertExpectations(suite.T())
	suite.secretPersistence.AssertExpectations(suite.T())
}

func (suite *FSNotifySuite) TestSyncSecretsFailureCase() {
	// prepare
	suite.secretPersistence.On("Load").Return(nil, errors.New("error"))

	// test
	isSuccess, actualErr := suite.fsNotify.SyncSecrets()

	// verify
	suite.False(isSuccess)
	suite.NoError(actualErr)
	suite.fsPersistence.AssertExpectations(suite.T())
	suite.secretPersistence.AssertExpectations(suite.T())
}

func (suite *FSNotifySuite) TestWatchConfigInvokesSyncSecrets() {
	// prepare
	existingState := getSerializedState("testPrivateKey", "")
	suite.secretPersistence.On("Load").Return(existingState, nil)
	suite.fsPersistence.On("Load").Return(existingState, nil)

	// test
	actualErr := suite.fsNotify.watchConfig()

	// verify
	suite.NoError(actualErr)
	suite.fsPersistence.AssertExpectations(suite.T())
	suite.secretPersistence.AssertExpectations(suite.T())
}

func getSerializedState(privateKey, activationId string) state.SerializedState {
	testState := &state.State{
		PrivateKey:   privateKey,
		ActivationId: activationId,
	}
	serializedState, _ := testState.Serialize()
	return serializedState
}
