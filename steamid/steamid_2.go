package steamid

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
)

// Steam2Pattern matches the classic STEAM_X:Y:Z format.
var Steam2Pattern = regexp.MustCompile(`^STEAM_([0-5]):([0-1]):(\d+)$`)

// ErrInvalidSteam2 is returned when the input string is not a valid Steam2 format.
var ErrInvalidSteam2 = fmt.Errorf("invalid Steam2 format")

// FromSteam2 parses strings like "STEAM_0:1:1234".
func FromSteam2(input string) (SteamID, error) {
	var matches = Steam2Pattern.FindStringSubmatch(input)
	if len(matches) != 4 {
		return 0, errors.Join(ErrInvalidSteam2, fmt.Errorf("invalid format for %s", input))
	}

	universe, err := strconv.ParseInt(matches[1], 10, 8)
	switch {
	case err != nil:
		return 0, errors.Join(ErrInvalidSteam2, fmt.Errorf("invalid universe value for %s: %w", input, err))

	case universe == int64(AccountUniverseInvalid):
		// Historically, STEAM_0 indicated UniversePublic.
		universe = int64(AccountUniversePublic)
	}

	authenticationServer, err := strconv.ParseUint(matches[2], 10, 32)
	if err != nil {
		return 0, errors.Join(ErrInvalidSteam2, fmt.Errorf("invalid authentication server value for %s: %w", input, err))
	}

	accountNumber, err := strconv.ParseUint(matches[3], 10, 32)
	if err != nil {
		return 0, errors.Join(ErrInvalidSteam2, fmt.Errorf("invalid account number value for %s: %w", input, err))
	}

	return NewSteamID((uint32(accountNumber)<<1)|uint32(authenticationServer), 1, uint8(universe), AccountTypeIndividual), nil
}

// MustFromSteam2 parses a SteamID from a string and panics if the string is not a valid SteamID.
func MustFromSteam2(input string) SteamID {
	sid, err := FromSteam2(input)
	if err != nil {
		panic(err)
	}
	return sid
}

// ToSteam2 returns the classic STEAM_X:Y:Z format if possible.
// If Universe is 0 or invalid, we treat it as UniversePublic (1).
func (sid SteamID) ToSteam2() string {
	var universe = sid.AccountUniverse()
	if universe == AccountUniverseInvalid {
		universe = AccountUniversePublic
	}
	return fmt.Sprintf("STEAM_%d:%d:%d", universe, sid.AccountID()&1, sid.AccountID()>>1)
}
