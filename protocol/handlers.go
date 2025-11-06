package protocol

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"hash/crc32"
	"io"
	"time"

	"github.com/lewisgibson/go-steam/crypto"
	"github.com/lewisgibson/go-steam/language/steam"
	"github.com/lewisgibson/go-steam/steamid"
	"google.golang.org/protobuf/proto"
)

// handleMessage handles a message from the server.
func (h *Protocol) handleMessage(ctx context.Context, header *Header, body *bytes.Reader) error {
	var headerJSONBytes []byte
	if header != nil {
		var err error
		headerJSONBytes, err = json.Marshal(header)
		if err != nil {
			return fmt.Errorf("failed to marshal header: %w", err)
		}
	}

	// Emit debug event for raw EMsg
	select {
	case h.events <- &EventMessageReceived{EMsg: header.EMsg, Header: string(headerJSONBytes)}:
	default:
	}

	switch header.EMsg {
	case steam.EMsg_k_EMsgChannelEncryptRequest:
		return h.handleChannelEncryptRequest(ctx, header, body)

	case steam.EMsg_k_EMsgChannelEncryptResult:
		return h.handleChannelEncryptResult(ctx, header, body)

	case steam.EMsg_k_EMsgClientLogOnResponse:
		return h.handleClientLogOnResponse(ctx, header, body)

	case steam.EMsg_k_EMsgMulti:
		return h.handleMulti(ctx, header, body)

	case steam.EMsg_k_EMsgClientItemAnnouncements:
		return h.handleClientItemAnnouncements(ctx, header, body)

	case steam.EMsg_k_EMsgClientUserNotifications:
		return h.handleClientUserNotifications(ctx, header, body)
	}

	return nil
}

// handleChannelEncryptRequest handles the channel encrypt request.
func (h *Protocol) handleChannelEncryptRequest(ctx context.Context, header *Header, body *bytes.Reader) error {
	// Protocol
	var protocol uint32
	switch err := binary.Read(body, binary.LittleEndian, &protocol); {
	case err != nil:
		return fmt.Errorf("failed to read protocol: %w", err)

	case protocol != 1:
		return fmt.Errorf("unsupported protocol: %d", protocol)
	}

	// Universe
	var universe uint32
	switch err := binary.Read(body, binary.LittleEndian, &universe); {
	case err != nil:
		return fmt.Errorf("failed to read universe: %w", err)

	case universe != 1:
		return fmt.Errorf("unsupported universe: %d", universe)
	}

	// Nonce
	var nonce = make([]byte, 16)
	if _, err := body.Read(nonce); err != nil {
		return fmt.Errorf("failed to read nonce: %w", err)
	}

	// Session Key
	plain, encrypted, err := crypto.GenerateSessionKey()
	if err != nil {
		return fmt.Errorf("failed to generate session key: %w", err)
	}

	// Temporarily store the session key.
	h.sessionKey = plain

	// Write the protocol.
	var buf = bytes.NewBuffer(nil)
	if err := binary.Write(buf, binary.LittleEndian, protocol); err != nil {
		return fmt.Errorf("failed to write protocol: %w", err)
	}

	// Write the length of the session key.
	if err := binary.Write(buf, binary.LittleEndian, uint32(len(encrypted))); err != nil {
		return fmt.Errorf("failed to write session key length: %w", err)
	}

	// Write the session key.
	if err := binary.Write(buf, binary.LittleEndian, encrypted); err != nil {
		return fmt.Errorf("failed to write session key: %w", err)
	}

	// Write the CRC32 checksum for the session key.
	if err := binary.Write(buf, binary.LittleEndian, crc32.ChecksumIEEE(encrypted)); err != nil {
		return fmt.Errorf("failed to write session key checksum: %w", err)
	}

	// Write the nonce.
	if err := binary.Write(buf, binary.LittleEndian, nonce); err != nil {
		return fmt.Errorf("failed to write empty data: %w", err)
	}

	var hdr = &Header{
		EMsg:        steam.EMsg_k_EMsgChannelEncryptResponse,
		TargetJobID: header.TargetJobID,
		SourceJobID: header.SourceJobID,
	}
	if err := h.Write(ctx, hdr, buf); err != nil {
		return fmt.Errorf("failed to write response: %w", err)
	}

	return nil
}

// handleChannelEncryptResult handles the channel encrypt result.
func (h *Protocol) handleChannelEncryptResult(ctx context.Context, _ *Header, body *bytes.Reader) error {
	var eresult uint32
	switch err := binary.Read(body, binary.LittleEndian, &eresult); {
	case err != nil:
		return fmt.Errorf("failed to read eresult: %w", err)

	case eresult != 1:
		return fmt.Errorf("failed to encrypt channel: %d", eresult)
	}

	h.connection.SetSessionKey(h.sessionKey)
	h.sessionKey = nil
	h.events <- &EventConnected{}

	return nil
}

// handleClientLogOnResponse handles the client log on response.
func (h *Protocol) handleClientLogOnResponse(ctx context.Context, header *Header, body *bytes.Reader) error {
	b, err := io.ReadAll(body)
	if err != nil {
		return fmt.Errorf("failed to read body: %w", err)
	}

	var resp steam.CMsgClientLogonResponse
	switch err := proto.Unmarshal(b, &resp); {
	case err != nil:
		return fmt.Errorf("failed to unmarshal body: %w", err)

	case resp.GetEresult() != 1:
		return fmt.Errorf("failed to log in: %d", resp.GetEresult())
	}

	h.startHeartbeat(ctx, time.Duration(*resp.HeartbeatSeconds)*time.Second)

	h.steamID = steamid.FromUint64(header.Proto.GetSteamid())
	h.sessionID = header.Proto.GetClientSessionid()
	h.events <- &EventLoggedOn{
		SteamID:   h.steamID,
		SessionID: h.sessionID,
	}

	if err := h.Send(ctx, &Header{
		EMsg: steam.EMsg_k_EMsgClientRequestItemAnnouncements,
		Proto: &steam.CMsgProtoBufHeader{
			Steamid:         proto.Uint64(h.steamID.Uint64()),
			ClientSessionid: proto.Int32(h.sessionID),
		},
	}, &steam.CMsgClientRequestItemAnnouncements{}); err != nil {
		return fmt.Errorf("failed to send item announcements request: %w", err)
	}

	return nil
}

// handleMulti handles k_EMsgMulti messages that contain multiple sub-messages
func (h *Protocol) handleMulti(ctx context.Context, header *Header, body *bytes.Reader) error {
	// Read
	b, err := io.ReadAll(body)
	if err != nil {
		return fmt.Errorf("failed to read body: %w", err)
	}

	// Unmarshal
	var resp steam.CMsgMulti
	switch err := proto.Unmarshal(b, &resp); {
	case err != nil:
		return fmt.Errorf("failed to unmarshal body: %w", err)

	case len(resp.GetMessageBody()) == 0:
		return fmt.Errorf("CMsgMulti contains empty message body")
	}

	// Optional Decompress
	var data = resp.GetMessageBody()
	if resp.GetSizeUnzipped() != 0 {
		gzipReader, err := gzip.NewReader(bytes.NewReader(resp.GetMessageBody()))
		if err != nil {
			return fmt.Errorf("failed to create gzip reader: %w", err)
		}
		defer gzipReader.Close()

		data, err = io.ReadAll(gzipReader)
		switch {
		case err != nil:
			return fmt.Errorf("failed to create gzip reader: %w", err)

		case uint32(len(data)) != resp.GetSizeUnzipped():
			return fmt.Errorf("decompressed size mismatch: expected %d, got %d", resp.GetSizeUnzipped(), len(data))
		}
	}

	// Handle Each Sub-Message
	// Each sub-message follows this format:
	// 4 bytes: sub-message size (NOT including the size field itself)
	// N bytes: sub-message data (EMsg + header + body)
	var reader = bytes.NewReader(data)
	for reader.Len() != 0 {
		// Size
		var size uint32
		switch err := binary.Read(reader, binary.LittleEndian, &size); {
		case errors.Is(err, io.EOF):
		case err != nil:
			return fmt.Errorf("failed to read sub-message size: %w", err)

		case size < 4:
			return fmt.Errorf("invalid sub-message size: %d", size)

		case uint32(reader.Len()) < size:
			return fmt.Errorf("insufficient data for sub-message: need %d bytes, have %d", size, reader.Len())
		}

		// Data
		var data = make([]byte, size)
		if _, err := io.ReadFull(reader, data); err != nil {
			return fmt.Errorf("failed to read sub-message data: %w", err)
		}

		if err := h.readMessage(ctx, bytes.NewReader(data)); err != nil {
			return fmt.Errorf("failed to read sub-message: %w", err)
		}
	}

	return nil
}

func (h *Protocol) handleClientItemAnnouncements(ctx context.Context, header *Header, body *bytes.Reader) error {
	// Read
	b, err := io.ReadAll(body)
	if err != nil {
		return fmt.Errorf("failed to read body: %w", err)
	}

	// Unmarshal
	var resp steam.CMsgClientItemAnnouncements
	if err := proto.Unmarshal(b, &resp); err != nil {
		return fmt.Errorf("failed to unmarshal body: %w", err)
	}

	select {
	case h.events <- &EventItemAnnouncements{
		CountNewItems: resp.GetCountNewItems(),
	}:
	default:
	}

	return nil
}

// handleClientUserNotifications handles user notifications including trade offers.
func (h *Protocol) handleClientUserNotifications(ctx context.Context, header *Header, body *bytes.Reader) error {
	// Read
	b, err := io.ReadAll(body)
	if err != nil {
		return fmt.Errorf("failed to read body: %w", err)
	}

	// Unmarshal
	var resp steam.CMsgClientUserNotifications
	if err := proto.Unmarshal(b, &resp); err != nil {
		return fmt.Errorf("failed to unmarshal body: %w", err)
	}

	for _, notification := range resp.GetNotifications() {
		select {
		case h.events <- &EventUserNotification{
			NotificationType: notification.GetUserNotificationType(),
			Count:            notification.GetCount(),
		}:
		default:
		}
	}

	return nil
}
