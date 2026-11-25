package cli

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
	"time"

	"github.com/spf13/cobra"

	"github.com/cloudpayout/supervisor/internal/browser"
	"github.com/cloudpayout/supervisor/internal/instance"
	"github.com/cloudpayout/supervisor/internal/memlimit"
	"github.com/cloudpayout/supervisor/internal/registry"
	"github.com/cloudpayout/supervisor/internal/session"
)

// newStartCmd is the operator start command (specs/005-start-command-syntax).
// The session path is an optional positional argument. Casino is read from the file.
// Omitting --headless opens a visible window. --help prints help; -h is a usage error.
func newStartCmd() *cobra.Command {
	var headless bool
	var noLightweight bool
	var memoryLimit string
	var threshold int
	cmd := &cobra.Command{
		Use:   "start [path]",
		Short: "Start an isolated browser",
		Long: "Starts one browser instance. Pass a session file path to open that casino's page. " +
			"Omit the path to open a blank page. The default is a visible window. " +
			"Pass --headless to run without a window with a 1280x720 layout using the project's chrome-headless-shell (CHROME_PATH is ignored for headless). " +
			"Headless starts use the lightweight Chrome profile by default (015 baseline plus phase-4 GPU/compositor tuning). " +
			"Each instance profile directory is removed when the run stops. " +
			"Headed starts use full Chrome (CHROME_PATH or PATH) and only headed-safe extras. " +
			"Pass --no-lightweight to skip all lightweight and RAM-tuning flags (headed and headless). " +
			"Memory figures are mebibytes (MiB), not decimal megabytes. " +
			"--memory-limit sets a status warning policy only (NOTE column); it does not cap Chrome RSS. " +
			"--memory-limit accepts M/MiB and G/GiB (rejects MB/GB). " +
			"--threshold is 1-100 and requires --memory-limit; omitted with a limit means 100.",
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// pflag treats an undefined -h as help. A hidden flag claims that
			// shorthand so -h is a usage error and stays off the help text (FR-007).
			if cmd.Flags().Changed("shorthand-h") {
				return fmt.Errorf("usage: unknown shorthand flag: 'h' in -h")
			}
			path := ""
			if len(args) == 1 {
				path = args[0]
			}
			limitBytes, thresholdPct, err := parseMemoryFlags(memoryLimit, threshold, cmd.Flags().Changed("threshold"))
			if err != nil {
				return err
			}
			return runStart(headless, noLightweight, limitBytes, thresholdPct, path)
		},
	}
	cmd.Flags().BoolVar(&headless, "headless", false, "Run Chrome without a window (1280x720 layout)")
	cmd.Flags().BoolVar(&noLightweight, "no-lightweight", false, "Skip lightweight Chrome extras (headed and headless)")
	cmd.Flags().StringVar(&memoryLimit, "memory-limit", "", "Max memory in mebibytes: M, MiB, G, or GiB (not MB/GB)")
	cmd.Flags().IntVar(&threshold, "threshold", 0, "Warn in status NOTE at this percent of --memory-limit (1-100)")
	cmd.Flags().Bool("help", false, "help for start")
	cmd.Flags().BoolP("shorthand-h", "h", false, "rejected")
	_ = cmd.Flags().MarkHidden("shorthand-h")
	return cmd
}

func parseMemoryFlags(memoryLimit string, threshold int, thresholdSet bool) (uint64, int, error) {
	if thresholdSet && memoryLimit == "" {
		return 0, 0, fmt.Errorf("usage: --threshold requires --memory-limit")
	}
	if memoryLimit == "" {
		if thresholdSet {
			return 0, 0, fmt.Errorf("usage: --threshold requires --memory-limit")
		}
		return 0, 0, nil
	}
	bytes, err := memlimit.ParseLimit(memoryLimit)
	if err != nil {
		return 0, 0, fmt.Errorf("usage: %w", err)
	}
	pct := 100
	if thresholdSet {
		if threshold < 1 || threshold > 100 {
			return 0, 0, fmt.Errorf("usage: --threshold must be between 1 and 100")
		}
		pct = threshold
	}
	return bytes, pct, nil
}

func runStart(headless, noLightweight bool, memoryLimitBytes uint64, thresholdPct int, sessionPath string) error {
	var casinoID string
	if sessionPath != "" {
		id, _, _, err := session.Parse(sessionPath, time.Now())
		if err != nil {
			return err
		}
		casinoID = id
	}
	if headless {
		if os.Getenv("CHROME_PATH") != "" {
			fmt.Fprintln(os.Stderr, "headless ignores CHROME_PATH; using bundled chrome-headless-shell")
		}
		if err := browser.PreflightHeadless(); err != nil {
			return err
		}
	} else if err := browser.Preflight(); err != nil {
		return err
	}
	layout, err := registry.Resolve()
	if err != nil {
		return err
	}
	store, err := openStore()
	if err != nil {
		return err
	}

	var id, profile, logPath string
	err = store.Update(func(doc *registry.Document) error {
		var err error
		id, err = uniqueID(doc)
		if err != nil {
			return err
		}
		profile = layout.ProfileDir(id)
		logPath = layout.LogPath(id)
		if err := browser.EnsureDirs(profile); err != nil {
			return err
		}
		inst := registry.Instance{
			ID:         id,
			State:      registry.StateStarting,
			CreatedAt:  time.Now().UTC(),
			ProfileDir: profile,
			LogPath:    logPath,
		}
		if headless {
			v := true
			inst.Headless = &v
		}
		if casinoID != "" {
			inst.Casino = casinoID
		}
		lw := !noLightweight
		inst.Lightweight = &lw
		if memoryLimitBytes > 0 {
			inst.MemoryLimitBytes = memoryLimitBytes
			inst.MemoryThresholdPercent = thresholdPct
		}
		if doc.SchemaVersion < registry.SchemaVersionV2 {
			doc.SchemaVersion = registry.SchemaVersionV2
		}
		doc.Instances = append(doc.Instances, inst)
		return nil
	})
	if err != nil {
		return err
	}

	logf, err := browser.LogFile(logPath)
	if err != nil {
		rollbackStart(store, id, profile)
		return err
	}
	_ = logf.Close()
	args := []string{"worker", "run", "--id", id, "--log-path", logPath}
	if headless {
		args = append(args, "--headless")
	}
	if sessionPath != "" {
		args = append(args, "--session-file", sessionPath)
	}
	devnull, err := os.OpenFile(os.DevNull, os.O_WRONLY, 0)
	if err != nil {
		rollbackStart(store, id, profile)
		return err
	}
	defer devnull.Close()
	worker := exec.Command(os.Args[0], args...)
	// The worker reopens the log and wraps its own stdout/stderr (FR-002b).
	worker.Stdout = devnull
	worker.Stderr = devnull
	worker.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
	if err := worker.Start(); err != nil {
		rollbackStart(store, id, profile)
		return fmt.Errorf("start worker: %w", err)
	}
	_ = worker.Process.Release()
	fmt.Println(id)
	return nil
}

func uniqueID(doc *registry.Document) (string, error) {
	for i := 0; i < 8; i++ {
		id, err := instance.NewID()
		if err != nil {
			return "", err
		}
		if doc.Find(id) < 0 {
			return id, nil
		}
	}
	return "", fmt.Errorf("could not allocate a unique instance id")
}

func rollbackStart(store registry.Store, id, profile string) {
	_ = store.Update(func(doc *registry.Document) error {
		i := doc.Find(id)
		if i >= 0 {
			doc.Instances = append(doc.Instances[:i], doc.Instances[i+1:]...)
		}
		return nil
	})
	browser.RemoveProfile(profile)
}
