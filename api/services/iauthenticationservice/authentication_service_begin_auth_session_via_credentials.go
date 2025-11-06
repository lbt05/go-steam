package iauthenticationservice

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/lewisgibson/go-steam/language/steam"
	"github.com/lewisgibson/go-steam/steamid"
)

// BeginAuthSessionViaCredentialsParameters describes the parameters for the BeginAuthSessionViaCredentials method.
type BeginAuthSessionViaCredentialsParameters struct {
	DeviceFriendlyName  string                                                `url:"device_friendly_name,omitempty"` // Required
	AccountName         string                                                `url:"account_name,omitempty"`         // Required
	EncryptedPassword   string                                                `url:"encrypted_password,omitempty"`   // Required
	EncryptionTimestamp uint64                                                `url:"encryption_timestamp,omitempty"` // Required
	RememberLogin       bool                                                  `url:"remember_login,omitempty"`       // Required (deprecated)
	PlatformType        int32                                                 `url:"platform_type,omitempty"`        // Required
	Persistence         int32                                                 `url:"persistence,omitempty"`          // Optional
	WebsiteID           string                                                `url:"website_id,omitempty"`           // Optional
	DeviceDetails       BeginAuthSessionViaCredentialsDeviceDetailsParameters `url:"device_details,omitempty"`       // Required
	GuardData           string                                                `url:"guard_data,omitempty"`           // Required
	Language            uint32                                                `url:"language,omitempty"`             // Required
	QosLevel            int32                                                 `url:"qos_level,omitempty"`            // Optional
}

// BeginAuthSessionViaCredentialsDeviceDetailsParameters represents device information for authentication.
type BeginAuthSessionViaCredentialsDeviceDetailsParameters struct {
	DeviceFriendlyName string `json:"device_friendly_name,omitempty"`
	PlatformType       int32  `json:"platform_type,omitempty"`
	OsType             int32  `json:"os_type,omitempty"`
	GamingDeviceType   uint32 `json:"gaming_device_type,omitempty"`
	ClientCount        uint32 `json:"client_count,omitempty"`
	MachineID          []byte `json:"machine_id,omitempty"`
}

// MarshalText implements encoding.TextMarshaler for URL encoding.
func (d BeginAuthSessionViaCredentialsDeviceDetailsParameters) MarshalText() ([]byte, error) {
	return json.Marshal(d)
}

// BeginAuthSessionViaCredentialsResponse describes the response for the BeginAuthSessionViaCredentials method.
type BeginAuthSessionViaCredentialsResponse struct {
	ClientID             uint64                `json:"client_id,string"`
	RequestID            string                `json:"request_id"`
	Interval             int32                 `json:"interval"`
	AllowedConfirmations []AllowedConfirmation `json:"allowed_confirmations"`
	SteamID              steamid.SteamID       `json:"steamid"`
	ExtendedErrorMessage string                `json:"extended_error_message"`
}

// AllowedConfirmation describes the allowed confirmation for the BeginAuthSessionViaCredentials method.
type AllowedConfirmation struct {
	ConfirmationType  steam.EAuthSessionGuardType `json:"confirmation_type"`
	AssociatedMessage string                      `json:"associated_message"`
}

// BeginAuthSessionViaCredentials begins an authentication session using credentials.
func (a *IAuthenticationService) BeginAuthSessionViaCredentials(ctx context.Context, params BeginAuthSessionViaCredentialsParameters) (*BeginAuthSessionViaCredentialsResponse, error) {
	var resBody struct {
		Response BeginAuthSessionViaCredentialsResponse `json:"response"`
	}
	if err := a.transport.Call(ctx, http.MethodPost, "IAuthenticationService", "BeginAuthSessionViaCredentials", 1, params, &resBody); err != nil {
		return nil, err
	}
	return &resBody.Response, nil
}
