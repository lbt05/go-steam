package steamid

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
)

// Steam64Pattern matches a 64-bit SteamID.
var Steam64Pattern = regexp.MustCompile(`^\d{17,18}$`)

// ErrInvalidSteam64 is returned when the input does not match the Steam64 pattern.
var ErrInvalidSteam64 = fmt.Errorf("invalid format for Steam64")

// FromSteam64 parses strings like "76561198065346589".
func FromSteam64(input string) (SteamID, error) {
	if !Steam64Pattern.MatchString(input) {
		// return ErrInvalidSteam64
		return 0, errors.Join(ErrInvalidSteam64, fmt.Errorf("value '%s' does not match pattern '%s'", input, Steam64Pattern))
	}

	id, err := strconv.ParseUint(input, 10, 64)
	if err != nil {
		return 0, errors.Join(ErrInvalidSteam64, err)
	}

	return SteamID(id), nil
}

// MustFromSteam64 parses a SteamID from a string and panics if the string is not a valid SteamID.
func MustFromSteam64(input string) SteamID {
	sid, err := FromSteam64(input)
	if err != nil {
		panic(err)
	}
	return sid
}

// ToSteam64 returns the SteamID as a 64-bit string.
func (sid SteamID) ToSteam64() string {
	return strconv.FormatUint(sid.Uint64(), 10)
}
