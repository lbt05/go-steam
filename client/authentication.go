package client

// Credentials describes the Steam account credentials used to perform the
// first half of the login flow via Client.BeginAuthSession. It is a pure
// configuration value; the actual authentication state is owned by
// Client.
type Credentials struct {
	Username     string
	Password     string
	SharedSecret string
}