#!/usr/bin/env python3
"""sqlite3 の .dump 出力から salt の表記ゆれを揃える。

sqlite3 のバージョンによって REAL の出力精度が変わる。
3.46 は 1.39999999999999991、3.54 は 1.4 と書き出すため、
環境が違うと意味のない差分が出てしまう。

menus の INSERT は 13 列で、11 番目（0 始まりで 10）が salt。
これを小数点 1 桁に丸め直す。
"""

import re
import sys

USAGE = "使い方: normalize-dump.py <sql-file>"


def round_salt(match: re.Match) -> str:
    parts = match.group(1).split(",")
    if len(parts) != 13:
        return match.group(0)
    try:
        value = float(parts[10])
    except ValueError:
        return match.group(0)
    # 1.0 -> 1, 1.4 -> 1.4
    parts[10] = f"{value:.1f}".rstrip("0").rstrip(".")
    return f"INSERT INTO menus VALUES({','.join(parts)});"


def main() -> int:
    if len(sys.argv) != 2:
        print(USAGE, file=sys.stderr)
        return 1

    path = sys.argv[1]
    with open(path, encoding="utf-8") as f:
        sql = f.read()

    normalized = re.sub(r"INSERT INTO menus VALUES\(([^;]*)\);", round_salt, sql)

    with open(path, "w", encoding="utf-8") as f:
        f.write(normalized)

    return 0


if __name__ == "__main__":
    sys.exit(main())
