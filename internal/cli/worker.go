package cli

import (
	"github.com/spf13/cobra"

	"github.com/cloudpayout/supervisor/internal/worker"
)

func newWorkerCmd() *cobra.Command {
	workerCmd := &cobra.Command{
		Use:    "worker",
		Short:  "Internal instance owner",
		Hidden: true,
	}
	var id string
	var headless bool
	var sessionFile string
	var logPath string
	run := &cobra.Command{
		Use:    "run",
		Hidden: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			store, err := openStore()
			if err != nil {
				return err
			}
			return worker.Run(store, id, sessionFile, logPath)
		},
	}
	run.Flags().StringVar(&id, "id", "", "instance id")
	run.Flags().BoolVar(&headless, "headless", false, "headless mode")
	// FR-001: the worker receives the session file path, not the document.
	run.Flags().StringVar(&sessionFile, "session-file", "", "path to the session file")
	run.Flags().StringVar(&logPath, "log-path", "", "instance log file")
	_ = run.MarkFlagRequired("id")
	_ = run.MarkFlagRequired("log-path")
	workerCmd.AddCommand(run)
	return workerCmd
}
