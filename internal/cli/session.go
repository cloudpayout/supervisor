package cli

import "github.com/spf13/cobra"

func newSessionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "session",
		Short: "Show metadata for a supervised instance",
		Long: `Show session metadata for one supervised instance.

  supervisor session info <instance-id>`,
	}
	cmd.AddCommand(newSessionInfoCmd())
	return cmd
}
