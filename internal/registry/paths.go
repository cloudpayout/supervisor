package registry

import (
	"os"
	"path/filepath"
)

// Layout resolves XDG state and data directories for Supervisor.
type Layout struct {
	StateDir string
	DataDir  string
}

// Resolve uses XDG_STATE_HOME, XDG_DATA_HOME, and HOME fallbacks.
func Resolve() (Layout, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return Layout{}, err
	}
	state := os.Getenv("XDG_STATE_HOME")
	if state == "" {
		state = filepath.Join(home, ".local", "state")
	}
	data := os.Getenv("XDG_DATA_HOME")
	if data == "" {
		data = filepath.Join(home, ".local", "share")
	}
	return Layout{
		StateDir: filepath.Join(state, "supervisor"),
		DataDir:  filepath.Join(data, "supervisor"),
	}, nil
}

func (l Layout) RegistryPath() string {
	return filepath.Join(l.StateDir, "instances.json")
}

func (l Layout) LogPath(id string) string {
	return filepath.Join(l.StateDir, "logs", id+".log")
}

func (l Layout) ProfileDir(id string) string {
	return filepath.Join(l.DataDir, "profiles", id)
}
