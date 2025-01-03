package config

import (
	"github.com/BurntSushi/toml"
)

type Config struct {
	Server   ServerConfig
	Log      LogConfig
	Database DatabaseConfig
	Frontend FrontendConfig
}

type ServerConfig struct {
	Host     string `toml:"host"`
	Port     int    `toml:"port"`
	BasePath string `toml:"basePath"`
}

type LogConfig struct {
	LogFilePath string `toml:"logFilePath"`
	MaxLogSize  int    `toml:"maxLogSize"` // MB
}

type DatabaseConfig struct {
	Model string `toml:"model"`
	Path  string `toml:"path"` // bolt file path
}

type FrontendConfig struct {
	Chartlist int `toml:"chartlist"`
}

// LoadConfig 从 TOML 配置文件加载配置
func LoadConfig(filePath string) (*Config, error) {
	var config Config
	if _, err := toml.DecodeFile(filePath, &config); err != nil {
		return nil, err
	}
	return &config, nil
}
