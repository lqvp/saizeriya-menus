// Package db は SQLite の読み書きを担当する。
package db

import (
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
)

// Row は menus テーブルの 1 行。
type Row struct {
	ID           int
	Name         string
	NameEn       *string
	NameZh       *string
	PriceNoTax   int
	PriceWithTax int
	Calorie      *int
	Salt         *float64
	CategoryID   int
	GenreID      *int
	IsAlcohol    int
	IconID       *int
	PreID        string
}

// Open は DB を開く。
func Open(path string) (*sql.DB, error) {
	conn, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, fmt.Errorf("DB を開けませんでした: %w", err)
	}
	if err := conn.Ping(); err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("DB に接続できませんでした: %w", err)
	}
	return conn, nil
}

// LoadRows は menus の全行を読み込む。
func LoadRows(conn *sql.DB) ([]Row, error) {
	rows, err := conn.Query(
		`SELECT id, name, name_en, name_zh, price, price_with_tax, calorie, salt,
		        category_id, genre_id, COALESCE(is_alcohol, 0), icon_id, COALESCE(pre_id, '')
		 FROM menus ORDER BY id`,
	)
	if err != nil {
		return nil, fmt.Errorf("menus の読み込みに失敗しました: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var out []Row
	for rows.Next() {
		var r Row
		if err := rows.Scan(
			&r.ID, &r.Name, &r.NameEn, &r.NameZh, &r.PriceNoTax, &r.PriceWithTax,
			&r.Calorie, &r.Salt, &r.CategoryID, &r.GenreID, &r.IsAlcohol,
			&r.IconID, &r.PreID,
		); err != nil {
			return nil, fmt.Errorf("行の読み込みに失敗しました: %w", err)
		}
		out = append(out, r)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("行の走査に失敗しました: %w", err)
	}
	return out, nil
}

// Execer は *sql.DB と *sql.Tx の共通部分。
type Execer interface {
	Exec(query string, args ...any) (sql.Result, error)
}

// UpdateNutrition は 1 行の価格・カロリー・塩分を更新する。
// 識別子（id / genre_id / pre_id / category_id）は変更しない。
func UpdateNutrition(
	conn Execer, id, priceWithTax int, calorie *int, salt *float64,
) error {
	priceNoTax := int(float64(priceWithTax)/1.1 + 0.5)
	_, err := conn.Exec(
		`UPDATE menus SET price = ?, price_with_tax = ?, calorie = ?, salt = ?
		 WHERE id = ?`,
		priceNoTax, priceWithTax, calorie, salt, id,
	)
	if err != nil {
		return fmt.Errorf("id=%d の更新に失敗しました: %w", id, err)
	}
	return nil
}
