package util

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
