package logfmt

import (
	"bufio"
	"os"
	"path/filepath"
	"sync"
)

// FileLog appends structured lines to one instance log.
type FileLog struct {
	mu      sync.Mutex
	f       *os.File
	secrets []string
}

// Open appends to path, creating parent directories.
func Open(path string, secrets []string) (*FileLog, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, err
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return nil, err
	}
	return &FileLog{f: f, secrets: secrets}, nil
}

func (l *FileLog) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.f == nil {
		return nil
	}
	err := l.f.Close()
	l.f = nil
	return err
}

// WriteLevel appends one structured line at level.
func (l *FileLog) WriteLevel(level, msg string) {
	msg = flatten(Redact(msg, l.secrets))
	if msg == "" {
		return
	}
	line := Format(level, msg) + "\n"
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.f == nil {
		return
	}
	_, _ = l.f.WriteString(line)
}

func (l *FileLog) Info(msg string)   { l.WriteLevel(LevelInfo, msg) }
func (l *FileLog) Notice(msg string) { l.WriteLevel(LevelNotice, msg) }
func (l *FileLog) Error(msg string)  { l.WriteLevel(LevelError, msg) }

// LastErrorInFile returns the message of the last structured ERROR line.
func LastErrorInFile(path string) (string, bool) {
	f, err := os.Open(path)
	if err != nil {
		return "", false
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 64*1024), 1024*1024)
	var msg string
	found := false
	for sc.Scan() {
		ok, _, level, m := ParseLine(sc.Text())
		if ok && level == LevelError {
			msg = m
			found = true
		}
	}
	return msg, found
}
