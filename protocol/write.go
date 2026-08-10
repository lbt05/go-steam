package protocol

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"

	"github.com/lbt05/go-steam/language/steam"
	"google.golang.org/protobuf/proto"
)

// Send sends a message to the server.
func (h *Protocol) Send(ctx context.Context, header *Header, body proto.Message) error {
	{
		var headerJSONBytes []byte
		if header != nil {
			var err error
			headerJSONBytes, err = json.Marshal(header)
			if err != nil {
				return fmt.Errorf("failed to marshal header: %w", err)
			}
		}

		var bodyJSONBytes []byte
		if body != nil {
			var err error
			bodyJSONBytes, err = json.Marshal(body)
			if err != nil {
				return fmt.Errorf("failed to marshal body: %w", err)
			}
		}

		select {
		case h.events <- &EventMessageSent{EMsg: header.EMsg, Header: string(headerJSONBytes), Body: string(bodyJSONBytes)}:
		default:
		}
	}

	b, err := proto.Marshal(body)
	if err != nil {
		return fmt.Errorf("failed to marshal protobuf: %w", err)
	}

	if err := h.Write(ctx, header, bytes.NewBuffer(b)); err != nil {
		return fmt.Errorf("failed to write response: %w", err)
	}

	return nil
}

func (h *Protocol) Write(ctx context.Context, header *Header, body *bytes.Buffer) error {
	var headerBuf = bytes.NewBuffer(nil)
	switch {
	case header.EMsg == steam.EMsg_k_EMsgChannelEncryptResponse:
		if err := binary.Write(headerBuf, binary.LittleEndian, uint32(header.EMsg)); err != nil {
			return fmt.Errorf("failed to write EMsg: %w", err)
		}

		if err := binary.Write(headerBuf, binary.LittleEndian, uint64(header.TargetJobID)); err != nil {
			return fmt.Errorf("failed to write TargetJobID: %w", err)
		}

		if err := binary.Write(headerBuf, binary.LittleEndian, uint64(header.SourceJobID)); err != nil {
			return fmt.Errorf("failed to write SourceJobID: %w", err)
		}

	case header.Proto != nil:
		switch header.EMsg {
		case steam.EMsg_k_EMsgClientHello, steam.EMsg_k_EMsgServiceMethodCallFromClientNonAuthed:
			break

		default:
			if header.Proto.ClientSessionid == nil {
				header.Proto.ClientSessionid = proto.Int32(0)
			}
			if header.Proto.Steamid == nil {
				header.Proto.Steamid = proto.Uint64(0)
			}
		}

		// Marshal the protobuf.
		b, err := proto.Marshal(header.Proto)
		if err != nil {
			return fmt.Errorf("failed to marshal protobuf: %w", err)
		}

		// Write the emsg to the header buffer
		if err := binary.Write(headerBuf, binary.LittleEndian, uint32(header.EMsg)|PROTO_MASK); err != nil {
			return fmt.Errorf("failed to write emsg: %w", err)
		}

		// Write the length of the protobuf to the header buffer
		if err := binary.Write(headerBuf, binary.LittleEndian, uint32(len(b))); err != nil {
			return fmt.Errorf("failed to write header length: %w", err)
		}

		// Write the protobuf to the header buffer
		if _, err := headerBuf.Write(b); err != nil {
			return fmt.Errorf("failed to write protobuf: %w", err)
		}

	default:
		// Write the EMsg to the header buffer
		if err := binary.Write(headerBuf, binary.LittleEndian, uint32(header.EMsg)); err != nil {
			return fmt.Errorf("failed to write EMsg: %w", err)
		}

		// Write 36 to the header buffer. This represents the
		if err := binary.Write(headerBuf, binary.LittleEndian, uint32(36)); err != nil {
			return fmt.Errorf("failed to write 36: %w", err)
		}

		// Write 2 to the header buffer. This represents the
		if err := binary.Write(headerBuf, binary.LittleEndian, uint16(2)); err != nil {
			return fmt.Errorf("failed to write 2: %w", err)
		}

		// Write the TargetJobID to the header buffer
		if err := binary.Write(headerBuf, binary.LittleEndian, uint64(header.TargetJobID)); err != nil {
			return fmt.Errorf("failed to write TargetJobID: %w", err)
		}

		// Write the SourceJobID to the header buffer
		if err := binary.Write(headerBuf, binary.LittleEndian, uint64(header.SourceJobID)); err != nil {
			return fmt.Errorf("failed to write SourceJobID: %w", err)
		}

		// Write 239 to the header buffer. This represents the
		if err := binary.Write(headerBuf, binary.LittleEndian, uint8(239)); err != nil {
			return fmt.Errorf("failed to write 239: %w", err)
		}
		// Write the SteamID to the header buffer
		if err := binary.Write(headerBuf, binary.LittleEndian, uint64(header.SteamID)); err != nil {
			return fmt.Errorf("failed to write SteamID: %w", err)
		}

		// Write the SessionID to the header buffer
		if err := binary.Write(headerBuf, binary.LittleEndian, uint32(header.SessionID)); err != nil {
			return fmt.Errorf("failed to write SessionID: %w", err)
		}
	}

	if err := h.connection.Send(ctx, append(headerBuf.Bytes(), body.Bytes()...)); err != nil {
		return fmt.Errorf("failed to send: %w", err)
	}

	return nil
}
