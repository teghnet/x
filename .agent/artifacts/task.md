# Unimatrix TUI Architecture Planning

## Current Objective
Implement the Unimatrix TUI application.

## Tasks

### Phase 1: TUI Foundation
- [/] Create `unimatrix/go.mod` as separate module
- [ ] Create `main.go` with flag parsing
- [ ] Implement Lipgloss styles (`internal/tui/styles.go`)
- [ ] Create root Bubble Tea model (`internal/tui/tui.go`)

### Phase 2: TUI Components
- [ ] Implement `tree.Model` with mock data
- [ ] Implement `preview.Model`
- [ ] Implement `header.Model` and `statusbar.Model`
- [ ] Wire up Browser view with pane switching

### Phase 3: Domain Models (Later)
### Phase 4: Local Sync (Later)
### Phase 5: External Connectors (Later)

## Completed
- [x] Analyze existing codebase and packages
- [x] Create implementation plan for TUI architecture
- [x] Update PLAN.md with architecture overview
- [x] Create /go-architect workflow
- [x] Get user approval on the plan
