package pdfprinter

import (
	"encoding/base64"
	"github.com/skip2/go-qrcode"
	"fmt"
	"os"
)

// ImageToBase64 membaca file gambar dan mengembalikan string base64 RAW
func ImageToBase64(filePath string) (string, error) {
	// 1. Baca file gambar dari disk
	bytes, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("gagal membaca gambar %s: %v", filePath, err)
	}

	// 2. Langsung encode ke Base64
	// Kita tidak perlu deteksi JPG/PNG di sini karena di HTML template 
	// kita sudah menulis prefix <img src="data:image/png;base64,...">
	encoded := base64.StdEncoding.EncodeToString(bytes)
	
	return encoded, nil
}

// [NEW] GenerateQRCodeBase64 membuat QR code dari text dan mengembalikannya sbg string Base64
func GenerateQRCodeBase64(text string) (string, error) {
	// Generate QR Code size 256x256
	var png []byte
	png, err := qrcode.Encode(text, qrcode.Medium, 256)
	if err != nil {
		return "", fmt.Errorf("failed to generate qr code: %w", err)
	}

	// Convert ke Base64
	base64Str := base64.StdEncoding.EncodeToString(png)
	return base64Str, nil
}