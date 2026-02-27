// Copyright (c) 2024-2026 Paweł Zaremba
// SPDX-License-Identifier: MIT

package paths_test

import (
	"strings"
	"testing"

	"github.com/teghnet/x/paths"
)

func TestPaths_App(t *testing.T) {
	tests := []struct {
		n string
		f func(app string) string
	}{
		{"paths.AppConfig", paths.AppConfig},
		{"paths.AppCache", paths.AppCache},
		{"paths.AppState", paths.AppState},
		{"paths.AppData", paths.AppData},
	}
	for _, tt := range tests {
		t.Run(tt.n, func(t *testing.T) {
			p := tt.f("testapp")
			if !strings.HasSuffix(p, "/testapp") {
				t.Errorf("expected path to end with `testapp` name, got: %s", p)
				return
			}
			t.Log(p)
		})
	}
	for _, tt := range tests {
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

func TestPaths_Profile(t *testing.T) {
	tests := []struct {
		n string
		f func(app, profile string) string
	}{
		{"paths.ProfileConfig", paths.ProfileConfig},
		{"paths.ProfileCache", paths.ProfileCache},
		{"paths.ProfileState", paths.ProfileState},
		{"paths.ProfileData", paths.ProfileData},
	}
	for _, tt := range tests {
		t.Run(tt.n, func(t *testing.T) {
			p := tt.f("testapp", "work")
			if !strings.HasSuffix(p, "/work") {
				t.Errorf("expected path to end with profile name `work`, got: %s", p)
				return
			}
			if !strings.Contains(p, "profiles") {
				t.Errorf("expected path to contain `profiles`, got: %s", p)
				return
			}
			t.Log(p)
		})
	}

	// Test panic on empty appName
	for _, tt := range tests {
		t.Run(tt.n+" empty app name", func(t *testing.T) {
			defer func() {
				if recover() == nil {
					t.Errorf("expected panic for empty appName")
				}
			}()
			tt.f("", "work")
		})
	}

	// Test panic on empty profileName
	for _, tt := range tests {
		t.Run(tt.n+" empty profile name", func(t *testing.T) {
			defer func() {
				if recover() == nil {
					t.Errorf("expected panic for empty profileName")
				}
			}()
			tt.f("testapp", "")
		})
	}
}
