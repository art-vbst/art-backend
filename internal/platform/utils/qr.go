package utils

import (
	"github.com/skip2/go-qrcode"
)

func GenerateQRCode(uri string) ([]byte, error) {
	qr, err := qrcode.New(uri, qrcode.Medium)
	if err != nil {
		return nil, err
	}
	return qr.PNG(256)
}
