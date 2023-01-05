// Package initializer contains init container related functionalities.
package initializer

import (
	"k8s.io/klog/v2"

	"github.com/aws/amazon-eks-connector/pkg/agent"
	"github.com/aws/amazon-eks-connector/pkg/config"
	"github.com/aws/amazon-eks-connector/pkg/state"
)

type Initializer interface {
	Initialize() error
}

func NewInitializer(
	activationConfig *config.ActivationConfig,
	secretPersistence state.Persistence,
	fsPersistence state.Persistence,
	registration agent.Registration) Initializer {
	return &ssmInitializer{
		activationConfig:  activationConfig,
		secretPersistence: secretPersistence,
		fsPersistence:     fsPersistence,
		registration:      registration,
	}
}

type ssmInitializer struct {
	activationConfig  *config.ActivationConfig
	secretPersistence state.Persistence
	fsPersistence     state.Persistence
	registration      agent.Registration
}

func (i *ssmInitializer) Initialize() error {
	klog.Infof("eks-connector initializer starts...")

	klog.Infof("loading persisted state from secrets...")
	serializedSecret, err := i.loadPreviousState()
	if err != nil {
		return err
	}

	if serializedSecret == nil {
		klog.Infof("registering as new instance")
		connectorState, err := i.registration.Register()
		if err != nil {
			return err
		}

		klog.Infof("serializing state information...")
		serializedSecret, err = connectorState.Serialize()
		if err != nil {
			return err
		}

		klog.Infof("persisting state information to secrets...")
		err = i.secretPersistence.Save(serializedSecret)
		if err != nil {
			return err
		}
	}

	klog.Infof("persisting state information to filesystem...")
	err = i.fsPersistence.Save(serializedSecret)

	return err
}

func (i *ssmInitializer) loadPreviousState() (state.SerializedState, error) {
	serializedSecret, err := i.secretPersistence.Load()
	if err != nil {
		return nil, err
	}
	if serializedSecret == nil {
		klog.Infof("eks connector state is not found in persistent store, performing new activation...")
		return nil, nil
	}
	klog.Infof("eks connector state is found in persistent store")
	connectorState, err := state.Deserialize(serializedSecret)
	if err != nil {
		klog.Errorf("eks connector state cannot be deserialized")
		return nil, err
	}
	if connectorState.ActivationId != "" {
		if connectorState.ActivationId != i.activationConfig.ID {
			klog.Warningf("ssm activation id mismatch! state: %s, config: %s", connectorState.ActivationId, i.activationConfig.ID)
			klog.Warningf("eks connector is discarding previous state and performing new activation...")
			return nil, nil
		}
	} else {
		klog.Warningf("ssm activation id is not available, state might be created by an earlier version of eks-connector")
	}

	klog.Infof("eks connector is inheriting previous state...")
	return serializedSecret, nil
}
