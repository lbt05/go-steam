package ieconservice

type ETradeOfferConfirmationMethod int32

const (
	ETradeOfferConfirmationMethodNone      ETradeOfferConfirmationMethod = 0
	ETradeOfferConfirmationMethodEmail     ETradeOfferConfirmationMethod = 1
	ETradeOfferConfirmationMethodMobileApp ETradeOfferConfirmationMethod = 2
)

type ETradeOfferState int32

const (
	ETradeOfferStateInvalid                  ETradeOfferState = 1
	ETradeOfferStateActive                   ETradeOfferState = 2
	ETradeOfferStateAccepted                 ETradeOfferState = 3
	ETradeOfferStateCountered                ETradeOfferState = 4
	ETradeOfferStateExpired                  ETradeOfferState = 5
	ETradeOfferStateCanceled                 ETradeOfferState = 6
	ETradeOfferStateDeclined                 ETradeOfferState = 7
	ETradeOfferStateInvalidItems             ETradeOfferState = 8
	ETradeOfferStateCreatedNeedsConfirmation ETradeOfferState = 9
	ETradeOfferStateCanceledBySecondFactor   ETradeOfferState = 10
	ETradeOfferStateInEscrow                 ETradeOfferState = 11
)
