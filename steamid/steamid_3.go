package steamid

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// Steam3Pattern is a regex pattern for parsing Steam3 IDs.
var Steam3Pattern = regexp.MustCompile(`^\[([iIUgGcCAa])?:?(\d+):(\d+)(:?\d*)?\]$`)

// ErrInvalidSteam3 is returned when the input does not match the Steam3 pattern.
var ErrInvalidSteam3 = fmt.Errorf("invalid format for Steam3")

// FromSteam3 parses strings like "[U:1:1234]", "[G:1:9999]", etc.
func FromSteam3(input string) (SteamID, error) {
	var matches = Steam3Pattern.FindStringSubmatch(input)
	if len(matches) < 4 {
		return 0, errors.Join(ErrInvalidSteam3, fmt.Errorf("value '%s' does not match pattern '%s'", input, Steam3Pattern))
	}

	accountType := MapSteamID3CharacterToAccountType(matches[1])

	universe, err := strconv.ParseUint(matches[2], 10, 8)
	if err != nil {
		return 0, errors.Join(ErrInvalidSteam3, fmt.Errorf("invalid universe value for %s: %w", input, err))
	}

	accountID, err := strconv.ParseUint(matches[3], 10, 32)
	if err != nil {
		return 0, errors.Join(ErrInvalidSteam3, fmt.Errorf("invalid account ID value for %s: %w", input, err))
	}

	return NewSteamID(uint32(accountID), defaultInstanceForAccountType(accountType), uint8(universe), accountType), nil
}

// MustFromSteam3 parses a SteamID from a string and panics if the string is not a valid SteamID.
func MustFromSteam3(input string) SteamID {
	sid, err := FromSteam3(input)
	if err != nil {
		panic(err)
	}
	return sid
}

// ToSteam3 returns the bracketed format, e.g. [U:1:12345].
func (sid SteamID) ToSteam3() string {
	return fmt.Sprintf(
		"[%s:%d:%d]",
		MapAccountTypeToSteamID3Character(sid.AccountType()),
		sid.AccountUniverse(),
		sid.AccountID(),
	)
}

// MapAccountTypeToSteamID3Character maps an AccountType to a single-character account type.
func MapAccountTypeToSteamID3Character(accountType AccountType) string {
	switch accountType {
	case AccountTypeIndividual:
		return "U"
	case AccountTypeClan:
		return "G"
	case AccountTypeAnonGameServer:
		return "A"
	case AccountTypeContentServer:
		return "C"
	case AccountTypeChat:
		return "T"
	case AccountTypeInvalid:
		return "I"
	default:
		return "U"
	}
}

// MapSteamID3CharacterToAccountType maps a single-character account type to an AccountType.
func MapSteamID3CharacterToAccountType(char string) AccountType {
	switch strings.ToUpper(char) {
	case "U":
		return AccountTypeIndividual
	case "G":
		return AccountTypeClan
	case "A":
		return AccountTypeAnonGameServer
	case "C":
		return AccountTypeContentServer
	case "T":
		return AccountTypeChat
	case "I":
		return AccountTypeInvalid
	default:
		return AccountTypeIndividual
	}
}

// defaultInstanceForAccountType returns the default instance for a given account type.
func defaultInstanceForAccountType(accountType AccountType) AccountInstance {
	switch accountType {
	case AccountTypeClan:
		return 0

	default:
		return 1
	}
}
