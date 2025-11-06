package client

import (
	"context"
	"fmt"
	"sync"

	"github.com/lewisgibson/go-steam/identity"
)

// Credentials is a wrapper around the identity package to make the client easier to use.
type Credentials struct {
	Username     string
	Password     string
	SharedSecret string

	mutex    sync.Mutex
	provider *identity.CredentialsIdentityProvider
}

// Identity simply creates a new credentials identity provider and returns the identity.
func (c *Credentials) Identity(ctx context.Context) (Identity, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if c.provider == nil {
		provider, err := identity.NewCredentialsIdentityProvider(identity.Credentials{
			AccountName:  c.Username,
			Password:     c.Password,
			SharedSecret: c.SharedSecret,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to create identity provider: %w", err)
		}
		c.provider = provider
	}
	return c.provider.Identity(ctx)
}
