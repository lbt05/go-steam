package client

import (
	"github.com/lewisgibson/go-steam/protocol"
)

func (c *Client) Events() <-chan protocol.Event {
	return c.events
}
