package app

import "context"

type Option func(*App)

func DefaultName(name string) Option {
	return func(a *App) {
		if a.Name == "" {
			a.Name = name
		}
	}
}
func OverrideName(name string) Option {
	return func(a *App) { a.Name = name }
}
func WithPreferWDStore(prefer bool) Option {
	return func(a *App) { a.PreferWDStore = prefer }
}

func WithPreferDotLocalStore(prefer bool) Option {
	return func(a *App) { a.PreferDotLocalStore = prefer }
}

func WithContext(ctx context.Context) Option {
	return func(a *App) { a.ctx = ctx }
}
