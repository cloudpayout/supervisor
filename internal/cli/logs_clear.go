package cli

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/cloudpayout/supervisor/internal/registry"
)

func newLogsClearCmd() *cobra.Command {
	var yes bool
	cmd := &cobra.Command{
		Use:   "clear",
		Short: "Delete eligible instance log files",
		Long: `Delete eligible *.log files under the supervisor logs directory after confirmation.

Logs for instances in starting, running, or stopping are never removed.
The instance registry is not modified. When standard input is not a terminal,
pass --yes to delete without a prompt. Declining leaves every log file unchanged.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runLogsClear(yes)
		},
	}
	cmd.Flags().BoolVar(&yes, "yes", false, "Delete eligible logs without prompting (required when stdin is not a terminal)")
	return cmd
}

func protectedState(s registry.State) bool {
	return s == registry.StateStarting || s == registry.StateRunning || s == registry.StateStopping
}

func classifyLogs(layout registry.Layout, store registry.Store) (eligible, skipped []string, err error) {
	dir := layout.LogsDir()
	st, statErr := os.Stat(dir)
	if statErr != nil {
		if os.IsNotExist(statErr) {
			return nil, nil, nil
		}
		return nil, nil, statErr
	}
	if !st.IsDir() {
		return nil, nil, fmt.Errorf("logs path is not a directory")
	}
	protected := map[string]struct{}{}
	if err := store.View(func(doc *registry.Document) error {
		for _, inst := range doc.Instances {
			if protectedState(inst.State) {
				protected[inst.ID] = struct{}{}
			}
		}
		return nil
	}); err != nil {
		return nil, nil, err
	}
	matches, err := filepath.Glob(filepath.Join(dir, "*.log"))
	if err != nil {
		return nil, nil, err
	}
	for _, path := range matches {
		info, err := os.Stat(path)
		if err != nil || !info.Mode().IsRegular() {
			continue
		}
		id := strings.TrimSuffix(filepath.Base(path), ".log")
		if _, ok := protected[id]; ok {
			skipped = append(skipped, path)
			continue
		}
		eligible = append(eligible, path)
	}
	return eligible, skipped, nil
}

func runLogsClear(yes bool) error {
	layout, err := registry.Resolve()
	if err != nil {
		fmt.Fprintf(os.Stderr, "supervisor: %v\n", err)
		os.Exit(1)
	}
	store, err := openStore()
	if err != nil {
		fmt.Fprintf(os.Stderr, "supervisor: %v\n", err)
		os.Exit(1)
	}
	eligible, skipped, err := classifyLogs(layout, store)
	if err != nil {
		fmt.Fprintf(os.Stderr, "supervisor: %v\n", err)
		os.Exit(1)
	}
	if !yes {
		if !isTTY(os.Stdin) {
			return fmt.Errorf("logs clear requires --yes when stdin is not a terminal")
		}
		fmt.Fprintf(os.Stderr, "This will delete %d eligible log file(s). %d log file(s) for starting, running, or stopping instances will be kept.\n", len(eligible), len(skipped))
		fmt.Fprint(os.Stderr, "Clear eligible log files? [y/N]: ")
		line, err := bufio.NewReader(os.Stdin).ReadString('\n')
		if err != nil && err != io.EOF {
			fmt.Fprintln(os.Stdout, "logs clear cancelled")
			return nil
		}
		if !confirmYes(line) {
			fmt.Fprintln(os.Stdout, "logs clear cancelled")
			return nil
		}
	}
	for _, path := range eligible {
		if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
			fmt.Fprintf(os.Stderr, "supervisor: %v\n", err)
			os.Exit(1)
		}
	}
	if len(eligible) == 0 {
		fmt.Printf("No eligible log files to clear (%d skipped).\n", len(skipped))
		return nil
	}
	fmt.Printf("Deleted %d log file(s); skipped %d active log file(s).\n", len(eligible), len(skipped))
	return nil
}
