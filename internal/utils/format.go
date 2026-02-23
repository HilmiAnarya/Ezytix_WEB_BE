package utils

import (
	"fmt"
	"strings"
)

func FormatDuration(totalMinutes int) string {
	hours := totalMinutes / 60
	minutes := totalMinutes % 60

	if hours > 0 {
		return fmt.Sprintf("%dj %dm", hours, minutes)
	}
	return fmt.Sprintf("%dm", minutes)
}

func FormatRupiah(amount float64) string {
	intAmount := int64(amount)
	
	sign := ""
	if intAmount < 0 {
		sign = "-"
		intAmount = -intAmount
	}

	str := fmt.Sprintf("%d", intAmount)
	n := len(str)
	
	var result strings.Builder
	for i, v := range str {
		if (n-i)%3 == 0 && i != 0 {
			result.WriteRune('.')
		}
		result.WriteRune(v)
	}

	return fmt.Sprintf("Rp %s%s", sign, result.String())
}