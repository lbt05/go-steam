package crypto

import (
	"bytes"
	"fmt"
)

// Pkcs7Pad applies PKCS7 padding to data so that its length becomes a multiple of blockSize.
func Pkcs7Pad(data []byte, blockSize int) []byte {
	// The data is padded with the remainder of the block size, repeated for the remainder.
	var pad = blockSize - (len(data) % blockSize)
	var padding = bytes.Repeat([]byte{byte(pad)}, pad)
	return append(data, padding...)
}

// Pkcs7Unpad removes PKCS7 padding from decrypted data.
func Pkcs7Unpad(data []byte, blockSize int) ([]byte, error) {
	// The data should be a multiple of the block size.
	if len(data) == 0 || len(data)%blockSize != 0 {
		return nil, fmt.Errorf("invalid padded data")
	}

	// The padding length is the last byte of the data.
	var padLen = int(data[len(data)-1])
	if padLen == 0 || padLen > blockSize {
		return nil, fmt.Errorf("invalid padding")
	}

	// Verify the padding.
	for i := len(data) - padLen; i < len(data); i++ {
		if data[i] != byte(padLen) {
			return nil, fmt.Errorf("invalid padding")
		}
	}

	// Return the unpadded data.
	return data[:len(data)-padLen], nil
}
