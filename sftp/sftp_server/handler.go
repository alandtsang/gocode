package sftpserver

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/alandtsang/gocode/sftp/config"
	"github.com/pkg/sftp"
)

// permissionHandler implements sftp.Handler with permission checking.
// It wraps the base filesystem operations with permission validation.
type permissionHandler struct {
	// user config
	userConfig *config.UserConfig
	// root path for this user
	rootPath string
	// whether this is a read-only operation
	readPerm bool
	// whether this allows write operations
	writePerm bool
	// whether this allows command operations
	cmdPerm bool
	// whether this allows list operations
	listPerm bool
}

// Fileread implements the sftp.Handler Fileread method with permission checking
// This method checks if the user has read permission before allowing file reading.
func (h *permissionHandler) Fileread(r *sftp.Request) (io.ReaderAt, error) {
	fmt.Printf("Fileread filepath: %s, readPerm: %v\n", r.Filepath, h.readPerm)
	if err := h.checkPermission(r.Filepath, h.readPerm); err != nil {
		fmt.Printf("Fileread permission denied for user: %s, filepath: %s\n", h.userConfig.Username, r.Filepath)
		return nil, err
	}
	// Implement actual file reading logic here
	return os.Open(filepath.Join(h.rootPath, r.Filepath))
}

// Filewrite implements the sftp.Handler Filewrite method with permission checking
func (h *permissionHandler) Filewrite(r *sftp.Request) (io.WriterAt, error) {
	fmt.Printf("Filewrite filepath: %s, writePerm: %v\n", r.Filepath, h.writePerm)
	if err := h.checkPermission(r.Filepath, h.writePerm); err != nil {
		fmt.Printf("Filewrite permission denied for user: %s, filepath: %s\n", h.userConfig.Username, r.Filepath)
		return nil, err
	}

	// Create parent directories if they don't exist
	fullPath := filepath.Join(h.rootPath, r.Filepath)
	if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
		return nil, err
	}
	return os.Create(fullPath)
}

// Filecmd implements the sftp.Handler Filecmd method with permission checking
func (h *permissionHandler) Filecmd(r *sftp.Request) error {
	fmt.Printf("Filecmd, cmdPerm: %v, method: %s\n", h.cmdPerm, r.Method)
	if err := h.checkPermission(r.Filepath, h.cmdPerm); err != nil {
		fmt.Printf("Filecmd permission denied for user: %s, filepath: %s, method: %s\n", h.userConfig.Username, r.Filepath, r.Method)
		return err
	}

	path := filepath.Join(h.rootPath, r.Filepath)

	// Implement file operations (remove, rename, etc.)
	switch r.Method {
	case "Remove":
		return os.Remove(path)
	case "Rename":
		targetPath := filepath.Join(h.rootPath, r.Target)
		return os.Rename(path, targetPath)
	case "Mkdir":
		return os.Mkdir(path, 0755)
	case "Stat":
		_, err := os.Stat(path)
		return err
	case "Lstat":
		_, err := os.Lstat(path)
		return err
	case "Readlink":
		_, err := os.Readlink(path)
		return err
	default:
		return sftp.ErrSSHFxOpUnsupported
	}
}

// Filelist implements the sftp.Handler Filelist method with permission checking
func (h *permissionHandler) Filelist(r *sftp.Request) (sftp.ListerAt, error) {
	fmt.Printf("Filelist, filepath: %s, listPerm: %v\n", r.Filepath, h.listPerm)
	if err := h.checkPermission(r.Filepath, h.listPerm); err != nil {
		fmt.Printf("Filelist permission denied for user: %s, filepath: %s\n", h.userConfig.Username, r.Filepath)
		return nil, err
	}

	f, err := os.Open(filepath.Join(h.rootPath, r.Filepath))
	if err != nil {
		return nil, err
	}
	defer f.Close()

	files, err := f.Readdir(-1)
	if err != nil {
		return nil, err
	}

	return dirLister(files), nil
}

// Add this custom type at the package level
type dirLister []os.FileInfo

// Implement sftp.ListerAt interface
func (l dirLister) ListAt(ls []os.FileInfo, offset int64) (int, error) {
	if offset >= int64(len(l)) {
		return 0, io.EOF
	}
	n := copy(ls, l[offset:])
	if n < len(ls) {
		return n, io.EOF
	}
	return n, nil
}

// checkPermission checks if the user has the required permission for the given path
func (h *permissionHandler) checkPermission(path string, hasPermission bool) error {
	if h.userConfig == nil {
		return os.ErrPermission
	}

	// Convert to absolute path within the user's root
	absPath := filepath.Join(h.rootPath, path)
	relPath, err := filepath.Rel(h.rootPath, absPath)
	if err != nil {
		return os.ErrPermission
	}

	fmt.Printf("checkPermission: absPath=%s, relPath=%s, hasPermission=%v\n", absPath, relPath, hasPermission)

	if !hasPermission {
		return os.ErrPermission
	}
	return nil
}
