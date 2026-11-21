package unit

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/cloudpayout/supervisor/internal/logfmt"
)

func TestFormatParseRoundTrip(t *testing.T) {
	line := logfmt.Format(logfmt.LevelError, "currency failed")
	ok, _, level, msg := logfmt.ParseLine(line)
	if !ok || level != "ERROR" || msg != "currency failed" {
		t.Fatalf("parse %q -> %v %s %s", line, ok, level, msg)
	}
	if strings.Contains(logfmt.Format(logfmt.LevelInfo, "a\nb"), "\n") {
		t.Fatal("message kept a newline")
	}
}

func TestKeepLine(t *testing.T) {
	errLine := logfmt.Format(logfmt.LevelError, "boom")
	note := logfmt.Format(logfmt.LevelNotice, "phase")
	info := logfmt.Format(logfmt.LevelInfo, "detail")
	if !logfmt.KeepLine(errLine, "error") || logfmt.KeepLine(note, "error") || logfmt.KeepLine(info, "error") {
		t.Fatal("error filter")
	}
	if !logfmt.KeepLine(errLine, "notice") || !logfmt.KeepLine(note, "notice") || logfmt.KeepLine(info, "notice") {
		t.Fatal("notice filter")
	}
	if logfmt.KeepLine("raw worker text", "error") {
		t.Fatal("unstructured line passed filter")
	}
}

func TestRedactAndLastError(t *testing.T) {
	if got := logfmt.Redact("cookie record-secret-value here", []string{"record-secret-value"}); strings.Contains(got, "record-secret-value") {
		t.Fatal(got)
	}
	dir := t.TempDir()
	path := filepath.Join(dir, "id.log")
	body := logfmt.Format(logfmt.LevelInfo, "hello") + "\n" + logfmt.Format(logfmt.LevelError, "final failure") + "\n"
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	msg, ok := logfmt.LastErrorInFile(path)
	if !ok || msg != "final failure" {
		t.Fatalf("last error %q %v", msg, ok)
	}
}
