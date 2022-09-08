package state

import (
	"os"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/aws/amazon-eks-connector/pkg/config"
)

const (
	testFSStateFileManifest        = "manifest file content"
	testFSStateFileFingerPrint     = "finger print file content"
	testFSStateFileRegistrationKey = "registration key file content"
)

func TestFileSystemPersistenceSuite(t *testing.T) {
	suite.Run(t, new(FileSystemPersistenceSuite))
}

type FileSystemPersistenceSuite struct {
	suite.Suite

	dirName     string
	persistence Persistence
}

func (suite *FileSystemPersistenceSuite) TestSaveAndLoad() {
	state := SerializedState{
		FileManifest:            testFSStateFileManifest,
		FileRegistrationKey:     testFSStateFileRegistrationKey,
		FileInstanceFingerprint: testFSStateFileFingerPrint,
	}

	err := suite.persistence.Save(state)
	suite.NoError(err)
	loadedState, err := suite.persistence.Load()
	suite.NoError(err)
	suite.Equal(state, loadedState)
}

func (suite *FileSystemPersistenceSuite) TestSaveAndLoadMissingManifest() {
	state := SerializedState{
		FileRegistrationKey:     testFSStateFileRegistrationKey,
		FileInstanceFingerprint: testFSStateFileFingerPrint,
		FileManifest:            "",
	}

	err := suite.persistence.Save(state)
	suite.NoError(err)
	loadedState, err := suite.persistence.Load()
	suite.NoError(err)
	suite.Equal(state, loadedState)
}

func (suite *FileSystemPersistenceSuite) SetupTest() {
	dir, err := os.MkdirTemp("", "eks_connector_vault")
	suite.NoError(err)
	suite.dirName = dir
	suite.persistence = NewFileSystemPersistence(&config.StateConfig{
		BaseDir: suite.dirName,
	})
}

func (suite *FileSystemPersistenceSuite) TearDownTest() {
	err := os.RemoveAll(suite.dirName)
	suite.NoError(err)
}
