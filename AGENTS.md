# AGENTS.md

## What this repo is

Dataset of Saizeriya menu items. The source of truth is the SQLite dump at `db/saizeriya.sql`. Output files (`saizeriya.csv`, `saizeriya.json`, `saizeriya.md`) are auto-generated — never edit them directly.

## Critical workflow

1. To change data: edit `db/saizeriya.sql` (or modify `db/saizeriya.db` then run `make dump`)
2. Output files regenerate automatically via CI on push when `db/**` or `tools/converter/**` change
3. To regenerate locally: restore DB first, then run the converter

## Commands

```bash
# Restore SQLite DB from dump (requires sqlite3)
make restore

# Dump current DB to SQL file
make dump

# Regenerate CSV/JSON/MD output files (requires Go 1.21+)
# Must run from tools/converter/ — it uses relative paths to ../../db/saizeriya.db
cd tools/converter && go run cmd/main.go --csv --json --md
```

## Directory layout

- `db/` — SQLite dump (`saizeriya.sql`); `.db` files are gitignored
- `tools/converter/` — Go program that reads DB and writes CSV/JSON/MD to repo root
- `tools/generate-menu-sql/` — Go program that generates SQL inserts from JSON
- `docs/database.md` — DB schema (4 tables: categories, genres, icons, menus)

## Release

Uses release-please (`simple` release type). Commits following Conventional Commits (`feat:`, `fix:`) trigger version bumps. Pre-1.0: `feat` bumps patch, `BREAKING CHANGE` bumps minor.

## Gotchas

- The converter uses hardcoded relative paths (`../../db/saizeriya.db`). Always run it from `tools/converter/`.
- `db/saizeriya.db` is gitignored. Run `make restore` before running the converter locally.
- Prices in the DB are in JPY (integer). `salt` is stored as `REAL` but displayed with one decimal.
