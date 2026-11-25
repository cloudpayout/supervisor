package unit

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/cloudpayout/supervisor/internal/worker"
)

func TestServeCaptureAndWatchSocket(t *testing.T) {
	dir := t.TempDir()
	sock := filepath.Join(dir, "cap.sock")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	png := []byte{0x89, 'P', 'N', 'G', 0x0d, 0x0a, 0x1a, 0x0a, 1, 2, 3}

	errCh := make(chan error, 1)
	go func() {
		errCh <- worker.ServeCapture(ctx, sock, worker.CaptureHandlers{
			Capture: func(context.Context) ([]byte, error) {
				return png, nil
			},
			Watch: func(conn net.Conn, fps int) {
				worker.RunWatchSession(conn, context.Background(), nil, fps, &worker.WatchGate{})
			},
		})
	}()

	capConn := dialSocket(t, sock)
	if _, err := io.WriteString(capConn, "CAPTURE\n"); err != nil {
		t.Fatal(err)
	}
	got, err := worker.ReadCaptureResponse(capConn)
	capConn.Close()
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(png) {
		t.Fatalf("png mismatch")
	}

	wConn := dialSocket(t, sock)
	if _, err := io.WriteString(wConn, "WATCH 1\n"); err != nil {
		t.Fatal(err)
	}
	br := bufio.NewReader(wConn)
	url, err := worker.ReadWatchOK(br)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(url, "http://127.0.0.1:") {
		t.Fatalf("url %q", url)
	}

	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		t.Fatalf("get stream: %v", err)
	}
	defer resp.Body.Close()
	if !strings.Contains(resp.Header.Get("Content-Type"), "multipart/x-mixed-replace") {
		t.Fatalf("content-type %q", resp.Header.Get("Content-Type"))
	}

	wConn.Close()
	cancel()
	select {
	case <-errCh:
	case <-time.After(3 * time.Second):
		t.Fatal("server did not stop")
	}
}

func TestBindLoopbackPort(t *testing.T) {
	ln, port, err := worker.BindLoopback()
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()
	if port < 8080 || port >= 8080+64 {
		t.Fatalf("port %d", port)
	}
	addr := ln.Addr().String()
	if !strings.HasPrefix(addr, "127.0.0.1:") {
		t.Fatalf("addr %q", addr)
	}
	ln2, port2, err := worker.BindLoopback()
	if err != nil {
		t.Fatal(err)
	}
	defer ln2.Close()
	if port2 == port {
		t.Fatalf("ports %d %d", port, port2)
	}
}

func TestParseWatchFPS(t *testing.T) {
	fps, ok := worker.ParseWatchFPS("WATCH 10\n")
	if !ok || fps != 10 {
		t.Fatalf("got %d %v", fps, ok)
	}
	_, ok = worker.ParseWatchFPS("CAPTURE\n")
	if ok {
		t.Fatal("CAPTURE should not parse as watch")
	}
	_, ok = worker.ParseWatchFPS(fmt.Sprintf("WATCH %d\n", 0))
	if ok {
		t.Fatal("0 invalid")
	}
}