package app

import (
	"errors"
	"os"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
)

func NewConf[T any]() (*T, error) {
	if err := godotenv.Load(); err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, err
	}
	conf, err := env.ParseAs[T]()
	if err != nil {
		return nil, err
	}
	return &conf, nil
}
