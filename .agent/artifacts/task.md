# Unimatrix TUI - Complete ✓

All 5 phases implemented:

| Phase | Package | Status |
|-------|---------|--------|
| 1-2 | `internal/tui/` | ✓ Bubble Tea UI |
| 3 | `internal/model/`, `internal/app/` | ✓ Domain models |
| 4 | `internal/connector/local.go`, `internal/sync/` | ✓ Local sync |
| 5 | `notion.go`, `obsidian.go`, `gdrive.go` | ✓ External connectors |

## Run
```bash
go -C unimatrix run .
```

## Connectors
- **Local**: Filesystem operations
- **Notion**: `NOTION_API_TOKEN` env var
- **Google Drive**: `GDRIVE_ACCESS_TOKEN` env var
- **Obsidian**: Vault with frontmatter + backlinks
