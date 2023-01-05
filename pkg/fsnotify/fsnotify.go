package fsnotify

import (
	"path"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/klog/v2"

	"github.com/aws/amazon-eks-connector/pkg/config"
	"github.com/aws/amazon-eks-connector/pkg/k8s"
	"github.com/aws/amazon-eks-connector/pkg/state"
)

// fsWatchProvider monitors local filesystem changes and updates Kubernetes secrets.
type fsWatchProvider struct {
	sync.RWMutex
	viper             *viper.Viper
	configFilePath    string
	fsPersistence     state.Persistence
	secretPersistence state.Persistence
}

// backoff, retries upto one minute
var backoff = wait.Backoff{
	Duration: 1 * time.Second,
	Factor:   2,
	Steps:    7,
}

// NewWatcher initiates fsWatchProvider to monitor SSM agent's key pair file
func NewWatcher(stateConfig *config.StateConfig) error {
	secret, err := k8s.NewSecretInCluster(stateConfig)
	if err != nil {
		return errors.Wrap(err, "could not read secrets when initializing fs watcher")
	}

	provider := &fsWatchProvider{
		viper:             viper.New(),
		configFilePath:    getConfigFilePath(stateConfig.BaseDir),
		secretPersistence: state.NewSecretPersistence(secret),
		fsPersistence:     state.NewFileSystemPersistence(stateConfig),
	}

	return provider.watchConfig()
}

// getConfigFilePath returns absolute path of RegistrationKey file
func getConfigFilePath(baseDir string) string {
	return path.Join(baseDir, state.FileRegistrationKey)
}

// watchConfig uses viper's fsnotify() to monitor agent files and update Kubernetes secret
func (fs *fsWatchProvider) watchConfig() error {
	fs.viper.SetConfigFile(fs.configFilePath)
	fs.viper.SetConfigType("json") // required when filename doesn't have any extension.

	// perform sync during container restart
	if err := wait.ExponentialBackoff(backoff, fs.SyncSecrets); err != nil {
		return errors.Wrap(err, "could not sync K8s secrets when initializing fs watcher")
	}

	fs.viper.WatchConfig()
	fs.viper.OnConfigChange(func(event fsnotify.Event) {
		klog.Infof("Changes received for config file %s. Operation: %s.", fs.configFilePath, event.Op)
		if err := wait.ExponentialBackoff(backoff, fs.SyncSecrets); err != nil {
			// TODO: Should kill process here?
			// If k8s secret is not updated then subsequent new ssm-agent containers will not be able to authenticate
			// with ssm backend service. Other option is to add pod event but failure here could be most probably
			// because of connectivity issue with APIServer.
			klog.Errorf("Failed to process updates for %s: %v", fs.configFilePath, err)
			return
		}
		klog.V(2).Infof("successfully updated K8s secret")
	})
	return nil
}

// SyncSecrets syncs local file content with K8s secret. Return value indicates whether ExponentialBackoff()
// should retry the operation or not.
func (fs *fsWatchProvider) SyncSecrets() (bool, error) {
	fs.Lock()
	defer fs.Unlock()

	existingState, err := fs.secretPersistence.Load()
	if err != nil {
		klog.Errorf("failed to load Kubernetes secret due to %v", err)
		return false, nil
	}

	newState, err := fs.fsPersistence.Load()
	if err != nil {
		klog.Errorf("failed to load agent's local file due to %v", err)
		return false, nil
	}

	if existingState[state.FileRegistrationKey] == newState[state.FileRegistrationKey] {
		// if K8s secrets and file content are same then don't perform any operation
		klog.Infof("Skip updating k8s secrets since key-pair did not change.")
		return true, nil
	}

	mergeState(existingState, newState)

	if err = fs.secretPersistence.Save(newState); err != nil {
		klog.Errorf("failed to save secret due to %v", err)
		return false, nil
	}
	klog.Infof("Updated kubernetes secrets with new key-pair")

	return true, nil
}

func mergeState(preexistingState, newState state.SerializedState) {
	// inherit EksConnectorConfig content since FS persistence does not have the information.
	newState[state.EksConnectorConfig] = preexistingState[state.EksConnectorConfig]
}
