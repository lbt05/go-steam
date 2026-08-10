package client

import (
	"github.com/lbt05/go-steam/protocol"
)

func (c *Client) Events() <-chan protocol.Event {
	return c.events
}
