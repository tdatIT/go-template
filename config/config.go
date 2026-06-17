package config

import (
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/spf13/viper"
)

type AppConfig struct {
	Server   Server
	Database Database
	Redis    Redis
	Logger   Logger
	Adapters Adapters
}

type Adapters struct {
	ProductService HTTPClient
}

type HTTPClient struct {
	BaseURL    string
	Timeout    time.Duration
	KeepAlive  time.Duration
	RetryCount int
	RetryWait  time.Duration
	Debug      bool
	APIKey     string
}

type Server struct {
	Debug             bool
	Name              string
	Version           string
	Port              string
	BaseURI           string
	RequestTimeout    time.Duration
	ReadTimeout       time.Duration
	WriteTimeout      time.Duration
	CtxDefaultTimeout time.Duration
}

type Database struct {
	Host            string
	Port            int
	UserName        string
	Password        string
	Database        string
	Schema          string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
	Options         string
	Debug           bool
}

type Redis struct {
	Mode       string
	Username   string
	Password   string
	MasterName string
	PoolSize   int
	Address    []string
}

type Logger struct {
	Level     string
	Endcoding string
}

func getDefaultConfig() string {
	return "/config/config"
}

func NewConfig() (*AppConfig, error) {
	path := os.Getenv("CONFIG_PATH")
	if path == "" {
		path = getDefaultConfig()
	}

	fmt.Printf("config path:%s\n", path)

	config := AppConfig{}
	v := viper.New()
	v.SetConfigName(path)
	v.AddConfigPath(".")
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	if err := v.ReadInConfig(); err != nil {
		slog.Error("unable to read config file", slog.String("error", err.Error()))
		return nil, err
	}

	err := v.Unmarshal(&config)
	if err != nil {
		slog.Error("unable to decode into struct", slog.String("error", err.Error()))
		return nil, err
	}

	return &config, nil
}
