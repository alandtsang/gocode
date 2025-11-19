package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Config config struct.
type Config struct {
	App        AppConfig        `yaml:"app"`
	Database   DatabaseConfig   `yaml:"database"`
	Redis      RedisConfig      `yaml:"redis"`
	Encryption EncryptionConfig `yaml:"encryption"`
}

// AppConfig app config struct.
type AppConfig struct {
	Name  string `yaml:"name"`
	Port  int    `yaml:"port"`
	Debug bool   `yaml:"debug"`
}

// DatabaseConfig database config struct.
type DatabaseConfig struct {
	Host              string `yaml:"host"`
	Port              int    `yaml:"port"`
	Username          string `yaml:"username"`
	Password          string `yaml:"password"` // encrypted password
	DecryptedPassword string `yaml:"-"`        // decrypted password, not serialized to yaml
	Name              string `yaml:"name"`
}

// RedisConfig redis config struct.
type RedisConfig struct {
	Host              string `yaml:"host"`
	Port              int    `yaml:"port"`
	Password          string `yaml:"password"` // encrypted password
	DecryptedPassword string `yaml:"-"`        // decrypted password, not serialized to yaml
}

// EncryptionConfig encryption config struct.
type EncryptionConfig struct {
	Algorithm string `yaml:"algorithm"`
}

// LoadConfig load config from file.
func LoadConfig(configPath string) (*Config, error) {
	// if configPath is empty, use default config path
	if configPath == "" {
		configPath = getDefaultConfigPath()
	}

	// read config file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// parse yaml
	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// decrypt secrets
	if err := config.DecryptSecrets(); err != nil {
		return nil, fmt.Errorf("failed to decrypt secrets: %w", err)
	}

	return &config, nil
}

// DecryptSecrets decrypt secrets.
func (c *Config) DecryptSecrets() error {
	var err error

	// decrypt database password
	if c.Database.Password != "" {
		c.Database.DecryptedPassword, err = Decrypt(c.Database.Password)
		if err != nil {
			return fmt.Errorf("failed to decrypt database password: %w", err)
		}
	}

	// decrypt redis password
	if c.Redis.Password != "" {
		c.Redis.DecryptedPassword, err = Decrypt(c.Redis.Password)
		if err != nil {
			return fmt.Errorf("failed to decrypt redis password: %w", err)
		}
	}

	return nil
}

// getDefaultConfigPath get default config file path
func getDefaultConfigPath() string {
	// try multiple possible config file locations
	possiblePaths := []string{
		"conf/conf.yaml",
		"conf.yaml",
		"./conf/conf.yaml",
		"../conf/conf.yaml",
		"../../conf/conf.yaml",
	}

	for _, path := range possiblePaths {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	// if not found, return default path
	return "conf/conf.yaml"
}
