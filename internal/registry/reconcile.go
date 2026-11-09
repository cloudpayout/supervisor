package registry

import (
	"time"

	"github.com/cloudpayout/supervisor/internal/proc"
)

// Reap describes one instance removed during reconciliation.
type Reap struct {
	ID         string
	ProfileDir string
	KillPIDs   []int
	// ProcessesGone is true when both owner and browser were already dead.
	ProcessesGone bool
}

// Plan returns removals for FR-013 and FR-009a/b. alive reports whether a PID exists.
func Plan(doc *Document, now time.Time, alive func(int) bool) []Reap {
	if alive == nil {
		alive = func(int) bool { return false }
	}
	var out []Reap
	for _, inst := range doc.Instances {
		ownerUp := inst.OwnerPID > 0 && alive(inst.OwnerPID)
		browserUp := inst.BrowserPID > 0 && alive(inst.BrowserPID)
		age := now.Sub(inst.CreatedAt)

		if inst.State == StateStarting && age > StartingTimeout {
			out = append(out, reapOf(inst, ownerUp, browserUp))
			continue
		}
		if inst.State == StateStarting && inst.OwnerPID > 0 && !ownerUp {
			out = append(out, reapOf(inst, false, browserUp))
			continue
		}
		if inst.State == StateStarting {
			continue
		}
		// running / stopping / stopped
		switch {
		case !ownerUp && browserUp:
			out = append(out, reapOf(inst, false, true))
		case ownerUp && inst.BrowserPID > 0 && !browserUp:
			out = append(out, reapOf(inst, true, false))
		case !ownerUp && !browserUp:
			r := reapOf(inst, false, false)
			r.ProcessesGone = true
			out = append(out, r)
		}
	}
	return out
}

func reapOf(inst Instance, killOwner, killBrowser bool) Reap {
	r := Reap{ID: inst.ID, ProfileDir: inst.ProfileDir}
	if killOwner && inst.OwnerPID > 0 {
		r.KillPIDs = append(r.KillPIDs, inst.OwnerPID)
	}
	if killBrowser && inst.BrowserPID > 0 {
		r.KillPIDs = append(r.KillPIDs, inst.BrowserPID)
	}
	if killBrowser {
		r.ProfileDir = inst.ProfileDir
	}
	return r
}

// Apply mutates doc by executing Plan and terminating orphan processes.
func Apply(doc *Document, now time.Time) []Reap {
	planned := Plan(doc, now, proc.Alive)
	if len(planned) == 0 {
		return nil
	}
	remove := map[string]struct{}{}
	for _, r := range planned {
		remove[r.ID] = struct{}{}
		for _, pid := range r.KillPIDs {
			proc.Terminate(pid)
		}
		if r.ProfileDir != "" {
			proc.KillByProfile(r.ProfileDir)
		}
	}
	next := doc.Instances[:0]
	for _, inst := range doc.Instances {
		if _, drop := remove[inst.ID]; drop {
			continue
		}
		next = append(next, inst)
	}
	doc.Instances = next
	return planned
}
