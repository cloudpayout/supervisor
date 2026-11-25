//go:build slow

package integration

import (
	"bufio"
	"bytes"
	"context"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"
)

func TestWatchCadence(t *testing.T) {
	requireChrome(t)
	bin := supervisorBin(t)
	env, regPath := isolatedEnv(t)
	id := startHeadless(t, bin, env)
	t.Cleanup(func() { stopID(t, bin, env, id) })
	waitRunning(t, regPath, id)

	rate := watchPartRate(t, bin, env, id, time.Minute)
	if rate < 0.8 || rate > 1.2 {
		t.Fatalf("default fps rate %.3f parts/s", rate)
	}
	rate = watchPartRate(t, bin, env, id, time.Minute, "--fps", "10")
	if rate < 8 || rate > 12 {
		t.Fatalf("--fps 10 rate %.3f parts/s", rate)
	}
}

func watchPartRate(t *testing.T, bin string, env []string, id string, window time.Duration, extra ...string) float64 {
	t.Helper()
	args := append([]string{"watch", id}, extra...)
	cmd := execCommand(bin, env, args...)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = cmd.Process.Signal(os.Interrupt)
		_ = cmd.Wait()
	}()

	line, err := bufio.NewReader(stdout).ReadString('\n')
	if err != nil {
		t.Fatal(err)
	}
	url := strings.TrimSpace(line)
	if !strings.HasPrefix(url, "http://127.0.0.1:") {
		t.Fatalf("url %q", url)
	}
	return multipartPartRate(t, url, window)
}

func multipartPartRate(t *testing.T, url string, window time.Duration) float64 {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), window+20*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		t.Fatal(err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if !strings.Contains(resp.Header.Get("Content-Type"), "multipart/x-mixed-replace") {
		t.Fatalf("content-type %q", resp.Header.Get("Content-Type"))
	}

	buf := make([]byte, 32<<10)
	var pending []byte
	var start time.Time
	n := 0
	marker := []byte("--frame")
	for {
		if !start.IsZero() && time.Since(start) >= window {
			break
		}
		nr, rerr := resp.Body.Read(buf)
		pending = append(pending, buf[:nr]...)
		for {
			i := bytes.Index(pending, marker)
			if i < 0 {
				if len(pending) > len(marker) {
					pending = pending[len(pending)-len(marker):]
				}
				break
			}
			if start.IsZero() {
				start = time.Now()
			}
			if time.Since(start) <= window {
				n++
			}
			pending = pending[i+len(marker):]
		}
		if rerr != nil {
			if rerr == io.EOF || ctx.Err() != nil {
				break
			}
			t.Fatalf("read stream: %v", rerr)
		}
	}
	if start.IsZero() {
		t.Fatal("no multipart frame within budget")
	}
	elapsed := time.Since(start)
	if elapsed < window {
		t.Fatalf("stream ended after %s with %d parts", elapsed, n)
	}
	return float64(n) / window.Seconds()
}
