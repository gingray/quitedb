package config

import (
	"github.com/caarlos0/env/v11"
)

type Config struct {
	HTTPServiceConfig
	StorageConfig
}

type HTTPServiceConfig struct {
	Port int `env:"HTTP_PORT" envDefault:"3000"`
}

type StorageConfig struct {
	StoragePath string `env:"STORAGE_PATH" envDefault:"./storage"`
}

func NewConfig() (*Config, error) {
	var cfg Config
	err := env.Parse(&cfg)
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}
