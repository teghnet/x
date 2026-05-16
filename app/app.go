package app

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/teghnet/x/paths"
)

type App struct {
	Name                string `env:"NAME,unset"`
	PreferWDStore       bool   `env:"PREFER_WD_STORE,unset"`
	PreferDotLocalStore bool   `env:"PREFER_DOT_LOCAL_STORE,unset"`

	xdg paths.XDG
	ctx context.Context
	cnc context.CancelCauseFunc
}

func (a *App) Init(opts ...Option) {
	for _, opt := range opts {
		opt(a)
	}
	if a.ctx == nil {
		a.ctx = context.Background()
	}
	a.ctx, a.cnc = context.WithCancelCause(a.ctx)
	a.xdg = paths.NewXDG(a.Name,
		paths.WithPreferWDStore(a.PreferWDStore),
		paths.WithPreferDotLocalStore(a.PreferDotLocalStore),
	)
}

func (a *App) Filename(elem ...string) string {
	return strings.Join(append([]string{a.Name}, elem...), "-")
}

// [io.Closer]

// Close implements [io.Closer]
func (a *App) Close() error {
	a.cnc(fmt.Errorf("%T closed", a))
	return nil
}

// [context.Context]

// Deadline implements [context.Context]
func (a *App) Deadline() (deadline time.Time, ok bool) {
	return a.ctx.Deadline()
}

// Done implements [context.Context]
func (a *App) Done() <-chan struct{} {
	return a.ctx.Done()
}

// Err implements [context.Context]
func (a *App) Err() error {
	return a.ctx.Err()
}

// Value implements [context.Context]
func (a *App) Value(key any) any {
	return a.ctx.Value(key)
}

// [paths.XDG]

// CachePath implements [paths.XDG]
func (a *App) CachePath(s ...string) string {
	return a.xdg.CachePath(s...)
}

// ConfigPath implements [paths.XDG]
func (a *App) ConfigPath(s ...string) string {
	return a.xdg.ConfigPath(s...)
}

// DataPath implements [paths.XDG]
func (a *App) DataPath(s ...string) string {
	return a.xdg.DataPath(s...)
}

// StatePath implements [paths.XDG]
func (a *App) StatePath(s ...string) string {
	return a.xdg.StatePath(s...)
}
