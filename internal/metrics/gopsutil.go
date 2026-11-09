package metrics

import (
	"github.com/shirou/gopsutil/v4/process"
)

// Snapshot is a point-in-time CPU and RSS aggregate for one instance.
type Snapshot struct {
	CPUPercent float64
	MemoryRSS  uint64
	OK         bool
}

// Aggregate sums CPU percent and RSS for owner PID and the Chrome process tree (FR-004a).
func Aggregate(ownerPID, browserPID int) Snapshot {
	seen := map[int32]struct{}{}
	var pids []int32
	add := func(pid int) {
		if pid <= 0 {
			return
		}
		id := int32(pid)
		if _, ok := seen[id]; ok {
			return
		}
		seen[id] = struct{}{}
		pids = append(pids, id)
		pids = append(pids, descendants(id, seen)...)
	}
	add(ownerPID)
	add(browserPID)
	if len(pids) == 0 {
		return Snapshot{}
	}
	var snap Snapshot
	snap.OK = true
	for _, pid := range pids {
		p, err := process.NewProcess(pid)
		if err != nil {
			snap.OK = false
			return snap
		}
		cpu, err := p.CPUPercent()
		if err != nil {
			snap.OK = false
			return snap
		}
		mem, err := p.MemoryInfo()
		if err != nil || mem == nil {
			snap.OK = false
			return snap
		}
		snap.CPUPercent += cpu
		snap.MemoryRSS += mem.RSS
	}
	return snap
}

func descendants(pid int32, seen map[int32]struct{}) []int32 {
	p, err := process.NewProcess(pid)
	if err != nil {
		return nil
	}
	kids, err := p.Children()
	if err != nil {
		return nil
	}
	var out []int32
	for _, k := range kids {
		if _, ok := seen[k.Pid]; ok {
			continue
		}
		seen[k.Pid] = struct{}{}
		out = append(out, k.Pid)
		out = append(out, descendants(k.Pid, seen)...)
	}
	return out
}
