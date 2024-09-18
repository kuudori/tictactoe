package config

import (
	"fmt"
	"github.com/ilyakaznacheev/cleanenv"
	"time"
)

type Config struct {
	Env  string `yaml:"env" env-default:"local"`
	GRPC GRPCConfig
}

type GRPCConfig struct {
	Port    int           `yaml:"port"`
	Timeout time.Duration `yaml:"timeout"`
}

var (
	instance *Config
)

func LoadConfig() (*Config, error) {
	cfg := &Config{}

	err := cleanenv.ReadConfig("env.yaml", cfg)
	if err != nil {
		return nil, fmt.Errorf("error reading config file: %w", err)
	}

	err = cleanenv.ReadEnv(cfg)
	if err != nil {
		return nil, fmt.Errorf("error reading environment variables: %w", err)
	}
	instance = cfg

	return cfg, nil
}

func GetConfig() *Config {
	return instance
}
