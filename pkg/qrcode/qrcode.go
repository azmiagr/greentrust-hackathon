package qrcode

import skipqrcode "github.com/skip2/go-qrcode"

func GeneratePNG(content string, size int) ([]byte, error) {
	if size <= 0 {
		size = 512
	}

	return skipqrcode.Encode(content, skipqrcode.Medium, size)
}
