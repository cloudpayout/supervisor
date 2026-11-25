package integration

import (
	"bufio"
	"context"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"
)

func TestWatchHeadlessPrintsURL(t *testing.T) {
	requireChrome(t)
	bin := supervisorBin(t)
	env, regPath := isolatedEnv(t)
	id := startHeadless(t, bin, env)
	t.Cleanup(func() { stopID(t, bin, env, id) })
	before := waitRunning(t, regPath, id)

	urlCh := make(chan string, 1)
	errCh := make(chan error, 1)
	go func() {
		cmd := execCommand(bin, env, "watch", id)
		stdout, err := cmd.StdoutPipe()
		if err != nil {
			errCh <- err
			return
		}
		if err := cmd.Start(); err != nil {
			errCh <- err
			return
		}
		br := bufio.NewReader(stdout)
		line, err := br.ReadString('\n')
		if err != nil {
			_ = cmd.Process.Kill()
			errCh <- err
			return
		}
		urlCh <- strings.TrimSpace(line)
		deadline := time.Now().Add(15 * time.Second)
		for time.Now().Before(deadline) {
			if !strings.HasPrefix(strings.TrimSpace(line), "http://127.0.0.1:") {
				errCh <- context.DeadlineExceeded
				_ = cmd.Process.Kill()
				return
			}
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			req, _ := http.NewRequestWithContext(ctx, http.MethodGet, strings.TrimSpace(line), nil)
			resp, err := http.DefaultClient.Do(req)
			cancel()
			if err == nil {
				ct := resp.Header.Get("Content-Type")
				resp.Body.Close()
				if strings.Contains(ct, "multipart/x-mixed-replace") {
					_ = cmd.Process.Signal(os.Interrupt)
					_ = cmd.Wait()
					errCh <- nil
					return
				}
			}
			time.Sleep(200 * time.Millisecond)
		}
		_ = cmd.Process.Kill()
		errCh <- context.DeadlineExceeded
	}()

	var url string
	select {
	case url = <-urlCh:
	case err := <-errCh:
		t.Fatal(err)
	case <-time.After(12 * time.Second):
		t.Fatal("timeout waiting for watch url")
	}
	if !strings.HasPrefix(url, "http://127.0.0.1:") {
		t.Fatalf("url %q", url)
	}

	select {
	case err := <-errCh:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(20 * time.Second):
		t.Fatal("watch did not finish after interrupt")
	}

	after := waitRunning(t, regPath, id)
	if after.OwnerPID != before.OwnerPID || after.BrowserPID != before.BrowserPID {
		t.Fatalf("pids changed: before %+v after %+v", before, after)
	}
}

func TestWatchEndsWhenInstanceStops(t *testing.T) {
	requireChrome(t)
	bin := supervisorBin(t)
	env, regPath := isolatedEnv(t)
	id := startHeadless(t, bin, env)
	waitRunning(t, regPath, id)

	cmd := execCommand(bin, env, "watch", id)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	br := bufio.NewReader(stdout)
	url, err := br.ReadString('\n')
	if err != nil {
		t.Fatal(err)
	}
	url = strings.TrimSpace(url)

	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()

	stopID(t, bin, env, id)

	select {
	case err := <-done:
		if err != nil && exitCode(err) != 0 {
			t.Fatalf("watch exit %v", err)
		}
	case <-time.After(6 * time.Second):
		t.Fatal("watch did not exit within 5s of stop")
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	resp, err := http.DefaultClient.Do(req)
	if err == nil {
		resp.Body.Close()
	}
	_ = io.Discard
}