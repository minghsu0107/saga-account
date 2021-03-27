package config

import (
	"os"

	"github.com/kelseyhightower/envconfig"
	"gopkg.in/yaml.v2"
)

// Config is a type for general configuration
type Config struct {
	HTTPPort  string     `yaml:"httpPort" envconfig:"HTTP_PORT"`
	GRPCPort  string     `yaml:"grpcPort" envconfig:"GRPC_PORT"`
	AppName   string     `yaml:"appName" envconfig:"APP_NAME"`
	GinMode   string     `yaml:"ginMode" envconfig:"GIN_MODE"`
	JWTConfig *JWTConfig `yaml:"jwtConfig"`
	DBConfig  *DBConfig  `yaml:"dbConfig"`
	Logger    *Logger
}

// JWTConfig is the type for jwt config
type JWTConfig struct {
	Secret                   string `yaml:"secret" envconfig:"JWT_SECRET"`
	AccessTokenExpireSecond  int64  `yaml:"accessTokenExpireSecond" envconfig:"JWT_ACCESS_TOKEN_EXPIRE_SECOND"`
	RefreshTokenExpireSecond int64  `yaml:"refreshTokenExpireSecond" envconfig:"JWT_REFRESH_TOKEN_EXPIRE_SECOND"`
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
