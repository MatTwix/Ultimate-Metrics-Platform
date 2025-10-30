package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

type Config struct {
	Env         string            `mapstructure:"env"`
	Server      ServerConfig      `mapstructure:"server"`
	Worker      WorkerConfig      `mapstructure:"worker"`
	Github      GithubConfig      `mapstructure:"github"`
	OpenWeather OpenWeatherConfig `mapstructure:"open_weather"`
	Broker      BrokerConfig      `mapstructure:"broker"`
	Urls        UrlsConfig        `mapstructure:"urls"`
}

type ServerConfig struct {
	Port        string        `mapstructure:"port"`
	Timeout     time.Duration `mapstructure:"timeout"`
	IdleTimeout time.Duration `mapstructure:"idle_timeout"`
}

type WorkerConfig struct {
	PollInterval time.Duration `mapstructure:"poll_interval"`
}

type GithubConfig struct {
	Token      string `mapstructure:"token"`
	Repository string `mapstructure:"repository"`
}

type OpenWeatherConfig struct {
	APIKey string `mapstructure:"api_key"`
	City   string `mapstructure:"city"`
}

type BrokerConfig struct {
	Type  string      `mapstructure:"type"`
	Kafka KafkaConfig `mapstructure:"kafka"`
}

type KafkaConfig struct {
	Brokers []string `mapstructure:"brokers"`
	Topic   string   `mapstructure:"topic"`
}

type UrlsConfig struct {
	CacheService string `mapstructure:"cache_service"`
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
