package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"gopkg.in/yaml.v3"
)

// Config represents the application configuration
type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
	GitHub   GitHubConfig   `yaml:"github"`
	Logging  LoggingConfig  `yaml:"logging"`
}

// ServerConfig represents the server configuration
type ServerConfig struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
}

// DatabaseConfig represents the database configuration
type DatabaseConfig struct {
	Type string `yaml:"type"` // sqlite or mysql
	Path string `yaml:"path"` // For SQLite
	// MySQL configuration (for future use)
	Host     string `yaml:"host,omitempty"`
	Port     int    `yaml:"port,omitempty"`
	Username string `yaml:"username,omitempty"`
	Password string `yaml:"password,omitempty"`
	Database string `yaml:"database,omitempty"`
}

// GitHubConfig represents the GitHub configuration
type GitHubConfig struct {
	RefreshInterval time.Duration `yaml:"refresh_interval"`
	ItemsPerFetch   int           `yaml:"items_per_fetch"`
}

// LoggingConfig represents the logging configuration
type LoggingConfig struct {
	Level  string `yaml:"level"`
	Format string `yaml:"format"`
}

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	return &Config{
		Server: ServerConfig{
			Host: "127.0.0.1",
			Port: 8080,
		},
		Database: DatabaseConfig{
			Type: "sqlite",
			Path: "data/github-repos.db",
		},
		GitHub: GitHubConfig{
			RefreshInterval: 30 * time.Minute,
			ItemsPerFetch:   10,
		},
		Logging: LoggingConfig{
			Level:  "info",
			Format: "text",
		},
	}
}

// Load loads the configuration from the specified file
func Load(configPath string) (*Config, error) {
	config := DefaultConfig()

	// If no config file is specified, use environment variables
	if configPath == "" {
		return loadFromEnv(config)
	}

	// Read the config file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Parse the config file
	if err := yaml.Unmarshal(data, config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return config, nil
}

// loadFromEnv loads configuration from environment variables
func loadFromEnv(config *Config) (*Config, error) {
	// Server configuration
	if host := os.Getenv("GHREPOS_HOST"); host != "" {
		config.Server.Host = host
	}
	if portStr := os.Getenv("GHREPOS_PORT"); portStr != "" {
		if port, err := strconv.Atoi(portStr); err == nil && port > 0 {
			config.Server.Port = port
		}
	}

	// Database configuration
	if dbType := os.Getenv("GHREPOS_DB_TYPE"); dbType != "" {
		config.Database.Type = dbType
	}
	if dbPath := os.Getenv("GHREPOS_DB_PATH"); dbPath != "" {
		config.Database.Path = dbPath
	}

	// GitHub configuration
	if refreshInterval := os.Getenv("GHREPOS_REFRESH_INTERVAL"); refreshInterval != "" {
		if duration, err := time.ParseDuration(refreshInterval); err == nil {
			config.GitHub.RefreshInterval = duration
		}
	}
	if itemsPerFetchStr := os.Getenv("GHREPOS_ITEMS_PER_FETCH"); itemsPerFetchStr != "" {
		if items, err := strconv.Atoi(itemsPerFetchStr); err == nil && items > 0 {
			config.GitHub.ItemsPerFetch = items
		}
	}

	// Logging configuration
	if logLevel := os.Getenv("GHREPOS_LOG_LEVEL"); logLevel != "" {
		config.Logging.Level = logLevel
	}
	if logFormat := os.Getenv("GHREPOS_LOG_FORMAT"); logFormat != "" {
		config.Logging.Format = logFormat
	}

	return config, nil
}
