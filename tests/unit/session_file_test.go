package unit

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/cloudpayout/supervisor/internal/session"
)

func TestSessionFileRules(t *testing.T) {
	now := time.Date(2026, 11, 16, 0, 0, 0, 0, time.UTC)
	dir := t.TempDir()
	secret := "runtime-secret-value"

	ok := filepath.Join(dir, "ok.json")
	write := func(name, body string) string {
		t.Helper()
		path := filepath.Join(dir, name)
		if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
			t.Fatal(err)
		}
		return path
	}
	_ = ok

	casinoID, got, err := session.Parse(write("full.json", `{
		"casino": "onewin",
		"cookies": [{
			"name": "session-id",
			"value": "`+secret+`",
			"domain": "127.0.0.1",
			"path": "/",
			"expires": "2026-12-01T00:00:00Z",
			"maxAge": 10,
			"secure": true,
			"httpOnly": true,
			"sameSite": "Lax",
			"partitioned": true
		}]
	}`), now)
	if err != nil {
		t.Fatal(err)
	}
	if casinoID != "onewin" || len(got) != 1 || got[0].Value != secret || got[0].Domain != "127.0.0.1" || got[0].Path != "/" {
		t.Fatalf("%+v", got)
	}
	if !got[0].Secure || !got[0].HTTPOnly || !got[0].Partitioned || got[0].SameSite != "Lax" {
		t.Fatalf("%+v", got[0])
	}
	if got[0].Expires == nil || !got[0].Expires.Equal(time.Date(2026, 12, 1, 0, 0, 0, 0, time.UTC)) {
		t.Fatalf("expires %+v", got[0].Expires)
	}

	_, aged, err := session.Parse(write("age.json", `{"casino":"onewin","cookies":[{"name":"a","value":"b","maxAge":90}]}`), now)
	if err != nil {
		t.Fatal(err)
	}
	wantAge := now.Add(90 * time.Second)
	if aged[0].Expires == nil || !aged[0].Expires.Equal(wantAge) {
		t.Fatalf("maxAge expiry %+v", aged[0].Expires)
	}

	_, two, err := session.Parse(write("two.json", `{"casino":"onewin","cookies":[{"name":"a","value":"b"},{"name":"c","value":"d"}]}`), now)
	if err != nil || len(two) != 2 {
		t.Fatalf("%v %+v", err, two)
	}

	for _, tc := range []struct {
		name string
		body string
	}{
		{"empty", `{"casino":"onewin","cookies":[]}`},
		{"missing casino", `{"cookies":[{"name":"a","value":"b"}]}`},
		{"empty casino", `{"casino":"","cookies":[{"name":"a","value":"b"}]}`},
		{"bad casino", `{"casino":"Onewin","cookies":[{"name":"a","value":"b"}]}`},
		{"url casino", `{"casino":"1win.com","cookies":[{"name":"a","value":"b"}]}`},
		{"unknown casino", `{"casino":"other","cookies":[{"name":"a","value":"b"}]}`},
		{"unknown field", `{"casino":"onewin","extra":true,"cookies":[{"name":"a","value":"b"}]}`},
		{"blank name", `{"casino":"onewin","cookies":[{"name":"","value":"b"}]}`},
		{"blank value", `{"casino":"onewin","cookies":[{"name":"a","value":""}]}`},
		{"bad expires", `{"casino":"onewin","cookies":[{"name":"a","value":"b","expires":"tomorrow"}]}`},
		{"fractional maxAge", `{"casino":"onewin","cookies":[{"name":"a","value":"b","maxAge":1.5}]}`},
		{"bad sameSite", `{"casino":"onewin","cookies":[{"name":"a","value":"b","sameSite":"Maybe"}]}`},
		{"unknown", `{"casino":"onewin","cookies":[{"name":"a","value":"` + secret + `","extra":true}]}`},
	} {
		_, _, err := session.Parse(write(tc.name+".json", tc.body), now)
		if err == nil {
			t.Fatalf("%s accepted", tc.name)
		}
		if strings.Contains(err.Error(), secret) {
			t.Fatalf("%s error contains the cookie value: %v", tc.name, err)
		}
	}
}
