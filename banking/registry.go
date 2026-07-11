package banking

import (
	"maps"
	"slices"
	"sync"
)

var (
	registryMu sync.RWMutex
	registry   = make(map[string]func() Parser)
)

// Register adds a new bank parser factory to the global registry under name.
// Registering the same name twice overwrites the previous factory.
func Register(name string, factory func() Parser) {
	registryMu.Lock()
	defer registryMu.Unlock()
	registry[name] = factory
}

// GetParser returns a freshly constructed parser registered under name. ok is
// false if no parser is registered under that name.
func GetParser(name string) (parser Parser, ok bool) {
	registryMu.RLock()
	factory, ok := registry[name]
	registryMu.RUnlock()
	if !ok {
		return nil, false
	}
	return factory(), true
}

// List returns the names of all registered parsers in lexical order.
func List() []string {
	registryMu.RLock()
	defer registryMu.RUnlock()
	return slices.Sorted(maps.Keys(registry))
}
