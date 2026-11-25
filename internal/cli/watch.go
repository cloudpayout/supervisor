package cli

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/spf13/cobra"

	"github.com/cloudpayout/supervisor/internal/registry"
	"github.com/cloudpayout/supervisor/internal/worker"
)

const watchDialBudget = 30 * time.Second

// newWatchCmd streams a local live view of a running instance (018 FR-001).
func newWatchCmd() *cobra.Command {
	var fpsRaw string
	cmd := &cobra.Command{
		Use:   "watch <instance-id>",
		Short: "Open a local live view of a running instance's page",
		Long: `Prints a loopback URL for a live JPEG stream of the page a running instance is showing.

The instance ID is required. The command does not prompt.

With no --fps flag, the stream targets 1 frame per second. --fps may be 1 through 30.

Open the printed URL in a browser on this machine. Interrupt the command to end watch
while the instance keeps running.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			fps, err := ParseWatchFPSFlag(fpsRaw)
			if err != nil {
				return err
			}
			return runWatch(fps, args[0])
		},
	}
	cmd.Flags().StringVar(&fpsRaw, "fps", "1", "Target frames per second (1-30)")
	return cmd
}

// ParseWatchFPSFlag parses --fps. Non-integers and values outside 1..30 are usage errors (018 FR-009).
func ParseWatchFPSFlag(raw string) (int, error) {
	n, err := strconv.Atoi(strings.TrimSpace(raw))
	if err != nil || ValidateWatchFPS(n) != nil {
		return 0, fmt.Errorf("usage: invalid --fps %q", raw)
	}
	return n, nil
}

// ValidateWatchFPS checks --fps is in 1..30 (018 FR-009).
func ValidateWatchFPS(fps int) error {
	if fps < 1 || fps > 30 {
		return fmt.Errorf("usage: invalid --fps %q", fmt.Sprint(fps))
	}
	return nil
}

func runWatch(fps int, rawID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), watchDialBudget)
	defer cancel()

	id := strings.TrimSpace(rawID)
	if id == "" {
		return fmt.Errorf("instance ID required")
	}
	if !singlePathSegment(id) {
		return fmt.Errorf("instance not found: %s", id)
	}

	store := mustStore()
	var inst *registry.Instance
	err := store.View(func(doc *registry.Document) error {
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
	if inst.State != registry.StateRunning {
		return fmt.Errorf("instance not running: %s", id)
	}

	layout, err := registry.Resolve()
	if err != nil {
		return fmt.Errorf("watch failed: %s: %w", id, err)
	}
	socketPath := layout.CaptureSocket(id)

	conn, err := dialWatchSocket(ctx, socketPath)
	if err != nil {
		return fmt.Errorf("watch failed: %s: %w", id, err)
	}
	defer conn.Close()

	if _, err := fmt.Fprintf(conn, "WATCH %d\n", fps); err != nil {
		return fmt.Errorf("watch failed: %s: %w", id, err)
	}
	br := bufio.NewReader(conn)
	url, err := worker.ReadWatchOK(br)
	if err != nil {
		return fmt.Errorf("watch failed: %s: %w", id, err)
	}
	fmt.Println(url)

	sig := make(chan os.Signal, 1)
	notifyWatch(sig)
	defer stopNotifyWatch(sig)

	done := make(chan struct{})
	go func() {
		_, _ = io.Copy(io.Discard, conn)
		close(done)
	}()

	select {
	case <-sig:
		conn.Close()
		<-done
		return nil
	case <-done:
		return nil
	}
}

func dialWatchSocket(ctx context.Context, socketPath string) (net.Conn, error) {
	var last error
	for {
		if err := ctx.Err(); err != nil {
			if last != nil {
				return nil, last
			}
			return nil, err
		}
		var dialer net.Dialer
		conn, err := dialer.DialContext(ctx, "unix", socketPath)
		if err == nil {
			return conn, nil
		}
		if !socketNotReady(err) {
			return nil, err
		}
		last = err
		timer := time.NewTimer(20 * time.Millisecond)
		select {
		case <-ctx.Done():
			timer.Stop()
			return nil, last
		case <-timer.C:
		}
	}
}

func notifyWatch(ch chan os.Signal) {
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
}

func stopNotifyWatch(ch chan os.Signal) {
	signal.Stop(ch)
}
