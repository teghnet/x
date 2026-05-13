// Copyright (c) 2026 Paweł Zaremba
// SPDX-License-Identifier: MIT

package jsonio

import (
	"io"
	"os"
	"slices"

	"charm.land/log/v2"

	"github.com/teghnet/x"
)

type Collection[T comparable] []T

func (cs *Collection[T]) Add(c T) {
	if slices.Contains(*cs, c) {
		log.Debugf("config already exists: %v", c)
		return
	}
	*cs = append(*cs, c)
}

func (cs *Collection[T]) Load(s string) error {
	file, err := os.Open(s)
	if err != nil {
		return err
	}
	defer x.ClosePrint(file)
	return cs.Read(file)
}
func (cs *Collection[T]) Store(s string) error {
	file, err := os.Create(s)
	if err != nil {
		panic(err)
	}
	defer x.ClosePrint(file)
	return cs.Write(file)
}
func (cs *Collection[T]) Read(file io.Reader) error {
	for r := range List[T](file) {
		if r.Err != nil {
			return r.Err
		}
		cs.Add(r.Val)
	}
	return nil
}
func (cs *Collection[T]) Write(w io.Writer) error {
	var err error
	for _, cc := range *cs {
		err = Write(w, cc)
		if err != nil {
			return err
		}
	}
	return nil
}
