// Package state provides types and functions for eks connector state management
package state

import (
	"encoding/json"
)

type State struct {
	ActivationId          string
	FingerPrint           string
	InstanceID            string
	PrivateKey            string
	PrivateKeyType        string
	PrivateKeyCreatedDate string
	Region                string
}

func Deserialize(serializedState SerializedState) (state *State, err error) {
	fingerprint := &instanceFingerprintState{}
	if err = unmarshalIfNotEmpty(serializedState[FileInstanceFingerprint], fingerprint); err != nil {
		return nil, err
	}
	regKey := &registrationKeyState{}
	if err = unmarshalIfNotEmpty(serializedState[FileRegistrationKey], regKey); err != nil {
		return nil, err
	}
	connectorConfig := &eksConnectorConfig{}
	if err = unmarshalIfNotEmpty(serializedState[EksConnectorConfig], connectorConfig); err != nil {
		return nil, err
	}
	state = &State{}

	state.ActivationId = connectorConfig.ActivationId
	state.FingerPrint = fingerprint.Fingerprint
	state.InstanceID = regKey.InstanceID
	state.PrivateKey = regKey.PrivateKey
	state.PrivateKeyType = regKey.PrivateKeyType
	state.PrivateKeyCreatedDate = regKey.PrivateKeyCreatedDate
	state.Region = regKey.Region
	return state, nil
}

func (state *State) Serialize() (serializedState SerializedState, err error) {
	serializedState = SerializedState{}

	serializedState[FileManifest], err = state.serializeManifest()
	if err != nil {
		return
	}

	serializedState[FileRegistrationKey], err = state.serializeRegistrationKey()
	if err != nil {
		return
	}

	serializedState[FileInstanceFingerprint], err = state.serializeInstanceFingerprint()
	if err != nil {
		return
	}

	serializedState[EksConnectorConfig], err = state.serializeEksConnectorConfig()
	if err != nil {
		return
	}

	return serializedState, nil
}

func (state *State) serializeEksConnectorConfig() (string, error) {
	config := &eksConnectorConfig{
		ActivationId: state.ActivationId,
	}
	return marshal(config)
}

func (state *State) serializeManifest() (string, error) {
	// Manifest is a static file.
	manifest := &manifestState{
		InstanceFingerprint: "/var/lib/amazon/ssm/Vault/Store/InstanceFingerprint",
		RegistrationKey:     "/var/lib/amazon/ssm/Vault/Store/RegistrationKey",
	}
	return marshal(manifest)
}

func (state *State) serializeInstanceFingerprint() (string, error) {
	fingerprint := &instanceFingerprintState{}
	fingerprint.Fingerprint = state.FingerPrint
	fingerprint.HardwareHash = make(map[string]string)
	fingerprint.SimilarityThreshold = -1

	return marshal(fingerprint)
}

func (state *State) serializeRegistrationKey() (string, error) {
	registrationKey := &registrationKeyState{}

	registrationKey.PrivateKey = state.PrivateKey
	registrationKey.PrivateKeyType = state.PrivateKeyType
	registrationKey.Region = state.Region
	registrationKey.PrivateKeyCreatedDate = state.PrivateKeyCreatedDate
	registrationKey.InstanceID = state.InstanceID
	registrationKey.AvailabilityZone = ""
	registrationKey.InstanceType = ""

	return marshal(registrationKey)
}

func marshal(v interface{}) (string, error) {
	data, err := json.Marshal(v)
	if err != nil {
		return "", nil
	}
	return string(data), nil
}

func unmarshalIfNotEmpty(text string, v interface{}) error {
	if text == "" {
		return nil
	}
	return json.Unmarshal([]byte(text), v)
}
