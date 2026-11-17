# saizeriya-menus

![Release](https://img.shields.io/github/v/release/ryohidaka/saizeriya-menus)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](https://opensource.org/licenses/MIT)

## 概要

サイゼリヤのメニュー一覧を提供するデータセット

> [!NOTE]
> 公序良俗に反する使用をしないでください。

> [!WARNING]
>
> - 収録データが古い場合があります。最新の情報は[公式サイトのメニュー](https://book.saizeriya.co.jp/library/menu1907/book/list)にてご確認ください。
> - 収録データは手入力のため誤ったデータが含まれる場合があります。あらかじめご了承ください。
> - 商品情報の著作権は株式会社 サイゼリヤ様に帰属します。

## データ

- [CSV](./saizeriya.csv)
- [JSON](./saizeriya.json)
- [Markdown](./saizeriya.md)

## データの更新

価格と栄養値は公式アレルギー情報 API から取得しています。GitHub Actions が毎週月曜 3:00 (JST) に自動で取り込みます。手元で実行する場合:

```bash
# DB を復元
sqlite3 db/saizeriya.db < db/saizeriya.sql

# 公式 API から価格・栄養値を取り込む（--dry-run で確認のみ）
cd tools/sync-menu
go run cmd/main.go -db ../../db/saizeriya.db --dry-run

# SQL に書き戻して生成物を更新
cd ../..
sqlite3 db/saizeriya.db ".output db/saizeriya.sql" ".dump"
cd tools/converter && go run cmd/main.go --csv --json --md
```

> [!NOTE]
> アレルギー情報 API はアレルゲン表示の対象商品しか返しません（酒類・ドリンクバー等が含まれない）。そのため `sync-menu` は既存行を残したまま価格と栄養値だけを更新し、行の追加・削除は行いません。

## [コントリビュートガイド](./CONTRIBUTING.md)

## 出典元

- [サイゼリヤ様公式サイト](https://book.saizeriya.co.jp/library/menu1907/book/list)
  - ID (注文時に使用する番号) / 価格 (税抜)
- [サイゼリヤ様公式アレルギー情報](https://allergy.saizeriya.co.jp/view?lg=ja)
  - 名称 / 価格 (税込) / カテゴリ名 / エネルギー(カロリー) / 食塩相当量

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
