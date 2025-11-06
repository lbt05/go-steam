package ieconservice_test

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	_ "embed"

	"github.com/lewisgibson/go-steam/api/services/ieconservice"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

//go:embed fixtures/trade_offers.json
var getTradeOffersResponseBody []byte

func TestIEconService_GetTradeOffers_Success(t *testing.T) {
	t.Parallel()

	// Arrange: create a mock transport that returns the trade offers response body.
	transport := NewMockTransport(gomock.NewController(t))
	transport.EXPECT().
		Call(gomock.Any(), http.MethodGet, "IEconService", "GetTradeOffers", 1, gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, verb, service, method string, version int, params, response any) error {
			return json.Unmarshal(getTradeOffersResponseBody, response)
		})

	// Arrange: create service
	service := ieconservice.NewIEconService(transport)

	// Act: get trade offers
	response, err := service.GetTradeOffers(t.Context(), ieconservice.GetTradeOffersParameters{})
	require.NoError(t, err)

	// Assert: trade offers should be returned
	require.NotNil(t, response)
	require.Len(t, response.Sent, 1)
	require.Equal(t, "8588538066", response.Sent[0].ID)
	require.Len(t, response.Received, 1)
	require.Equal(t, "8588570967", response.Received[0].ID)
}
