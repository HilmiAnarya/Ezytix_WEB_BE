package utils

import (
	"fmt"
	"strings"
)

// FormatDuration mengubah menit (int) menjadi format string "Xj Ym"
func FormatDuration(totalMinutes int) string {
	hours := totalMinutes / 60
	minutes := totalMinutes % 60

	if hours > 0 {
		return fmt.Sprintf("%dj %dm", hours, minutes)
	}
	return fmt.Sprintf("%dm", minutes)
}

// FormatRupiah mengubah float64 menjadi format mata uang IDR
// Contoh: 1500000 -> "Rp 1.500.000"
func FormatRupiah(amount float64) string {
	// Konversi ke integer untuk membuang desimal (biasanya tiket tidak pakai sen)
	intAmount := int64(amount)
	
	// Jika negatif, tangani tanda minus
	sign := ""
	if intAmount < 0 {
		sign = "-"
		intAmount = -intAmount
	}

	// Konversi ke string
	str := fmt.Sprintf("%d", intAmount)
	n := len(str)
	
	// Sisipkan titik setiap 3 digit dari belakang
	var result strings.Builder
	for i, v := range str {
		if (n-i)%3 == 0 && i != 0 {
			result.WriteRune('.')
		}
		result.WriteRune(v)
	}

	return fmt.Sprintf("Rp %s%s", sign, result.String())
}