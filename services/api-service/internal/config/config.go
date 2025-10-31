package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

type Config struct {
	Env      string         `mapstructure:"env"`
	Server   ServerConfig   `mapstructure:"server"`
	Postgres PostgresConfig `mapstructure:"postgres"`
	GRPC     GRPCConfig     `mapstructure:"grpc"`
}

type ServerConfig struct {
	Port        string        `mapstructure:"port"`
	Timeout     time.Duration `mapstructure:"timeout"`
	IdleTimeout time.Duration `mapstructure:"idle_timeout"`
}

type PostgresConfig struct {
	Host     string     `mapstructure:"host"`
	Port     string     `mapstructure:"port"`
	User     string     `mapstructure:"user"`
	Password string     `mapstructure:"password"`
	DBName   string     `mapstructure:"dbname"`
	SSLMode  string     `mapstructure:"sslmode"`
	Pool     PoolConfig `mapstructure:"pool"`
}

type PoolConfig struct {
	MaxOpenConns    int           `mapstructure:"max_open_conns"`
	MaxIdleConns    int           `mapstructure:"max_idle_conns"`
	ConnMaxLifetime time.Duration `mapstructure:"conn_max_lifetime"`
	ConnMaxIdleTime time.Duration `mapstrucutre:"conn_max_idle_time"`
}

type GRPCConfig struct {
	Port string `mapstructure:"port"`
}

func LoadConfig(path string) (*Config, error) {
	viper.SetConfigFile(path)

	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config

	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &cfg, nil
}

func WatchConfig(onChange func(e fsnotify.Event)) {
	viper.WatchConfig()
	viper.OnConfigChange(onChange)
}
