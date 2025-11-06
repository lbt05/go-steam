package protocol

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"

	"github.com/lewisgibson/go-steam/language/steam"
	"google.golang.org/protobuf/proto"
)

func (h *Protocol) readMessage(ctx context.Context, reader *bytes.Reader) error {
	// The first four bytes are the EMsg.
	var rawEMsg uint32
	if err := binary.Read(reader, binary.LittleEndian, &rawEMsg); err != nil {
		return fmt.Errorf("failed to read message type: %w", err)
	}

	// Create the header.
	var header = &Header{
		EMsg: steam.EMsg(rawEMsg &^ PROTO_MASK),
	}

	// The header is then read based on the EMsg and proto flag.
	if err := h.readHeader(header, reader, rawEMsg&PROTO_MASK != 0); err != nil {
		return fmt.Errorf("failed to read message header: %w", err)
	}

	// The message is then handled based on the EMsg.
	if err := h.handleMessage(ctx, header, reader); err != nil {
		return fmt.Errorf("failed to handle message: %w", err)
	}

	return nil
}

// readHeader reads the header from the reader.
func (h *Protocol) readHeader(header *Header, reader *bytes.Reader, isProtobuf bool) error {
	// Protobuf Header
	if isProtobuf {
		return h.readProtobufHeader(header, reader)
	}

	// Standard Header
	if header.EMsg == steam.EMsg_k_EMsgChannelEncryptRequest || header.EMsg == steam.EMsg_k_EMsgChannelEncryptResult {
		return h.readStandardHeader(header, reader)
	}

	// Extended Header
	return h.readExtendedHeader(header, reader)
}

// readProtobufHeader reads a protobuf header from the reader.
func (h *Protocol) readProtobufHeader(header *Header, reader *bytes.Reader) error {
	// The first four bytes are the length of the header.
	var headerLength uint32
	if err := binary.Read(reader, binary.LittleEndian, &headerLength); err != nil {
		return fmt.Errorf("failed to read header length: %w", err)
	}

	// Read the header bytes.
	var headerBytes = make([]byte, headerLength)
	if n, err := reader.Read(headerBytes); err != nil {
		return fmt.Errorf("failed to read protobuf header data: %w", err)
	} else if n != int(headerLength) {
		return fmt.Errorf("incomplete protobuf header read: expected %d bytes, got %d", headerLength, n)
	}

	// Unmarshal the header.
	var headerProto steam.CMsgProtoBufHeader
	if err := proto.Unmarshal(headerBytes, &headerProto); err != nil {
		return fmt.Errorf("failed to unmarshal protobuf header: %w", err)
	}
	header.Proto = &headerProto

	// Set the header fields.
	header.TargetJobID = headerProto.GetJobidTarget()
	header.SourceJobID = headerProto.GetJobidSource()
	header.SteamID = headerProto.GetSteamid()
	header.SessionID = headerProto.GetClientSessionid()

	return nil
}

// readStandardHeader reads a standard header from the reader.
func (h *Protocol) readStandardHeader(header *Header, reader *bytes.Reader) error {
	// Ensure we have enough data for standard header (16 bytes: 8 + 8)
	if reader.Len() < 16 {
		return fmt.Errorf("insufficient data for standard header: need 16 bytes, have %d", reader.Len())
	}

	// TargetJobID (8 bytes)
	if err := binary.Read(reader, binary.LittleEndian, &header.TargetJobID); err != nil {
		return fmt.Errorf("failed to read TargetJobID: %w", err)
	}

	// SourceJobID (8 bytes)
	if err := binary.Read(reader, binary.LittleEndian, &header.SourceJobID); err != nil {
		return fmt.Errorf("failed to read SourceJobID: %w", err)
	}

	return nil
}

// readExtendedHeader reads an extended header from the reader.
func (h *Protocol) readExtendedHeader(header *Header, reader *bytes.Reader) error {
	var headerSize uint8
	if err := binary.Read(reader, binary.LittleEndian, &headerSize); err != nil {
		return fmt.Errorf("failed to read header size: %w", err)
	}

	var headerVersion uint16
	if err := binary.Read(reader, binary.LittleEndian, &headerVersion); err != nil {
		return fmt.Errorf("failed to read header version: %w", err)
	}

	switch headerSize {
	case 0:
		// No additional header data

	case 20:
		// Extended header: TargetJobID (8) + SourceJobID (8) + canary (1) + padding (1)
		if err := binary.Read(reader, binary.LittleEndian, &header.TargetJobID); err != nil {
			return fmt.Errorf("failed to read TargetJobID: %w", err)
		}

		if err := binary.Read(reader, binary.LittleEndian, &header.SourceJobID); err != nil {
			return fmt.Errorf("failed to read SourceJobID: %w", err)
		}

		var canary uint8
		if err := binary.Read(reader, binary.LittleEndian, &canary); err != nil {
			return fmt.Errorf("failed to read canary: %w", err)
		}

		// Skip padding byte
		var padding uint8
		if err := binary.Read(reader, binary.LittleEndian, &padding); err != nil {
			return fmt.Errorf("failed to read padding: %w", err)
		}

	case 36:
		// Full extended header: TargetJobID (8) + SourceJobID (8) + canary (1) + SteamID (8) + SessionID (4) + padding (7)
		if err := binary.Read(reader, binary.LittleEndian, &header.TargetJobID); err != nil {
			return fmt.Errorf("failed to read TargetJobID: %w", err)
		}

		if err := binary.Read(reader, binary.LittleEndian, &header.SourceJobID); err != nil {
			return fmt.Errorf("failed to read SourceJobID: %w", err)
		}

		var canary uint8
		if err := binary.Read(reader, binary.LittleEndian, &canary); err != nil {
			return fmt.Errorf("failed to read canary: %w", err)
		}

		if err := binary.Read(reader, binary.LittleEndian, &header.SteamID); err != nil {
			return fmt.Errorf("failed to read SteamID: %w", err)
		}

		if err := binary.Read(reader, binary.LittleEndian, &header.SessionID); err != nil {
			return fmt.Errorf("failed to read SessionID: %w", err)
		}

		// Skip remaining padding bytes (3 bytes to align to 4-byte boundary)
		var padding [3]uint8
		if err := binary.Read(reader, binary.LittleEndian, &padding); err != nil {
			return fmt.Errorf("failed to read padding: %w", err)
		}

	default:
		return fmt.Errorf("unsupported header size: %d", headerSize)
	}

	return nil
}
