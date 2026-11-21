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
	"time"

	"github.com/cloudpayout/supervisor/internal/registry"
)

type standInHit struct {
	cookie string
	shown  bool
}

type standInEvent struct {
	kind  string
	value string
}

type standIn struct {
	url        string
	name       string
	value      string
	blockReady bool
	slowEntry  bool
	mu         sync.Mutex
	hits       []standInHit
	paths      []string
	events     []standInEvent
	srv        *http.Server
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
	if r.URL.Path == "/supervisor-event" {
		s.mu.Lock()
		s.events = append(s.events, standInEvent{
			kind:  r.URL.Query().Get("kind"),
			value: r.URL.Query().Get("value"),
		})
		s.mu.Unlock()
		w.WriteHeader(http.StatusNoContent)
		return
	}
	s.mu.Lock()
	s.paths = append(s.paths, r.URL.Path)
	s.mu.Unlock()
	if r.URL.Query().Get("frame") == "1" {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprint(w, `<!doctype html><html><body><div class="game-logo">aviator</div></body></html>`)
		return
	}
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
	block := "false"
	if s.blockReady {
		block = "true"
	}
	delay := 0
	if s.slowEntry {
		delay = 2500
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintf(w, `<!doctype html><html><head><style>html,body{margin:0;width:100%%;height:100%%;background:%s}</style></head><body>
%s
<button data-supervisor-currency-open type="button">currency</button>
<div id="currencies" hidden>
<button data-supervisor-currency type="button">USD</button>
<button data-supervisor-currency type="button">MXN</button>
</div>
<button data-supervisor-play type="button">Play</button>
<button data-supervisor-demo type="button">Modo demo</button>
<iframe data-supervisor-game-frame id="game" style="width:20px;height:20px;border:0" src="about:blank"></iframe>
<script>
function ping(kind, value) {
  fetch('/supervisor-event?kind=' + encodeURIComponent(kind) + '&value=' + encodeURIComponent(value || ''));
}
document.querySelector('[data-supervisor-currency-open]').onclick = () => { document.getElementById('currencies').hidden = false; };
document.querySelectorAll('[data-supervisor-currency]').forEach((b) => {
  b.addEventListener('click', () => ping('currency', (b.textContent || '').trim()));
});
const block = %s;
const delay = %d;
function reveal() {
  if (block) return;
  const frame = document.getElementById('game');
  const src = new URL(location.href);
  src.searchParams.set('frame', '1');
  src.searchParams.set('marker', 'spribelaunch.com/aviator');
  setTimeout(() => { frame.src = src.toString(); }, delay);
}
document.querySelector('[data-supervisor-play]').onclick = () => { ping('play', ''); reveal(); };
document.querySelector('[data-supervisor-demo]').onclick = () => { ping('demo', ''); reveal(); };
</script>
</body></html>`, color, body, block, delay)
}

func (s *standIn) hitsCopy() []standInHit {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]standInHit, len(s.hits))
	copy(out, s.hits)
	return out
}

func (s *standIn) pathsCopy() []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]string, len(s.paths))
	copy(out, s.paths)
	return out
}

func (s *standIn) eventsCopy() []standInEvent {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]standInEvent, len(s.events))
	copy(out, s.events)
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

func waitEntry(t *testing.T, regPath, id, want string) registry.Instance {
	t.Helper()
	deadline := time.Now().Add(70 * time.Second)
	for time.Now().Before(deadline) {
		doc := readDoc(t, regPath)
		found := false
		for _, inst := range doc.Instances {
			if inst.ID != id {
				continue
			}
			found = true
			if inst.Entry == want {
				return inst
			}
		}
		if !found && want != "" {
			t.Fatalf("instance %s left the registry before entry %s\nlog:\n%s", id, want, debugLog(t, regPath, id))
		}
		time.Sleep(200 * time.Millisecond)
	}
	t.Fatalf("instance %s did not reach entry %s", id, want)
	return registry.Instance{}
}

func debugLog(t *testing.T, regPath, id string) string {
	t.Helper()
	// Registry lives at <state>/supervisor/instances.json.
	logPath := filepath.Join(filepath.Dir(regPath), "logs", id+".log")
	raw, err := os.ReadFile(logPath)
	if err != nil {
		return err.Error()
	}
	return string(raw)
}

func waitGone(t *testing.T, regPath, id string) {
	t.Helper()
	deadline := time.Now().Add(70 * time.Second)
	for time.Now().Before(deadline) {
		doc := readDoc(t, regPath)
		present := false
		for _, inst := range doc.Instances {
			if inst.ID == id {
				present = true
			}
		}
		if !present {
			return
		}
		time.Sleep(200 * time.Millisecond)
	}
	t.Fatalf("instance %s still in registry", id)
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
