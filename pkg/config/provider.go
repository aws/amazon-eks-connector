package config

import "github.com/spf13/viper"

type Provider interface {
	Get() (*Config, error)
}

func NewProvider() Provider {
	return &viperProvider{}
}

type viperProvider struct {
}

func (p *viperProvider) Get() (*Config, error) {
	cfg := &Config{}
	err := viper.Unmarshal(cfg)
	if err != nil {
		return nil, err
	}
	return cfg, nil
}
