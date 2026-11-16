package cli

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/spf13/cobra"

	"github.com/cloudpayout/supervisor/internal/registry"
)

const screenshotBudget = 15 * time.Second

var pngMagic = []byte{0x89, 'P', 'N', 'G'}

// newScreenshotCmd saves one PNG of a running instance (FR-001, FR-002, FR-013).
func newScreenshotCmd() *cobra.Command {
	var output string
	cmd := &cobra.Command{
		Use:   "screenshot <instance-id>",
		Short: "Save a PNG of a running instance's current page",
		Long: `Saves one PNG of the page a running instance is showing.

The instance ID is required. The command does not prompt.

With no output path, the file is screenshot-<id>-<YYYYMMDD-HHMMSS>.png in the
current directory. A name that is already taken gets a -2, -3, … suffix.

Use --output or -o to write a relative or absolute path instead. An existing
file is not replaced.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runScreenshot(cmd.Flags().Changed("output"), output, args[0])
		},
	}
	cmd.Flags().StringVarP(&output, "output", "o", "", "Relative or absolute PNG path (optional)")
	return cmd
}

// runScreenshot captures one running instance within 15 seconds (FR-004, FR-012).
// A blank ID is a usage error (FR-003). A missing row is not found, and any
// other state is not running (FR-010).
func runScreenshot(outputSet bool, output, rawID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), screenshotBudget)
	defer cancel()

	id := strings.TrimSpace(rawID)
	if id == "" {
		return fmt.Errorf("instance ID required")
	}
	if outputSet && strings.TrimSpace(output) == "" {
		return fmt.Errorf("output path required")
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

	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("screenshot failed: %s: %w", id, err)
	}
	var dest, printed string
	if outputSet {
		dest, printed, err = ResolveScreenshotOutput(cwd, output)
		if err != nil {
			return fmt.Errorf("screenshot failed: %s: %w", id, err)
		}
	}

	layout, err := registry.Resolve()
	if err != nil {
		return fmt.Errorf("screenshot failed: %s: %w", id, err)
	}
	png, err := fetchPNG(ctx, layout.CaptureSocket(id))
	if err != nil {
		return fmt.Errorf("screenshot failed: %s: %w", id, err)
	}

	if !outputSet {
		dest, printed, err = PublishDefaultPNG(cwd, id, time.Now(), png)
		if err != nil {
			return fmt.Errorf("screenshot failed: %s: %w", id, err)
		}
	} else if err := PublishPNG(dest, png); err != nil {
		return fmt.Errorf("screenshot failed: %s: %w", id, err)
	}
	fmt.Println(printed)
	return nil
}

func singlePathSegment(id string) bool {
	return id != "." && id != ".." && filepath.Base(id) == id && !strings.Contains(id, string(filepath.Separator))
}

// ResolveScreenshotOutput maps a relative or absolute --output value to a
// destination file (FR-007). printed is the trimmed flag value. The destination
// must not exist, must not be a directory, and its parent must exist.
func ResolveScreenshotOutput(cwd, flag string) (dest, printed string, err error) {
	printed = strings.TrimSpace(flag)
	if printed == "" {
		return "", "", errors.New("output path required")
	}
	if filepath.IsAbs(printed) {
		dest = filepath.Clean(printed)
	} else {
		dest = filepath.Clean(filepath.Join(cwd, printed))
	}
	if err := checkNewFile(dest); err != nil {
		return "", "", err
	}
	return dest, printed, nil
}

func checkNewFile(dest string) error {
	parent := filepath.Dir(dest)
	st, err := os.Stat(parent)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("output parent missing: %s", parent)
		}
		return err
	}
	if !st.IsDir() {
		return fmt.Errorf("output parent is not a directory: %s", parent)
	}
	st, err = os.Stat(dest)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	if st.IsDir() {
		return fmt.Errorf("output is a directory: %s", dest)
	}
	return fmt.Errorf("output already exists: %s", dest)
}

// DefaultScreenshotName is screenshot-<id>-<YYYYMMDD-HHMMSS>.png in local time (FR-006).
func DefaultScreenshotName(id string, now time.Time) string {
	return fmt.Sprintf("screenshot-%s-%s.png", id, now.Local().Format("20060102-150405"))
}

func suffixedPNGName(name string, n int) string {
	if n <= 1 {
		return name
	}
	ext := filepath.Ext(name)
	base := strings.TrimSuffix(name, ext)
	return fmt.Sprintf("%s-%d%s", base, n, ext)
}

// PublishDefaultPNG writes png to a new screenshot-<id>-<time>.png in dir (FR-006).
// The returned printed path is the relative file name. A taken name gets a
// -2, -3, … suffix instead of replacing the existing file.
func PublishDefaultPNG(dir, id string, now time.Time, png []byte) (dest, printed string, err error) {
	base := DefaultScreenshotName(id, now)
	var last error
	for n := 1; n < 1000; n++ {
		printed = suffixedPNGName(base, n)
		dest = filepath.Join(dir, printed)
		err = PublishPNG(dest, png)
		if err == nil {
			return dest, printed, nil
		}
		last = err
		if !errors.Is(err, os.ErrExist) {
			return "", "", err
		}
	}
	return "", "", last
}

// PublishPNG writes png to dest via a temporary file and os.Link (FR-011).
// An existing dest is left unchanged. The temporary file is removed on
// success and on failure.
func PublishPNG(dest string, png []byte) error {
	if len(png) < len(pngMagic) || !bytes.HasPrefix(png, pngMagic) {
		return errors.New("payload is not a PNG")
	}
	dir := filepath.Dir(dest)
	tmp, err := os.CreateTemp(dir, ".screenshot-*.partial")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	linked := false
	defer func() {
		if !linked {
			os.Remove(tmpName)
		}
	}()
	if _, err := tmp.Write(png); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if err := os.Link(tmpName, dest); err != nil {
		if errors.Is(err, os.ErrExist) {
			return err
		}
		return err
	}
	linked = true
	os.Remove(tmpName)
	return nil
}

// fetchPNG dials the capture socket until it accepts or ctx ends (FR-012).
// A socket that is not listening yet is retried. A response from the worker is not.
func fetchPNG(ctx context.Context, socketPath string) ([]byte, error) {
	var last error
	for {
		if err := ctx.Err(); err != nil {
			if last != nil {
				return nil, last
			}
			return nil, err
		}
		png, err := fetchPNGOnce(ctx, socketPath)
		if err == nil || !socketNotReady(err) {
			return png, err
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

func fetchPNGOnce(ctx context.Context, socketPath string) ([]byte, error) {
	var dialer net.Dialer
	conn, err := dialer.DialContext(ctx, "unix", socketPath)
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	if deadline, ok := ctx.Deadline(); ok {
		_ = conn.SetDeadline(deadline)
	}
	if _, err := io.WriteString(conn, "CAPTURE\n"); err != nil {
		return nil, err
	}
	png, err := readCapture(bufio.NewReader(conn))
	if err != nil {
		return nil, err
	}
	return png, nil
}

func socketNotReady(err error) bool {
	return errors.Is(err, os.ErrNotExist) || errors.Is(err, syscall.ENOENT) || errors.Is(err, syscall.ECONNREFUSED)
}

func readCapture(br *bufio.Reader) ([]byte, error) {
	line, err := br.ReadString('\n')
	if err != nil {
		return nil, err
	}
	line = strings.TrimRight(line, "\r\n")
	if strings.HasPrefix(line, "ERR ") || line == "ERR" {
		return nil, errors.New(strings.TrimSpace(strings.TrimPrefix(line, "ERR")))
	}
	var n int
	if _, err := fmt.Sscanf(line, "OK %d", &n); err != nil || n <= 0 || n > 32<<20 {
		return nil, errors.New("bad capture response")
	}
	buf := make([]byte, n)
	if _, err := io.ReadFull(br, buf); err != nil {
		return nil, err
	}
	if !bytes.HasPrefix(buf, pngMagic) {
		return nil, errors.New("payload is not a PNG")
	}
	return buf, nil
}
