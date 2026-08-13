// Command steam-login signs in to Steam from the terminal.
//
// It uses the two-step login flow implemented by github.com/lbt05/go-steam:
//  1. BeginAuthSession encrypts the password with Steam's RSA key and opens
//     an auth session.
//  2. SubmitSteamGuardCode delivers the Steam Guard code (TOTP, email, or
//     mobile) that Steam demands; the code type is chosen automatically from
//     the session's AllowedConfirmations.
//  3. Connect + Logon open the CM connection and complete the logon.
//
// Credentials are read interactively by default. -username, -password and
// -shared-secret may be supplied to skip the corresponding prompts. After
// EventLoggedOn the program idles and prints protocol events until Ctrl-C.
//
// When -session-file points at a usable path, the SteamID and refresh token
// obtained after a successful login are written there (mode 0600) and a
// subsequent launch restores them via Client.SetIdentity, skipping the
// password prompt and Steam Guard entirely.
package main

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/lbt05/go-steam/api/services/iauthenticationservice"
	"github.com/lbt05/go-steam/api/services/iplayerservice"
	"github.com/lbt05/go-steam/api/transports"
	"github.com/lbt05/go-steam/client"
	"github.com/lbt05/go-steam/language/steam"
	"github.com/lbt05/go-steam/protocol"
	"github.com/lbt05/go-steam/steamid"
	"github.com/lbt05/go-steam/totp"
	"golang.org/x/term"
)

// savedSession is the on-disk representation of a previously-cached login
// session. It is written by saveSession after a successful logon and read
// by restoreSession on the next launch.
type savedSession struct {
	SteamID      uint64 `json:"steam_id"`
	RefreshToken string `json:"refresh_token"`
	Username     string `json:"username,omitempty"`
}

func defaultSessionPath() string {
	if dir, err := os.UserConfigDir(); err == nil && dir != "" {
		return filepath.Join(dir, "go-steam-example", "session.json")
	}
	if home, err := os.UserHomeDir(); err == nil {
		return filepath.Join(home, ".go-steam-example-session.json")
	}
	return "session.json"
}

// loadSession reads a previously-persisted session from path. The returned
// error is suitable for direct comparison against os.ErrNotExist by the
// caller; other errors indicate a malformed or partially written file.
func loadSession(path string) (*savedSession, error) {
	if path == "" {
		return nil, errors.New("session file path is empty")
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var s savedSession
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, err
	}
	if s.SteamID == 0 || s.RefreshToken == "" {
		return nil, errors.New("session file is empty or incomplete")
	}
	return &s, nil
}

// saveSession writes the currently-cached identity to disk so that the next
// launch can skip the full Steam Guard flow. The file is created with mode
// 0600 because the refresh token is a long-lived credential.
func saveSession(c *client.Client, path, username string) error {
	if path == "" {
		return nil
	}
	id, err := c.Identity(context.Background())
	if err != nil {
		return err
	}
	data, err := json.MarshalIndent(savedSession{
		SteamID:      id.SteamID().Uint64(),
		RefreshToken: id.RefreshToken(),
		Username:     username,
	}, "", "  ")
	if err != nil {
		return err
	}
	if dir := filepath.Dir(path); dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0o700); err != nil {
			return err
		}
	}
	return os.WriteFile(path, data, 0o600)
}

// restoreSession loads path, validates it, and installs the identity on c
// via Client.SetIdentity so that subsequent Connect/Logon skip the auth flow.
func restoreSession(c *client.Client, path string) (*savedSession, error) {
	s, err := loadSession(path)
	if err != nil {
		return nil, err
	}
	if _, err := c.SetIdentity(steamid.SteamID(s.SteamID), s.RefreshToken); err != nil {
		return nil, err
	}
	return s, nil
}

func main() {
	var (
		username     = flag.String("username", "", "Steam account name (skips the prompt)")
		password     = flag.String("password", "", "Steam account password (skips the prompt; prefer the prompt)")
		sharedSecret = flag.String("shared-secret", "", "Base64 TOTP shared secret for Steam Guard (skips the code prompt)")
		apiKey       = flag.String("api-key", "", "Steam Web API key (https://steamcommunity.com/dev/apikey); required for GetOwnedGames")
		verbose      = flag.Bool("v", false, "Print every protocol event (MessageSent / MessageReceived)")
		sessionFile  = flag.String("session-file", defaultSessionPath(), "Path to persist the session token (auto-created on first successful login); empty disables persistence")
	)
	flag.Parse()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	c, err := client.NewClient(client.WithReconnect(true))
	if err != nil {
		log.Fatalf("create client: %v", err)
	}

	// Try to restore a previously-saved session before prompting for
	// credentials. If anything is wrong (file missing, malformed,
	// SetIdentity rejected the values) we fall through to the full
	// two-step login flow.
	if restored, err := restoreSession(c, *sessionFile); err == nil {
		fmt.Printf("restored session for %s (SteamID=%d) from %s\n",
			restored.Username, restored.SteamID, *sessionFile)
		if *username == "" {
			*username = restored.Username
		}
	} else {
		if *sessionFile != "" && !errors.Is(err, os.ErrNotExist) {
			fmt.Fprintf(os.Stderr, "restore session: %v\n", err)
		}
		if err := runFullLogin(ctx, c, *username, *password, *sharedSecret); err != nil {
			log.Fatalf("login: %v", err)
		}
	}

	connectDone := make(chan error, 1)
	go func() {
		connectDone <- c.Connect(ctx)
	}()
	defer func() {
		if err := c.Disconnect(); err != nil {
			log.Printf("disconnect: %v", err)
		}
		<-connectDone
	}()

	fmt.Println("waiting for events (Ctrl-C to quit)...")
	for {
		select {
		case <-ctx.Done():
			fmt.Println("\ninterrupted, shutting down...")
			return

		case err := <-connectDone:
			if err != nil && !errors.Is(err, context.Canceled) {
				log.Fatalf("connection terminated: %v", err)
			}
			return

		case event := <-c.Events():
			handleEvent(c, event, *verbose, *apiKey, *sessionFile, *username)
		}
	}
}

// runFullLogin performs the interactive two-step login: it prompts for any
// missing credentials and drives BeginAuthSession + the Steam Guard second
// half. Callers should only invoke it when restoreSession fails.
func runFullLogin(ctx context.Context, c *client.Client, username, password, sharedSecret string) error {
	reader := bufio.NewReader(os.Stdin)
	if username == "" {
		username = prompt(reader, "Steam username: ")
	}
	if password == "" {
		v, err := promptPassword("Steam password: ")
		if err != nil {
			return fmt.Errorf("read password: %w", err)
		}
		password = v
	}

	creds := &client.Credentials{
		Username:     username,
		Password:     password,
		SharedSecret: sharedSecret,
	}

	session, err := c.BeginAuthSession(ctx, creds)
	if err != nil {
		return fmt.Errorf("begin auth session: %w", err)
	}
	fmt.Printf("auth session opened: client_id=%d interval=%s\n",
		session.ClientID, session.Interval)

	logonCode, err := resolveSteamGuardCode(reader, sharedSecret, session.AllowedConfirmations)
	switch {
	case err == nil:
		if _, err := c.SubmitSteamGuardCode(ctx, session, logonCode); err != nil {
			return fmt.Errorf("submit Steam Guard code: %w", err)
		}
	case errors.Is(err, errNeedsExternalConfirmation):
		if _, err := c.WaitForConfirmation(ctx, session); err != nil {
			return fmt.Errorf("wait for confirmation: %w", err)
		}
	default:
		return fmt.Errorf("prepare Steam Guard code: %w", err)
	}
	fmt.Println("Steam Guard accepted; tokens cached.")
	return nil
}

func handleEvent(c *client.Client, event protocol.Event, verbose bool, apiKey, sessionFile, username string) {
	switch e := event.(type) {
	case *protocol.EventSteamGuardChallenge:
		fmt.Printf("Steam Guard required (%d confirmations allowed)\n",
			len(e.AllowedConfirmations))
		for _, c := range e.AllowedConfirmations {
			fmt.Printf("  - %s\n", guardTypeLabel(c.ConfirmationType))
		}

	case *protocol.EventConnected:
		fmt.Println("connected to a Steam CM server")
		// Logon needs the protocol context; use Background here so it
		// survives the EventConnected handling and isn't tied to a
		// separate cancel scope.
		if err := c.Logon(context.Background()); err != nil {
			fmt.Fprintf(os.Stderr, "logon: %v\n", err)
		}

	case *protocol.EventLoggedOn:
		fmt.Printf("logged on: SteamID=%d SessionID=%d\n", e.SteamID, e.SessionID)
		if err := saveSession(c, sessionFile, username); err != nil {
			fmt.Fprintf(os.Stderr, "save session: %v\n", err)
		}
		go fetchAndPrintOwnedGames(context.Background(), uint64(e.SteamID), apiKey)

	case *protocol.EventError:
		fmt.Fprintf(os.Stderr, "protocol error: %v\n", e.Err)

	case *protocol.EventItemAnnouncements:
		fmt.Printf("item announcements: %d new\n", e.CountNewItems)

	case *protocol.EventUserNotification:
		fmt.Printf("user notification: type=%d count=%d\n",
			e.NotificationType, e.Count)

	case *protocol.EventMessageSent:
		if verbose {
			fmt.Printf("[out] %s %s\n", e.EMsg, e.Header)
		}

	case *protocol.EventMessageReceived:
		if verbose {
			fmt.Printf("[in]  %s %s\n", e.EMsg, e.Header)
		}

	default:
		if verbose {
			fmt.Printf("[event] %T %+v\n", event, event)
		}
	}
}

func fetchAndPrintOwnedGames(ctx context.Context, steamID uint64, apiKey string) {
	if apiKey == "" {
		fmt.Println("(skipping GetOwnedGames: no -api-key provided; register one at https://steamcommunity.com/dev/apikey)")
		return
	}

	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	playerService := iplayerservice.NewIPlayerService(transports.NewHTTPTransport())
	games, err := playerService.GetOwnedGames(ctx, iplayerservice.GetOwnedGamesParameters{
		APIKey:                 apiKey,
		SteamID:                steamID,
		IncludeAppInfo:         true,
		IncludePlayedFreeGames: true,
		Language:               "schinese",
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "GetOwnedGames: %v\n", err)
		return
	}

	printOwnedGames(ctx, *games)
}

func printOwnedGames(ctx context.Context, r iplayerservice.GetOwnedGamesResponse) {
	fmt.Printf("owned games: %d\n", r.GameCount)
	for _, g := range r.Games {
		last := "never"
		if g.RTimeLastPlayed != 0 {
			last = time.Unix(int64(g.RTimeLastPlayed), 0).Format("2006-01-02")
		}
		coverTag := urlStatus(ctx, g.GameCoverURL)
		logoTag := urlStatus(ctx, g.GameLogoURL)
		fmt.Printf("  - appid=%d name=%q playtime_forever=%dm last_played=%s img=%s cover=%s%s logo=%s%s\n",
			g.AppID, g.Name, g.PlaytimeForever, last, g.GameIconURL,
			g.GameCoverURL, coverTag, g.GameLogoURL, logoTag)
	}
}

func urlStatus(parent context.Context, raw string) string {
	if raw == "" {
		return ""
	}
	ctx, cancel := context.WithTimeout(parent, 10*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, raw, nil)
	if err != nil {
		return " [INVALID]"
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return " [INVALID]"
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		return " [INVALID]"
	}
	return ""
}

// errNeedsExternalConfirmation signals that Steam Guard should be resolved by
// calling Client.WaitForConfirmation (push approval in Steam Mobile or an
// email-link approval) rather than by submitting a code.
var errNeedsExternalConfirmation = errors.New("needs external confirmation")

func resolveSteamGuardCode(
	reader *bufio.Reader,
	sharedSecret string,
	allowed []iauthenticationservice.AllowedConfirmation,
) (string, error) {
	hasDeviceCode := confirmationAllowed(allowed,
		steam.EAuthSessionGuardType_k_EAuthSessionGuardType_DeviceCode)
	hasDeviceConfirmation := confirmationAllowed(allowed,
		steam.EAuthSessionGuardType_k_EAuthSessionGuardType_DeviceConfirmation)

	if sharedSecret != "" {
		if !hasDeviceCode {
			return "", fmt.Errorf(
				"shared secret supplied but Steam did not allow DeviceCode (TOTP); allowed: %s",
				joinGuardTypes(allowed))
		}
		return totp.CreateAuthenticationCode(sharedSecret, time.Now())
	}

	// No shared secret. If Steam Mobile push (DeviceConfirmation) is allowed,
	// the user can't generate a TOTP code from here — wait for them to approve
	// on their phone. The caller will route to Client.WaitForConfirmation.
	if hasDeviceConfirmation {
		fmt.Printf("Steam Guard required. Allowed methods: %s\n", joinGuardTypes(allowed))
		fmt.Println(externalConfirmationPrompt(allowed))
		return "", errNeedsExternalConfirmation
	}

	// No push available — fall back to a code prompt (email 2FA, etc.).
	if !confirmationAllowed(allowed, steam.EAuthSessionGuardType_k_EAuthSessionGuardType_EmailCode) &&
		!confirmationAllowed(allowed, steam.EAuthSessionGuardType_k_EAuthSessionGuardType_MachineToken) {
		return "", fmt.Errorf(
			"no usable Steam Guard method; allowed: %s",
			joinGuardTypes(allowed))
	}
	fmt.Printf("Steam Guard required. Allowed methods: %s\n", joinGuardTypes(allowed))
	fmt.Print("Enter Steam Guard code: ")
	code, err := reader.ReadString('\n')
	if err != nil && !errors.Is(err, io.EOF) {
		return "", err
	}
	return strings.TrimSpace(code), nil
}

func prompt(reader *bufio.Reader, label string) string {
	fmt.Print(label)
	line, err := reader.ReadString('\n')
	if err != nil {
		log.Fatalf("read input: %v", err)
	}
	return strings.TrimSpace(line)
}

func promptPassword(label string) (string, error) {
	fmt.Print(label)
	fd := int(os.Stdin.Fd())
	if !term.IsTerminal(fd) {
		// Fall back to a plain read when stdin is not a TTY (e.g. piped input).
		line, err := bufio.NewReader(os.Stdin).ReadString('\n')
		if err != nil && !errors.Is(err, io.EOF) {
			return "", err
		}
		return strings.TrimSpace(line), nil
	}
	bytes, err := term.ReadPassword(fd)
	fmt.Println() // move past the echo-less prompt
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func confirmationAllowed(
	allowed []iauthenticationservice.AllowedConfirmation,
	want steam.EAuthSessionGuardType,
) bool {
	for _, c := range allowed {
		if c.ConfirmationType == want {
			return true
		}
	}
	return false
}

func externalConfirmationPrompt(allowed []iauthenticationservice.AllowedConfirmation) string {
	if confirmationAllowed(allowed, steam.EAuthSessionGuardType_k_EAuthSessionGuardType_DeviceConfirmation) {
		return "Steam Guard: please approve the login in your Steam Mobile app..."
	}
	if confirmationAllowed(allowed, steam.EAuthSessionGuardType_k_EAuthSessionGuardType_EmailConfirmation) {
		return "Steam Guard: please click the approval link in the email sent to you..."
	}
	return "Steam Guard: waiting for external confirmation..."
}

func joinGuardTypes(allowed []iauthenticationservice.AllowedConfirmation) string {
	if len(allowed) == 0 {
		return "<none>"
	}
	parts := make([]string, 0, len(allowed))
	for _, c := range allowed {
		parts = append(parts, guardTypeLabel(c.ConfirmationType))
	}
	return strings.Join(parts, ", ")
}

func guardTypeLabel(t steam.EAuthSessionGuardType) string {
	const prefix = "k_EAuthSessionGuardType_"
	name := t.String()
	if strings.HasPrefix(name, prefix) {
		return strings.TrimPrefix(name, prefix)
	}
	return name
}
