// Copyright (c) 2026 Paweł Zaremba
// SPDX-License-Identifier: MIT

package paths

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestNewXDG_defaults(t *testing.T) {
	t.Chdir(t.TempDir())

	x := NewXDG("testapp")
	if got, want := x.ConfigPath(), AppConfig("testapp"); got != want {
		t.Errorf("ConfigPath() = %q, want %q", got, want)
	}
	if got, want := x.CachePath(), AppCache("testapp"); got != want {
		t.Errorf("CachePath() = %q, want %q", got, want)
	}
	if got, want := x.DataPath(), AppData("testapp"); got != want {
		t.Errorf("DataPath() = %q, want %q", got, want)
	}
	if got, want := x.StatePath(), AppState("testapp"); got != want {
		t.Errorf("StatePath() = %q, want %q", got, want)
	}
	if got, want := x.RuntimePath(), AppRuntime("testapp"); got != want {
		t.Errorf("RuntimePath() = %q, want %q", got, want)
	}
}

func TestNewXDG_preferWDStore(t *testing.T) {
	wd := t.TempDir()
	t.Chdir(wd)

	x := NewXDG("testapp", WithPreferWDStore(true))

	for _, tt := range []struct {
		name string
		got  string
		dir  string
	}{
		{"ConfigPath", x.ConfigPath(), dirConfig},
		{"CachePath", x.CachePath(), dirCache},
		{"DataPath", x.DataPath(), dirData},
		{"StatePath", x.StatePath(), dirState},
		{"RuntimePath", x.RuntimePath(), dirRuntime},
	} {
		want := filepath.Join(wd, "."+tt.dir)
		if tt.got != want {
			t.Errorf("%s() = %q, want %q", tt.name, tt.got, want)
		}
		if !isDir(want) {
			t.Errorf("expected %q to have been created", want)
		}
	}
}

func TestNewXDG_preferDotLocalStore(t *testing.T) {
	wd := t.TempDir()
	t.Chdir(wd)

	x := NewXDG("testapp", WithPreferDotLocalStore(true))

	root := filepath.Join(wd, ".testapp")
	for _, tt := range []struct {
		name string
		got  string
		dir  string
	}{
		{"ConfigPath", x.ConfigPath(), dirConfig},
		{"CachePath", x.CachePath(), dirCache},
		{"DataPath", x.DataPath(), dirData},
		{"StatePath", x.StatePath(), dirState},
		{"RuntimePath", x.RuntimePath(), dirRuntime},
	} {
		want := filepath.Join(root, tt.dir)
		if tt.got != want {
			t.Errorf("%s() = %q, want %q", tt.name, tt.got, want)
		}
		if !isDir(want) {
			t.Errorf("expected %q to have been created", want)
		}
	}
}

func TestNewXDG_preferWDStoreWinsOverDotLocal(t *testing.T) {
	wd := t.TempDir()
	t.Chdir(wd)

	x := NewXDG("testapp", WithPreferWDStore(true), WithPreferDotLocalStore(true))

	want := filepath.Join(wd, ".config")
	if got := x.ConfigPath(); got != want {
		t.Errorf("ConfigPath() = %q, want %q (WithPreferWDStore should win)", got, want)
	}
	if isDir(filepath.Join(wd, ".testapp")) {
		t.Errorf("expected .testapp not to have been created when both options are set")
	}
}

func TestXDG_envOverrides(t *testing.T) {
	t.Chdir(t.TempDir())

	dataHome := t.TempDir()
	stateHome := t.TempDir()
	t.Setenv("XDG_DATA_HOME", dataHome)
	t.Setenv("XDG_STATE_HOME", stateHome)

	x := NewXDG("testapp")
	if got, want := x.DataPath(), filepath.Join(dataHome, "testapp"); got != want {
		t.Errorf("DataPath() = %q, want %q", got, want)
	}
	if got, want := x.StatePath(), filepath.Join(stateHome, "testapp"); got != want {
		t.Errorf("StatePath() = %q, want %q", got, want)
	}
}

func TestAppData_relativeXDGDataHomePanics(t *testing.T) {
	t.Chdir(t.TempDir())
	t.Setenv("XDG_DATA_HOME", "relative/path")

	defer func() {
		if recover() == nil {
			t.Errorf("expected panic for relative XDG_DATA_HOME")
		}
	}()
	AppData("testapp")
}

func TestAppState_relativeXDGStateHomePanics(t *testing.T) {
	t.Chdir(t.TempDir())
	t.Setenv("XDG_STATE_HOME", "relative/path")

	defer func() {
		if recover() == nil {
			t.Errorf("expected panic for relative XDG_STATE_HOME")
		}
	}()
	AppState("testapp")
}

func TestAppRuntime_envOverride(t *testing.T) {
	t.Chdir(t.TempDir())

	runtimeHome := t.TempDir()
	t.Setenv("XDG_RUNTIME_DIR", runtimeHome)

	if got, want := AppRuntime("testapp"), filepath.Join(runtimeHome, "testapp"); got != want {
		t.Errorf("AppRuntime() = %q, want %q", got, want)
	}
}

func TestAppRuntime_relativeXDGRuntimeDirPanics(t *testing.T) {
	t.Chdir(t.TempDir())
	t.Setenv("XDG_RUNTIME_DIR", "relative/path")

	defer func() {
		if recover() == nil {
			t.Errorf("expected panic for relative XDG_RUNTIME_DIR")
		}
	}()
	AppRuntime("testapp")
}

func TestAppRuntime_fallsBackWhenUnset(t *testing.T) {
	t.Chdir(t.TempDir())
	t.Setenv("XDG_RUNTIME_DIR", "")

	got := AppRuntime("testapp")
	if !filepath.IsAbs(got) {
		t.Errorf("AppRuntime() = %q, want an absolute path", got)
	}
	if !strings.HasSuffix(got, string(filepath.Separator)+"testapp") {
		t.Errorf("AppRuntime() = %q, want it to end with /testapp", got)
	}
}
