package db

import (
	"database/sql"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

// salt に float64 をそのまま渡すと 1.4 が 1.39999999999999991 として
// 保存される。UpdateNutrition が丸めていることを確認する。
func TestUpdateNutritionRoundsSalt(t *testing.T) {
	conn, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = conn.Close() }()

	_, err = conn.Exec(`CREATE TABLE menus (
		id INTEGER PRIMARY KEY,
		name TEXT,
		price INTEGER,
		price_with_tax INTEGER,
		calorie INTEGER,
		salt REAL
	)`)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := conn.Exec(
		`INSERT INTO menus VALUES (1202, '小エビのサラダ', 0, 0, 0, 0)`,
	); err != nil {
		t.Fatal(err)
	}

	salt := 1.4
	calorie := 196
	if err := UpdateNutrition(conn, 1202, 336, 370, &calorie, &salt); err != nil {
		t.Fatal(err)
	}

	var gotSalt float64
	var gotCalorie, gotPrice, gotPriceNoTax int
	if err := conn.QueryRow(
		`SELECT salt, calorie, price, price_with_tax FROM menus WHERE id = 1202`,
	).Scan(&gotSalt, &gotCalorie, &gotPriceNoTax, &gotPrice); err != nil {
		t.Fatal(err)
	}

	// 文字列化して 1.4 になること（1.3999999999999999 ではない）
	var asText string
	if err := conn.QueryRow(`SELECT CAST(salt AS TEXT) FROM menus WHERE id = 1202`).
		Scan(&asText); err != nil {
		t.Fatal(err)
	}
	if asText != "1.4" {
		t.Errorf("salt = %s, want 1.4", asText)
	}
	if gotCalorie != 196 {
		t.Errorf("calorie = %d, want 196", gotCalorie)
	}
	// 370 / 1.1 = 336.36... → 336
	if gotPriceNoTax != 336 {
		t.Errorf("price = %d, want 336", gotPriceNoTax)
	}
}

// 税込価格が変わらない行では、呼び出し側が渡した税抜価格をそのまま
// 保存することを確認する。350 から逆算すると 318 になるが、実際の
// 税抜価格は 319 であり、既知の値を壊してはいけない。
func TestUpdateNutritionKeepsGivenNetPrice(t *testing.T) {
	conn, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = conn.Close() }()

	if _, err := conn.Exec(`CREATE TABLE menus (
		id INTEGER PRIMARY KEY, name TEXT, price INTEGER,
		price_with_tax INTEGER, calorie INTEGER, salt REAL
	)`); err != nil {
		t.Fatal(err)
	}
	if _, err := conn.Exec(
		`INSERT INTO menus VALUES (1208, 'グリーンサラダ', 319, 350, 166, 0.7)`,
	); err != nil {
		t.Fatal(err)
	}

	// 税込は 350 のまま、カロリーだけ更新する
	calorie := 170
	salt := 0.7
	if err := UpdateNutrition(conn, 1208, 319, 350, &calorie, &salt); err != nil {
		t.Fatal(err)
	}

	var net, gross int
	if err := conn.QueryRow(
		`SELECT price, price_with_tax FROM menus WHERE id = 1208`,
	).Scan(&net, &gross); err != nil {
		t.Fatal(err)
	}
	if net != 319 {
		t.Errorf("price = %d, want 319 (must not be recomputed from gross)", net)
	}
	if gross != 350 {
		t.Errorf("price_with_tax = %d, want 350", gross)
	}
}
