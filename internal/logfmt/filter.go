package logfmt

// KeepLine reports whether a structured line meets the minimum severity.
// min is "error" or "notice". Unstructured lines are excluded.
func KeepLine(line, min string) bool {
	ok, _, level, _ := ParseLine(stringsTrim(line))
	if !ok {
		return false
	}
	switch min {
	case "error":
		return level == LevelError
	case "notice":
		return level == LevelError || level == LevelNotice
	default:
		return false
	}
}

func stringsTrim(line string) string {
	for len(line) > 0 && (line[len(line)-1] == '\n' || line[len(line)-1] == '\r') {
		line = line[:len(line)-1]
	}
	return line
}
