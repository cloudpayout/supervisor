package cli

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/cloudpayout/supervisor/internal/browser"
	"github.com/cloudpayout/supervisor/internal/proc"
	"github.com/cloudpayout/supervisor/internal/registry"
)

const stopAllToken = "all"

type stopOutcome int

const (
	outcomeNotFound stopOutcome = iota
	outcomeAlready
	outcomeStopped
)

// stopOne is the per-instance stop used by bulk mode. Tests may replace it.
var stopOne = stopInstance

func newStopCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "stop <instance-id|all>",
		Short: "Stop one instance by ID, or every tracked instance",
		Long: `Stop a supervised instance by ID, or stop every tracked instance.

  supervisor stop <instance-id>
  supervisor stop all

The token all is reserved for bulk stop and is not an instance ID.
The command does not prompt and does not read standard input.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runStop(strings.TrimSpace(args[0]))
		},
	}
}

func runStop(id string) error {
	if id == "" {
		return fmt.Errorf("instance ID required")
	}
	if id == stopAllToken {
		return runStopAll()
	}
	_, err := stopInstance(id)
	return err
}

func stopInstance(id string) (stopOutcome, error) {
	store := mustStore()
	var before *registry.Instance
	var after *registry.Instance
	var gone bool
	err := store.Update(func(doc *registry.Document) error {
		if i := doc.Find(id); i >= 0 {
			cp := doc.Instances[i]
			before = &cp
		}
		reaped := registry.Apply(doc, time.Now().UTC())
		for _, r := range reaped {
			if r.ID == id && r.ProcessesGone {
				gone = true
			}
		}
		if i := doc.Find(id); i >= 0 {
			cp := doc.Instances[i]
			after = &cp
		}
		return nil
	})
	if err != nil {
		return 0, err
	}
	if before == nil {
		fmt.Fprintf(os.Stderr, "instance not found: %s\n", id)
		return outcomeNotFound, nil
	}
	if after == nil {
		if gone {
			fmt.Fprintf(os.Stderr, "instance already stopped: %s\n", id)
			return outcomeAlready, nil
		}
		fmt.Fprintf(os.Stderr, "stopped instance %s\n", id)
		return outcomeStopped, nil
	}

	proc.Terminate(after.OwnerPID)
	proc.Terminate(after.BrowserPID)
	proc.KillByProfile(after.ProfileDir)
	// FR-007: remove the profile only after the browser has been asked to exit.
	// Feature 008 FR-009b: do not delete the instance log file.
	browser.RemoveProfile(after.ProfileDir)
	err = store.Update(func(doc *registry.Document) error {
		i := doc.Find(id)
		if i >= 0 {
			doc.Instances = append(doc.Instances[:i], doc.Instances[i+1:]...)
		}
		return nil
	})
	if err != nil {
		return 0, err
	}
	fmt.Fprintf(os.Stderr, "stopped instance %s\n", id)
	return outcomeStopped, nil
}

func runStopAll() error {
	store := mustStore()
	var ids []string
	err := store.Update(func(doc *registry.Document) error {
		registry.Apply(doc, time.Now().UTC())
		ids = make([]string, len(doc.Instances))
		for i, inst := range doc.Instances {
			ids[i] = inst.ID
		}
		return nil
	})
	if err != nil {
		return err
	}
	if len(ids) == 0 {
		fmt.Fprintf(os.Stderr, "No instances are running.\n")
		return nil
	}

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)
	defer signal.Stop(sig)

	stopped, skipped := 0, 0
	var failMsgs []string
	done := 0
	interrupted := false
	pause := stopAllPause()

	for i, id := range ids {
		select {
		case <-sig:
			interrupted = true
		default:
		}
		if interrupted {
			break
		}
		outcome, stopErr := stopOne(id)
		done++
		if stopErr != nil {
			// Same text single-ID stop surfaces (the error itself); no extra prefix.
			failMsgs = append(failMsgs, stopErr.Error())
		} else {
			switch outcome {
			case outcomeStopped:
				stopped++
			default:
				skipped++
			}
		}
		if pause > 0 && i < len(ids)-1 {
			select {
			case <-sig:
				interrupted = true
			case <-time.After(pause):
			}
		}
	}
	if interrupted {
		fmt.Fprintf(os.Stderr, "stop all interrupted: completed %d of %d\n", done, len(ids))
		return fmt.Errorf("stop all interrupted")
	}
	fmt.Fprintf(os.Stderr, "stop all: stopped %d, skipped %d (already stopped or not found)\n", stopped, skipped)
	if len(failMsgs) > 0 {
		return fmt.Errorf("%s", strings.Join(failMsgs, "\n"))
	}
	return nil
}

// stopAllPause honors SUPERVISOR_STOP_ALL_PAUSE so tests can interrupt between instances.
func stopAllPause() time.Duration {
	v := os.Getenv("SUPERVISOR_STOP_ALL_PAUSE")
	if v == "" {
		return 0
	}
	d, err := time.ParseDuration(v)
	if err != nil || d < 0 {
		return 0
	}
	return d
}
