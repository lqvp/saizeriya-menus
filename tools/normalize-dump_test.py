"""normalize-dump.py のテスト。

標準ライブラリだけで動くよう unittest を使う。
    python3 -m unittest discover -s tools -p 'normalize-dump_test.py'
"""

import subprocess
import sys
import tempfile
import unittest
from pathlib import Path

SCRIPT = Path(__file__).parent / "normalize-dump.py"


def normalize(sql: str) -> str:
    with tempfile.NamedTemporaryFile("w", suffix=".sql", delete=False) as f:
        f.write(sql)
        path = f.name
    subprocess.run([sys.executable, str(SCRIPT), path], check=True)
    return Path(path).read_text()


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
        src = "INSERT INTO categories VALUES(1,'グランド','Grand Menu','主');"
        self.assertEqual(normalize(src), src)

    def test_leaves_null_salt_alone(self):
        """salt が NULL の行はそのまま。"""
        src = (
            "INSERT INTO menus VALUES(3301,'生ビール','Beer','啤酒',"
            "364,400,NULL,NULL,1,33,1,25,NULL);"
        )
        self.assertEqual(normalize(src), src)


if __name__ == "__main__":
    unittest.main()
