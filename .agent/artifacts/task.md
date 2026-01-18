# Unimatrix TUI Implementation

## Completed ✓

### Phase 1: TUI Foundation
- [x] `go.mod` - Separate module with Bubble Tea dependencies
- [x] `main.go` - Entry point with flag parsing
- [x] `internal/tui/styles.go` - Borg-themed Lipgloss styles
- [x] `internal/tui/tui.go` - Root Bubble Tea model

### Phase 2: TUI Components
- [x] `components/tree.go` - Hierarchical file browser
- [x] `components/preview.go` - File preview pane
- [x] `components/header.go` - Title bar
- [x] `components/statusbar.go` - Key bindings

### Phase 3: Domain Models
- [x] `internal/model/node.go` - File/folder node
- [x] `internal/model/link.go` - Sync link with strategies
- [x] `internal/model/profile.go` - Sync profiles
- [x] `internal/app/app.go` - Config management

### Phase 4: Local Sync
- [x] `internal/connector/connector.go` - Interface
- [x] `internal/connector/local.go` - Local filesystem
- [x] `internal/sync/engine.go` - Sync engine

## Next: Phase 5
- [ ] Notion connector
- [ ] Google Drive connector
- [ ] Obsidian vault support

## Run
```bash
go -C unimatrix run .
```
