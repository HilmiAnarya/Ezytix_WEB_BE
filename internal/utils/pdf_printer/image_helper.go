package pdfprinter

import (
	"encoding/base64"
	"github.com/skip2/go-qrcode"
	"fmt"
	"os"
)

func ImageToBase64(filePath string) (string, error) {
	bytes, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("gagal membaca gambar %s: %v", filePath, err)
	}
	encoded := base64.StdEncoding.EncodeToString(bytes)
	
	return encoded, nil
}

func GenerateQRCodeBase64(text string) (string, error) {
	var png []byte
	png, err := qrcode.Encode(text, qrcode.Medium, 256)
	if err != nil {
		return "", fmt.Errorf("failed to generate qr code: %w", err)
	}

	base64Str := base64.StdEncoding.EncodeToString(png)
	return base64Str, nil
}