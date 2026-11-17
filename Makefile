DB_NAME=db/saizeriya.db
SQL_NAME=db/saizeriya.sql

.PHONY: restore dump sync check

# 既存 DB を消してから作り直す。
# 途中で失敗したら壊れた DB を残さないよう、成功時だけ置き換える。
restore:
	@rm -f $(DB_NAME).tmp
	@sqlite3 $(DB_NAME).tmp < $(SQL_NAME)
	@mv $(DB_NAME).tmp $(DB_NAME)
	@echo "$(DB_NAME) を復元しました"

dump:
	sqlite3 $(DB_NAME) ".output $(SQL_NAME)" ".dump"
	python3 tools/normalize-dump.py $(SQL_NAME)

# CI と同じ手順で同期し、差分が出ないことを確認する。
# 差分が出たら「コミットすべき変更がある」か「正規化が効いていない」。
sync:
	cd tools/sync-menu && go run cmd/main.go -db ../../$(DB_NAME)

check: restore sync dump
	@git diff --stat -- $(SQL_NAME) || true
