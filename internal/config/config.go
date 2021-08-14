package config

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// Config main config.
type Config struct {
	Exchange Exchange `yaml:"exchange"`
	Broker   Broker   `yaml:"broker"`
}

// Exchange stock exchange service config.
type Exchange struct {
	Tickers  []string      `yaml:"tickers"`
	Interval time.Duration `yaml:"interval"`
}

// Broker broker config.
type Broker struct {
	DB DB `yaml:"db"`
}

// DB Postgres config.
type DB struct {
	Host     string `yaml:"host"`
	Port     int32  `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	DBName   string `yaml:"db_name"`
	TimeZone string `yaml:"time_zone"`
}

// Load loads configs from a file.
func Load(name string) (Config, error) {
	_, currentPath, _, _ := runtime.Caller(0)
	root := strings.Split(currentPath, "trading")[0]
	configPath := filepath.Join(root, "trading", "configs", fmt.Sprintf("%s.yaml", name))

	f, err := os.Open(configPath)
	if err != nil {
		return Config{}, err
	}
	defer func() {
		_ = f.Close()
	}()

	cfg := Config{}

	decoder := yaml.NewDecoder(f)
	if err := decoder.Decode(&cfg); err != nil {
		return Config{}, err
	}

	return cfg, nil
}
