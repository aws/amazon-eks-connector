package config

import "github.com/spf13/viper"

type Provider interface {
	Get() (*Config, error)
}

func NewProvider(v *viper.Viper) Provider {
	return &viperProvider{viperFlag: v}
}

type viperProvider struct {
	viperFlag *viper.Viper
}

func (p *viperProvider) Get() (*Config, error) {
	cfg := &Config{}
	err := p.viperFlag.Unmarshal(cfg)
	if err != nil {
		return nil, err
	}
	return cfg, nil
}
