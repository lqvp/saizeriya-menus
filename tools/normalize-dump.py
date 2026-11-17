#!/usr/bin/env python3
"""sqlite3 の .dump 出力から salt の表記ゆれを揃える。

sqlite3 のバージョンによって REAL の出力精度が変わる。
3.46 は 1.39999999999999991、3.54 は 1.4 と書き出すため、
環境が違うと意味のない差分が出てしまう。

menus の INSERT は 13 列で、8 番目（0 始まりで 7）が salt:
  id, name, name_en, name_zh, price, price_with_tax, calorie, salt,
  category_id, genre_id, is_alcohol, icon_id, pre_id
これを小数点 1 桁に丸め直す。

商品名にカンマが入っても列がずれないよう、SQL の文字列リテラルを
意識して分割する。列数が想定と合わない行は、静かに素通しすると
「正規化したつもりでできていない」状態になるため、エラーで止める。

併せて、sqlite3 のバージョンによって出入りするゴミも落とす:

- `PRAGMA writable_schema=ON/OFF` と `CREATE TABLE sqlite_sequence` は
  3.54 系が `.dump` に出すが 3.46 系は出さない。どちらでも復元結果は
  同じ（sqlite_sequence は SQLite が内部で作る）ため、常に落として揃える。
- `DELETE FROM sqlite_sequence` の直後に同じ行が重複して並ぶことがある
  （categories=6 と icons=31 が 2 回ずつ）。重複を畳んでおく。
"""

import re
import sys

USAGE = "使い方: normalize-dump.py <sql-file>"

# menus テーブルの列数
MENUS_COLUMNS = 13
# salt の位置（0 始まり）
SALT_INDEX = 7

# INSERT INTO menus VALUES(...); を拾う。
# 値の中に ')' や ';' が入りうるので、開始位置から閉じ括弧までを
# 素朴に切り出すのではなく、1 文字ずつクォート状態を追って探す。
INSERT_PREFIX = "INSERT INTO menus VALUES("


class NormalizeError(Exception):
    """正規化できない行に当たったときに送出する。"""


def split_sql_values(raw: str) -> list[str]:
    """VALUES(...) の中身を、クォートを意識してカンマで分割する。

    SQL の文字列リテラルは ' で囲み、内側の ' は '' で表す。
    """
    values: list[str] = []
    buf: list[str] = []
    in_quote = False
    i = 0
    while i < len(raw):
        ch = raw[i]
        if in_quote:
            buf.append(ch)
            if ch == "'":
                # '' はエスケープされた引用符（文字列は継続する）
                if i + 1 < len(raw) and raw[i + 1] == "'":
                    buf.append(raw[i + 1])
                    i += 2
                    continue
                in_quote = False
            i += 1
            continue

        if ch == "'":
            in_quote = True
            buf.append(ch)
        elif ch == ",":
            values.append("".join(buf))
            buf = []
        else:
            buf.append(ch)
        i += 1

    if in_quote:
        raise NormalizeError("文字列リテラルが閉じていません")
    values.append("".join(buf))
    return values


def find_statement_end(sql: str, start: int) -> int:
    """start から始まる VALUES(...) の閉じ括弧の位置を返す。"""
    depth = 0
    in_quote = False
    i = start
    while i < len(sql):
        ch = sql[i]
        if in_quote:
            if ch == "'":
                if i + 1 < len(sql) and sql[i + 1] == "'":
                    i += 2
                    continue
                in_quote = False
            i += 1
            continue
        if ch == "'":
            in_quote = True
        elif ch == "(":
            depth += 1
        elif ch == ")":
            depth -= 1
            if depth == 0:
                return i
        i += 1
    raise NormalizeError("VALUES(...) の閉じ括弧が見つかりません")


def format_salt(raw: str) -> str:
    """salt の値を小数点 1 桁に丸めた表記へ変換する。

    NULL や数値でない値は、そのまま返す。
    """
    text = raw.strip()
    if text.upper() == "NULL":
        return raw
    try:
        value = float(text)
    except ValueError:
        return raw
    # 1.0 -> 1, 1.4 -> 1.4
    return f"{value:.1f}".rstrip("0").rstrip(".")


def normalize_sql(sql: str) -> tuple[str, int]:
    """SQL 全体を正規化し、(変換後, 変換した行数) を返す。"""
    out: list[str] = []
    pos = 0
    converted = 0

    while True:
        start = sql.find(INSERT_PREFIX, pos)
        if start == -1:
            out.append(sql[pos:])
            break

        values_start = start + len(INSERT_PREFIX) - 1  # '(' の位置
        try:
            end = find_statement_end(sql, values_start)
        except NormalizeError as exc:
            line = sql.count("\n", 0, start) + 1
            raise NormalizeError(f"{line} 行目: {exc}") from exc

        raw_values = sql[values_start + 1 : end]
        values = split_sql_values(raw_values)
        if len(values) != MENUS_COLUMNS:
            line = sql.count("\n", 0, start) + 1
            raise NormalizeError(
                f"{line} 行目: menus の値が {len(values)} 個です"
                f"（{MENUS_COLUMNS} 個のはず）: {raw_values[:80]}..."
            )

        new_salt = format_salt(values[SALT_INDEX])
        if new_salt != values[SALT_INDEX]:
            converted += 1
        values[SALT_INDEX] = new_salt

        out.append(sql[pos : values_start + 1])
        out.append(",".join(values))
        out.append(")")
        pos = end + 1

    return drop_version_noise("".join(out)), converted


# sqlite3 のバージョン差で出入りする行。意味を持たないので常に落とす。
NOISE_PATTERNS = (
    re.compile(r"^PRAGMA writable_schema=(?:ON|OFF);\s*$\n?", re.MULTILINE),
    re.compile(r"^CREATE TABLE IF NOT EXISTS sqlite_sequence\(name,seq\);\s*$\n?", re.MULTILINE),
)


def drop_version_noise(sql: str) -> str:
    """バージョン差によるノイズ行を落とし、sqlite_sequence の重複を畳む。"""
    for pattern in NOISE_PATTERNS:
        sql = pattern.sub("", sql)

    # DELETE FROM sqlite_sequence の直後に同じ INSERT が繰り返されている場合、
    # 初出だけを残す。sqlite_sequence は主キーを持たず重複行を許すが、
    # 同じ内容を二度書いても意味は無い。
    marker = "DELETE FROM sqlite_sequence;\n"
    head, sep, tail = sql.partition(marker)
    if not sep:
        return sql

    body, commit_sep, commit = tail.partition("COMMIT;")
    seen: set[str] = set()
    kept: list[str] = []
    for line in body.splitlines(keepends=True):
        stripped = line.strip()
        if stripped.startswith("INSERT INTO sqlite_sequence VALUES("):
            if stripped in seen:
                continue
            seen.add(stripped)
        kept.append(line)
    return head + sep + "".join(kept) + commit_sep + commit


def main() -> int:
    if len(sys.argv) != 2:
        print(USAGE, file=sys.stderr)
        return 1

    path = sys.argv[1]
    with open(path, encoding="utf-8") as f:
        sql = f.read()

    try:
        normalized, converted = normalize_sql(sql)
    except NormalizeError as exc:
        print(f"正規化に失敗しました: {exc}", file=sys.stderr)
        return 1

    with open(path, "w", encoding="utf-8") as f:
        f.write(normalized)

    print(f"salt を正規化しました: {converted} 行")
    return 0


if __name__ == "__main__":
    sys.exit(main())
