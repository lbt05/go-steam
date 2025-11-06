package protocol

import (
	"github.com/lewisgibson/go-steam/language/steam"
	"github.com/lewisgibson/go-steam/steamid"
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
