package logfmt

import (
	"regexp"
	"strings"
	"time"
)

const (
	LevelError  = "ERROR"
	LevelNotice = "NOTICE"
	LevelInfo   = "INFO"
)

var linePattern = regexp.MustCompile(`^(\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}(?:\.\d+)?Z) (ERROR|NOTICE|INFO) (.*)$`)

// Format builds one structured log line without a trailing newline in the message.
func Format(level, msg string) string {
	msg = flatten(msg)
	ts := time.Now().UTC().Format(time.RFC3339Nano)
	return ts + " " + level + " " + msg
}

func flatten(msg string) string {
	msg = strings.ReplaceAll(msg, "\r\n", " ")
	msg = strings.ReplaceAll(msg, "\n", " ")
	msg = strings.ReplaceAll(msg, "\r", " ")
	return strings.TrimSpace(msg)
}

// ParseLine reports whether line matches the canonical structured pattern.
func ParseLine(line string) (ok bool, ts time.Time, level, msg string) {
	line = strings.TrimRight(line, "\r")
	m := linePattern.FindStringSubmatch(line)
	if m == nil {
		return false, time.Time{}, "", ""
	}
	parsed, err := time.Parse(time.RFC3339Nano, m[1])
	if err != nil {
		return false, time.Time{}, "", ""
	}
	return true, parsed, m[2], m[3]
}
