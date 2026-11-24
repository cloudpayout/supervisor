package memlimit

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

const (
	mib = 1024 * 1024
	gib = 1024 * mib
)

// ParseLimit parses a binary size string into bytes.
// Accepts M/MiB (mebibytes) and G/GiB (gibibytes). Rejects MB/GB, bare numbers, and non-positive values.
func ParseLimit(s string) (uint64, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, fmt.Errorf("invalid memory limit %q: empty", s)
	}
	i := 0
	for i < len(s) && (unicode.IsDigit(rune(s[i])) || s[i] == '.') {
		i++
	}
	if i == 0 || i == len(s) {
		return 0, fmt.Errorf("invalid memory limit %q: expected a positive size with suffix M, MiB, G, or GiB", s)
	}
	num := s[:i]
	suf := s[i:]
	if strings.Contains(num, ".") {
		return 0, fmt.Errorf("invalid memory limit %q: fractional sizes are not accepted", s)
	}
	n, err := strconv.ParseUint(num, 10, 64)
	if err != nil || n == 0 {
		return 0, fmt.Errorf("invalid memory limit %q: expected a positive size with suffix M, MiB, G, or GiB", s)
	}
	var mult uint64
	switch suf {
	case "MiB", "M":
		mult = mib
	case "GiB", "G":
		mult = gib
	case "MB", "GB":
		return 0, fmt.Errorf("invalid memory limit %q: decimal MB/GB are not accepted; use MiB/M or GiB/G", s)
	default:
		return 0, fmt.Errorf("invalid memory limit %q: expected suffix M, MiB, G, or GiB", s)
	}
	if n > ^uint64(0)/mult {
		return 0, fmt.Errorf("invalid memory limit %q: overflow", s)
	}
	return n * mult, nil
}
