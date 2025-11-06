package main

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand/v2"
	"os"
	"os/signal"

	"github.com/lewisgibson/go-steam/client"
	"github.com/lewisgibson/go-steam/protocol"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	credentialsStr, ok := os.LookupEnv("GO_STEAM_CREDENTIALS")
	if !ok {
		fmt.Println("GO_STEAM_CREDENTIALS is not set")
		return
	}

	var credentialsList []struct {
		Username     string `json:"username"`
		Password     string `json:"password"`
		SharedSecret string `json:"shared_secret"`
	}
	if err := json.Unmarshal([]byte(credentialsStr), &credentialsList); err != nil {
		fmt.Printf("failed to unmarshal credentials: %v\n", err)
		return
	}
	credentials := credentialsList[rand.IntN(len(credentialsList))] //nolint:gosec

	c, err := client.NewClient(&client.Credentials{
		Username:     credentials.Username,
		Password:     credentials.Password,
		SharedSecret: credentials.SharedSecret,
	})
	if err != nil {
		fmt.Printf("failed to create client: %v\n", err)
		return
	}

	go func() {
		if err := c.Connect(ctx); err != nil {
			fmt.Printf("failed to connect: %v\n", err)
		}
	}()
	defer func() {
		fmt.Printf("Disconnecting...\n")
		if err := c.Disconnect(); err != nil {
			fmt.Printf("failed to disconnect: %v\n", err)
			return
		}
		fmt.Printf("Disconnected!\n")
	}()

	for {
		select {
		case <-ctx.Done():
			return

		case event := <-c.Events():
			switch event := event.(type) {
			case *protocol.EventMessageSent:
				fmt.Printf("-> %s %s %s\n", event.EMsg.String(), event.Header, event.Body)

			case *protocol.EventMessageReceived:
				fmt.Printf("<- %s %s\n", event.EMsg.String(), event.Header)

			case *protocol.EventConnected:
				fmt.Printf("Connected!\n")

				if err := c.Logon(ctx); err != nil {
					fmt.Printf("failed to login: %v\n", err)
				}

			case *protocol.EventLoggedOn:
				fmt.Printf("Logged on as %s (%d)!\n", credentials.Username, event.SteamID)

			case *protocol.EventItemAnnouncements:
				fmt.Printf("Item announcements: Count=%d\n", event.CountNewItems)

			case *protocol.EventUserNotification:
				fmt.Printf("User notification: Type=%d Count=%d\n", event.NotificationType, event.Count)
			}
		}
	}
}
