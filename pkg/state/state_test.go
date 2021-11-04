package state

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

func TestStateSuite(t *testing.T) {
	suite.Run(t, new(StateSuite))
}

type StateSuite struct {
	suite.Suite
}

func (suite *StateSuite) TestSerializeAndDeserialize() {
	expectedState := testState()
	serializedState, err := expectedState.Serialize()
	suite.NoError(err)

	actualState, err := Deserialize(serializedState)
	suite.NoError(err)

	suite.Equal(expectedState, actualState)
}

func (suite *StateSuite) TestDeserializeNoEksConnectorConfig() {
	expectedState := testState()
	expectedState.ActivationId = ""
	serializedState, err := expectedState.Serialize()
	suite.NoError(err)
	delete(serializedState, EksConnectorConfig)
	suite.Empty(serializedState[EksConnectorConfig])

	actualState, err := Deserialize(serializedState)

	suite.NoError(err)
	suite.Empty(actualState.ActivationId)
	suite.Equal(expectedState, actualState)
}

func testState() *State {
	return &State{
		ActivationId:          "f4423803-dd4a-4994-8fcd-b7d6105b3c43",
		FingerPrint:           "49e29d36-9096-4fe4-bb65-0357e67cdc70",
		InstanceID:            "eks_c:my-cluster_e13b1367a2b4",
		PrivateKey:            "a3ViZXJuZXRlcyBpcyBhd2Vzb21l",
		PrivateKeyType:        "plaintext",
		PrivateKeyCreatedDate: "2021-10-05 05:27:47.693369915 +0000 UTC",
		Region:                "mars-east-1",
	}
}
