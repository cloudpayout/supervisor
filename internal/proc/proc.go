package proc

import (
	"fmt"
	"os"
	"strings"
	"syscall"
	"time"

	"github.com/shirou/gopsutil/v4/process"
)

// Alive reports whether pid is a running process. Zombies count as dead so a
// parent that has not yet waited does not keep an instance registered.
func Alive(pid int) bool {
	if pid <= 0 {
		return false
	}
	data, err := os.ReadFile(fmt.Sprintf("/proc/%d/stat", pid))
	if err != nil {
		return false
	}
	text := string(data)
	i := strings.LastIndex(text, ")")
	if i < 0 || i+2 >= len(text) {
		return false
	}
	state := text[i+2]
	return state != 'Z' && state != 'X'
}

// Terminate sends SIGTERM, then SIGKILL if the process is still alive.
func Terminate(pid int) {
	if pid <= 0 || !Alive(pid) {
		return
	}
	_ = syscall.Kill(pid, syscall.SIGTERM)
	_ = syscall.Kill(-pid, syscall.SIGTERM)
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		if !Alive(pid) {
			return
		}
		time.Sleep(100 * time.Millisecond)
	}
	_ = syscall.Kill(pid, syscall.SIGKILL)
	_ = syscall.Kill(-pid, syscall.SIGKILL)
}

// KillByProfile SIGKILLs processes whose command line contains the profile path.
func KillByProfile(profile string) {
	if profile == "" {
		return
	}
	procs, err := process.Processes()
	if err != nil {
		return
	}
	var pids []int
	for _, p := range procs {
		cmd, err := p.Cmdline()
		if err != nil || !strings.Contains(cmd, profile) {
			continue
		}
		pids = append(pids, int(p.Pid))
	}
	for _, pid := range pids {
		_ = syscall.Kill(pid, syscall.SIGTERM)
	}
	time.Sleep(200 * time.Millisecond)
	for _, pid := range pids {
		if Alive(pid) {
			_ = syscall.Kill(pid, syscall.SIGKILL)
		}
	}
}
