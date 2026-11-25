package worker

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
)

const (
	watchBoundary     = "frame"
	watchPortStart    = 8080
	watchPortAttempts = 64
)

// WatchGate ensures at most one WATCH session per worker (018 FR-013).
type WatchGate struct {
	mu     sync.Mutex
	active bool
}

// TryBegin reports whether this worker may start a new watch session.
func (g *WatchGate) TryBegin() bool {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.active {
		return false
	}
	g.active = true
	return true
}

// End clears the active watch session.
func (g *WatchGate) End() {
	g.mu.Lock()
	g.active = false
	g.mu.Unlock()
}

// BindLoopback tries 127.0.0.1:8080 and up to 63 further ports (018 FR-014).
func BindLoopback() (net.Listener, int, error) {
	for port := watchPortStart; port < watchPortStart+watchPortAttempts; port++ {
		ln, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", port))
		if err == nil {
			return ln, port, nil
		}
	}
	return nil, 0, fmt.Errorf("no loopback port")
}

// ParseWatchFPS parses the fps argument on a WATCH line.
func ParseWatchFPS(line string) (int, bool) {
	line = strings.TrimRight(line, "\r\n")
	if !strings.HasPrefix(line, "WATCH ") {
		return 0, false
	}
	fps, err := strconv.Atoi(strings.TrimSpace(strings.TrimPrefix(line, "WATCH ")))
	if err != nil || fps < 1 || fps > 30 {
		return 0, false
	}
	return fps, true
}

type latestJPEG struct {
	v atomic.Value // []byte
}

func (l *latestJPEG) store(b []byte) {
	if len(b) == 0 {
		return
	}
	c := make([]byte, len(b))
	copy(c, b)
	l.v.Store(c)
}

func (l *latestJPEG) load() []byte {
	x := l.v.Load()
	if x == nil {
		return nil
	}
	return x.([]byte)
}

// RunWatchSession serves MJPEG on loopback and holds conn until teardown (018 FR-004, FR-005).
func RunWatchSession(conn net.Conn, browserCtx context.Context, captureMu *sync.Mutex, fps int, gate *WatchGate) {
	defer conn.Close()
	if gate != nil && !gate.TryBegin() {
		fmt.Fprintf(conn, "ERR already watching\n")
		return
	}
	if gate != nil {
		defer gate.End()
	}

	ln, port, err := BindLoopback()
	if err != nil {
		fmt.Fprintf(conn, "ERR %s\n", oneLine(err.Error()))
		return
	}

	watchCtx, cancelWatch := context.WithCancel(browserCtx)
	defer cancelWatch()

	latest := &latestJPEG{}
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		streamMJPEG(w, watchCtx, fps, latest)
	})
	srv := &http.Server{Handler: mux}
	go func() { _ = srv.Serve(ln) }()

	scDone := make(chan struct{})
	go func() {
		defer close(scDone)
		_ = runScreencast(watchCtx, browserCtx, latest, captureMu)
	}()

	url := fmt.Sprintf("http://127.0.0.1:%d/", port)
	if _, err := fmt.Fprintf(conn, "OK %s\n", url); err != nil {
		cancelWatch()
		_ = srv.Close()
		<-scDone
		return
	}

	holdConnUntilDone(conn, browserCtx)

	cancelWatch()
	_ = srv.Close()
	<-scDone
}

func holdConnUntilDone(conn net.Conn, browserCtx context.Context) {
	done := make(chan struct{})
	go func() {
		_, _ = io.Copy(io.Discard, conn)
		close(done)
	}()
	select {
	case <-done:
	case <-browserCtx.Done():
		_ = conn.Close()
		<-done
	}
}

func streamMJPEG(w http.ResponseWriter, ctx context.Context, fps int, latest *latestJPEG) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "multipart/x-mixed-replace; boundary="+watchBoundary)
	w.WriteHeader(http.StatusOK)
	flusher.Flush()

	interval := time.Second / time.Duration(fps)
	if interval < time.Millisecond {
		interval = time.Millisecond
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			jpeg := latest.load()
			if jpeg == nil {
				continue
			}
			if err := writeMJPEGPart(w, jpeg); err != nil {
				return
			}
			flusher.Flush()
		}
	}
}

func writeMJPEGPart(w io.Writer, jpeg []byte) error {
	if _, err := fmt.Fprintf(w, "--%s\r\nContent-Type: image/jpeg\r\nContent-Length: %d\r\n\r\n", watchBoundary, len(jpeg)); err != nil {
		return err
	}
	if _, err := w.Write(jpeg); err != nil {
		return err
	}
	_, err := fmt.Fprint(w, "\r\n")
	return err
}

func runScreencast(watchCtx, taskCtx context.Context, latest *latestJPEG, captureMu *sync.Mutex) error {
	framesDone := make(chan struct{})
	go func() {
		defer close(framesDone)
		for ev, err := range chromedp.Events(watchCtx, page.ScreencastFrame) {
			if err != nil {
				return
			}
			sid := ev.SessionID
			data := append([]byte(nil), ev.Data...)
			go func() {
				_, _ = chromedp.Call(taskCtx, page.ScreencastFrameAck, page.ScreencastFrameAckParams{SessionID: sid})
			}()
			latest.store(data)
			_ = captureMu // screencast does not hold capture mutex (018 FR-012)
		}
	}()

	q := int64(80)
	if _, err := chromedp.Call(taskCtx, page.StartScreencast, page.StartScreencastParams{
		Format:  page.StartScreencastFormatJpeg,
		Quality: &q,
	}); err != nil {
		return err
	}
	defer func() {
		_, _ = chromedp.Call(taskCtx, page.StopScreencast, cdp.Empty{})
	}()

	<-watchCtx.Done()
	<-framesDone
	return nil
}

// ReadWatchOK reads `OK <url>` from the worker after WATCH.
func ReadWatchOK(r *bufio.Reader) (string, error) {
	line, err := r.ReadString('\n')
	if err != nil {
		return "", err
	}
	line = strings.TrimRight(line, "\r\n")
	if strings.HasPrefix(line, "ERR ") || line == "ERR" {
		return "", fmt.Errorf("%s", strings.TrimSpace(strings.TrimPrefix(line, "ERR")))
	}
	if !strings.HasPrefix(line, "OK ") {
		return "", fmt.Errorf("bad response")
	}
	url := strings.TrimSpace(strings.TrimPrefix(line, "OK "))
	if url == "" {
		return "", fmt.Errorf("bad response")
	}
	return url, nil
}