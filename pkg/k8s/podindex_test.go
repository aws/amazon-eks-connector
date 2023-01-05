package k8s

import (
	"os"
	"testing"

	"github.com/stretchr/testify/suite"
)

func TestPodIndexSuite(t *testing.T) {
	suite.Run(t, new(PodIndexSuite))
}

type PodIndexSuite struct {
	suite.Suite
	provider PodIndexProvider
}

func (suite *PodIndexSuite) SetupTest() {
	suite.provider = NewPodIndexProvider()
}

func (suite *PodIndexSuite) TearDownTest() {
	err := os.Unsetenv(EnvPodName)
	suite.NoError(err)
}

func (suite *PodIndexSuite) TestGetNoEnvVar() {
	index, err := suite.provider.Get()

	suite.Error(err)
	suite.Empty(index)
}

func (suite *PodIndexSuite) TestGetBadPodName() {
	badPodNames := []string{
		// ends with -
		"foo-",
		// empty
		"",
		// no -
		"asdf",
	}

	for _, podName := range badPodNames {
		os.Setenv(EnvPodName, podName)

		index, err := suite.provider.Get()

		suite.Error(err)
		suite.Empty(index)
	}
}

func (suite *PodIndexSuite) TestGetStatefulSetName() {
	podNames := map[string]string{
		"foo-bar":         "bar",
		"foo-bar-dar":     "dar",
		"eks-connector-0": "0",
	}

	for podName, expectedIndex := range podNames {
		os.Setenv(EnvPodName, podName)

		index, err := suite.provider.Get()

		suite.NoError(err)
		suite.Equal(expectedIndex, index)
	}
}
