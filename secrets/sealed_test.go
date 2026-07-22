package secrets

import (
	"path/filepath"
	"testing"
)

func TestSealedProviderRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "secrets.sealed")
	want := map[string]string{
		"stripe.api_key":    "sk_test_123",
		"wfirma.access_key": "ak_abc",
	}
	if err := SealFile(path, "correct horse battery staple", want); err != nil {
		t.Fatalf("SealFile: %v", err)
	}

	p, err := NewSealedProvider(path, "correct horse battery staple")
	if err != nil {
		t.Fatalf("NewSealedProvider: %v", err)
	}
	for k, v := range want {
		got, err := p.Get(k)
		if err != nil {
			t.Fatalf("Get(%q): %v", k, err)
		}
		if got != v {
			t.Errorf("Get(%q) = %q, want %q", k, got, v)
		}
	}

	if _, err := p.Get("missing.key"); err == nil {
		t.Error("Get(missing.key) = nil error, want error")
	}
}

func TestSealedProviderWrongPassphrase(t *testing.T) {
	path := filepath.Join(t.TempDir(), "secrets.sealed")
	if err := SealFile(path, "right", map[string]string{"a": "b"}); err != nil {
		t.Fatalf("SealFile: %v", err)
	}

	if _, err := NewSealedProvider(path, "wrong"); err == nil {
		t.Error("NewSealedProvider with wrong passphrase = nil error, want error")
	}
}

func TestSealedProviderCorruptFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "secrets.sealed")
	if _, err := NewSealedProvider(path, "whatever"); err == nil {
		t.Error("NewSealedProvider on missing file = nil error, want error")
	}
}
