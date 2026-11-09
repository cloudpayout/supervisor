package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/cloudpayout/supervisor/internal/registry"
)

func Execute() error {
	root := newRoot()
	err := root.Execute()
	if err == nil {
		return nil
	}
	if isUsageError(err) {
		fmt.Fprintln(os.Stderr, err)
		_ = root.Usage()
		os.Exit(2)
	}
	return err
}

func isUsageError(err error) bool {
	msg := err.Error()
	return strings.Contains(msg, "mutually exclusive") ||
		strings.Contains(msg, "were all set") ||
		strings.Contains(msg, "unknown flag") ||
		strings.Contains(msg, "unknown command") ||
		strings.Contains(msg, "required flag")
}

func newRoot() *cobra.Command {
	root := &cobra.Command{
		Use:           "supervisor",
		Short:         "Manage isolated chromedp browser instances",
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	root.AddCommand(newStartCmd())
	root.AddCommand(newStatusCmd())
	root.AddCommand(newStopCmd())
	root.AddCommand(newWorkerCmd())
	return root
}

func openStore() (registry.Store, error) {
	layout, err := registry.Resolve()
	if err != nil {
		return registry.Store{}, err
	}
	if err := os.MkdirAll(layout.StateDir, 0o755); err != nil {
		return registry.Store{}, err
	}
	return registry.Store{Path: layout.RegistryPath()}, nil
}

func mustStore() registry.Store {
	s, err := openStore()
	if err != nil {
		fmt.Fprintf(os.Stderr, "supervisor: %v\n", err)
		os.Exit(1)
	}
	return s
}
