package memlimit

import "fmt"

// FormatMiBOneDecimal formats bytes as one-decimal mebibytes (1024²), e.g. "245.3 MiB".
func FormatMiBOneDecimal(bytes uint64) string {
	mib := float64(bytes) / (1024 * 1024)
	return fmt.Sprintf("%.1f MiB", mib)
}
