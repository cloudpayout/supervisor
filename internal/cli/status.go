package cli

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"

	"github.com/cloudpayout/supervisor/internal/memlimit"
	"github.com/cloudpayout/supervisor/internal/metrics"
	"github.com/cloudpayout/supervisor/internal/registry"
)

const statusInterval = 2 * time.Second

func newStatusCmd() *cobra.Command {
	var follow bool
	cmd := &cobra.Command{
		Use:   "status",
		Args:  cobra.NoArgs,
		Short: "Print a snapshot of tracked instances and resource usage",
		Long: `Print instance state and CPU/memory once and exit.

MEMORY is mebibytes (MiB), not decimal megabytes. Without --memory-limit the column is current use only (for example 245.3 MiB). With a limit it is current/limit (for example 245.3 MiB/512.0 MiB). While starting, current use is unavailable (unavailable/512.0 MiB when a limit is set).

NOTE is empty unless aggregated memory is at or above the warning line, then it shows "threshold exceeded". The instance stays running.

Use --follow or -f to refresh the table about every 2 seconds until interrupted.
Full-screen clearing applies only in follow mode on an interactive terminal.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if follow {
				return runStatusFollow()
			}
			return runStatusOnce()
		},
	}
	cmd.Flags().BoolVarP(&follow, "follow", "f", false, "Refresh output every ~2s until interrupted")
	return cmd
}

func runStatusOnce() error {
	store := mustStore()
	return renderStatus(store, false)
}

func runStatusFollow() error {
	store := mustStore()
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(sig)

	ticker := time.NewTicker(statusInterval)
	defer ticker.Stop()
	for {
		if err := renderStatus(store, isTTY(os.Stdout)); err != nil {
			return err
		}
		ticker.Reset(statusInterval)
		select {
		case <-sig:
			fmt.Println()
			return nil
		case <-ticker.C:
		}
	}
}

func renderStatus(store registry.Store, clearScreen bool) error {
	var rows []registry.Instance
	err := store.Update(func(doc *registry.Document) error {
		registry.Apply(doc, time.Now().UTC())
		rows = append([]registry.Instance(nil), doc.Instances...)
		return nil
	})
	if err != nil {
		return err
	}
	if clearScreen && isTTY(os.Stdout) {
		fmt.Print("\033[H\033[2J")
	}
	if len(rows) == 0 {
		fmt.Println("No instances are running.")
		return nil
	}
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tSTATE\tENTRY\tCPU%\tMEMORY\tNOTE")
	shown := 0
	for _, inst := range rows {
		if inst.State != registry.StateStarting && inst.State != registry.StateRunning {
			continue
		}
		shown++
		cpu := "unavailable"
		var rss uint64
		ok := false
		if inst.State == registry.StateRunning {
			snap := metrics.Aggregate(inst.OwnerPID, inst.BrowserPID)
			if snap.OK {
				cpu = fmt.Sprintf("%.1f", snap.CPUPercent)
				rss = snap.MemoryRSS
				ok = true
			}
		}
		mem := formatMemory(inst, ok, rss)
		note := noteFor(inst, ok, rss)
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n", inst.ID, inst.State, inst.Entry, cpu, mem, note)
	}
	if shown == 0 {
		fmt.Println("No instances are running.")
		return nil
	}
	return w.Flush()
}

func formatMemory(inst registry.Instance, metricsOK bool, rss uint64) string {
	if inst.MemoryLimitBytes == 0 {
		if !metricsOK {
			return "unavailable"
		}
		return memlimit.FormatMiBOneDecimal(rss)
	}
	limit := memlimit.FormatMiBOneDecimal(inst.MemoryLimitBytes)
	if !metricsOK {
		return "unavailable/" + limit
	}
	return memlimit.FormatMiBOneDecimal(rss) + "/" + limit
}

func noteFor(inst registry.Instance, metricsOK bool, rss uint64) string {
	if inst.State != registry.StateRunning || !metricsOK || inst.MemoryLimitBytes == 0 {
		return ""
	}
	warn := inst.WarningBytes()
	if warn > 0 && rss >= warn {
		return registry.NoteThresholdExceeded
	}
	return ""
}

func isTTY(f *os.File) bool {
	st, err := f.Stat()
	if err != nil {
		return false
	}
	return st.Mode()&os.ModeCharDevice != 0
}
