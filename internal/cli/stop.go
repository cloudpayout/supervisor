package cli

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/cloudpayout/supervisor/internal/proc"
	"github.com/cloudpayout/supervisor/internal/registry"
)

func newStopCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "stop <instance-id>",
		Short: "Stop a browser instance by ID",
		Long:  "Stop the supervised instance with the given ID. No interactive prompt is used.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runStop(strings.TrimSpace(args[0]))
		},
	}
}

func runStop(id string) error {
	if id == "" {
		return fmt.Errorf("instance ID required")
	}

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
		return err
	}
	if before == nil {
		fmt.Fprintf(os.Stderr, "instance not found: %s\n", id)
		return nil
	}
	if after == nil {
		if gone {
			fmt.Fprintf(os.Stderr, "instance already stopped: %s\n", id)
			return nil
		}
		fmt.Fprintf(os.Stderr, "stopped instance %s\n", id)
		return nil
	}

	proc.Terminate(after.OwnerPID)
	proc.Terminate(after.BrowserPID)
	proc.KillByProfile(after.ProfileDir)
	err = store.Update(func(doc *registry.Document) error {
		i := doc.Find(id)
		if i >= 0 {
			doc.Instances = append(doc.Instances[:i], doc.Instances[i+1:]...)
		}
		return nil
	})
	if err != nil {
		return err
	}
	fmt.Fprintf(os.Stderr, "stopped instance %s\n", id)
	return nil
}
