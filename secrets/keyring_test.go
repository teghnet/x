package secrets

import "testing"

// TestKeyringProvider hits the real OS keychain. There's no fake to swap in
// (that's the point of this provider), so skip where no backend is
// available instead of failing CI on headless/minimal Linux runners.
func TestKeyringProvider(t *testing.T) {
	p := KeyringProvider{service: "accd-test"}
	const key, val = "unit.test.key", "unit-test-value"

	if err := p.Set(key, val); err != nil {
		t.Skipf("no OS keychain backend available: %v", err)
	}
	defer p.Delete(key)

	got, err := p.Get(key)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got != val {
		t.Errorf("Get() = %q, want %q", got, val)
	}

	if err := p.Delete(key); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if _, err := p.Get(key); err == nil {
		t.Error("Get after Delete = nil error, want error")
	}
}
