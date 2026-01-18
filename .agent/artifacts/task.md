# Unimatrix TUI Implementation

## Phase 1: TUI Foundation ✓
- [x] Create `unimatrix/go.mod` as separate module
- [x] Create `main.go` with flag parsing
- [x] Implement Lipgloss styles (`internal/tui/styles.go`)
- [x] Create root Bubble Tea model (`internal/tui/tui.go`)

## Phase 2: TUI Components ✓
- [x] Implement `tree.Model` with mock data
- [x] Implement `preview.Model`
- [x] Implement `header.Model` and `statusbar.Model`
- [x] Wire up Browser view with pane switching

## Next: Phase 3: Domain Models
- [ ] Create `model.Node`, `model.Link`, `model.Profile`
- [ ] Create `internal/app/` with config management
- [ ] Connect TUI to app layer
