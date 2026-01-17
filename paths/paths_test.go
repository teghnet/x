package paths_test

import (
	"strings"
	"testing"

	"github.com/teghnet/x/paths"
)

func TestPaths_App(t *testing.T) {
	testCases := []struct {
		n string
		f func(app string) string
	}{
		{"paths.AppConfig", paths.AppConfig},
		{"paths.AppCache", paths.AppCache},
		{"paths.AppState", paths.AppState},
		{"paths.AppData", paths.AppData},
	}
	for _, tt := range testCases {
		t.Run(tt.n, func(t *testing.T) {
			p := tt.f("testapp")
			if !strings.HasSuffix(p, "/testapp") {
				t.Errorf("expected path to end with `testapp` name, got: %s", p)
				return
			}
			t.Log(p)
		})
	}
	for _, tt := range testCases {
		t.Run(tt.n+" empty app name", func(t *testing.T) {
			defer func() {
				if recover() == nil {
					t.Errorf("expected panic")
				}
			}()
			tt.f("")
		})
	}
}
