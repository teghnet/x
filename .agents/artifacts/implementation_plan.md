# Profile-Aware Path Functions

Add profile support to the paths package, allowing apps to organize their config/cache/data/state directories per profile.

## Design

### Directory Structure

Given `appName="myapp"` and `profileName="work"`:

| Function | Path |
|----------|------|
| `ProfileConfig` | `~/.config/myapp/profiles/work` |
| `ProfileCache` | `~/.cache/myapp/profiles/work` |
| `ProfileData` | `~/.local/share/myapp/profiles/work` |
| `ProfileState` | `~/.local/state/myapp/profiles/work` |

For local development (`.local/` or `.myapp/` exists in cwd):

| Function | Path |
|----------|------|
| `ProfileConfig` | `.local/<profileName>` or `.myapp/<profileName>` |
| `ProfileCache` | `.local/<profileName>/cache` or `.myapp/<profileName>/cache` |
| `ProfileData` | `.local/<profileName>/data` or `.myapp/<profileName>/data` |
| `ProfileState` | `.local/<profileName>/state` or `.myapp/<profileName>/state` |

### Function Signatures

```go
func ProfileConfig(appName, profileName string) string
func ProfileCache(appName, profileName string) string
func ProfileData(appName, profileName string) string
func ProfileState(appName, profileName string) string
```

Both `appName` and `profileName` must be non-empty (panic otherwise).

---

## Proposed Changes

### [MODIFY] [app.go](file:///home/box/Projects/teghnet/x/paths/app.go)

Add four new functions with custom local path handling:

```go
func ProfileConfig(appName, profileName string) string {
    if appName == "" {
        panic("appName must be non-empty")
    }
    if profileName == "" {
        panic("profileName must be non-empty")
    }

    if wd, err := os.Getwd(); err == nil && !wdIsHome() {
        // Local dev: .local/<profileName>
        dir := filepath.Join(wd, ".local", profileName)
        if info, err := os.Stat(dir); err == nil && info.IsDir() {
            return dir
        }
        // Local dev: .<appName>/<profileName>
        dir = filepath.Join(wd, "."+appName, profileName)
        if info, err := os.Stat(dir); err == nil && info.IsDir() {
            return dir
        }
    }

    // System: ~/.config/<appName>/profiles/<profileName>
    return filepath.Join(AppConfig(appName), "profiles", profileName)
}

func ProfileCache(appName, profileName string) string {
    if appName == "" {
        panic("appName must be non-empty")
    }
    if profileName == "" {
        panic("profileName must be non-empty")
    }

    if wd, err := os.Getwd(); err == nil && !wdIsHome() {
        dir := filepath.Join(wd, ".local", profileName, "cache")
        if info, err := os.Stat(dir); err == nil && info.IsDir() {
            return dir
        }
        dir = filepath.Join(wd, "."+appName, profileName, "cache")
        if info, err := os.Stat(dir); err == nil && info.IsDir() {
            return dir
        }
    }

    return filepath.Join(AppCache(appName), "profiles", profileName)
}

func ProfileData(appName, profileName string) string {
    if appName == "" {
        panic("appName must be non-empty")
    }
    if profileName == "" {
        panic("profileName must be non-empty")
    }

    if wd, err := os.Getwd(); err == nil && !wdIsHome() {
        dir := filepath.Join(wd, ".local", profileName, "data")
        if info, err := os.Stat(dir); err == nil && info.IsDir() {
            return dir
        }
        dir = filepath.Join(wd, "."+appName, profileName, "data")
        if info, err := os.Stat(dir); err == nil && info.IsDir() {
            return dir
        }
    }

    return filepath.Join(AppData(appName), "profiles", profileName)
}

func ProfileState(appName, profileName string) string {
    if appName == "" {
        panic("appName must be non-empty")
    }
    if profileName == "" {
        panic("profileName must be non-empty")
    }

    if wd, err := os.Getwd(); err == nil && !wdIsHome() {
        dir := filepath.Join(wd, ".local", profileName, "state")
        if info, err := os.Stat(dir); err == nil && info.IsDir() {
            return dir
        }
        dir = filepath.Join(wd, "."+appName, profileName, "state")
        if info, err := os.Stat(dir); err == nil && info.IsDir() {
            return dir
        }
    }

    return filepath.Join(AppState(appName), "profiles", profileName)
}
```

### [MODIFY] [paths_test.go](file:///home/box/Projects/teghnet/x/paths/paths_test.go)

Add tests for Profile* functions:
- Verify path ends with `profiles/<profileName>`
- Verify panic on empty profileName
- Verify panic on empty appName

---

## Verification Plan

### Automated Tests

```bash
go test -v ./paths/...
```

Tests will verify:
1. Correct path structure: `<appDir>/profiles/<profileName>`
2. Panic on empty `appName`
3. Panic on empty `profileName`
