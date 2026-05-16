// Copyright (c) 2026 Paweł Zaremba
// SPDX-License-Identifier: MIT

package jsonio

import (
	"errors"
	"fmt"
	"os"
	"slices"

	"charm.land/log/v2"

	"github.com/teghnet/x"
)

type Collection[T comparable] []T

var errItemExists = errors.New("item already exists")
var errItemOfTypeExists = func(t any) error { return fmt.Errorf("%w: %v", errItemExists, t) }

func (c *Collection[T]) Add(t T) error {
	if slices.Contains(*c, t) {
		return errItemOfTypeExists(t)
	}
	*c = append(*c, t)
	return nil
}
func (c *Collection[T]) Add1(t T) {
	if err := c.Add(t); err != nil {
		if errors.Is(err, errItemExists) {
			return
		}
		log.Error(err)
	}
}
func Collect[T comparable](path string) (Collection[T], error) {
	r, err := os.Open(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			log.Debug(err)
			return nil, nil
		}
		return nil, err
	}
	defer x.ClosePrint(r)
	var v Collection[T]
	for res := range List[T](r) {
		if res.Err != nil {
			log.Debug(res.Err)
			continue
		}
		v = append(v, res.Val)
	}
	return v, nil
}

func Collect1[T comparable](path string) (Collection[T], error) {
	r, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer x.ClosePrint(r)
	var v Collection[T]
	for res := range List[T](r) {
		if res.Err != nil {
			return nil, res.Err
		}
		v = append(v, res.Val)
	}
	return v, nil
}

func Save[T comparable](path string, c Collection[T]) error {
	w, err := os.Create(path)
	if err != nil {
		return err
	}
	defer x.ClosePrint(w)
	for _, cc := range c {
		err = Write(w, cc)
		if err != nil {
			return err
		}
	}
	return nil
}
