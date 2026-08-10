package ieconservice

import (
	"context"
	"net/http"

	"github.com/lbt05/go-steam/internal/unixtime"
	"github.com/lbt05/go-steam/steamid"
)

// GetTradeOffersParameters describes the parameters for the GetTradeOffers method.
type GetTradeOffersParameters struct {
	AccessToken          string `url:"access_token,omitempty"`           // Steamworks Web API access token.
	APIKey               string `url:"key,omitempty"`                    // Steamworks Web API user authentication key.
	Language             string `url:"language,omitempty"`               // The language to use when loading item display data.
	GetSentOffers        bool   `url:"get_sent_offers,omitempty"`        // Include sent offers.
	GetReceivedOffers    bool   `url:"get_received_offers,omitempty"`    // Include received offers.
	GetDescriptions      bool   `url:"get_descriptions,omitempty"`       // Include item display data for returned trade offers.
	ActiveOnly           bool   `url:"active_only,omitempty"`            // Include only active offers.
	TimeHistoricalCutoff uint32 `url:"time_historical_cutoff,omitempty"` // When active_only is set, include offers updated since this time.
	HistoricalOnly       bool   `url:"historical_only,omitempty"`        // Include only historical offers.
}

// GetTradeOffersResponse describes the response for the GetTradeOffers method.
type GetTradeOffersResponse struct {
	Sent       []TradeOffer `json:"trade_offers_sent"`     // Sent is the list of sent offers.
	Received   []TradeOffer `json:"trade_offers_received"` // Received is the list of received offers.
	NextCursor int          `json:"next_cursor"`           // NextCursor is the cursor to use to get the next page of offers.
}

type TradeOffer struct {
	ID           string           `json:"tradeofferid"`      // ID is the unique identifier for the offer.
	State        ETradeOfferState `json:"trade_offer_state"` // State is the current state of the offer.
	Eresult      int              `json:"eresult"`           // Eresult is the end result of the offer.
	Message      string           `json:"message"`           // Message is the message associated with the offer.
	Counterparty steamid.SteamID  `json:"accountid_other"`   // Counterparty is the SteamID of the counterparty.

	ConfirmationMethod ETradeOfferConfirmationMethod `json:"confirmation_method"`

	ItemsToGive    []TradeOfferItem `json:"items_to_give"`
	ItemsToReceive []TradeOfferItem `json:"items_to_receive"`

	IsOurOffer        bool `json:"is_our_offer"`         // IsOurOffer is true when we created the offer.
	FromRealTimeTrade bool `json:"from_real_time_trade"` // FromRealTimeTrade is true when the offer was created from a real-time trade modal.

	CreatedAt    unixtime.UnixTime `json:"time_created"`    // CreatedAt is the time the offer was created.
	UpdatedAt    unixtime.UnixTime `json:"time_updated"`    // UpdatedAt is the time the offer was last updated.
	ExpiresAt    unixtime.UnixTime `json:"expiration_time"` // ExpiresAt is the time the offer expires.
	SettlesAt    unixtime.UnixTime `json:"settlement_date"` // SettlesAt is the time the offer will no longer be reversible.
	EscrowEndsAt unixtime.UnixTime `json:"escrow_end_date"` // EscrowEndsAt is the time the offer will execute if escrow is required (no 2fa).
}

type TradeOfferItem struct {
	AppID      int    `json:"appid"`
	ContextID  string `json:"contextid"`
	AssetID    string `json:"assetid"`
	ClassID    string `json:"classid"`
	InstanceID string `json:"instanceid"`
	Amount     string `json:"amount"`
	Missing    bool   `json:"missing"`
	EstUsd     string `json:"est_usd"`
}

// GetTradeOffers gets trade offers.
func (a *IEconService) GetTradeOffers(ctx context.Context, params GetTradeOffersParameters) (*GetTradeOffersResponse, error) {
	if params.Language == "" {
		params.Language = "en-US"
	}
	var resBody struct {
		Response GetTradeOffersResponse `json:"response"`
	}
	if err := a.transport.Call(ctx, http.MethodGet, "IEconService", "GetTradeOffers", 1, params, &resBody); err != nil {
		return nil, err
	}
	return &resBody.Response, nil
}
