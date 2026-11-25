package registry

import "time"

const (
	SchemaVersion   = 1
	SchemaVersionV2 = 2
)

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
	// Casino is the registered identifier, such as onewin. A blank start omits it.
	// It is not a cookie, a session path, or a page address (FR-005, FR-006).
	Casino string `json:"casino,omitempty"`
	// Entry is entering or ready while onewin Aviator entry is in progress or done.
	// A blank start omits it. Failure removes the row instead of storing a reason.
	Entry string `json:"entry,omitempty"`
	// MemoryLimitBytes is the optional --memory-limit in bytes (mebibyte parsing). Minimum 1 when set.
	MemoryLimitBytes uint64 `json:"memory_limit_bytes,omitempty"`
	// MemoryThresholdPercent is 1–100 when a memory limit is set.
	MemoryThresholdPercent int `json:"memory_threshold_percent,omitempty"`
	// Lightweight is nil on legacy rows (worker treats as enabled). 013 start writes true or false.
	// Pointer so false is persisted (bool omitempty would drop false).
	Lightweight *bool `json:"lightweight,omitempty"`
}

// LightweightEnabled reports whether 013 lightweight extras apply.
// A missing field (legacy registry) is treated as enabled.
func (inst Instance) LightweightEnabled() bool {
	if inst.Lightweight == nil {
		return true
	}
	return *inst.Lightweight
}

// NoteThresholdExceeded is the fixed status NOTE when RSS is at or above the warning line.
const NoteThresholdExceeded = "threshold exceeded"

// WarningBytes is floor(limit * percent / 100). Zero when no limit is configured.
func (inst Instance) WarningBytes() uint64 {
	if inst.MemoryLimitBytes == 0 || inst.MemoryThresholdPercent <= 0 {
		return 0
	}
	return inst.MemoryLimitBytes * uint64(inst.MemoryThresholdPercent) / 100
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
