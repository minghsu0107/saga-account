package config

import (
	"os"

	"github.com/kelseyhightower/envconfig"
	"gopkg.in/yaml.v2"
)

// Config is a type for general configuration
type Config struct {
	HTTPPort  string    `yaml:"httpPort" envconfig:"HTTP_PORT"`
	GRPCPort  string    `yaml:"grpcPort" envconfig:"GRPC_PORT"`
	AppName   string    `yaml:"appName" envconfig:"APP_NAME"`
	GinMode   string    `yaml:"ginMode" envconfig:"GIN_MODE"`
	JWTSecret string    `yaml:"jwtSecret" envconfig:"JWT_SECRET"`
	DBConfig  *DBConfig `yaml:"dbConfig"`
	Logger    *Logger
}

// DBConfig is the type for database config
type DBConfig struct {
	Dsn          string `yaml:"dsn" envconfig:"DB_DSN"`
	MaxIdleConns int    `yaml:"maxIdleConns" envconfig:"DB_MAX_IDLE_CONNS"`
	MaxOpenConns int    `yaml:"maxOpenConns" envconfig:"DB_MAX_OPEN_CONNS"`
}

// NewConfig is a factory for Config instance
func NewConfig() (*Config, error) {
	var config Config
	if err := readFile(&config); err != nil {
		return nil, err
	}
	if err := readEnv(&config); err != nil {
		return nil, err
	}
	config.Logger = newLogger(config.AppName, config.GinMode)
	return &config, nil
}

func readFile(config *Config) error {
	f, err := os.Open("config.yml")
	if err != nil {
		return err
	}
	defer f.Close()

	decoder := yaml.NewDecoder(f)
	err = decoder.Decode(config)
	if err != nil {
		return err
	}
	return nil
}

func readEnv(config *Config) error {
	err := envconfig.Process("", config)
	if err != nil {
		return err
	}
	return nil
}
