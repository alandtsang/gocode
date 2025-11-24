package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

// Config represents the application configuration
type Config struct {
	Server ServerConfig `yaml:"server"`
	Users  []UserConfig `yaml:"users"`
}

// ServerConfig represents the SFTP server configuration
type ServerConfig struct {
	// Host to bind the SFTP server
	Host string `yaml:"host"`
	// Port to bind the SFTP server
	Port int `yaml:"port"`
	// Root path for SFTP filesystem
	RootPath string `yaml:"root_path"`
}

// UserConfig represents a user configuration
type UserConfig struct {
	// Username for the SFTP user
	Username string `yaml:"username"`
	// Password for the SFTP user
	Password string `yaml:"password"`
	// Permissions for the SFTP user
	Permissions []Permissions `yaml:"permissions"`
}

// Permissions represents user permissions
type Permissions struct {
	// Path is the directory path this permission applies to
	Path string `yaml:"path"`
	// Read permission
	Read bool `yaml:"read"`
	// Write permission
	Write bool `yaml:"write"`
	// Command permission
	Cmd bool `yaml:"cmd"`
	// List permission
	List bool `yaml:"list"`
}

// LoadConfig loads the configuration from a YAML file
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var config Config
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}
