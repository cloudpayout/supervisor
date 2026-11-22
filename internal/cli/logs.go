package cli

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/cloudpayout/supervisor/internal/logfmt"
	"github.com/cloudpayout/supervisor/internal/registry"
)

func newLogsCmd() *cobra.Command {
	var follow bool
	var recent bool
	var level string
	cmd := &cobra.Command{
		Use:   "logs [instance-id]",
		Short: "Show instance logs or recent untracked logs",
		Long: `Print an instance log, list logs for instances that are no longer in status, or delete eligible log files.

  supervisor logs <instance-id>
  supervisor logs <instance-id> --follow
  supervisor logs <instance-id> --level error
  supervisor logs --recent
  supervisor logs clear`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if recent && follow {
				return fmt.Errorf("usage: --follow cannot be combined with --recent")
			}
			if recent && level != "" {
				return fmt.Errorf("usage: --level cannot be combined with --recent")
			}
			if level != "" && level != "error" && level != "notice" {
				return fmt.Errorf("usage: invalid --level %q", level)
			}
			if recent {
				if len(args) != 0 {
					return fmt.Errorf("usage: logs --recent does not take an instance ID")
				}
				return runLogsRecent()
			}
			if len(args) != 1 || strings.TrimSpace(args[0]) == "" {
				return fmt.Errorf("usage: instance ID required")
			}
			return runLogs(strings.TrimSpace(args[0]), follow, level)
		},
	}
	cmd.Flags().BoolVarP(&follow, "follow", "f", false, "Print new lines until interrupted")
	cmd.Flags().BoolVar(&recent, "recent", false, "List logs for instances that are not in status")
	cmd.Flags().StringVar(&level, "level", "", "Minimum severity: error or notice")
	cmd.AddCommand(newLogsClearCmd())
	return cmd
}

func runLogs(id string, follow bool, level string) error {
	layout, err := registry.Resolve()
	if err != nil {
		return err
	}
	store := mustStore()
	path, tracked, err := resolveLogPath(store, layout, id)
	if err != nil {
		return err
	}
	if !tracked {
		fmt.Fprintf(os.Stderr, "instance %s is no longer in status\n", id)
	}
	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("log not found: %s", id)
	}
	defer f.Close()
	return streamLog(f, follow, level)
}

func resolveLogPath(store registry.Store, layout registry.Layout, id string) (string, bool, error) {
	canonical := layout.LogPath(id)
	var tracked bool
	var registered string
	err := store.View(func(doc *registry.Document) error {
		i := doc.Find(id)
		if i < 0 {
			return nil
		}
		tracked = true
		registered = doc.Instances[i].LogPath
		return nil
	})
	if err != nil {
		return "", false, err
	}
	if tracked && registered != "" {
		if st, err := os.Stat(registered); err == nil && !st.IsDir() {
			return registered, true, nil
		}
	}
	if st, err := os.Stat(canonical); err == nil && !st.IsDir() {
		return canonical, tracked, nil
	}
	return "", tracked, fmt.Errorf("log not found: %s", id)
}

func streamLog(f *os.File, follow bool, level string) error {
	var sig chan os.Signal
	if follow {
		sig = make(chan os.Signal, 1)
		notifyLog(sig)
		defer stopNotifyLog(sig)
	}
	r := bufio.NewReader(f)
	var pending string
	for {
		if follow {
			select {
			case <-sig:
				return nil
			default:
			}
		}
		chunk, err := r.ReadString('\n')
		pending += chunk
		complete := strings.HasSuffix(pending, "\n") || (err == io.EOF && !follow && pending != "")
		if complete && pending != "" {
			emitLogLine(pending, level)
			pending = ""
		}
		if err == nil {
			continue
		}
		if err != io.EOF {
			return err
		}
		if !follow {
			return nil
		}
		select {
		case <-sig:
			return nil
		default:
		}
		select {
		case <-sig:
			return nil
		case <-time.After(500 * time.Millisecond):
		}
	}
}

func emitLogLine(line, level string) {
	text := strings.TrimRight(line, "\r\n")
	if text == "" && strings.TrimSpace(line) == "" {
		return
	}
	if level != "" && !logfmt.KeepLine(text, level) {
		return
	}
	fmt.Println(text)
}

func runLogsRecent() error {
	layout, err := registry.Resolve()
	if err != nil {
		return err
	}
	store := mustStore()
	tracked := map[string]struct{}{}
	if err := store.View(func(doc *registry.Document) error {
		for _, inst := range doc.Instances {
			tracked[inst.ID] = struct{}{}
		}
		return nil
	}); err != nil {
		return err
	}
	matches, err := filepath.Glob(filepath.Join(layout.LogsDir(), "*.log"))
	if err != nil {
		return err
	}
	type row struct {
		id    string
		when  time.Time
		error string
	}
	var rows []row
	for _, path := range matches {
		id := strings.TrimSuffix(filepath.Base(path), ".log")
		if _, ok := tracked[id]; ok {
			continue
		}
		st, err := os.Stat(path)
		if err != nil || st.IsDir() {
			continue
		}
		summary := "no error recorded"
		if msg, ok := logfmt.LastErrorInFile(path); ok && msg != "" {
			summary = truncateRunes(msg, 80)
		}
		rows = append(rows, row{id: id, when: st.ModTime().UTC(), error: summary})
	}
	sort.Slice(rows, func(i, j int) bool {
		if rows[i].when.Equal(rows[j].when) {
			return rows[i].id > rows[j].id
		}
		return rows[i].when.After(rows[j].when)
	})
	omitted := 0
	if len(rows) > 50 {
		omitted = len(rows) - 50
		rows = rows[:50]
	}
	if len(rows) == 0 {
		fmt.Println("No orphaned logs.")
		return nil
	}
	fmt.Printf("ID\tLAST_ACTIVITY_UTC\tLAST_ERROR\n")
	for _, row := range rows {
		fmt.Printf("%s\t%s\t%s\n", row.id, row.when.Format(time.RFC3339), row.error)
	}
	if omitted > 0 {
		fmt.Printf("%d additional orphaned logs were not shown\n", omitted)
	}
	return nil
}

func truncateRunes(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	if n <= 3 {
		return string(r[:n])
	}
	return string(r[:n-3]) + "..."
}
