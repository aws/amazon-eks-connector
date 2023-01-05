package k8s

import (
	"testing"

	"github.com/stretchr/testify/suite"
	"k8s.io/client-go/kubernetes/fake"
	k8sTesting "k8s.io/client-go/testing"
)

const (
	testSecretName      = "top"
	testSecretNamespace = "eks-connector-unit-test"
)

func TestSecretSuite(t *testing.T) {
	suite.Run(t, new(SecretSuite))
}

type SecretSuite struct {
	suite.Suite

	k8sClient *fake.Clientset
	secret    Secret
}

func (suite *SecretSuite) SetupTest() {
	suite.k8sClient = fake.NewSimpleClientset()
	suite.secret = NewSecret(testSecretName, testSecretNamespace, suite.k8sClient)
}

func (suite *SecretSuite) TestGetNotFound() {
	data, err := suite.secret.Get()

	suite.NoError(err)
	suite.Nil(data)
	actions := suite.k8sClient.Actions()
	suite.Len(actions, 1)
	suite.assertSecretAPICall(actions[0], "get")
}

func (suite *SecretSuite) TestSetAndGet() {
	data := map[string][]byte{
		"Key1": []byte("Data1"),
		"Key2": []byte("Data2"),
	}

	err := suite.secret.Put(data)
	suite.NoError(err)
	savedData, err := suite.secret.Get()
	suite.NoError(err)
	suite.Equal(data, savedData)
	actions := suite.k8sClient.Actions()
	suite.Len(actions, 3)
	suite.assertSecretAPICall(actions[0], "get")
	suite.assertSecretAPICall(actions[1], "create")
	suite.assertSecretAPICall(actions[2], "get")
}

func (suite *SecretSuite) TestSetPreexistingSecret() {
	// prepare
	data1 := map[string][]byte{
		"Key1": []byte("Data1"),
	}
	err := suite.secret.Put(data1)
	suite.NoError(err)

	// test
	data2 := map[string][]byte{
		"Key2": []byte("Data2"),
	}
	err = suite.secret.Put(data2)
	suite.NoError(err)

	// verify
	savedData, err := suite.secret.Get()
	suite.NoError(err)
	suite.Equal(data2, savedData)

	actions := suite.k8sClient.Actions()
	suite.Len(actions, 5)
	// first Save, get and create
	suite.assertSecretAPICall(actions[0], "get")
	suite.assertSecretAPICall(actions[1], "create")
	// second Save, get and update
	suite.assertSecretAPICall(actions[2], "get")
	suite.assertSecretAPICall(actions[3], "update")
	// final get
	suite.assertSecretAPICall(actions[4], "get")
}

func (suite *SecretSuite) assertSecretAPICall(action k8sTesting.Action, verb string) {
	suite.Equal(testSecretNamespace, action.GetNamespace())
	suite.Equal("secrets", action.GetResource().Resource)
	suite.Equal("v1", action.GetResource().Version)
	suite.Equal(verb, action.GetVerb())
	suite.Equal("", action.GetResource().Group)
}
