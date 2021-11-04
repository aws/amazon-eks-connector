// Package config contains structs to hold eks connector configurations
package config

// Config is the whole configuration of eks-connector.
type Config struct {
	AgentConfig      *AgentConfig      `mapstructure:"agent"`
	ProxyConfig      *ProxyConfig      `mapstructure:"proxy"`
	WatcherConfig    *WatcherConfig    `mapstructure:"watcher"`
	ActivationConfig *ActivationConfig `mapstructure:"activation"`
	StateConfig      *StateConfig      `mapstructure:"state"`
}

type SocketType string

const (
	TCP  SocketType = "tcp"
	Unix SocketType = "unix"
)

// AgentConfig is the sub-configuration for ssm agent.
type AgentConfig struct {
	// Region is the AWS region code that ssm agent will connect to.
	Region string `mapstructure:"region"`

	// Endpoint is the endpoint that ssm agent will connect to.
	// If not set, ssm agent will connect to the default endpoint of the region.
	Endpoint string `mapstructure:"endpoint"`
}

// ProxyConfig is the sub-configuration for api server proxy.
type ProxyConfig struct {
	SocketType    SocketType `mapstructure:"socketType"`
	SocketAddress string     `mapstructure:"socketAddr"`

	TargetHost     string `mapstructure:"targetHost"`
	TargetProtocol string `mapstructure:"targetProtocol"`
}

// WatcherConfig is the sub-configuration for ssm agent watcher.
type WatcherConfig struct {
}

// ActivationConfig is the sub-configuration for ssm agent activation.
type ActivationConfig struct {
	Code string `mapstructure:"code"`
	ID   string `mapstructure:"id"`
}

// StateConfig is the sub-configuration for storing states of EKS connector.
type StateConfig struct {
	// BaseDir is the SSM agent Vault dir that contains SSM agent state.
	BaseDir string `mapstructure:"baseDir"`
	// SecretNamePrefix is the prefix of secret name that contains EKS connector state.
	// EKS connector Pod ordinal index in StatefulSet is appended.
	SecretNamePrefix string `mapstructure:"secretNamePrefix"`
	// SecretNamespace is the namespace of secret that container EKS connector state.
	SecretNamespace string `mapstructure:"secretNamespace"`
}
