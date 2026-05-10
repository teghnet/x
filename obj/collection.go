// Copyright (c) 2026 Paweł Zaremba
// SPDX-License-Identifier: MIT

package obj

import (
	"io"
	"os"
	"slices"

	"charm.land/log/v2"

	"github.com/teghnet/x"
	"github.com/teghnet/x/jsonio"
)

type Collection[T comparable] []T

func (cs *Collection[T]) Add(c T) {
	if slices.Contains(*cs, c) {
		log.Warnf("config already exists: %v", c)
		return
	}
	*cs = append(*cs, c)
}

type filePath func(...string) string

func (cs *Collection[T]) Load(s filePath) error {
	file, err := os.Open(s())
	if err != nil {
		return err
	}
	defer x.ClosePrint(file)
	return cs.Read(file)
}
func (cs *Collection[T]) Read(file io.Reader) error {
	for cc, err := range jsonio.ReadJSONList[T](file) {
		if err != nil {
			return err
		}
		cs.Add(cc)
	}
	return nil
}
func (cs Collection[T]) Store(s filePath) error {
	file, err := os.Create(s())
	if err != nil {
		panic(err)
	}
	defer x.ClosePrint(file)
	return cs.Write(file)
}
func (cs Collection[T]) Write(w io.Writer) error {
	var err error
	for _, cc := range cs {
		err = jsonio.Write(w, cc)
		if err != nil {
			return err
		}
	}
	return nil
}
