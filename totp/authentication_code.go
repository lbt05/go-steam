package totp

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"fmt"
	"time"
)

// CreateAuthenticationCode creates a new authentication code using the shared secret and the current time.
// The shared secret must be base64 encoded.
func CreateAuthenticationCode(sharedSecret string, time time.Time) (string, error) {
	// Decode the shared secret from base64.
	sharedSecretBytes, err := base64.StdEncoding.DecodeString(sharedSecret)
	if err != nil {
		return "", fmt.Errorf("failed to decode shared secret: %w", err)
	}

	// Write four zeros to a new buffer.
	var buffer = bytes.NewBuffer(nil)
	if err := binary.Write(buffer, binary.BigEndian, uint32(0)); err != nil {
		return "", fmt.Errorf("failed to write zeros to buffer: %w", err)
	}

	// Write the time-step to the buffer. Each time-step lasts 30 seconds.
	if err := binary.Write(buffer, binary.BigEndian, uint32(time.Unix()/30)); err != nil {
		return "", fmt.Errorf("failed to write time step to buffer: %w", err)
	}

	// Create a new HMAC-SHA1 hash with the shared secret.
	var hash = hmac.New(sha1.New, sharedSecretBytes)
	if _, err := hash.Write(buffer.Bytes()); err != nil {
		return "", fmt.Errorf("failed to write to SHA-1 hash: %w", err)
	}

	// Calculate the checksum.
	checksum := hash.Sum(nil)

	// Determine the dynamic offset for truncation as defined in the HOTP/TOTP specifications.
	// We use the lower 4 bits of the last byte of the HMAC digest (masked with 0x0F) to obtain a value between 0 and 15.
	// This dynamic offset selects the starting position for a 4-byte segment in the digest, adding unpredictability
	// to the extraction process and ensuring a more evenly distributed use of the hash's entropy.
	var dynamicOffset = int(checksum[len(checksum)-1] & 0x0F)
	if dynamicOffset+4 > len(checksum) {
		return "", errors.New("dynamic bytes exceed checksum length")
	}

	// Convert the selected 4-byte segment (starting at dynamicOffset) from big-endian order
	// into a 32-bit unsigned integer. Then mask out the most significant bit to produce a 31-bit
	// positive integer, as required by the HOTP/TOTP standard. The mask ensures that the result is always positive.
	binaryCode := binary.BigEndian.Uint32(checksum[dynamicOffset:dynamicOffset+4]) & 0x7fffffff

	// charSet contains the characters that are allowed in the authentication code.
	// The characters are chosen to avoid confusion between similar looking characters.
	const charSet = "23456789BCDFGHJKMNPQRTVWXY"

	// Generate a 5-character authentication code by converting the binary code into a custom base defined by charSet.
	// For each of the 5 characters, we take the remainder of binaryCode divided by the length of charSet to determine
	// the index of the corresponding character. Then, we divide binaryCode by the length of charSet to shift to the
	// next digit. This process extracts one digit at a time from binaryCode, resulting in a human-friendly code.
	var code = make([]byte, 5)
	for i := 0; i < 5; i++ {
		code[i] = charSet[binaryCode%uint32(len(charSet))]
		binaryCode /= uint32(len(charSet))
	}

	return string(code), nil
}
