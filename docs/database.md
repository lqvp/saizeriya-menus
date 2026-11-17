# データベース構成

## ER 図

```mermaid
erDiagram
    CATEGORIES {
        INTEGER id PK
        TEXT name
        TEXT name_en
        TEXT name_zh
    }

    GENRES {
        INTEGER id PK
        TEXT name
        TEXT pre_id
    }

    ICONS {
        INTEGER id PK
        TEXT icon
    }

    MENUS {
        INTEGER id PK
        TEXT name
        TEXT name_en
        TEXT name_zh
        INTEGER price
        INTEGER price_with_tax
        INTEGER calorie
        INTEGER salt
        INTEGER category_id FK
        INTEGER genre_id FK
        INTEGER icon_id FK
        TEXT pre_id
    }

    CATEGORIES ||--o{ MENUS: "has"
    GENRES ||--o{ MENUS: "has"
    ICONS ||--o{ MENUS: "has"
```

## データ構造

### カテゴリマスタ (`categories`)

| KEY 名    | 説明               | データ型  | サンプル     |
| --------- | ------------------ | --------- | ------------ |
| `id`      | カテゴリ ID        | `INTEGER` | `1`          |
| `name`    | カテゴリ名         | `TEXT`    | `グランド`   |
| `name_en` | カテゴリ名(英語)   | `TEXT`    | `Grand Menu` |
| `name_zh` | カテゴリ名(中国語) | `TEXT`    | `主`         |

### ジャンルマスタ (`genres`)

| KEY 名   | 説明          | データ型  | サンプル |
| -------- | ------------- | --------- | -------- |
| `id`     | ジャンル ID   | `INTEGER` | `12`     |
| `name`   | ジャンル名    | `TEXT`    | `サラダ` |
| `pre_id` | 旧ジャンル ID | `TEXT`    | `SA`     |

### アイコンマスタ (`icons`)

| KEY 名 | 説明        | データ型  | サンプル |
| ------ | ----------- | --------- | -------- |
| `id`   | アイコン ID | `INTEGER` | `1`      |
| `icon` | アイコン    | `TEXT`    | `🥗`    |

### 商品マスタ (`menus`)

| KEY 名           | 説明                        | データ型  | サンプル         |
| ---------------- | --------------------------- | --------- | ---------------- |
| `id`             | メニュー ID                 | `INTEGER` | `1202`           |
| `name`           | メニュー名                  | `TEXT`    | `小エビのサラダ` |
| `name_en`        | メニュー名(英語)            | `TEXT`    | `Shrimp Salad`   |
| `name_zh`        | メニュー名(中国語)          | `TEXT`    | `小甜虾沙拉`     |
| `price`          | 税抜価格                    | `INTEGER` | `319`            |
| `price_with_tax` | 税込価格                    | `INTEGER` | `350`            |
| `calorie`        | エネルギー(カロリー) (kcal) | `INTEGER` | `192`            |
| `salt`           | 食塩相当量 (g)              | `INTEGER`※ | `1.5`            |
| `category_id`    | カテゴリ ID                 | `INTEGER` | `1`              |
| `genre_id`       | ジャンル ID                 | `INTEGER` | `12`             |
| `icon_id`        | アイコン ID                 | `INTEGER` | `6`              |
| `pre_id`         | 旧メニュー ID               | `TEXT`    | `SA02`           |

※ `salt` は DDL 上 `INTEGER` と宣言されているが、実データは `1.4` のような
小数を含むため、SQLite の型親和性により **REAL として保存される**
（`SELECT typeof(salt)` で確認できる）。宣言と実体がずれているが、
DDL を変更すると保存型が変わり、下流の読み取りに影響しうるため、
このPRでは扱わない。変更する場合は別途マイグレーションとして検討すること。

> [!WARNING]
> `pre_id` は「商品を一意に指す ID」ではない。テーブルをまたぐと意味が変わるうえ、
> `menus` の中でも次の 2 種類が混在している。
>
> | 対象              | 意味                                | 例           |
> | ----------------- | ----------------------------------- | ------------ |
> | 1xxx〜8xxx        | 商品 SKU コード                     | `SA02`       |
> | 32xx / 39xx       | デザートのペアキー（同じ値で 2 行） | `DE01`       |
>
> さらに `AA06` のように、**別商品が同じ `pre_id` を共有している**行も実在する
> （`1406 小エビのカクテル` と `1425 柔ら青豆の温サラダ`）。
> `pre_id` を手がかりに値をコピーする処理を書くときは、`category_id` と
> 正規化後の `name` でも一致を確認すること。
