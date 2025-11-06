package steamid

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

// SteamID represents a 64-bit Steam identifier, combining:
//   - AccountID (32 bits)
//   - AccountInstance (20 bits)
//   - AccountType (4 bits)
//   - AccountUniverse (4 bits)
type SteamID uint64

// AccountID is the 32-bit account identifier.
type AccountID = uint32

// AccountInstance provides more information about the account type.
type AccountInstance = uint32

const (
	AccountInstanceClanMask     AccountInstance = 1 << 19 // 0x80000 in hex
	AccountInstanceLobbyMask    AccountInstance = 1 << 18 // 0x40000 in hex
	AccountInstanceMMSLobbyMask AccountInstance = 1 << 17 // 0x20000 in hex
)

// AccountType represents the type of account.
type AccountType = uint8

const (
	AccountTypeInvalid        AccountType = 0
	AccountTypeIndividual     AccountType = 1
	AccountTypeMultiseat      AccountType = 2
	AccountTypeGameServer     AccountType = 3
	AccountTypeAnonGameServer AccountType = 4
	AccountTypePending        AccountType = 5
	AccountTypeContentServer  AccountType = 6
	AccountTypeClan           AccountType = 7
	AccountTypeChat           AccountType = 8
	AccountTypeP2PSuperSeeder AccountType = 9
	AccountTypeAnonUser       AccountType = 10
)

// AccountUniverse represents the universe of the account.
type AccountUniverse = uint8

const (
	AccountUniverseInvalid  AccountUniverse = 0
	AccountUniversePublic   AccountUniverse = 1
	AccountUniverseBeta     AccountUniverse = 2
	AccountUniverseInternal AccountUniverse = 3
	AccountUniverseDev      AccountUniverse = 4
	AccountUniverseMax      AccountUniverse = 5
)

// NewSteamID creates a new SteamID with the given values.
func NewSteamID(accountID AccountID, accountInstance AccountInstance, accountType AccountType, accountUniverse AccountUniverse) SteamID {
	return SteamID(0).
		WithAccountID(accountID).
		WithAccountInstance(accountInstance).
		WithAccountType(accountType).
		WithAccountUniverse(accountUniverse)
}

// FromUint64 creates a new SteamID from a 64-bit unsigned integer.
func FromUint64(steamID uint64) SteamID {
	return SteamID(steamID)
}

// FromString parses a SteamID from a string.
// The string can be in Steam2, Steam3, or raw 64-bit format.
// If the string is not a valid SteamID, an error will be returned.
func FromString(input string) (SteamID, error) {
	switch input = strings.TrimSpace(input); {
	case Steam2Pattern.MatchString(input):
		return FromSteam2(input)

	case Steam3Pattern.MatchString(input):
		return FromSteam3(input)

	case Steam64Pattern.MatchString(input):
		id64, err := strconv.ParseUint(input, 10, 64)
		if err != nil {
			return 0, fmt.Errorf("invalid SteamID format: %s", input)
		}
		return SteamID(id64), nil

	default:
		return 0, fmt.Errorf("invalid SteamID format: %s", input)
	}
}

// MustFromString parses a SteamID from a string and panics if the string is not a valid SteamID.
func MustFromString(input string) SteamID {
	sid, err := FromString(input)
	if err != nil {
		panic(err)
	}
	return sid
}

// Uint64 returns the SteamID as a 64-bit unsigned integer.
func (sid SteamID) Uint64() uint64 {
	return uint64(sid)
}

// String implements the fmt.Stringer interface.
// It returns the SteamID in Steam2 format if it is an individual account.
// Otherwise, it returns the 64-bit SteamID.
func (sid SteamID) String() string {
	if sid.AccountType() == AccountTypeIndividual {
		return sid.ToSteam2()
	}
	return sid.ToSteam64()
}

// IsValid returns true if the SteamID is valid.
func (sid SteamID) IsValid() bool {
	return sid.AccountUniverse() != AccountUniverseInvalid &&
		sid.AccountType() != AccountTypeInvalid &&
		sid.AccountID() != 0
}

// ConvertClanToChat converts a clan account to a chat account.
// If the account is not a clan account, it will return the same SteamID.
func (sid SteamID) ConvertClanToChat() SteamID {
	// The account type is not a clan, so we don't need to do anything.
	if sid.AccountType() != AccountTypeClan {
		return sid
	}

	// Set the account type to a chat and set the instance to a clan.
	return sid.WithAccountType(AccountTypeChat).WithAccountInstance(AccountInstanceClanMask)
}

// ConvertChatToClan converts a chat account to a clan account.
// If the account is not a chat account, it will return the same SteamID.
func (sid SteamID) ConvertChatToClan() SteamID {
	// The account type is not a chat, so we don't need to do anything.
	if sid.AccountType() != AccountTypeChat {
		return sid
	}

	// If the account instance is not a clan, we don't need to do anything.
	if sid.AccountInstance() != AccountInstanceClanMask {
		return sid
	}

	// Set the account type to a clan and clear the instance.
	return sid.WithAccountType(AccountTypeClan).WithAccountInstance(0)
}

// EncodeValues implements the query.Encoder interface for SteamID.
func (sid SteamID) EncodeValues(key string, v *url.Values) error {
	v.Add(key, sid.ToSteam64())
	return nil
}

// MarshalJSON implements the json.Marshaler interface.
func (sid SteamID) MarshalJSON() ([]byte, error) {
	return json.Marshal(sid.String())
}

// UnmarshalJSON implements the json.Unmarshaler interface.
// It tries as an integer first, then as a string.
func (sid *SteamID) UnmarshalJSON(b []byte) error {
	var i uint64
	if err := json.Unmarshal(b, &i); err == nil {
		// It's a SteamID64 if it's greater than the SteamID64 base value.
		if i > 76561197960265728 {
			*sid = FromUint64(i)
		} else {
			*sid = NewSteamID(uint32(i), 1, AccountTypeIndividual, AccountUniversePublic)
		}
		return nil
	}

	var s string
	if err := json.Unmarshal(b, &s); err == nil {
		*sid, err = FromString(s)
		return err
	}

	return errors.New("invalid SteamID")
}
