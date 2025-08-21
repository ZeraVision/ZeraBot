package util

import "strings"

// Truncate shortens a string to the specified length, adding "..." if truncated.
// If the string is shorter than or equal to maxLen, it's returned as-is.
func Truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return "..."[:maxLen]
	}
	return s[:maxLen-3] + "..."
}

// EscapeMarkdown escapes only the essential Markdown characters that could break parsing
func EscapeMarkdown(s string) string {
	// Only escape characters that are commonly used in Markdown formatting
	// and could cause parsing issues if not escaped
	replacements := map[string]string{
		"\\": "\\\\", // Backslash must always be escaped
		"*":  "\\*",  // Asterisk for bold/italic
		"_":  "\\_",  // Underscore for italic
		"[":  "\\[",  // Square bracket for links
		"]":  "\\]",  // Square bracket for links
		"`":  "\\`",  // Backtick for code
	}

	result := s
	for old, new := range replacements {
		result = strings.ReplaceAll(result, old, new)
	}

	return result
}
