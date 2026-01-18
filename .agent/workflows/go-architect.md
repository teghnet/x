---
description: Plan and architect a Go TUI application
---

# Go TUI Architect Workflow

Use this workflow when planning a new TUI application or major architectural changes.

## Steps

1. **Analyze Existing Codebase**
   - Review the project structure with `list_dir`
   - Check for existing packages that can be reused
   - Review `go.mod` for module name and Go version

2. **Create Task Checklist**
   - Create `.agent/artifacts/task.md` with planning checklist
   - Mark items as `[/]` in progress, `[x]` completed

3. **Write Implementation Plan**
   - Create `.agent/artifacts/implementation_plan.md` with:
     - Project context and goals
     - Directory structure diagram
     - Core component designs (with Go interface signatures)
     - Domain models
     - Dependency list
     - Implementation phases
     - Verification plan
     - User review items

4. **Update Project PLAN.md**
   - Add architecture overview section
   - Link to detailed implementation plan
   - Include package structure summary

5. **Request User Review**
   - Use `notify_user` with `PathsToReview` pointing to:
     - Implementation plan
     - Updated PLAN.md
   - Set `BlockedOnUser: true` to wait for approval

## TUI Architecture Patterns

When designing Go TUI applications, follow these patterns:

### Directory Structure
```
<app>/
├── cmd/<app>/main.go        # Entry point
├── internal/
│   ├── app/                 # Application orchestration
│   ├── tui/                 # Bubble Tea layer
│   │   ├── tui.go           # Root model
│   │   ├── styles.go        # Lipgloss styles
│   │   ├── keys.go          # Key bindings
│   │   ├── components/      # Reusable UI pieces
│   │   └── views/           # Full-screen modes
│   ├── <domain>/            # Core business logic
│   └── model/               # Domain types
```

### Recommended Libraries
- **TUI Framework**: `github.com/charmbracelet/bubbletea`
- **Styling**: `github.com/charmbracelet/lipgloss`
- **Components**: `github.com/charmbracelet/bubbles`
- **CLI Commands**: `github.com/spf13/cobra`
- **Configuration**: `github.com/spf13/viper`

### Design Principles
- Separate UI from business logic
- Use interfaces for external dependencies
- Inject configuration, don't derive internally
- Use events/messages for component communication
