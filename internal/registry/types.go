package registry

import "time"

const SchemaVersion = 1

type State string

const (
	StateStarting State = "starting"
	StateRunning  State = "running"
	StateStopping State = "stopping"
	StateStopped  State = "stopped"
)

// StartingTimeout is how long a non-running registration may remain (FR-013).
const StartingTimeout = 60 * time.Second

// Instance is one tracked browser session.
type Instance struct {
	ID         string     `json:"id"`
	State      State      `json:"state"`
	CreatedAt  time.Time  `json:"created_at"`
	ProfileDir string     `json:"profile_dir"`
	OwnerPID   int        `json:"owner_pid,omitempty"`
	BrowserPID int        `json:"browser_pid,omitempty"`
	LogPath    string     `json:"log_path,omitempty"`
	Headless   *bool      `json:"headless,omitempty"`
	LastSeenAt *time.Time `json:"last_seen_at,omitempty"`
}

// Document is the on-disk registry (schema_version const 1).
type Document struct {
	SchemaVersion int        `json:"schema_version"`
	Instances     []Instance `json:"instances"`
}

func (d *Document) normalize() {
	if d.SchemaVersion == 0 {
		d.SchemaVersion = SchemaVersion
	}
	if d.Instances == nil {
		d.Instances = []Instance{}
	}
}

// Find returns the index of id or -1.
func (d *Document) Find(id string) int {
	return d.find(id)
}

func (d *Document) find(id string) int {
	for i := range d.Instances {
		if d.Instances[i].ID == id {
			return i
		}
	}
	return -1
}
