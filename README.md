# GosQLit

Terminal UI SQL client built with Bubble Tea. Multi-database support, encrypted credential storage.

## Features

- **Multi-database**: PostgreSQL (MySQL, SQLite coming)
- **Encrypted storage**: AES-256-GCM with master password
- **Schema browser**: Tree view with schemas → tables
- **Query editor**: Multi-line SQL editor
- **Results**: Paginated tables (1000 row limit, 50/page)
- **Query control**: Execute (Ctrl+Enter), cancel (Ctrl+K)
- **Connections**: Save/edit/delete, multiple connections

## Install

```bash
go build -o gosqlit
./gosqlit
```

## Usage

1. Set master password on first run
2. Add connection (press `n` in explorer)
3. Select connection (Enter)
4. Browse schemas (left panel, ↑↓←→)
5. Click table → auto-generates SELECT
6. Edit query, execute (Ctrl+Enter)

## Keyboard shortcuts

**Explorer:**
- `n` - New connection
- `e` - Edit connection
- `d` - Delete connection
- `Enter` - Connect
- `q` / `Ctrl+C` - Quit

**Connected view:**
- `Tab` - Cycle focus (editor → results → browser)
- `Ctrl+Enter` - Execute query
- `Ctrl+K` - Cancel running query
- `Ctrl+W` - Close tab
- `F5` - Refresh schemas

## Config

Encrypted config stored at `~/.gosqlit/config.encrypted`

Master password never saved to disk.

## Requirements

- Go 1.21+
- PostgreSQL database (for testing)
