package secrets

import (
	"encoding/json"
	"fmt"
	"os"
)

func NewFileProvider(path string) *FileProvider {
	return &FileProvider{path: path}
}

type FileProvider struct{ path string }

func (p FileProvider) Get(key string) (string, error) {
	b, err := os.ReadFile(p.path)
	if err != nil {
		return "", fmt.Errorf("secrets: read %s: %w", p.path, err)
	}
	m := map[string]string{}
	if err := json.Unmarshal(b, &m); err != nil {
		return "", fmt.Errorf("secrets: parse %s: %w", p.path, err)
	}
	v, ok := m[key]
	if !ok || v == "" {
		return "", fmt.Errorf("secrets: key %q not in %s", key, p.path)
	}
	return v, nil
}
