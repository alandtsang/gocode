package sftpserver

import (
	"encoding/hex"
	"fmt"
	"log"
	"net"
	"path/filepath"

	"github.com/alandtsang/gocode/sftp/config"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

const (
	extraDataUserKey        = "user"
	extraDataKeyIDKey       = "keyID"
	extraDataLoginMethodKey = "login_method"
)

// SFTPServer represents the SFTP server implementation
type SFTPServer struct {
	// Configuration for the SFTP server
	cfg *config.Config
	// SSH configuration for authentication and connection handling
	sshConfig *ssh.ServerConfig
}

// NewSFTPServer creates a new SFTP server
func NewSFTPServer(cfg *config.Config) (*SFTPServer, error) {
	server := &SFTPServer{
		cfg: cfg,
	}

	if err := server.setupSSHConfig(); err != nil {
		return nil, err
	}

	return server, nil
}

// setupSSHConfig sets up the SSH configuration for the server
func (s *SFTPServer) setupSSHConfig() error {
	s.sshConfig = &ssh.ServerConfig{
		NoClientAuth:     false,
		PasswordCallback: s.authenticateUser,
		// MaxAuthTries:     3,
		// ServerVersion:    "SSH-2.0-SFTP_SERVER_1.0",
	}

	// Generate host key (in production, load from file)
	privateKey, err := generateHostKey()
	if err != nil {
		return fmt.Errorf("failed to generate host key: %w", err)
	}

	s.sshConfig.AddHostKey(privateKey)
	return nil
}

func (s *SFTPServer) authenticateUser(conn ssh.ConnMetadata, password []byte) (*ssh.Permissions, error) {
	userMap := make(map[string]string)
	for i := range s.cfg.Users {
		user := &s.cfg.Users[i]
		userMap[user.Username] = user.Password
	}

	// Check if user exists and password matches
	if expectedPass, ok := userMap[conn.User()]; ok {
		if expectedPass == string(password) {
			return &ssh.Permissions{
				Extensions: map[string]string{
					"username": conn.User(),
				},
			}, nil
		}
	}
	return nil, fmt.Errorf("password rejected for %q", conn.User())
}

// Start starts the SFTP server
// This method listens for incoming SSH connections and handles them in separate goroutines.
// It will block until the server is stopped.
// The server will log information about incoming connections and handle them asynchronously.
func (s *SFTPServer) Start() error {
	// Start listening
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", s.cfg.Server.Port))
	if err != nil {
		return fmt.Errorf("failed to listen on port %d: %w", s.cfg.Server.Port, err)
	}
	log.Printf("SFTP server listening on port %d", s.cfg.Server.Port)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Failed to accept connection: %v", err)
			continue
		}
		go s.handleConnection(conn)
	}
}

// Stop stops the SFTP server
// This method provides a placeholder for server shutdown functionality.
// In a production implementation, this would gracefully close all connections and stop the listener.
func (s *SFTPServer) Stop() error {
	// TODO: Implement proper server shutdown
	log.Println("SFTP server stopping...")
	return nil
}

// GetStatus returns the current status of the SFTP server
// This method provides a placeholder for retrieving server status information.
// In a production implementation, this would return actual server metrics and statistics.
func (s *SFTPServer) GetStatus() map[string]interface{} {
	// TODO: Implement SFTP server status retrieval
	return map[string]interface{}{
		"status":                    "not implemented",
		"connections":               0,
		"uptime":                    "not implemented",
		"version":                   "0.1.0",
		"started_at":                "not implemented",
		"max_connections":           0,
		"current_connections":       0,
		"total_connections":         0,
		"active_sessions":           0,
		"total_transfers":           0,
		"total_errors":              0,
		"total_bytes_transferred":   0,
		"average_transfer_speed":    0.0,
		"peak_connections":          0,
		"current_bandwidth":         0.0,
		"total_bandwidth":           0,
		"current_transfer_rate":     0.0,
		"total_active_users":        0,
		"total_files_processed":     0,
		"total_directories_created": 0,
		"total_bytes_read":          0,
		"total_bytes_written":       0,
		"total_commands_executed":   0,
		"total_logins":              0,
		"total_sessions":            0,
	}
}

// handleConnection handles an individual SFTP connection
// This function manages the SSH and SFTP session for a single client connection.
// It handles authentication, channel establishment, and SFTP operations for each client.
func (s *SFTPServer) handleConnection(conn net.Conn) {
	defer conn.Close()

	// Perform SSH handshake
	sshConn, chans, reqs, err := ssh.NewServerConn(conn, s.sshConfig)
	if err != nil {
		log.Printf("Failed to handshake: %v", err)
		return
	}
	defer sshConn.Close()

	connectionID := hex.EncodeToString(sshConn.SessionID())

	log.Printf("New connection from %s (%s)", sshConn.RemoteAddr(), sshConn.User())
	log.Printf("Connection ID: %s", connectionID)

	// Discard global requests
	go ssh.DiscardRequests(reqs)

	// Handle channels
	for newChannel := range chans {
		if newChannel.ChannelType() != "session" {
			newChannel.Reject(ssh.UnknownChannelType, "unknown channel type")
			continue
		}

		channel, requests, err := newChannel.Accept()
		if err != nil {
			log.Printf("Failed to accept channel: %v", err)
			continue
		}

		go s.handleChannelRequests(channel, requests, sshConn)
	}
}

// handleChannelRequests handles the SSH channel requests for SFTP
// It processes subsystem requests to start the SFTP server for the connected client.
// The SFTP server is configured with permission handlers based on the connected user.
func (s *SFTPServer) handleChannelRequests(channel ssh.Channel, requests <-chan *ssh.Request, sshConn *ssh.ServerConn) {
	defer channel.Close()

	for req := range requests {
		switch req.Type {
		case "subsystem":
			if string(req.Payload[4:]) == "sftp" {
				req.Reply(true, nil)

				// Create a new SFTP server with our custom handler
				server := sftp.NewRequestServer(channel, s.newPermissionHandler(sshConn.User()))
				defer server.Close()

				log.Printf("Starting SFTP server for user %s", sshConn.User())
				if err := server.Serve(); err != nil {
					log.Printf("SFTP server completed with error: %v", err)
				}
				return
			}
		default:
			log.Printf("Received unknown request type %v", req.Type)
		}
	}
}

// newPermissionHandler creates a new SFTP handler with permission checking
func (s *SFTPServer) newPermissionHandler(username string) sftp.Handlers {
	userConfig := s.getUserConfig(username)
	if userConfig == nil {
		// Return a handler that denies all operations
		denyHandler := &permissionHandler{
			userConfig: nil,
			rootPath:   "",
			readPerm:   false,
			writePerm:  false,
			cmdPerm:    false,
			listPerm:   false,
		}
		return sftp.Handlers{
			FileGet:  denyHandler,
			FilePut:  denyHandler,
			FileCmd:  denyHandler,
			FileList: denyHandler,
		}
	}

	ph := &permissionHandler{
		userConfig: userConfig,
		rootPath:   filepath.Join(s.cfg.Server.RootPath, username),
		readPerm:   userConfig.Permissions[0].Read,
		writePerm:  userConfig.Permissions[0].Write,
		cmdPerm:    userConfig.Permissions[0].Cmd,
		listPerm:   userConfig.Permissions[0].List,
	}
	fmt.Printf("=== ph: %+v\n", ph)
	return sftp.Handlers{
		FileGet:  ph,
		FilePut:  ph,
		FileCmd:  ph,
		FileList: ph,
	}
}

func (s *SFTPServer) getUserConfig(username string) *config.UserConfig {
	for i := range s.cfg.Users {
		if s.cfg.Users[i].Username == username {
			return &s.cfg.Users[i]
		}
	}
	return nil
}
