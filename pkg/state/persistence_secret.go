package state

import (
	"github.com/aws/amazon-eks-connector/pkg/k8s"
)

const (
	SecretKeyManifest        = "manifest"
	SecretKeyRegistrationKey = "regkey"
	SecretKeyFingerprint     = "fingerprint"
	SecretKeyConnectorConfig = "connector-config"
)

type SecretPersistence struct {
	secret k8s.Secret
}

func NewSecretPersistence(secret k8s.Secret) Persistence {
	return &SecretPersistence{
		secret: secret,
	}
}

func (p *SecretPersistence) Load() (state SerializedState, err error) {
	state = SerializedState{}
	data, err := p.secret.Get()
	if err != nil {
		return
	}
	state = secretToState(data)
	return
}

func (p *SecretPersistence) Save(state SerializedState) (err error) {
	data := stateToSecret(state)
	err = p.secret.Put(data)
	return
}

func stateToSecret(state SerializedState) map[string][]byte {
	return map[string][]byte{
		SecretKeyManifest:        []byte(state[FileManifest]),
		SecretKeyFingerprint:     []byte(state[FileInstanceFingerprint]),
		SecretKeyRegistrationKey: []byte(state[FileRegistrationKey]),
		SecretKeyConnectorConfig: []byte(state[EksConnectorConfig]),
	}
}

func secretToState(secret map[string][]byte) SerializedState {
	if len(secret) == 0 {
		return nil
	}
	return SerializedState{
		FileManifest:            string(secret[SecretKeyManifest]),
		FileInstanceFingerprint: string(secret[SecretKeyFingerprint]),
		FileRegistrationKey:     string(secret[SecretKeyRegistrationKey]),
		EksConnectorConfig:      string(secret[SecretKeyConnectorConfig]),
	}
}
