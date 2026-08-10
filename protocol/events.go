package protocol

import (
	"github.com/lbt05/go-steam/api/services/iauthenticationservice"
	"github.com/lbt05/go-steam/identity"
	"github.com/lbt05/go-steam/language/steam"
	"github.com/lbt05/go-steam/steamid"
)

// Events returns the events channel.
func (h *Protocol) Events() chan<- Event {
	return h.events
}

type Event interface{ isEvent() }

type EventMessageSent struct {
	EMsg   steam.EMsg
	Header string
	Body   string
}

func (e *EventMessageSent) isEvent() {}

type EventMessageReceived struct {
	EMsg   steam.EMsg
	Header string
}

func (e *EventMessageReceived) isEvent() {}

type EventConnected struct{}

func (e *EventConnected) isEvent() {}

type EventLoggedOn struct {
	SteamID   steamid.SteamID
	SessionID int32
}

func (e *EventLoggedOn) isEvent() {}

type EventError struct {
	Err error
}

func (e *EventError) isEvent() {}

type EventItemAnnouncements struct {
	CountNewItems uint32
}

func (e *EventItemAnnouncements) isEvent() {}

type EventUserNotification struct {
	NotificationType uint32
	Count            uint32
}

func (e *EventUserNotification) isEvent() {}

// EventSteamGuardChallenge is emitted after Client.BeginAuthSession succeeds.
// It carries the AuthSession returned by the server plus the Steam Guard
// confirmation types the user can choose between when calling
// Client.SubmitSteamGuardCode.
type EventSteamGuardChallenge struct {
	Session              *identity.AuthSession
	AllowedConfirmations []iauthenticationservice.AllowedConfirmation
}

func (e *EventSteamGuardChallenge) isEvent() {}
