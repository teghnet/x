package app

import (
	"context"
	"fmt"
	"strings"

	"github.com/teghnet/x/paths"
)

const (
	nameProfileDB = "profiles.jsonl"
	nameProfile   = "profile.json"
)

type App struct {
	Name                string `env:"NAME,unset"`
	PreferWDStore       bool   `env:"PREFER_WD_STORE,unset"`
	PreferDotLocalStore bool   `env:"PREFER_DOT_LOCAL_STORE,unset"`

	context.Context
	cancel context.CancelCauseFunc

	paths.XDG
}
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
func (a *App) Init(ctx context.Context, opts ...Option) {
	for _, opt := range opts {
		opt(a)
	}
	a.Context, a.cancel = context.WithCancelCause(ctx)
	a.XDG = paths.NewXDG(a.Name,
		paths.WithCurrentDirsPreference(a.PreferWDStore),
		paths.WithLocalDirsPreference(a.PreferDotLocalStore),
	)
}
func (a App) Filename(elem ...string) string {
	return strings.Join(append([]string{a.Name}, elem...), "-")
}
func (a App) ProfileFilename() string {
	return a.Filename(nameProfile)
}
func (a App) ProfileDBFilename() string {
	return a.Filename(nameProfileDB)
}

// func (a App) Context() context.Context {
// 	return a.Context
// }

// func (a App) Done() <-chan struct{} {
// 	return a.Context.Done()
// }

func (a App) Close() error {
	a.cancel(fmt.Errorf("%T closed", a))
	return nil
}

// func (a App) Config(s ...string) string {
// 	return a.XDG.Config(s...)
// }
// func (a App) Data(s ...string) string {
// 	return a.XDG.Data(s...)
// }
// func (a App) Cache(s ...string) string {
// 	return a.XDG.Cache(s...)
// }
// func (a App) State(s ...string) string {
// 	return a.XDG.State(s...)
// }

func NewDefaultApp(ctx context.Context) (*DefaultApp, error) {
	app, err := NewConf[DefaultApp]()
	if err != nil {
		return nil, err
	}
	app.Init(ctx)
	return app, nil
}

type DefaultApp struct {
	App `json:"-" envPrefix:"APP_"`
}
