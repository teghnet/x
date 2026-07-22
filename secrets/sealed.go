package secrets

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"golang.org/x/crypto/argon2"
)

const (
	sealedSaltLen = 16
	sealedKeyLen  = 32 // AES-256

	// argon2id parameters. Deliberately heavier than OWASP's interactive
	// login minimum since this runs once at process start, not per request.
	argonTime    = 3
	argonMemory  = 64 * 1024 // KiB (64 MiB)
	argonThreads = 4
)

// SealedProvider holds secrets decrypted from a file at open time: JSON
// encrypted with AES-256-GCM under a key derived from a passphrase via
// argon2id. The file carries its own salt and nonce, so it's safe to keep on
// disk or back up — only the passphrase (never stored) unlocks it.
type SealedProvider struct {
	secrets map[string]string
}

// NewSealedProvider reads and decrypts the sealed file at path using
// passphrase. Decryption (and the argon2id key derivation) happens once,
// here; Get is a plain map lookup so repeated calls don't re-pay that cost.
func NewSealedProvider(path, passphrase string) (*SealedProvider, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("secrets: read %s: %w", path, err)
	}
	if len(raw) < sealedSaltLen {
		return nil, fmt.Errorf("secrets: %s is too short to be a sealed file", path)
	}
	salt, box := raw[:sealedSaltLen], raw[sealedSaltLen:]

	gcm, err := sealedGCM(passphrase, salt)
	if err != nil {
		return nil, err
	}
	if len(box) < gcm.NonceSize() {
		return nil, fmt.Errorf("secrets: %s is too short to be a sealed file", path)
	}
	nonce, ciphertext := box[:gcm.NonceSize()], box[gcm.NonceSize():]

	plain, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("secrets: decrypt %s: wrong passphrase or corrupt file", path)
	}

	m := map[string]string{}
	if err := json.Unmarshal(plain, &m); err != nil {
		return nil, fmt.Errorf("secrets: parse decrypted %s: %w", path, err)
	}
	return &SealedProvider{secrets: m}, nil
}

func (p *SealedProvider) Get(key string) (string, error) {
	v, ok := p.secrets[key]
	if !ok || v == "" {
		return "", fmt.Errorf("secrets: key %q not sealed", key)
	}
	return v, nil
}

// SealFile encrypts secrets as JSON and writes the sealed file to path
// (chmod 0600), generating a fresh random salt and nonce. Use it to create
// or rotate a sealed secrets file; accd itself only ever reads one.
func SealFile(path, passphrase string, secrets map[string]string) error {
	plain, err := json.Marshal(secrets)
	if err != nil {
		return fmt.Errorf("secrets: marshal: %w", err)
	}

	salt := make([]byte, sealedSaltLen)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		return fmt.Errorf("secrets: generate salt: %w", err)
	}
	gcm, err := sealedGCM(passphrase, salt)
	if err != nil {
		return err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return fmt.Errorf("secrets: generate nonce: %w", err)
	}

	box := gcm.Seal(nonce, nonce, plain, nil) // nonce || ciphertext || tag
	out := append(salt, box...)
	if err := os.WriteFile(path, out, 0o600); err != nil {
		return fmt.Errorf("secrets: write %s: %w", path, err)
	}
	return nil
}

func sealedGCM(passphrase string, salt []byte) (cipher.AEAD, error) {
	key := argon2.IDKey([]byte(passphrase), salt, argonTime, argonMemory, argonThreads, sealedKeyLen)
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("secrets: new cipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("secrets: new gcm: %w", err)
	}
	return gcm, nil
}
