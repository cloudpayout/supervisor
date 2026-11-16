package integration

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

type standInHit struct {
	cookie string
	shown  bool
}

type standIn struct {
	url   string
	name  string
	value string
	mu    sync.Mutex
	hits  []standInHit
	srv   *http.Server
}

func newStandIn(t *testing.T, name, value string) *standIn {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	s := &standIn{name: name, value: value}
	mux := http.NewServeMux()
	mux.HandleFunc("/", s.serve)
	s.srv = &http.Server{Handler: mux}
	s.url = "http://" + ln.Addr().String() + "/"
	go func() { _ = s.srv.Serve(ln) }()
	t.Cleanup(func() { _ = s.srv.Close() })
	return s
}

func (s *standIn) serve(w http.ResponseWriter, r *http.Request) {
	header := r.Header.Get("Cookie")
	shown := cookiePresent(header, s.name, s.value)
	s.mu.Lock()
	s.hits = append(s.hits, standInHit{cookie: header, shown: shown})
	s.mu.Unlock()
	color := "#0000ff"
	body := ""
	if shown {
		color = "#ff0000"
		body = "balance"
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintf(w, `<!doctype html><html><head><style>html,body{margin:0;width:100%%;height:100%%;background:%s}</style></head><body>%s</body></html>`, color, body)
}

func (s *standIn) hitsCopy() []standInHit {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]standInHit, len(s.hits))
	copy(out, s.hits)
	return out
}

func cookiePresent(header, name, value string) bool {
	for _, part := range strings.Split(header, ";") {
		part = strings.TrimSpace(part)
		if part == name+"="+value {
			return true
		}
	}
	return false
}

func writeSession(t *testing.T, body string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "session.json")
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func assertStatusRunning(t *testing.T, bin string, env []string, id string) {
	t.Helper()
	out := runStatusOnce(t, bin, env)
	if !strings.Contains(out, id) || !strings.Contains(out, "running") {
		t.Fatalf("status missing running %s: %s", id, out)
	}
}

func assertStatusGone(t *testing.T, bin string, env []string, id string) {
	t.Helper()
	out := runStatusOnce(t, bin, env)
	if strings.Contains(out, id) {
		t.Fatalf("status still lists %s: %s", id, out)
	}
}

func profilesDir(env []string) string {
	for _, item := range env {
		if strings.HasPrefix(item, "XDG_DATA_HOME=") {
			return filepath.Join(strings.TrimPrefix(item, "XDG_DATA_HOME="), "supervisor", "profiles")
		}
	}
	return ""
}
