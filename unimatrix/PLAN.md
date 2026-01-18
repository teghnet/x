# Application Plan

A CLI tool for syncing files between APIs and systems (Google Drive, Obsidian vault, Notion documents, local files).

- It is a separate Go module inside teghnet/x repo.
- The main package should be in `unimatrix` directory.
- All other packages should be in `unimatrix/internal` directory.

## Naming: `unimatrix`

**Theme:** Star Trek Borg. A Unimatrix is an organizational structure in the Borg Collective that coordinates thousands of vessels and billions of drones in perfect synchronization — fitting for a tool that keeps disparate file systems in sync.

**Naming Conventions to Follow:**
- Commands and terminology should lean into Borg vocabulary
- Examples: `sync`, `link`, `sever`, `status`, `assimilate`
- Profiles could be numbered like Unimatrices (e.g., `--zero`, `--one`)
- `unimatrix-zero` was a special resistance group in Trek — could be reserved for a primary/default profile
- Error/status messages can reference Borg phrases ("Resistance is futile", "Collective link established", etc.) but keep it functional first

**Short Aliases:** `uni` or `umx`

**Vibe:** Nerdy, playful, but still a professional-feeling tool.

## Architecture Overview

See [detailed implementation plan](../.agent/artifacts/implementation_plan.md) for full architecture.

### Package Structure
```
unimatrix/
├── go.mod                  # Separate module
├── main.go                 # Entry point (flag parsing, no Cobra)
├── PLAN.md
└── internal/
    ├── app/                # Application orchestration
    ├── tui/                # Bubble Tea TUI layer
    │   ├── components/     # Reusable UI components
    │   └── views/          # Full-screen views
    ├── sync/               # Sync engine (diff, conflict, strategy)
    ├── connector/          # Storage connectors interface
    └── model/              # Domain models (Node, Link, Profile)
```

### Key Components
- **App**: Lifecycle management, config via `paths.AppConfig("unimatrix")`
- **TUI**: Bubble Tea + Lipgloss with Borg styling
- **Sync Engine**: Decoupled sync logic with OneWay/BiDirectional/Mirror strategies
- **Connectors**: Abstract interface for Local, Notion, Google Drive, Obsidian

### Implementation Phases
1. **TUI Foundation** - Lipgloss styles, root Bubble Tea model
2. **TUI Components** - Tree, preview, header, statusbar panes
3. **Domain Models** - Node, Link, Profile; app layer
4. **Local Sync** - LocalConnector + sync engine
5. **External Connectors** - Notion, Google Drive, Obsidian