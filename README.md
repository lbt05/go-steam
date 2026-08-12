# go-steam

[![Build Workflow](https://github.com/lbt05/go-steam/actions/workflows/build.yaml/badge.svg)](https://github.com/lbt05/go-steam/actions/workflows/build.yaml)
[![Pkg Go Dev](https://pkg.go.dev/badge/github.com/lbt05/go-steam)](https://pkg.go.dev/github.com/lbt05/go-steam)

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

-   [Discussions](https://github.com/lbt05/go-steam/discussions)
-   [Reference](https://pkg.go.dev/github.com/lbt05/go-steam)
-   [Examples](examples/)

## Installation

```sh
go get github.com/lbt05/go-steam
```

## Quickstart

### Two-Step Login

Login is a two-step flow. First call `BeginAuthSession` to encrypt your
password and open a Steam authentication session; Steam will then ask for a
Steam Guard code (TOTP, email or mobile). Submit it with
`SubmitSteamGuardCode`, then connect and log on as usual.

```go
package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/lbt05/go-steam/client"
	"github.com/lbt05/go-steam/protocol"
	"github.com/lbt05/go-steam/totp"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	// Create a Steam client. No credentials yet — they are passed to
	// BeginAuthSession below so the login flow can be driven explicitly.
	c, err := client.NewClient()
	if err != nil {
		log.Fatal(err)
	}

	// Step 1: open an auth session.
	session, err := c.BeginAuthSession(ctx, &client.Credentials{
		Username: "your_username",
		Password: "your_password",
		// SharedSecret is optional. Leave empty if you need to be prompted
		// for an email/mobile Steam Guard code.
		SharedSecret: "your_base64_shared_secret",
	})
	if err != nil {
		log.Fatal(err)
	}

	// Step 2: produce and submit the Steam Guard code. The SharedSecret and
	// a user-entered code are mutually exclusive — pick one path. SubmitSteamGuardCode
	// chooses the correct code type (DeviceCode for TOTP, EmailCode for
	// email 2FA, etc.) automatically from session.AllowedConfirmations.
	code, err := steamGuardCode("your_base64_shared_secret")
	if err != nil {
		log.Fatal(err)
	}
	if _, err := c.SubmitSteamGuardCode(ctx, session, code); err != nil {
		log.Fatal(err)
	}

	// Connect to Steam and log on once the auth session has completed.
	go func() {
		if err := c.Connect(ctx); err != nil {
			log.Printf("Connection error: %v", err)
		}
	}()
	defer c.Disconnect()

	for {
		select {
		case <-ctx.Done():
			return
		case event := <-c.Events():
			switch e := event.(type) {
			case *protocol.EventSteamGuardChallenge:
				fmt.Printf("Steam Guard required (%d confirmations allowed)\n", len(e.AllowedConfirmations))
			case *protocol.EventConnected:
				fmt.Println("Connected to Steam!")
				if err := c.Logon(ctx); err != nil {
					log.Printf("Logon error: %v", err)
				}
			case *protocol.EventLoggedOn:
				fmt.Printf("Logged on! SteamID: %d\n", e.SteamID)
				return
			case *protocol.EventError:
				log.Printf("Error: %v", e.Err)
			}
		}
	}
}

// steamGuardCode returns the Steam Guard code to submit. With a shared
// secret it derives a fresh TOTP code; otherwise it prompts the user on
// stdin for an email or mobile code.
func steamGuardCode(sharedSecret string) (string, error) {
	if sharedSecret != "" {
		return totp.CreateAuthenticationCode(sharedSecret, time.Now())
	}
	fmt.Print("Enter Steam Guard code: ")
	line, _ := bufio.NewReader(os.Stdin).ReadString('\n')
	return strings.TrimSpace(line), nil
}
```

A complete, runnable version of this example lives in
[`examples/simple-client/main.go`](examples/simple-client/main.go). The
[`examples/simple`](examples/simple/) directory contains a richer demo that
reads multiple accounts from a JSON environment variable and pretty-prints
all protocol events.

### Steam ID Manipulation

```go
package main

import (
	"fmt"
	"log"

	"github.com/lbt05/go-steam/steamid"
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

	"github.com/lbt05/go-steam/totp"
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
	"time"

	"github.com/lbt05/go-steam/client"
	"github.com/lbt05/go-steam/connection"
)

func main() {
	dialer := &net.Dialer{
		Timeout: 10 * time.Second,
	}

	c, err := client.NewClient(
		client.WithDialer(dialer),
		client.WithReconnect(true), // Enable automatic reconnection
	)
	if err != nil {
		log.Fatal(err)
	}

	// Credentials are passed to BeginAuthSession, not to NewClient.
	_ = c
}
```

## API Reference

### Client Package

-   `NewClient(opts ...ClientOption) (*Client, error)` - Create a new Steam client
-   `BeginAuthSession(ctx, *Credentials) (*identity.AuthSession, error)` - First half of login; encrypts the password and opens an auth session
-   `SubmitSteamGuardCode(ctx, *identity.AuthSession, code string) (Identity, error)` - Second half of login; submits a Steam Guard code (TOTP, email, mobile, etc.) and caches the resulting identity. The code type is chosen automatically from `session.AllowedConfirmations`.
-   `Identity(ctx) (Identity, error)` - Returns the identity cached by the most recent `SubmitSteamGuardCode` call
-   `Connect(ctx context.Context) error` - Connect to Steam servers
-   `Disconnect() error` - Disconnect from Steam
-   `Logon(ctx context.Context) error` - Log on to Steam using the cached identity
-   `Events() <-chan protocol.Event` - Get event channel

### Client Options

-   `WithAPI(api API)` - Set custom Steam Directory API client
-   `WithAuthAPI(api identity.API)` - Set custom identity API used during the two-step authentication flow
-   `WithDialer(dialer connection.Dialer)` - Set custom network dialer
-   `WithReconnect(reconnect bool)` - Enable/disable automatic reconnection

### Protocol Events

-   `EventConnected` - Connected to Steam server
-   `EventLoggedOn` - Successfully logged on
-   `EventSteamGuardChallenge` - Emitted after `BeginAuthSession`; carries the `identity.AuthSession` and the allowed Steam Guard confirmation types
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

### Web API Services

Each Web API service is reachable through `*api.API`. Reach the player service
and call:

```go
games, err := apiClient.IPlayerService().GetOwnedGames(ctx, iplayerservice.GetOwnedGamesParameters{
    SteamID:        session.SteamID.Uint64(),
    IncludeAppInfo: true,
})
```

-   `(*api.API).IPlayerService().GetOwnedGames(ctx, GetOwnedGamesParameters) (*GetOwnedGamesResponse, error)` - List the games owned by the given SteamID. Requires the player's game details to be set to Public.

See the per-service `*_test.go` files (e.g. `api/services/ieconservice/econ_service_get_trade_offers_test.go`) for usage patterns with a custom `transports.Transport`.

## Architecture

### Package Structure

```
go-steam/
├── api/                 # Steam Web API client
│   ├── services/        # Service implementations
│   │   ├── iauthenticationservice/
│   │   ├── ieconservice/
│   │   ├── iplayerservice/
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

-   [Simple Client](examples/simple-client/) - Minimal two-step login (SharedSecret or interactive email code)
-   [Full Demo](examples/simple/) - Multi-account demo with verbose protocol event tracing

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
