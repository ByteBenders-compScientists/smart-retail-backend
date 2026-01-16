package utils

// ParseUUID returns the provided string. Kept for backward compatibility
// with earlier uint-based helpers; prefer using UUID strings directly.
func ParseUUID(s string) string {
	return s
}
