# Unimatrix TUI Architecture Plan

A CLI/TUI tool for syncing files between APIs and systems (Google Drive, Obsidian vault, Notion documents, local files). Built with Star Trek Borg theming.

## Project Context

- **Module**: `github.com/teghnet/x` (Go 1.25)
- **Location**: `unimatrix/` as separate Go module with internal packages at `unimatrix/internal/`
- **Reusable packages**: `paths`, `fsio`, `file` from parent module
- **Run commands**: Use `go -C unimatrix <command>` from project root

---

## TUI Framework Options

| Framework | Pros | Cons |
|-----------|------|------|
| **Bubble Tea** | Elm Architecture, excellent ecosystem (Lipgloss, Bubbles), most popular | Learning curve for Elm pattern |
| **tview** | Rich widgets (tables, forms, lists), simpler imperative API | Less composable than Bubble Tea |
| **gocui** | Minimalist, low-level control | More boilerplate, fewer built-in widgets |
| **termdash** | Great for dashboards with charts | Less suited for interactive UIs |
| **pterm** | Beautiful console output, progress bars | Not a full TUI framework |

**Recommendation**: Bubble Tea + Lipgloss for the modern composable architecture and Borg-themed styling.

---

## TUI Design

### Layout Structure

```
┌─────────────────────────────────────────────────────────────────────┐
│ [HEADER] Unimatrix Zero • Status: Connected • Last Sync: 2m ago     │
├─────────────────────────────────────────────┬───────────────────────┤
│                                             │                       │
│   [TREE PANE]                               │   [PREVIEW PANE]      │
│                                             │                       │
│   ▼ Local/Documents                         │   README.md           │
│     ├── projects/                           │   ───────────────     │
│     │   ├── unimatrix/                      │   # Unimatrix         │
│     │   │   └── README.md  ↔                │                       │
│     │   └── notes.md       →                │   A sync tool for...  │
│     └── archive/                            │                       │
│   ▶ Notion/Workspace                        │                       │
│   ▶ Google Drive/Shared                     │                       │
│                                             │                       │
├─────────────────────────────────────────────┴───────────────────────┤
│ [STATUS BAR] j/k:navigate • Enter:expand • s:sync • ?:help          │
└─────────────────────────────────────────────────────────────────────┘
```

### Panes & Components

| Component | Description | Bubble Tea Model |
|-----------|-------------|------------------|
| **Header** | App title, connection status, last sync time | `header.Model` |
| **Tree Pane** | Hierarchical file browser with sync indicators | `tree.Model` |
| **Preview Pane** | File content preview, diff view for conflicts | `preview.Model` |
| **Status Bar** | Key bindings, current action, progress | `statusbar.Model` |
| **Dialog** | Modal for confirmations, inputs | `dialog.Model` |
| **Help Overlay** | Full keybinding reference | `help.Model` |

### Sync Status Indicators

| Icon | Meaning |
|------|---------|
| `↔` | Bidirectional sync (in sync) |
| `→` | Pending upload to target |
| `←` | Pending download from source |
| `⚠` | Conflict detected |
| `✓` | Successfully synced |
| `○` | Not linked |

### Views (Full-Screen Modes)

1. **Browser View** - Default, tree + preview layout
2. **Sync View** - Progress during sync operation
3. **Diff View** - Side-by-side conflict resolution
4. **Settings View** - Profile/connector configuration

### Key Bindings

| Key | Action |
|-----|--------|
| `j/k` or `↓/↑` | Navigate up/down |
| `h/l` or `←/→` | Collapse/expand or switch pane |
| `Enter` | Toggle expand / Open file |
| `Tab` | Switch between panes |
| `s` | Sync selected item |
| `S` | Sync all |
| `d` | Show diff for conflicts |
| `r` | Refresh tree |
| `p` | Toggle preview pane |
| `?` | Show help |
| `q` | Quit |

---

## Directory Structure

```
unimatrix/
├── go.mod                  # Separate module
├── main.go                 # Entry point (no Cobra, simple flag parsing)
├── PLAN.md
└── internal/
    ├── app/                # Application orchestration
    │   ├── app.go          # App struct, lifecycle
    │   └── config.go       # Config loading/saving
    ├── tui/                # Bubble Tea layer
    │   ├── tui.go          # Root model, tea.Program
    │   ├── styles.go       # Lipgloss Borg theme
    │   ├── keys.go         # Key bindings
    │   ├── components/
    │   │   ├── header.go
    │   │   ├── tree.go
    │   │   ├── preview.go
    │   │   ├── statusbar.go
    │   │   ├── dialog.go
    │   │   └── help.go
    │   └── views/
    │       ├── browser.go
    │       ├── sync.go
    │       ├── diff.go
    │       └── settings.go
    ├── sync/               # Core sync engine
    │   ├── engine.go
    │   ├── diff.go
    │   ├── conflict.go
    │   └── strategy.go
    ├── connector/          # Storage connectors
    │   ├── connector.go    # Interface
    │   ├── local.go
    │   ├── notion.go
    │   ├── gdrive.go
    │   └── obsidian.go
    └── model/              # Domain models
        ├── node.go
        ├── link.go
        └── profile.go
```

---

## Implementation Phases

#### Phase 1: TUI Foundation
1. Create `unimatrix/go.mod` as separate module
2. Create `main.go` with flag parsing (no Cobra)
3. Implement Lipgloss styles (`internal/tui/styles.go`)
4. Create root Bubble Tea model (`internal/tui/tui.go`)

#### Phase 2: TUI Components
1. Implement `tree.Model` with mock data
2. Implement `preview.Model`
3. Implement `header.Model` and `statusbar.Model`
4. Wire up Browser view with pane switching

#### Phase 3: Domain Models
1. Create `model.Node`, `model.Link`, `model.Profile`
2. Create `internal/app/` with config management
3. Connect TUI to app layer

#### Phase 4: Local Sync
1. Implement `LocalConnector`
2. Implement `sync.Engine` with diff logic
3. Wire sync operations to TUI

#### Phase 5: External Connectors
1. Notion connector
2. Google Drive connector
3. Obsidian vault support

---

## Dependencies

| Package | Purpose |
|---------|---------|
| `github.com/charmbracelet/bubbletea` | TUI framework |
| `github.com/charmbracelet/lipgloss` | Styling |
| `github.com/charmbracelet/bubbles` | Common components |
| `google.golang.org/api/drive/v3` | Google Drive (Phase 5) |

---

## Verification

```bash
# Build check from project root
go -C unimatrix build .

# Run tests
go -C unimatrix test ./internal/...

# Run the application
go -C unimatrix run . --help
go -C unimatrix run . --profile zero
```

---

## User Review Required

> [!IMPORTANT]
> Please confirm:
> 1. **Framework choice**: Bubble Tea + Lipgloss acceptable, or prefer tview/gocui?
> 2. **Layout design**: Is the tree + preview pane layout correct?
> 3. **Phase order**: Starting with TUI (Phases 1-2), then domain/sync (3-4)?
