package command

import (
	"flag"

	"github.com/teghnet/x"
	"github.com/teghnet/x/osio"
)

func XMLPassthrough(args []string) error {
	var asHtml *bool
	i, o := defaultIO("XMLPassthrough", args, func(f *flag.FlagSet) {
		asHtml = f.Bool("html", false, "as HTML")
	})

	r, err := osio.DynamicReader(i)
	if err != nil {
		return err
	}
	defer x.ClosePrint(r)

	w, err := osio.DynamicWriter(o, false)
	if err != nil {
		return err
	}
	defer x.ClosePrint(w)

	return osio.TrimXML(r, w, *asHtml)
}
