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
func (cs *Collection[T]) Load(s string) {
	file, err := os.Open(s)
	if err != nil {
		panic(err)
	}
	defer x.ClosePrint(file)
	cs.Read(file)
}
func (cs *Collection[T]) Read(file *os.File) {
	for cc, err := range jsonio.ReadJSONList[T](file) {
		if err != nil {
			log.Warn(err)
			continue
		}
		cs.Add(cc)
	}
}
func (cs Collection[T]) Store(s string) {
	file, err := os.Create(s)
	if err != nil {
		panic(err)
	}
	defer x.ClosePrint(file)
	cs.Write(file)
}
func (cs Collection[T]) Write(w io.Writer) {
	var err error
	for _, cc := range cs {
		err = jsonio.WritePrettyJSON(w, cc)
		if err != nil {
			log.Warn(err)
			continue
		}
	}
}
