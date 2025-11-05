package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

type Config struct {
	Env     string       `mapstructure:"env"`
	Mongo   MongoConfig  `mapstructure:"mongo"`
	Urls    UrlsConfig   `mapstructure:"urls"`
	Metrics []MetricInfo `mapstructure:"metrics"`
	Server  ServerConfig `mapstructure:"server"`
}

type ServerConfig struct {
	Port        string        `mapstructure:"port"`
	Timeout     time.Duration `mapstructure:"timeout"`
	IdleTimeout time.Duration `mapstructure:"idle_timeout"`
}

type MongoConfig struct {
	Host       string `mapstructure:"host"`
	User       string `mapstructure:"user"`
	Password   string `mapstructure:"password"`
	Port       string `mapstructure:"port"`
	DBName     string `mapstructure:"dbname"`
	Collection string `mapstructure:"collection"`
}

func (c *MongoConfig) URI() string {
	return fmt.Sprintf("mongodb://%s:%s@%s:%s", c.User, c.Password, c.Host, c.Port)
}

type UrlsConfig struct {
	ApiService string `mapstructure:"api_service"`
}

type MetricInfo struct {
	Source string   `mapstructure:"source"`
	Names  []string `mapstructure:"names"`
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
