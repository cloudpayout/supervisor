package cli

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"

	"github.com/cloudpayout/supervisor/internal/metrics"
	"github.com/cloudpayout/supervisor/internal/registry"
)

const statusInterval = 2 * time.Second

func newStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "List tracked instances and refresh resource usage",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runStatus()
		},
	}
}

func runStatus() error {
	store := mustStore()
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(sig)

	ticker := time.NewTicker(statusInterval)
	defer ticker.Stop()
	for {
		if err := renderStatus(store); err != nil {
			return err
		}
		select {
		case <-sig:
			fmt.Println()
			return nil
		case <-ticker.C:
		}
	}
}

func renderStatus(store registry.Store) error {
	var rows []registry.Instance
	err := store.Update(func(doc *registry.Document) error {
		registry.Apply(doc, time.Now().UTC())
		rows = append([]registry.Instance(nil), doc.Instances...)
		return nil
	})
	if err != nil {
		return err
	}
	if isTTY(os.Stdout) {
		fmt.Print("\033[H\033[2J")
	}
	if len(rows) == 0 {
		fmt.Println("No instances are running.")
		return nil
	}
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tSTATE\tCPU%\tMEMORY")
	shown := 0
	for _, inst := range rows {
		if inst.State != registry.StateStarting && inst.State != registry.StateRunning {
			continue
		}
		shown++
		cpu, mem := "unavailable", "unavailable"
		if inst.State == registry.StateRunning {
			snap := metrics.Aggregate(inst.OwnerPID, inst.BrowserPID)
			if snap.OK {
				cpu = fmt.Sprintf("%.1f", snap.CPUPercent)
				mem = fmt.Sprintf("%.1f MiB", float64(snap.MemoryRSS)/1024/1024)
			}
		}
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", inst.ID, inst.State, cpu, mem)
	}
	if shown == 0 {
		fmt.Println("No instances are running.")
		return nil
	}
	return w.Flush()
}

func isTTY(f *os.File) bool {
	st, err := f.Stat()
	if err != nil {
		return false
	}
	return st.Mode()&os.ModeCharDevice != 0
}
