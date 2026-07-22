package secrets

import (
	"fmt"

	"github.com/zalando/go-keyring"
)

// KeyringProvider stores each secret as its own entry in the OS keychain —
// macOS Keychain, Windows Credential Manager, or a Secret Service /
// libsecret-compatible store on Linux — keyed by (Service, key). Nothing
// touches disk under accd's control, so there's no file to protect or lose.
type KeyringProvider struct {
	// Service namespaces entries so accd's secrets don't collide with other
	// apps' keychain entries. Defaults to "accd" if empty.
	Service string
}

func (p KeyringProvider) service() string {
	if p.Service == "" {
		return "github.com/teghnet/accd"
	}
	return p.Service
}

func (p KeyringProvider) Get(key string) (string, error) {
	v, err := keyring.Get(p.service(), key)
	if err != nil {
		return "", fmt.Errorf("secrets: keyring get %q: %w", key, err)
	}
	return v, nil
}

// Set writes or overwrites the keychain entry for key. Use it from a one-off
// script (or a future `accd secrets set` subcommand) to provision a machine;
// accd itself only reads.
func (p KeyringProvider) Set(key, value string) error {
	if err := keyring.Set(p.service(), key, value); err != nil {
		return fmt.Errorf("secrets: keyring set %q: %w", key, err)
	}
	return nil
}

// Delete removes the keychain entry for key.
func (p KeyringProvider) Delete(key string) error {
	if err := keyring.Delete(p.service(), key); err != nil {
		return fmt.Errorf("secrets: keyring delete %q: %w", key, err)
	}
	return nil
}
