package cli

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/cloudpayout/supervisor/internal/casino"
	"github.com/cloudpayout/supervisor/internal/registry"
)

// sessionDateLayout is UTC YYYY-MM-DD HH:MM (FR-009).
const sessionDateLayout = "2006-01-02 15:04"

func newSessionInfoCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "info [instance-id]",
		Short: "Print casino, game, strategy, and start time for an instance",
		Long: `Print four lines for one instance after reconciling the registry.

  casino: onewin
  game: aviator
  strategy: null
  date: 2026-11-26 09:15

strategy is always null in this version. date is UTC as YYYY-MM-DD HH:MM.
A missing instance ID is a usage error. An unknown ID is not found.
Success writes only those lines to standard output. Failures leave standard output empty.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id := strings.TrimSpace(args[0])
			if id == "" {
				return fmt.Errorf("instance ID required")
			}
			return runSessionInfo(id)
		},
	}
}

// FormatSessionInfo renders the four operator lines (specs/019-session-data FR-005).
func FormatSessionInfo(inst registry.Instance) string {
	casinoID := "null"
	game := "null"
	if inst.Casino != "" {
		casinoID = inst.Casino
		if found, err := casino.Lookup(inst.Casino); err == nil && found.Game != "" {
			game = found.Game
		}
	}
	date := inst.CreatedAt.UTC().Format(sessionDateLayout)
	return fmt.Sprintf("casino: %s\ngame: %s\nstrategy: null\ndate: %s\n", casinoID, game, date)
}

func runSessionInfo(id string) error {
	if !singlePathSegment(id) {
		return fmt.Errorf("instance not found: %s", id)
	}
	store := mustStore()
	var inst *registry.Instance
	err := store.Update(func(doc *registry.Document) error {
		registry.Apply(doc, time.Now().UTC())
		i := doc.Find(id)
		if i < 0 {
			return nil
		}
		cp := doc.Instances[i]
		inst = &cp
		return nil
	})
	if err != nil {
		return err
	}
	if inst == nil {
		return fmt.Errorf("instance not found: %s", id)
	}
	_, err = fmt.Fprint(os.Stdout, FormatSessionInfo(*inst))
	return err
}
