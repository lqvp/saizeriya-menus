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

# Dump current DB to SQL file and normalize it (see "Gotchas")
make dump

# Run the full CI-equivalent pipeline and show any diff
make check

# Regenerate CSV/JSON/MD output files (requires Go 1.21+)
# Must run from tools/converter/ — it uses relative paths to ../../db/saizeriya.db
cd tools/converter && go run cmd/main.go --csv --json --md

# Tests
cd tools/sync-menu && go test ./...
cd tools && python3 -m unittest normalize-dump_test
```

## Directory layout

- `db/` — SQLite dump (`saizeriya.sql`); `.db` files are gitignored
- `tools/converter/` — Go program that reads DB and writes CSV/JSON/MD to repo root
- `tools/generate-menu-sql/` — Go program that generates SQL inserts from JSON
- `tools/sync-menu/` — Fetches prices/nutrition from the official allergy API and
  updates existing rows. Never adds or removes rows: the API omits alcohol and
  other items, so regenerating from it would drop ~80 rows.
- `docs/database.md` — DB schema (4 tables: categories, genres, icons, menus)

## Release

Uses release-please (`simple` release type). Commits following Conventional Commits (`feat:`, `fix:`) trigger version bumps. Pre-1.0: `feat` bumps patch, `BREAKING CHANGE` bumps minor.

## Automated updates

`.github/workflows/update.yml` runs `tools/sync-menu` every Monday 03:00 JST,
then regenerates the output files and commits them. It needs
`permissions: contents: write` to push.

## Gotchas

- The converter uses hardcoded relative paths (`../../db/saizeriya.db`). Always run it from `tools/converter/`.
- `db/saizeriya.db` is gitignored. Run `make restore` before running the converter locally.
- Prices in the DB are in JPY (integer). `salt` is stored as `REAL` but displayed with one decimal.
- **Always normalize a fresh dump with `tools/normalize-dump.py`.** `sqlite3` 3.46 writes
  `1.39999999999999991` where 3.54 writes `1.4`, and 3.54 adds `PRAGMA writable_schema`
  and a duplicate `sqlite_sequence` block. Without normalization, an unchanged dataset
  produces a diff and the weekly job commits noise. `make dump` does this for you.
- `menus.pre_id` is **not** a unique product ID. It is a SKU code for most rows but a
  *pair key* for the 32xx/39xx desserts, and `AA06` is shared by two different products
  (1406, 1425). Any `pre_id`-based logic must also match `category_id` and normalized `name`.
- `sync-menu` never adds or removes rows and refuses to write when the match rate drops
  below `-min-match-rate` (default 50%). A sudden drop means the API changed.
- Ambiguous matches (several API items for one DB row) are **skipped and reported**,
  never assigned positionally — response order is not guaranteed.
