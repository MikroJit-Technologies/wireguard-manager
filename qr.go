package main

import (
	"fmt"

	qrcode "github.com/skip2/go-qrcode"
)

func generateQR(content string) ([]byte, error) {
	png, err := qrcode.Encode(content, qrcode.Medium, 512)
	if err != nil {
		return nil, fmt.Errorf("qr encode: %w", err)
	}
	return png, nil
}
