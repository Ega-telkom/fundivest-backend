// internal/pkg/qrcode/qrcode.go
package qrcode

import (
    "encoding/base64"
    
    "github.com/skip2/go-qrcode"
)

// Generate QR code dan return base64 string untuk embed di HTML
func Generate(content string) (string, error) {
    qr, err := qrcode.New(content, qrcode.Medium)
    if err != nil {
        return "", err
    }
    
    // Generate PNG dengan size 256x256
    pngBytes, err := qr.PNG(256)
    if err != nil {
        return "", err
    }
    
    // Encode ke base64 untuk embed di HTML
    encoded := base64.StdEncoding.EncodeToString(pngBytes)
    return "data:image/png;base64," + encoded, nil
}

// GenerateBytes return raw PNG bytes
func GenerateBytes(content string, size int) ([]byte, error) {
    return qrcode.Encode(content, qrcode.Medium, size)
}