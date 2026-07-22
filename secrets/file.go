package secrets

import (
	"encoding/json"
	"fmt"
	"os"
)

type FileProvider struct{ Path string }

func (p FileProvider) Get(key string) (string, error) {
	b, err := os.ReadFile(p.Path)
	if err != nil {
		return "", fmt.Errorf("secrets: read %s: %w", p.Path, err)
	}
	m := map[string]string{}
	if err := json.Unmarshal(b, &m); err != nil {
		return "", fmt.Errorf("secrets: parse %s: %w", p.Path, err)
	}
	v, ok := m[key]
	if !ok || v == "" {
		return "", fmt.Errorf("secrets: key %q not in %s", key, p.Path)
	}
	return v, nil
}
