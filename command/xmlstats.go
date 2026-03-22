package command

import (
	"maps"
	"slices"
	"strings"

	"github.com/teghnet/x"
	"github.com/teghnet/x/osio"
)

func XMLStats(args []string) error {
	i, o := defaultIO("XMLStats", args)

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

	counter := make(map[string]int64)
	dicts := make(map[string][]string)
	for k, v := range osio.XMLDicts(r) {
		if k == "" {
			continue
		}
		if counter[k+":"+v] == 0 {
			dicts[k] = append(dicts[k], v)
		}
		counter[k+":"+v]++
	}
	fPrint(w, "pole;wartość;licznik\n")
	keys := slices.Collect(maps.Keys(dicts))
	slices.Sort(keys)
	for _, k := range keys {
		fPrint(w, k, "\n")
	}
	// reDate := regexp.MustCompile(`^\d{4}-\d{1,2}-\d{1,2}$`)
	for k, vals := range dicts {
		if strings.TrimSpace(k) == "" {
			continue
		}
		fPrintf(w, "%s;;%d\n", k, len(vals))
		for _, v := range vals {
			// if _, err := strconv.ParseFloat(v, 64); err == nil || reDate.MatchString(v) {
			// 	continue
			// }
			fPrintf(w, "%s;%s;%d\n", k, v, counter[k+":"+v])
		}
	}
	return nil
}
