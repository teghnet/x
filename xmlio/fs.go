// Copyright (c) 2026 Paweł Zaremba
// SPDX-License-Identifier: MIT

package xmlio

import (
	"io/fs"
	"iter"

	"github.com/teghnet/x"
)

func XML[T any](fsfs fs.FS, name string) (T, error) {
	f, err := fsfs.Open(name)
	if err != nil {
		return *new(T), err
	}
	return ReadXML[T](f)
}

func XMLs[T any](fsfs fs.FS, name, elementName string) iter.Seq2[T, error] {
	return func(yield func(T, error) bool) {
		f, err := fsfs.Open(name)
		if err != nil {
			yield(*new(T), err)
			return
		}
		defer x.ClosePrint(f)
		ReadXMLs[T](f, elementName)(yield)
	}
}
