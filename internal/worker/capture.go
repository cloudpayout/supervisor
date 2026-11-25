package worker

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
	"sync"
	"time"
)

const maxPNG = 32 << 20

var pngMagic = []byte{0x89, 'P', 'N', 'G'}

// CaptureHandlers serves CAPTURE and optional WATCH on one unix socket (003, 018).
type CaptureHandlers struct {
	Capture func(context.Context) ([]byte, error)
	Watch   func(conn net.Conn, fps int)
	Gate    *WatchGate
}

// ServeCapture listens on a Unix socket and serves CAPTURE and WATCH requests.
// The socket directory is 0700 and the socket is 0600. Captures of one instance
// are serialized. A payload over 32 MiB is refused. The socket file is removed
// when ctx is cancelled or the listener stops.
func ServeCapture(ctx context.Context, socketPath string, h CaptureHandlers) error {
	capture := h.Capture
	if err := os.MkdirAll(dirOf(socketPath), 0o700); err != nil {
		return err
	}
	if err := os.Remove(socketPath); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	ln, err := net.Listen("unix", socketPath)
	if err != nil {
		return err
	}
	if err := os.Chmod(socketPath, 0o600); err != nil {
		ln.Close()
		os.Remove(socketPath)
		return err
	}
	go func() {
		<-ctx.Done()
		ln.Close()
	}()
	defer os.Remove(socketPath)

	var mu sync.Mutex
	for {
		conn, err := ln.Accept()
		if err != nil {
			if ctx.Err() != nil {
				return nil
			}
			return err
		}
		go serveSocketConn(conn, &mu, capture, h.Watch, h.Gate)
	}
}

func serveSocketConn(conn net.Conn, mu *sync.Mutex, capture func(context.Context) ([]byte, error), watch func(net.Conn, int), gate *WatchGate) {
	br := bufio.NewReader(conn)
	line, err := br.ReadString('\n')
	if err != nil {
		conn.Close()
		return
	}
	trim := strings.TrimRight(line, "\r\n")
	if trim == "CAPTURE" {
		serveCaptureConn(conn, mu, capture)
		return
	}
	if fps, ok := ParseWatchFPS(line); ok {
		if watch == nil {
			fmt.Fprintf(conn, "ERR watch unavailable\n")
			conn.Close()
			return
		}
		watch(conn, fps)
		return
	}
	fmt.Fprintf(conn, "ERR bad request\n")
	conn.Close()
}

func serveCaptureConn(conn net.Conn, mu *sync.Mutex, capture func(context.Context) ([]byte, error)) {
	defer conn.Close()
	_ = conn.SetDeadline(time.Now().Add(12 * time.Second))
	mu.Lock()
	png, err := capture(context.Background())
	mu.Unlock()
	if err != nil {
		fmt.Fprintf(conn, "ERR %s\n", oneLine(err.Error()))
		return
	}
	if len(png) == 0 || len(png) > maxPNG || !bytes.HasPrefix(png, pngMagic) {
		fmt.Fprintf(conn, "ERR invalid png\n")
		return
	}
	if _, err := fmt.Fprintf(conn, "OK %d\n", len(png)); err != nil {
		return
	}
	_, _ = conn.Write(png)
}

func oneLine(s string) string {
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.ReplaceAll(s, "\r", " ")
	return strings.TrimSpace(s)
}

func dirOf(path string) string {
	i := strings.LastIndex(path, "/")
	if i <= 0 {
		return "."
	}
	return path[:i]
}

// ReadCaptureResponse reads an OK payload or an ERR line from r.
func ReadCaptureResponse(r io.Reader) ([]byte, error) {
	br := bufio.NewReader(r)
	line, err := br.ReadString('\n')
	if err != nil {
		return nil, err
	}
	line = strings.TrimRight(line, "\r\n")
	if strings.HasPrefix(line, "ERR ") || line == "ERR" {
		reason := strings.TrimPrefix(line, "ERR")
		return nil, fmt.Errorf("%s", strings.TrimSpace(reason))
	}
	if !strings.HasPrefix(line, "OK ") {
		return nil, fmt.Errorf("bad response")
	}
	var n int
	if _, err := fmt.Sscanf(line, "OK %d", &n); err != nil || n < 0 || n > maxPNG {
		return nil, fmt.Errorf("bad response")
	}
	buf := make([]byte, n)
	if _, err := io.ReadFull(br, buf); err != nil {
		return nil, err
	}
	if !bytes.HasPrefix(buf, pngMagic) {
		return nil, fmt.Errorf("payload is not a PNG")
	}
	return buf, nil
}
