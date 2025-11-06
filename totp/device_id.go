package totp

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
)

// CreateDeviceID returns a standardized device ID based on a SteamID.
func CreateDeviceID(steamID64 string) string {
	var sum = sha1.Sum([]byte(steamID64))
	var hash = hex.EncodeToString(sum[:])
	return fmt.Sprintf("android:%s-%s-%s-%s-%s", hash[:8], hash[8:12], hash[12:16], hash[16:20], hash[20:32])
}
