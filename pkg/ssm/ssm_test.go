package ssm

import (
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/aws/amazon-eks-connector/pkg/config"
)

const (
	testActivationID   = "1897a40d-ce57-42f3-8228-25aeaf9dc4f1"
	testActivationCode = "testActivationCode"
	testInstanceID     = "166440b2-5829-4e85-8b4c-248098e4c3be"
	testFingerPrint    = "7a68a33a-7a1e-4db3-bd0c-581e5e252319"
	testPublicKey      = "335MFSysbgFKC9jQtuVQ"
	testPublicKeyType  = "plaintext"
	testRegion         = "mars-northeast-2"
)

func TestAnonymousServiceSuite(t *testing.T) {
	suite.Run(t, new(AnonymousServiceSuite))
}

type AnonymousServiceSuite struct {
	suite.Suite

	requester *MockAWSRequester
	request   *MockAWSRequest
	ssm       Client
}

func (suite *AnonymousServiceSuite) SetupTest() {
	suite.requester = &MockAWSRequester{}
	suite.request = &MockAWSRequest{}
	suite.ssm = &sdkClient{
		agentConfig: &config.AgentConfig{
			Region: testRegion,
		},
		sdk: suite.requester,
	}
}

func (suite *AnonymousServiceSuite) TestRegisterManagedInstance() {
	// prepare
	expectedOperation := NewExpectedOperation()
	expectedParam := NewExpectedParam()
	suite.request.On("Send").Return(nil)
	suite.requester.On("NewRequest", expectedOperation, expectedParam,
		mock.AnythingOfType("*ssm.registerManagedInstanceOutput")).
		Run(func(args mock.Arguments) {
			instance := testInstanceID
			output := args[2].(*registerManagedInstanceOutput)
			output.InstanceId = &instance
		}).
		Return(suite.request)

	// test
	instanceID, err := suite.ssm.RegisterManagedInstance(testActivationID, testActivationCode, testPublicKey, testPublicKeyType, testFingerPrint)

	// verify
	suite.NoError(err)
	suite.Equal(testInstanceID, instanceID)
	suite.requester.AssertExpectations(suite.T())
	suite.request.AssertExpectations(suite.T())
}

func (suite *AnonymousServiceSuite) TestRegisterManagedInstanceError() {
	// prepare
	expectedOperation := NewExpectedOperation()
	expectedParam := NewExpectedParam()
	suite.request.On("Send").Return(errors.New("AWS service error"))
	suite.requester.On("NewRequest", expectedOperation, expectedParam,
		mock.AnythingOfType("*ssm.registerManagedInstanceOutput")).
		Return(suite.request)

	// test
	_, err := suite.ssm.RegisterManagedInstance(testActivationID, testActivationCode, testPublicKey, testPublicKeyType, testFingerPrint)

	// verify
	suite.Error(err)
	suite.requester.AssertExpectations(suite.T())
	suite.request.AssertExpectations(suite.T())
}

func (suite *AnonymousServiceSuite) TestRegion() {
	region := suite.ssm.Region()

	suite.Equal(testRegion, region)
}

func NewExpectedParam() *registerManagedInstanceInput {
	return &registerManagedInstanceInput{
		ActivationId:   aws.String(testActivationID),
		ActivationCode: aws.String(testActivationCode),
		PublicKey:      aws.String(testPublicKey),
		PublicKeyType:  aws.String(testPublicKeyType),
		Fingerprint:    aws.String(testFingerPrint),
	}
}

func NewExpectedOperation() *request.Operation {
	return &request.Operation{
		Name:       operationRegisterManagedInstance,
		HTTPMethod: methodPost,
		HTTPPath:   "/",
	}
}
