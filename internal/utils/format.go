package utils

import "fmt"

// FormatDuration mengubah menit (int) menjadi format string "Xj Ym"
func FormatDuration(totalMinutes int) string {
	hours := totalMinutes / 60
	minutes := totalMinutes % 60

	if hours > 0 {
		return fmt.Sprintf("%dj %dm", hours, minutes)
	}
	return fmt.Sprintf("%dm", minutes)
}