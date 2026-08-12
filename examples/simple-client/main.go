// Package main demonstrates a minimal two-step Steam login using go-steam.
//
// Set STEAM_USERNAME, STEAM_PASSWORD and (optionally) STEAM_SHARED_SECRET in
// the environment before running:
//
//	export STEAM_USERNAME=your_username
//	export STEAM_PASSWORD=your_password
//	export STEAM_SHARED_SECRET=your_base64_shared_secret
//	go run ./examples/simple-client
//
// The program picks the right second-half flow based on what Steam allows:
//   - SharedSecret set   → TOTP code via SubmitSteamGuardCode (DeviceCode).
//   - Push or email-link → WaitForConfirmation (no code to type; the user
//                          approves on the Steam Mobile app or via email).
//   - Otherwise          → prompt on stdin for an email code, then
//                          SubmitSteamGuardCode (EmailCode).
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
	"github.com/lbt05/go-steam/identity"
	"github.com/lbt05/go-steam/protocol"
	"github.com/lbt05/go-steam/totp"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	username := os.Getenv("STEAM_USERNAME")
	password := os.Getenv("STEAM_PASSWORD")
	sharedSecret := os.Getenv("STEAM_SHARED_SECRET")
	if username == "" || password == "" {
		log.Fatal("STEAM_USERNAME and STEAM_PASSWORD must be set")
	}

	c, err := client.NewClient()
	if err != nil {
		log.Fatalf("create client: %v", err)
	}

	// First half: encrypt the password and open a Steam auth session.
	session, err := c.BeginAuthSession(ctx, &client.Credentials{
		Username:     username,
		Password:     password,
		SharedSecret: sharedSecret,
	})
	if err != nil {
		log.Fatalf("begin auth session: %v", err)
	}
	fmt.Printf("auth session opened: client_id=%d interval=%s\n", session.ClientID, session.Interval)

	// Second half: dispatch on what Steam allows and what the caller has.
	if err := authorize(ctx, c, session, sharedSecret); err != nil {
		log.Fatalf("authorize: %v", err)
	}
	fmt.Println("authorized; tokens cached.")

	// Drive the connection and logon in the background.
	go func() {
		if err := c.Connect(ctx); err != nil {
			log.Printf("connect: %v", err)
		}
	}()
	defer func() {
		if err := c.Disconnect(); err != nil {
			log.Printf("disconnect: %v", err)
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return
		case event := <-c.Events():
			switch e := event.(type) {
			case *protocol.EventSteamGuardChallenge:
				fmt.Printf("Steam Guard required (%d confirmations allowed)\n", len(e.AllowedConfirmations))
			case *protocol.EventConnected:
				fmt.Println("connected to Steam CM server")
				if err := c.Logon(ctx); err != nil {
					log.Printf("logon: %v", err)
				}
			case *protocol.EventLoggedOn:
				fmt.Printf("logged on: SteamID=%d SessionID=%d\n", e.SteamID, e.SessionID)
				return
			case *protocol.EventError:
				log.Printf("protocol error: %v", e.Err)
			}
		}
	}
}

// authorize runs the second half of the login flow, dispatching to the
// appropriate client method based on what Steam allows in the session and
// what credentials the caller has supplied. SharedSecret implies TOTP
// (DeviceCode). Otherwise, if Steam allows an external approval flow
// (DeviceConfirmation/EmailConfirmation), we wait for the user to tap
// Approve in the Steam Mobile app or click the email link. Otherwise we
// prompt on stdin for an email or mobile code.
func authorize(ctx context.Context, c *client.Client, session *identity.AuthSession, sharedSecret string) error {
	if sharedSecret != "" {
		code, err := totp.CreateAuthenticationCode(sharedSecret, time.Now())
		if err != nil {
			return fmt.Errorf("generate TOTP code: %w", err)
		}
		if _, err := c.SubmitSteamGuardCode(ctx, session, code); err != nil {
			return err
		}
		return nil
	}

	// No SharedSecret — try external confirmation first (mobile push /
	// email link). If Steam doesn't allow that, fall back to prompting for
	// an email code on stdin.
	if _, err := c.WaitForConfirmation(ctx, session); err == nil {
		fmt.Println("approved externally; tokens cached.")
		return nil
	} else if !strings.Contains(err.Error(), "no allowed external confirmation") {
		return err
	}

	fmt.Print("Enter Steam Guard code: ")
	line, _ := bufio.NewReader(os.Stdin).ReadString('\n')
	code := strings.TrimSpace(line)
	if _, err := c.SubmitSteamGuardCode(ctx, session, code); err != nil {
		return err
	}
	return nil
}