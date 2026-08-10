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
// Without STEAM_SHARED_SECRET the program prompts on stdin for an email or
// mobile Steam Guard code.
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

	// Second half: produce the Steam Guard code and submit it. With a
	// SharedSecret the TOTP code is generated automatically (no prompt);
	// otherwise the user is asked to type in an email or mobile code.
	code, err := steamGuardCode(sharedSecret)
	if err != nil {
		log.Fatalf("prepare Steam Guard code: %v", err)
	}
	if _, err := c.SubmitSteamGuardCode(ctx, session, code); err != nil {
		log.Fatalf("submit Steam Guard code: %v", err)
	}
	fmt.Println("Steam Guard code accepted; tokens cached.")

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

	// Consume events until interrupted.
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

// steamGuardCode returns the Steam Guard code to submit. The two-step login
// flow is mutually exclusive: a SharedSecret means a TOTP code is generated
// automatically (no user input), while an empty SharedSecret means the user
// is prompted on stdin for an email or mobile code.
func steamGuardCode(sharedSecret string) (string, error) {
	if sharedSecret != "" {
		// TOTP path. Steam's TOTP uses a 30-second step. A small clock
		// skew between this machine and Steam is usually tolerated; for
		// stricter reliability, query Steam's server time via
		// ITwoFactorService.QueryTime and pass it to
		// totp.CreateAuthenticationCode instead of time.Now().
		return totp.CreateAuthenticationCode(sharedSecret, time.Now())
	}

	// Email/mobile code path: prompt the user.
	fmt.Print("Enter Steam Guard code: ")
	line, _ := bufio.NewReader(os.Stdin).ReadString('\n')
	return strings.TrimSpace(line), nil
}