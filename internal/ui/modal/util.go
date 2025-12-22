package modal

import "strings"

// isControlKey checks if string is a control sequence
func isControlKey(s string) bool {
	if len(s) == 0 {
		return true
	}
	// Check for escape sequences
	if strings.HasPrefix(s, "\x1b") || strings.HasPrefix(s, "\x00") {
		return true
	}
	// Allow printable ASCII and common unicode
	for _, r := range s {
		if r < 32 && r != 9 && r != 10 && r != 13 { // Allow tab, newline, carriage return
			return true
		}
	}
	return false
}

// stripPasteMarkers removes terminal bracketed paste mode markers
func stripPasteMarkers(s string) string {
	// Remove bracketed paste start/end sequences
	s = strings.Trim(s, "[")
	s = strings.Trim(s, "]")

	// Remove other common escape sequences
	s = strings.ReplaceAll(s, "\x1b", "")
	return s
}
