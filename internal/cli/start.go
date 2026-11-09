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
	"github.com/cloudpayout/supervisor/internal/registry"
)

func newStartCmd() *cobra.Command {
	var headed, headless bool
	cmd := &cobra.Command{
		Use:   "start",
		Short: "Start an isolated browser on a blank page",
		Long:  "Starts one chromedp instance on about:blank. Default is headed. --headed and --headless are mutually exclusive.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runStart(headed, headless)
		},
	}
	cmd.Flags().BoolVar(&headed, "headed", false, "Show a visible Chrome window (default when neither flag is set)")
	cmd.Flags().BoolVar(&headless, "headless", false, "Run Chrome without a window")
	cmd.MarkFlagsMutuallyExclusive("headed", "headless")
	return cmd
}

func runStart(headed, headless bool) error {
	if err := browser.Preflight(); err != nil {
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

	useHeadless := headless && !headed
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
		if useHeadless {
			v := true
			inst.Headless = &v
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
	args := []string{"worker", "run", "--id", id}
	if useHeadless {
		args = append(args, "--headless")
	}
	worker := exec.Command(os.Args[0], args...)
	worker.Stdout = logf
	worker.Stderr = logf
	worker.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
	if err := worker.Start(); err != nil {
		logf.Close()
		rollbackStart(store, id, profile)
		return fmt.Errorf("start worker: %w", err)
	}
	_ = logf.Close()
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
