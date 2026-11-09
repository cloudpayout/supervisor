package worker

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/shirou/gopsutil/v4/process"

	"github.com/cloudpayout/supervisor/internal/browser"
	"github.com/cloudpayout/supervisor/internal/registry"
)

// Run owns one chromedp session until the browser exits or SIGTERM arrives.
func Run(reg registry.Store, id string) error {
	var inst registry.Instance
	err := reg.View(func(doc *registry.Document) error {
		i := doc.Find(id)
		if i < 0 {
			return fmt.Errorf("instance %s not in registry", id)
		}
		inst = doc.Instances[i]
		return nil
	})
	if err != nil {
		return err
	}
	headless := false
	if inst.Headless != nil {
		headless = *inst.Headless
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	allocCtx, allocCancel := chromedp.NewExecAllocator(ctx, browser.AllocatorOptions(inst.ProfileDir, headless)...)
	defer allocCancel()
	taskCtx, taskCancel := chromedp.NewContext(allocCtx)
	defer taskCancel()

	if err := chromedp.Do(taskCtx, chromedp.Navigate("about:blank")); err != nil {
		_ = remove(reg, id)
		return fmt.Errorf("open about:blank: %w", err)
	}

	browserPID, err := findBrowserPID(inst.ProfileDir)
	if err != nil {
		_ = remove(reg, id)
		return err
	}
	owner := os.Getpid()
	now := time.Now().UTC()
	err = reg.Update(func(doc *registry.Document) error {
		i := doc.Find(id)
		if i < 0 {
			return fmt.Errorf("instance %s disappeared", id)
		}
		doc.Instances[i].State = registry.StateRunning
		doc.Instances[i].OwnerPID = owner
		doc.Instances[i].BrowserPID = browserPID
		doc.Instances[i].LastSeenAt = &now
		return nil
	})
	if err != nil {
		_ = remove(reg, id)
		return err
	}

	<-taskCtx.Done()
	_ = remove(reg, id)
	return nil
}

func remove(reg registry.Store, id string) error {
	return reg.Update(func(doc *registry.Document) error {
		i := doc.Find(id)
		if i < 0 {
			return nil
		}
		doc.Instances = append(doc.Instances[:i], doc.Instances[i+1:]...)
		return nil
	})
}

func findBrowserPID(profile string) (int, error) {
	deadline := time.Now().Add(20 * time.Second)
	for time.Now().Before(deadline) {
		procs, err := process.Processes()
		if err != nil {
			time.Sleep(200 * time.Millisecond)
			continue
		}
		for _, p := range procs {
			cmd, err := p.Cmdline()
			if err != nil || !strings.Contains(cmd, profile) {
				continue
			}
			if strings.Contains(cmd, "--type=") {
				continue
			}
			return int(p.Pid), nil
		}
		time.Sleep(200 * time.Millisecond)
	}
	return 0, fmt.Errorf("chrome process for profile %s not found", profile)
}
