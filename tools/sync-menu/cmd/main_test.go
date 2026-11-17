package main

import (
	"database/sql"
	"testing"

	_ "github.com/mattn/go-sqlite3"

	"sync-menu/internal/db"
	"sync-menu/internal/match"
)

// 栄養値の比較は nil 安全でなければならない。
// 片方だけ nil のときに panic したり、誤って「同じ」と判定してはいけない。
func TestSameInt(t *testing.T) {
	three := 3
	other := 5

	if !sameInt(nil, nil) {
		t.Error("nil, nil should be equal")
	}
	if sameInt(&three, nil) {
		t.Error("3 and nil should differ")
	}
	if sameInt(nil, &three) {
		t.Error("nil and 3 should differ")
	}
	if !sameInt(&three, &three) {
		t.Error("3 and 3 should be equal")
	}
	if sameInt(&three, &other) {
		t.Error("3 and 5 should differ")
	}
}

func TestSameFloat(t *testing.T) {
	one := 1.4
	other := 0.9

	if !sameFloat(nil, nil) {
		t.Error("nil, nil should be equal")
	}
	if sameFloat(&one, nil) {
		t.Error("1.4 and nil should differ")
	}
	if !sameFloat(&one, &one) {
		t.Error("1.4 and 1.4 should be equal")
	}
	if sameFloat(&one, &other) {
		t.Error("1.4 and 0.9 should differ")
	}
}

// 価格が同じでも、栄養値が違えば変更として拾うこと。
// 拾い損ねると API のカロリー更新が黙って捨てられる。
func TestCollectChangesIncludesNutritionOnly(t *testing.T) {
	cal := 124
	newCal := 130
	salt := 1.3
	rows := []db.Row{
		{ID: 1406, Name: "小エビのカクテル", CategoryID: 1, PriceWithTax: 320, Calorie: &cal, Salt: &salt},
	}
	values := map[int]match.Resolved{
		1406: {PriceWithTax: 320, Calorie: &newCal, Salt: &salt},
	}

	changes, byID := collectChanges(rows, values)

	if len(changes) != 1 {
		t.Fatalf("len(changes) = %d, want 1", len(changes))
	}
	if changes[0].from != changes[0].to {
		t.Errorf("price should be unchanged: %d -> %d", changes[0].from, changes[0].to)
	}
	if len(changes[0].fields) != 1 || changes[0].fields[0] != "kcal" {
		t.Errorf("fields = %v, want [kcal]", changes[0].fields)
	}
	if _, ok := byID[1406]; !ok {
		t.Error("byID should contain 1406")
	}
}

// 何も変わらない行は変更に含めないこと（冪等性の担保）。
func TestCollectChangesSkipsUnchanged(t *testing.T) {
	cal := 124
	salt := 1.3
	rows := []db.Row{
		{ID: 1406, Name: "小エビのカクテル", CategoryID: 1, PriceWithTax: 320, Calorie: &cal, Salt: &salt},
	}
	c := cal
	s := salt
	values := map[int]match.Resolved{
		1406: {PriceWithTax: 320, Calorie: &c, Salt: &s},
	}

	changes, _ := collectChanges(rows, values)

	if len(changes) != 0 {
		t.Errorf("changes = %+v, want none", changes)
	}
}

// 価格が変わったときは price が先頭に来ること（表示順の担保）。
func TestCollectChangesOrdersPriceFirst(t *testing.T) {
	cal := 124
	newCal := 130
	salt := 1.3
	newSalt := 1.4
	rows := []db.Row{
		{ID: 1406, Name: "小エビのカクテル", CategoryID: 1, PriceWithTax: 320, Calorie: &cal, Salt: &salt},
	}
	values := map[int]match.Resolved{
		1406: {PriceWithTax: 370, Calorie: &newCal, Salt: &newSalt},
	}

	changes, _ := collectChanges(rows, values)

	if len(changes) != 1 {
		t.Fatalf("len(changes) = %d, want 1", len(changes))
	}
	got := changes[0].fields
	want := []string{"price", "kcal", "salt"}
	if len(got) != len(want) {
		t.Fatalf("fields = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("fields = %v, want %v", got, want)
			break
		}
	}
}

// 未知の id は無視すること。
func TestCollectChangesIgnoresUnknownID(t *testing.T) {
	cal := 100
	rows := []db.Row{
		{ID: 1, Name: "既知", CategoryID: 1, PriceWithTax: 100, Calorie: &cal},
	}
	values := map[int]match.Resolved{
		999: {PriceWithTax: 200, Calorie: &cal},
	}

	changes, _ := collectChanges(rows, values)

	if len(changes) != 0 {
		t.Errorf("changes = %+v, want none", changes)
	}
}

// 税込価格が変わったときだけ税抜を計算し直すこと。
func TestApplyRecomputesNetPriceOnlyWhenGrossChanges(t *testing.T) {
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
		`INSERT INTO menus VALUES (1208, 'グリーンサラダ', 319, 350, 166, 0.7),
		                        (1202, '小エビのサラダ', 336, 370, 196, 1.4)`,
	); err != nil {
		t.Fatal(err)
	}

	cal := 166
	newSalt := 0.8
	rows := []db.Row{
		// 税込は 350 のまま、塩分だけ変える
		{ID: 1208, Name: "グリーンサラダ", CategoryID: 1, PriceNoTax: 319, PriceWithTax: 350, Calorie: &cal},
		// 税込が 370 → 380 に変わる
		{ID: 1202, Name: "小エビのサラダ", CategoryID: 1, PriceNoTax: 336, PriceWithTax: 370, Calorie: &cal},
	}
	values := map[int]match.Resolved{
		1208: {PriceWithTax: 350, Calorie: &cal, Salt: &newSalt},
		1202: {PriceWithTax: 380, Calorie: &cal},
	}

	changes, byID := collectChanges(rows, values)
	if err := apply(conn, changes, byID, values); err != nil {
		t.Fatalf("apply() error = %v", err)
	}

	var net, gross int
	if err := conn.QueryRow(
		`SELECT price, price_with_tax FROM menus WHERE id = 1208`,
	).Scan(&net, &gross); err != nil {
		t.Fatal(err)
	}
	// 税込が変わらないので既知の税抜 319 を保つ（318 にしない）
	if net != 319 || gross != 350 {
		t.Errorf("1208 = %d/%d, want 319/350", net, gross)
	}

	// 380 / 1.1 = 345.45... → 345
	if err := conn.QueryRow(
		`SELECT price, price_with_tax FROM menus WHERE id = 1202`,
	).Scan(&net, &gross); err != nil {
		t.Fatal(err)
	}
	if net != 345 || gross != 380 {
		t.Errorf("1202 = %d/%d, want 345/380", net, gross)
	}
}
