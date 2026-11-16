package unit

import (
	"bytes"
	"context"
	"io"
	"net"
	"path/filepath"
	"testing"
	"time"

	"github.com/cloudpayout/supervisor/internal/worker"
)

func TestServeCapturePNG(t *testing.T) {
	dir := t.TempDir()
	sock := filepath.Join(dir, "cap.sock")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	png := []byte{0x89, 'P', 'N', 'G', 0x0d, 0x0a, 0x1a, 0x0a, 1, 2, 3}
	errCh := make(chan error, 1)
	go func() {
		errCh <- worker.ServeCapture(ctx, sock, func(context.Context) ([]byte, error) {
			return png, nil
		})
	}()

	conn := dialSocket(t, sock)
	if _, err := io.WriteString(conn, "CAPTURE\n"); err != nil {
		t.Fatal(err)
	}
	got, err := worker.ReadCaptureResponse(conn)
	conn.Close()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, png) {
		t.Fatalf("png = %v", got)
	}

	conn = dialSocket(t, sock)
	if _, err := io.WriteString(conn, "NOPE\n"); err != nil {
		t.Fatal(err)
	}
	if _, err := worker.ReadCaptureResponse(conn); err == nil {
		t.Fatal("expected ERR for a bad request")
	}
	conn.Close()

	cancel()
	select {
	case <-errCh:
	case <-time.After(2 * time.Second):
		t.Fatal("ServeCapture did not stop")
	}
}

func TestServeCaptureRejectsOversizedPNG(t *testing.T) {
	dir := t.TempDir()
	sock := filepath.Join(dir, "cap.sock")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		_ = worker.ServeCapture(ctx, sock, func(context.Context) ([]byte, error) {
			buf := make([]byte, 32<<20+1)
			copy(buf, []byte{0x89, 'P', 'N', 'G'})
			return buf, nil
		})
	}()

	conn := dialSocket(t, sock)
	defer conn.Close()
	if _, err := io.WriteString(conn, "CAPTURE\n"); err != nil {
		t.Fatal(err)
	}
	if _, err := worker.ReadCaptureResponse(conn); err == nil {
		t.Fatal("expected ERR for a PNG over 32 MiB")
	}
}

func dialSocket(t *testing.T, sock string) net.Conn {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		conn, err := net.Dial("unix", sock)
		if err == nil {
			return conn
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatalf("socket %s not ready", sock)
	return nil
}
