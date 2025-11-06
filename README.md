# go-steam

[![Build Workflow](https://github.com/lewisgibson/go-steam/actions/workflows/build.yaml/badge.svg)](https://github.com/lewisgibson/go-steam/actions/workflows/build.yaml)
[![Pkg Go Dev](https://pkg.go.dev/badge/github.com/lewisgibson/go-steam)](https://pkg.go.dev/github.com/lewisgibson/go-steam)

> ⚠️ **Warning**: This library is under active development and is expected to have breaking changes. Use at your own risk.

A Go library for interacting with the Steam network protocol. Connect to Steam, authenticate, and communicate with Steam services using the native protocol.

## Features

-   ✅ **Steam Protocol Implementation**: Full support for Steam's binary protocol
-   ✅ **Authentication**: Login with username/password and Steam Guard (TOTP)
-   ✅ **TCP Connection Management**: Automatic server discovery and connection
-   ✅ **Event-Driven Architecture**: React to protocol events in real-time
-   ✅ **Protocol Encryption**: Secure communication with session key exchange
-   ✅ **Steam ID Support**: Parse and create Steam IDs in multiple formats (SteamID64, SteamID2, SteamID3)
-   ✅ **Steam API Services**: Built-in support for authentication, trading, and directory services
-   ✅ **Protobuf Messages**: Auto-generated from official Steam protobufs
-   ✅ **Auto-Reconnect**: Optional automatic reconnection on disconnect
-   ✅ **Context Support**: Proper context handling for cancellation and timeouts
-   ✅ **Thread-Safe**: Concurrent-safe operations throughout

## Resources

-   [Discussions](https://github.com/lewisgibson/go-steam/discussions)
-   [Reference](https://pkg.go.dev/github.com/lewisgibson/go-steam)
-   [Examples](examples/)

## Installation

```sh
go get github.com/lewisgibson/go-steam
```

## Quickstart

### Basic Usage

```go
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/lewisgibson/go-steam/client"
	"github.com/lewisgibson/go-steam/protocol"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	// Create a new Steam client with credentials
	c, err := client.NewClient(&client.Credentials{
		Username:     "your_username",
		Password:     "your_password",
		SharedSecret: "your_shared_secret", // Base64-encoded Steam Guard secret
	})
	if err != nil {
		log.Fatal(err)
	}

	// Connect to Steam in a goroutine
	go func() {
		if err := c.Connect(ctx); err != nil {
			log.Printf("Connection error: %v", err)
		}
	}()

	// Handle events
	for {
		select {
		case <-ctx.Done():
			return

		case event := <-c.Events():
			switch event := event.(type) {
			case *protocol.EventConnected:
				fmt.Println("Connected to Steam!")

				// Log on after connecting
				if err := c.Logon(ctx); err != nil {
					log.Printf("Logon error: %v", err)
				}

			case *protocol.EventLoggedOn:
				fmt.Printf("Logged on! SteamID: %d\n", event.SteamID)

			case *protocol.EventMessageReceived:
				fmt.Printf("Received: %s\n", event.EMsg.String())

			case *protocol.EventError:
				log.Printf("Error: %v", event.Err)
			}
		}
	}
}
```

### Steam ID Manipulation

```go
package main

import (
	"fmt"
	"log"

	"github.com/lewisgibson/go-steam/steamid"
)

func main() {
	// Parse from SteamID64
	id, err := steamid.ParseSteamID("76561198006409530")
	if err != nil {
		log.Fatal(err)
	}

	// Convert to different formats
	fmt.Printf("SteamID64: %s\n", id.SteamID64())     // 76561198006409530
	fmt.Printf("SteamID2:  %s\n", id.SteamID2())      // STEAM_0:0:23071901
	fmt.Printf("SteamID3:  %s\n", id.SteamID3())      // [U:1:46143802]

	// Get components
	fmt.Printf("Account ID: %d\n", id.AccountID())
	fmt.Printf("Universe:   %d\n", id.Universe())
	fmt.Printf("Type:       %d\n", id.Type())
}
```

### TOTP Authentication Code

```go
package main

import (
	"fmt"
	"log"
	"time"

	"github.com/lewisgibson/go-steam/totp"
)

func main() {
	sharedSecret := "your_base64_shared_secret"

	// Generate authentication code
	code, err := totp.CreateAuthenticationCode(sharedSecret, time.Now())
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Steam Guard Code: %s\n", code)
}
```

### Client with Options

```go
package main

import (
	"net"

	"github.com/lewisgibson/go-steam/client"
	"github.com/lewisgibson/go-steam/connection"
)

func main() {
	// Create custom dialer
	dialer := &net.Dialer{
		Timeout: 10 * time.Second,
	}

	// Create client with options
	c, err := client.NewClient(
		&client.Credentials{
			Username:     "your_username",
			Password:     "your_password",
			SharedSecret: "your_shared_secret",
		},
		client.WithDialer(dialer),
		client.WithReconnect(true), // Enable automatic reconnection
	)
	if err != nil {
		log.Fatal(err)
	}

	// ... use client
}
```

## API Reference

### Client Package

-   `NewClient(identity IdentityProvider, opts ...ClientOption) (*Client, error)` - Create a new Steam client
-   `Connect(ctx context.Context) error` - Connect to Steam servers
-   `Disconnect() error` - Disconnect from Steam
-   `Logon(ctx context.Context) error` - Log on to Steam
-   `Events() <-chan protocol.Event` - Get event channel

### Client Options

-   `WithAPI(api API)` - Set custom API client
-   `WithDialer(dialer connection.Dialer)` - Set custom network dialer
-   `WithReconnect(reconnect bool)` - Enable/disable automatic reconnection

### Protocol Events

-   `EventConnected` - Connected to Steam server
-   `EventLoggedOn` - Successfully logged on
-   `EventMessageSent` - Protocol message sent
-   `EventMessageReceived` - Protocol message received
-   `EventItemAnnouncements` - Item notifications
-   `EventUserNotification` - User notifications
-   `EventError` - Error occurred

### SteamID Package

-   `ParseSteamID(s string) (SteamID, error)` - Parse from any format
-   `SteamID64() string` - Convert to SteamID64 format
-   `SteamID2() string` - Convert to SteamID2 format
-   `SteamID3() string` - Convert to SteamID3 format
-   `AccountID() AccountID` - Get account ID component
-   `Universe() AccountUniverse` - Get universe component
-   `Type() AccountType` - Get account type component

### TOTP Package

-   `CreateAuthenticationCode(sharedSecret string, time time.Time) (string, error)` - Generate Steam Guard code
-   `CreateDeviceID(steamID uint64) string` - Generate device ID from SteamID

## Architecture

### Package Structure

```
go-steam/
├── api/                 # Steam Web API client
│   ├── services/        # Service implementations
│   │   ├── iauthenticationservice/
│   │   ├── ieconservice/
│   │   ├── isteamdirectory/
│   │   └── itwofactorservice/
│   └── transports/      # HTTP transport layer
├── client/              # High-level Steam client
├── connection/          # TCP connection management
├── crypto/              # Steam cryptography (RSA, AES)
├── identity/            # Authentication identity providers
├── language/            # Auto-generated protobuf messages
├── protocol/            # Steam protocol implementation
├── steamid/             # Steam ID parsing and formatting
└── totp/                # TOTP/Steam Guard implementation
```

### Connection Flow

1. **Server Discovery**: Query Steam Directory API for available CM servers
2. **TCP Connection**: Establish connection to lowest-load server
3. **Session Key Exchange**: Perform encrypted session setup
4. **Authentication**: Send logon message with credentials
5. **Event Loop**: Handle incoming/outgoing protocol messages

## Examples

See the [examples directory](examples/) for more detailed usage examples:

-   [Simple Client](examples/simple/) - Basic connection and authentication

## Development

### Building

```bash
make build
```

### Testing

```bash
make test
```

### Benchmarking

```bash
make bench
```

### Generating Protobufs

```bash
make language
```

This will pull the latest protobufs from Valve and regenerate the Go code.

## Known Limitations

-   This library is **under active development**
-   **Breaking changes** are expected in future releases
-   Not all Steam protocol features are implemented yet
-   API stability is not guaranteed

## License

This project is licensed under the MIT License - see the [LICENSE.md](LICENSE.md) file for details.

## Acknowledgments

-   Steam Protocol implementation based on official Steam client behavior
-   Protobuf definitions from [SteamDatabase/Protobufs](https://github.com/SteamDatabase/Protobufs)
