package secrets

import (
	"fmt"
	"os"
)

const (
	EnvPrefix            = "ACCD_SECRET"
	EnvSecretsPassphrase = "ACCD_SECRETS_PASSPHRASE"
)

type EnvProvider struct{ Prefix string } // e.g. "ACCD_SECRET_"

func (p EnvProvider) Get(key string) (string, error) {
	env := p.Prefix + toEnv(key)
	if v := os.Getenv(env); v != "" {
		return v, nil
	}
	return "", fmt.Errorf("secrets: env %q not set", env)
}

func toEnv(key string) string {
	b := []byte(key)
	for i, c := range b {
		switch {
		case c >= 'a' && c <= 'z':
			b[i] = c - 32
		case c == '.' || c == '-':
			b[i] = '_'
		}
	}
	return string(b)
}
