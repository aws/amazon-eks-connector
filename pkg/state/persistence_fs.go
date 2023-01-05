package state

import (
	"os"
	"path"
	"path/filepath"

	"github.com/aws/amazon-eks-connector/pkg/config"
)

type FileSystemPersistence struct {
	stateConfig *config.StateConfig
}

func NewFileSystemPersistence(stateConfig *config.StateConfig) Persistence {
	return &FileSystemPersistence{
		stateConfig: stateConfig,
	}
}

func (p *FileSystemPersistence) Load() (SerializedState, error) {
	state := SerializedState{}

	err := p.readFileIntoState(state, FileManifest)
	if err != nil {
		return nil, err
	}
	err = p.readFileIntoState(state, FileInstanceFingerprint)
	if err != nil {
		return nil, err
	}
	err = p.readFileIntoState(state, FileRegistrationKey)
	if err != nil {
		return nil, err
	}
	return state, nil
}
func (p *FileSystemPersistence) Save(state SerializedState) error {
	err := p.writeFileFromState(state, FileManifest)
	if err != nil {
		return err
	}
	err = p.writeFileFromState(state, FileInstanceFingerprint)
	if err != nil {
		return err
	}
	return p.writeFileFromState(state, FileRegistrationKey)
}

func (p *FileSystemPersistence) readFileIntoState(state SerializedState, file string) error {
	filePath := path.Join(p.stateConfig.BaseDir, file)
	data, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}
	state[file] = string(data)
	return nil
}

func (p *FileSystemPersistence) writeFileFromState(state SerializedState, file string) error {
	filePath := path.Join(p.stateConfig.BaseDir, file)
	err := os.MkdirAll(filepath.Dir(filePath), 0700)
	if err != nil {
		return err
	}
	return os.WriteFile(filePath, []byte(state[file]), 0600)
}
