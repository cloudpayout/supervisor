package registry

import (
	"encoding/json"
	"fmt"
	"os"
	"syscall"
)

// Store is a JSON registry guarded by an exclusive or shared flock.
type Store struct {
	Path string
}

func (s Store) Update(fn func(*Document) error) error {
	return s.withLock(syscall.LOCK_EX, fn)
}

func (s Store) View(fn func(*Document) error) error {
	return s.withLock(syscall.LOCK_SH, fn)
}

func (s Store) withLock(how int, fn func(*Document) error) error {
	if err := os.MkdirAll(dirOf(s.Path), 0o755); err != nil {
		return err
	}
	f, err := os.OpenFile(s.Path, os.O_RDWR|os.O_CREATE, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()
	if err := syscall.Flock(int(f.Fd()), how); err != nil {
		return fmt.Errorf("lock registry: %w", err)
	}
	defer syscall.Flock(int(f.Fd()), syscall.LOCK_UN)

	doc, err := readDoc(f)
	if err != nil {
		return err
	}
	if err := fn(doc); err != nil {
		return err
	}
	if how&syscall.LOCK_SH != 0 && how&syscall.LOCK_EX == 0 {
		return nil
	}
	return writeDoc(f, doc)
}

func readDoc(f *os.File) (*Document, error) {
	st, err := f.Stat()
	if err != nil {
		return nil, err
	}
	doc := &Document{SchemaVersion: SchemaVersion, Instances: []Instance{}}
	if st.Size() == 0 {
		return doc, nil
	}
	dec := json.NewDecoder(f)
	if err := dec.Decode(doc); err != nil {
		return nil, fmt.Errorf("decode registry: %w", err)
	}
	doc.normalize()
	if _, err := f.Seek(0, 0); err != nil {
		return nil, err
	}
	return doc, nil
}

func writeDoc(f *os.File, doc *Document) error {
	doc.normalize()
	raw, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		return err
	}
	raw = append(raw, '\n')
	if err := f.Truncate(0); err != nil {
		return err
	}
	if _, err := f.Seek(0, 0); err != nil {
		return err
	}
	if _, err := f.Write(raw); err != nil {
		return err
	}
	return f.Sync()
}

func dirOf(path string) string {
	for i := len(path) - 1; i >= 0; i-- {
		if path[i] == '/' {
			if i == 0 {
				return "/"
			}
			return path[:i]
		}
	}
	return "."
}
