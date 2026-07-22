// Package secrets is the port for retrieving API credentials. You will hold
// live Stripe and wFirma keys, so credentials must not sit in config files or
// source.
//
// Five providers ship here, all satisfying the same Provider interface:
//   - EnvProvider: reads from environment variables. Good for dev and for
//     hosted deployments that inject secrets via the platform.
//   - FileProvider: reads a JSON file (chmod 0600). Convenient locally but
//     PLAINTEXT — acceptable only on a single-user machine you control.
//   - SealedProvider (sealed.go): a hardened FileProvider — the JSON is
//     encrypted with AES-256-GCM under a key derived from a passphrase via
//     argon2id, so the file is safe to keep on disk or back up.
//   - KeyringProvider (keyring.go): backed by the OS keychain
//     (github.com/zalando/go-keyring) — no file at all.
//   - KeepassXCProvider (keepassxc.go): backed by a local KeepassXC
//     instance via its Secret Service Integration setting — same
//     mechanism as KeyringProvider, kept distinct for discoverability.
package secrets

import (
	"context"
	"fmt"
	"os"
)

const (
	File      = "file"
	KeepassXC = "keepassXC"
	Keyring   = "keyring"
	Sealed    = "sealed"
)

var (
	_ Provider = EnvProvider{}
	_ Provider = FileProvider{}
	_ Provider = (*KeepassXCProvider)(nil)
	_ Provider = KeyringProvider{}
	_ Provider = (*SealedProvider)(nil)
)

type Provider interface {
	// Get returns the secret for key (e.g. "stripe.api_key"), or an error if
	// it is missing.
	Get(key string) (string, error)
}

type Config struct {
	Context context.Context

	FilePath       string `json:"file_path"`
	KeepassPath    string `json:"keepass_path"`
	KeyringService string `json:"keyring_service"`
	SealedPath     string `json:"sealed_path"`
}

// OpenSecrets picks a secrets.Provider from cfg.SecretsProvider. "sealed"
// needs a passphrase, which is itself a secret: it comes from
// ACCD_SECRETS_PASSPHRASE (env), never from config.json.
func OpenSecrets(cfg Config) (Provider, error) {
	switch {
	case cfg.FilePath != "":
		return FileProvider{Path: cfg.FilePath}, nil
	case cfg.KeepassPath != "":
		if cfg.Context == nil {
			cfg.Context = context.Background()
		}
		return NewKeepassXCProvider(cfg.Context, cfg.KeepassPath), nil
	case cfg.KeyringService != "":
		return KeyringProvider{Service: cfg.KeyringService}, nil
	case cfg.SealedPath != "":
		pass := os.Getenv(EnvSecretsPassphrase)
		if pass == "" {
			return nil, fmt.Errorf("requires %s", EnvSecretsPassphrase)
		}
		return NewSealedProvider(cfg.SealedPath, pass)
	default:
		return EnvProvider{Prefix: EnvPrefix + "_"}, nil
	}
}
