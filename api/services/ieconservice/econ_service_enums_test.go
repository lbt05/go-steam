package ieconservice_test

import (
	"testing"

	"github.com/lbt05/go-steam/api/services/ieconservice"
	"github.com/stretchr/testify/assert"
)

func TestETradeOfferConfirmationMethod(t *testing.T) {
	t.Parallel()

	assert.Equalf(t, 0, int(ieconservice.ETradeOfferConfirmationMethodNone), "These values are constants and should not be modified.")
	assert.Equalf(t, 1, int(ieconservice.ETradeOfferConfirmationMethodEmail), "These values are constants and should not be modified.")
	assert.Equalf(t, 2, int(ieconservice.ETradeOfferConfirmationMethodMobileApp), "These values are constants and should not be modified.")
}

func TestETradeOfferState(t *testing.T) {
	t.Parallel()

	assert.Equalf(t, 1, int(ieconservice.ETradeOfferStateInvalid), "These values are constants and should not be modified.")
	assert.Equalf(t, 2, int(ieconservice.ETradeOfferStateActive), "These values are constants and should not be modified.")
	assert.Equalf(t, 3, int(ieconservice.ETradeOfferStateAccepted), "These values are constants and should not be modified.")
	assert.Equalf(t, 4, int(ieconservice.ETradeOfferStateCountered), "These values are constants and should not be modified.")
	assert.Equalf(t, 5, int(ieconservice.ETradeOfferStateExpired), "These values are constants and should not be modified.")
	assert.Equalf(t, 6, int(ieconservice.ETradeOfferStateCanceled), "These values are constants and should not be modified.")
	assert.Equalf(t, 7, int(ieconservice.ETradeOfferStateDeclined), "These values are constants and should not be modified.")
	assert.Equalf(t, 8, int(ieconservice.ETradeOfferStateInvalidItems), "These values are constants and should not be modified.")
	assert.Equalf(t, 9, int(ieconservice.ETradeOfferStateCreatedNeedsConfirmation), "These values are constants and should not be modified.")
	assert.Equalf(t, 10, int(ieconservice.ETradeOfferStateCanceledBySecondFactor), "These values are constants and should not be modified.")
	assert.Equalf(t, 11, int(ieconservice.ETradeOfferStateInEscrow), "These values are constants and should not be modified.")
}
