package xmlio

import (
	"encoding/xml"
	"io"
	"iter"

	"github.com/teghnet/x/osio"
)

func ReadXML[T any](r io.Reader) (T, error) {
	var v T
	return v, xml.NewDecoder(r).Decode(&v)
}
func ReadXMLs[T any](r io.Reader, elementName string) iter.Seq2[T, error] {
	return func(yield func(T, error) bool) {
		dec := xml.NewDecoder(r)
		for t := range osio.Tokens(dec) {
			switch el := t.(type) {
			case xml.StartElement:
				var v T
				if el.Name.Local != elementName {
					continue
				}
				if !yield(v, dec.DecodeElement(&v, &el)) {
					return
				}
			}
		}
	}
}
