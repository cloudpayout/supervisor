package logfmt

import (
	"bufio"
	"os"
	"syscall"
)

// AttachProcessStreams sends subsequent writes to stdout and stderr through
// structured INFO lines on the instance log. The returned stop closes the
// capture pipe.
func AttachProcessStreams(l *FileLog) (func(), error) {
	r, w, err := os.Pipe()
	if err != nil {
		return nil, err
	}
	if err := syscall.Dup2(int(w.Fd()), 1); err != nil {
		r.Close()
		w.Close()
		return nil, err
	}
	if err := syscall.Dup2(int(w.Fd()), 2); err != nil {
		r.Close()
		w.Close()
		return nil, err
	}
	_ = w.Close()
	go func() {
		defer r.Close()
		sc := bufio.NewScanner(r)
		sc.Buffer(make([]byte, 64*1024), 1024*1024)
		for sc.Scan() {
			l.Info(sc.Text())
		}
	}()
	return func() {}, nil
}
