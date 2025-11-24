package main

import (
	"log"
	"os"

	"github.com/alandtsang/gocode/sftp/config"
	sftpserver "github.com/alandtsang/gocode/sftp/sftp_server"
)

func main() {
	log.Println("SFTP server starting...")

	// Get configuration
	configPath := "../../sftp/conf/conf.yaml"
	if len(os.Args) > 1 {
		configPath = os.Args[1]
	}

	// Load SFTP configuration
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	log.Printf("Loaded config: %+v", cfg)

	if err := ensureDirectories(cfg); err != nil {
		log.Fatalf("Failed to create required directories: %v", err)
	}

	server, err := sftpserver.NewSFTPServer(cfg)
	if err != nil {
		log.Fatalf("Failed to create SFTP server: %v", err)
	}

	if err = server.Start(); err != nil {
		log.Fatalf("Failed to start SFTP server: %v", err)
	}
}

func ensureDirectories(cfg *config.Config) error {
	dirs := []string{
		cfg.Server.RootPath,
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			log.Printf("Warning: failed to create directory %s: %v", dir, err)
		}
	}

	log.Printf("Ensured directories: %v", dirs)
	return nil
}
