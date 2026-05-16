package app

import "context"

func NewDefaultApp(ctx context.Context) (*DefaultApp, error) {
	app, err := NewConf[DefaultApp]()
	if err != nil {
		return nil, err
	}
	app.Init(WithContext(ctx))
	return app, nil
}

type DefaultApp struct {
	App `json:"-" envPrefix:"APP_"`
}
