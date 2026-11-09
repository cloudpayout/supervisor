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
	run := &cobra.Command{
		Use:    "run",
		Hidden: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			store, err := openStore()
			if err != nil {
				return err
			}
			return worker.Run(store, id)
		},
	}
	run.Flags().StringVar(&id, "id", "", "instance id")
	run.Flags().BoolVar(&headless, "headless", false, "headless mode")
	_ = run.MarkFlagRequired("id")
	workerCmd.AddCommand(run)
	return workerCmd
}
