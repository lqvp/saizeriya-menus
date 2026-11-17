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

# CI と同じ手順で同期する。
sync:
	cd tools/sync-menu && go run cmd/main.go -db ../../$(DB_NAME)

# restore → sync → dump を通し、差分が出ないことを確認する。
#
# 前提ターゲットに分けると `make -j` で並行実行され、順序が壊れる
# （dump が restore より先に走るなど）。1 つのレシピにまとめて
# 必ずこの順で実行されるようにしている。
check:
	$(MAKE) restore
	$(MAKE) sync
	$(MAKE) dump
	@echo "--- 差分（空なら冪等） ---"
	@git diff --stat -- $(SQL_NAME) || true
