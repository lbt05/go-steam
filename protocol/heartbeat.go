package protocol

import (
	"context"
	"time"

	"github.com/lewisgibson/go-steam/language/steam"
	"google.golang.org/protobuf/proto"
)

// startHeartbeat starts the heartbeat loop.
func (h *Protocol) startHeartbeat(ctx context.Context, interval time.Duration) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	// If a heartbeat is already running, stop it.
	if h.stopHeartbeat != nil {
		close(h.stopHeartbeat)
	}

	h.stopHeartbeat = make(chan struct{}, 1)

	go func() {
		var t = time.NewTicker(interval)
		defer t.Stop()
		for {
			select {
			case <-ctx.Done():
				return

			case <-h.stopHeartbeat:
				return

			case <-t.C:
				if err := h.Send(ctx, &Header{EMsg: steam.EMsg_k_EMsgClientHeartBeat, Proto: &steam.CMsgProtoBufHeader{
					Steamid:         proto.Uint64(h.steamID.Uint64()),
					ClientSessionid: proto.Int32(h.sessionID),
					JobidSource:     proto.Uint64(JOBID_NONE),
					JobidTarget:     proto.Uint64(JOBID_NONE),
				}}, &steam.CMsgClientHeartBeat{}); err != nil {
					h.events <- &EventError{Err: err}
				}
			}
		}
	}()
}
