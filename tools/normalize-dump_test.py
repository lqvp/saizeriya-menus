"""normalize-dump.py のテスト。

標準ライブラリだけで動くよう unittest を使う。
    cd tools && python3 -m unittest normalize-dump_test

`discover -s tools -p 'normalize-dump_test.py'` ではモジュール名に
ハイフンを含むため 0 件しか拾えない（テストしていないのに緑になる）。
"""

import subprocess
import sys
import tempfile
import unittest
from pathlib import Path

SCRIPT = Path(__file__).parent / "normalize-dump.py"


def run_script(sql: str) -> subprocess.CompletedProcess:
    with tempfile.NamedTemporaryFile("w", suffix=".sql", delete=False) as f:
        f.write(sql)
        path = f.name
    result = subprocess.run(
        [sys.executable, str(SCRIPT), path],
        capture_output=True,
        text=True,
        check=False,
    )
    result.output_text = Path(path).read_text()  # type: ignore[attr-defined]
    return result


def normalize(sql: str) -> str:
    result = run_script(sql)
    if result.returncode != 0:
        raise AssertionError(f"script failed: {result.stderr}")
    return result.output_text  # type: ignore[attr-defined]


class NormalizeDumpTest(unittest.TestCase):
    def test_rounds_long_float(self):
        """sqlite3 3.46 が書き出す精度いっぱいの値を丸める。"""
        src = (
            "INSERT INTO menus VALUES(1202,'小エビのサラダ','Shrimp Salad','小甜虾沙拉',"
            "336,370,196,1.39999999999999991,1,12,0,6,'SA02');"
        )
        self.assertIn("196,1.4,1,12,0,6", normalize(src))

    def test_keeps_integer_salt(self):
        """小数部が 0 のときは末尾を落とす。"""
        src = (
            "INSERT INTO menus VALUES(1301,'コーンクリームスープ','Corn','玉米',"
            "145,160,150,1.0,1,13,0,2,'SU01');"
        )
        self.assertIn("150,1,1,13,0,2", normalize(src))

    def test_leaves_other_tables_alone(self):
        """menus 以外の INSERT は触らない。"""
        src = (
            "INSERT INTO categories VALUES(1,'グランド','Grand Menu','主');\n"
            "INSERT INTO menus VALUES(1,'あ','a','a',1,2,3,1,1,1,0,1,'X1');\n"
        )
        self.assertEqual(normalize(src), src)

    def test_fails_on_empty_file(self):
        """空ファイルは「正常終了」させない。

        途中で .dump が落ちた場合に、空に近い SQL を書き戻す事故を防ぐ。
        """
        result = run_script("")
        self.assertEqual(result.returncode, 1)
        self.assertIn("1 行もありません", result.stderr)

    def test_fails_when_no_menus_insert(self):
        """menus の INSERT が無いファイルも失敗させる。"""
        src = "INSERT INTO categories VALUES(1,'グランド','Grand Menu','主');\n"
        result = run_script(src)
        self.assertEqual(result.returncode, 1)
        self.assertIn("1 行もありません", result.stderr)

    def test_fails_on_missing_file(self):
        """存在しないパスはトレースバックではなく整形されたエラーにする。"""
        result = subprocess.run(
            [sys.executable, str(SCRIPT), "/nonexistent/path/to.sql"],
            capture_output=True,
            text=True,
            check=False,
        )
        self.assertEqual(result.returncode, 1)
        self.assertIn("読めませんでした", result.stderr)
        self.assertNotIn("Traceback", result.stderr)

    def test_does_not_touch_file_on_failure(self):
        """失敗したときはファイルを書き換えない（部分破壊を避ける）。"""
        broken = "INSERT INTO menus VALUES(1,'あ',1,2);\n"
        with tempfile.NamedTemporaryFile("w", suffix=".sql", delete=False) as f:
            f.write(broken)
            path = f.name
        result = subprocess.run(
            [sys.executable, str(SCRIPT), path],
            capture_output=True,
            text=True,
            check=False,
        )
        self.assertEqual(result.returncode, 1)
        self.assertEqual(Path(path).read_text(), broken)

    def test_leaves_null_salt_alone(self):
        """salt が NULL の行はそのまま。"""
        src = (
            "INSERT INTO menus VALUES(3301,'生ビール','Beer','啤酒',"
            "364,400,NULL,NULL,1,33,1,25,NULL);"
        )
        self.assertEqual(normalize(src), src)

    def test_handles_comma_inside_quoted_value(self):
        """商品名にカンマが入っても列がずれない。"""
        src = (
            "INSERT INTO menus VALUES(1307,'たまねぎのズッパ','Onion soup, Zuppa','洋葱面包汤',"
            "291,320,278,2.39999999999999991,1,13,0,2,'SU07');"
        )
        got = normalize(src)
        self.assertIn("278,2.4,1,13,0,2,'SU07'", got)
        self.assertIn("'Onion soup, Zuppa'", got)

    def test_handles_escaped_quote(self):
        """'' でエスケープされた引用符を文字列の終端と誤認しない。"""
        src = (
            "INSERT INTO menus VALUES(1406,'小エビのサラダ','Shrimp with \"SAIZERIYA\" Dressing',"
            "'虾佐“SAIZERIYA”酱汁',291,320,124,1.29999999999999999,1,14,0,6,'AA06');"
        )
        src = src.replace('"SAIZERIYA"', "''SAIZERIYA''")
        got = normalize(src)
        self.assertIn("124,1.3,1,14,0,6,'AA06'", got)

    def test_handles_parenthesis_inside_value(self):
        """値の中の括弧を VALUES(...) の終端と誤認しない。"""
        src = (
            "INSERT INTO menus VALUES(1452,'アロスティチーニ（ラムの串焼き）(Wサイズ)',"
            "'Arrosticini','意式羊肉串',600,660,300,1.59999999999999999,1,14,0,6,'AA52');"
        )
        got = normalize(src)
        self.assertIn("300,1.6,1,14,0,6,'AA52'", got)

    def test_multiple_rows(self):
        """複数行をすべて処理する。"""
        src = (
            "INSERT INTO menus VALUES(1,'あ','a','a',1,2,3,1.19999999999999995,1,1,0,1,'X1');\n"
            "INSERT INTO menus VALUES(2,'い','i','i',1,2,3,2.89999999999999991,1,1,0,1,'X2');\n"
        )
        got = normalize(src)
        self.assertIn(",3,1.2,1,1,0,1,'X1'", got)
        self.assertIn(",3,2.9,1,1,0,1,'X2'", got)

    def test_fails_on_wrong_column_count(self):
        """列数が合わない行は素通しせず、終了コード 1 で止まる。"""
        src = "INSERT INTO menus VALUES(1202,'小エビのサラダ',336,370);"
        result = run_script(src)
        self.assertEqual(result.returncode, 1)
        self.assertIn("menus の値が 4 個です", result.stderr)

    def test_fails_on_unterminated_statement(self):
        """閉じ括弧が無い行も止まる。"""
        src = "INSERT INTO menus VALUES(1202,'小エビのサラダ',336,370"
        result = run_script(src)
        self.assertEqual(result.returncode, 1)

    def test_reports_converted_count(self):
        """変換した行数を出力する（CI ログで気づけるようにする）。"""
        src = (
            "INSERT INTO menus VALUES(1,'あ','a','a',1,2,3,1.19999999999999995,1,1,0,1,'X1');\n"
        )
        result = run_script(src)
        self.assertEqual(result.returncode, 0)
        self.assertIn("1 行", result.stdout)

    def test_requires_argument(self):
        """引数が無ければ使い方を出して失敗する。"""
        result = subprocess.run(
            [sys.executable, str(SCRIPT)],
            capture_output=True,
            text=True,
            check=False,
        )
        self.assertEqual(result.returncode, 1)
        self.assertIn("使い方", result.stderr)

    def test_drops_pragma_noise(self):
        """sqlite3 3.54 が出す PRAGMA / CREATE TABLE を落とす。

        3.46 は出さないため、そのままだとバージョン差で毎回差分になる。
        """
        src = (
            "INSERT INTO menus VALUES(1,'あ','a','a',1,2,3,1,1,1,0,1,'X1');\n"
            "PRAGMA writable_schema=ON;\n"
            "CREATE TABLE IF NOT EXISTS sqlite_sequence(name,seq);\n"
            "DELETE FROM sqlite_sequence;\n"
            "INSERT INTO sqlite_sequence VALUES('categories',6);\n"
            "PRAGMA writable_schema=OFF;\n"
            "COMMIT;\n"
        )
        got = normalize(src)
        self.assertNotIn("PRAGMA", got)
        self.assertNotIn("CREATE TABLE", got)
        self.assertIn("INSERT INTO sqlite_sequence VALUES('categories',6);", got)

    def test_dedupes_repeated_sqlite_sequence_rows(self):
        """同じ sqlite_sequence 行が重複していたら 1 つに畳む。"""
        src = (
            "INSERT INTO menus VALUES(1,'あ','a','a',1,2,3,1,1,1,0,1,'X1');\n"
            "DELETE FROM sqlite_sequence;\n"
            "INSERT INTO sqlite_sequence VALUES('categories',6);\n"
            "INSERT INTO sqlite_sequence VALUES('icons',31);\n"
            "INSERT INTO sqlite_sequence VALUES('categories',6);\n"
            "INSERT INTO sqlite_sequence VALUES('icons',31);\n"
            "COMMIT;\n"
        )
        got = normalize(src)
        self.assertEqual(got.count("VALUES('categories',6)"), 1)
        self.assertEqual(got.count("VALUES('icons',31)"), 1)

    def test_keeps_distinct_sqlite_sequence_rows(self):
        """値が違えば畳まずに残す。"""
        src = (
            "INSERT INTO menus VALUES(1,'あ','a','a',1,2,3,1,1,1,0,1,'X1');\n"
            "DELETE FROM sqlite_sequence;\n"
            "INSERT INTO sqlite_sequence VALUES('categories',6);\n"
            "INSERT INTO sqlite_sequence VALUES('categories',7);\n"
            "COMMIT;\n"
        )
        got = normalize(src)
        self.assertIn("VALUES('categories',6)", got)
        self.assertIn("VALUES('categories',7)", got)


if __name__ == "__main__":
    unittest.main()
