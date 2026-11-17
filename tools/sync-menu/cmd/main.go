// sync-menu は公式アレルギー情報 API から価格・栄養値を取得し、
// 既存の DB 行へ反映する。
//
// このコマンドは行の追加・削除をしない。API にはアレルゲン表示の対象商品
// しか含まれず（酒類・ドリンクバー等が欠ける）、そのまま再生成すると
// 既存行が大量に消えるため。
package main

import (
	"flag"
	"fmt"
	"os"

	"sync-menu/internal/db"
	"sync-menu/internal/fetch"
	"sync-menu/internal/match"
)

func main() {
	dbPath := flag.String("db", "../../db/saizeriya.db", "SQLite データベースのパス")
	dryRun := flag.Bool("dry-run", false, "更新内容を表示するだけで書き込まない")
	flag.Parse()

	if err := run(*dbPath, *dryRun); err != nil {
		fmt.Fprintf(os.Stderr, "エラー: %v\n", err)
		os.Exit(1)
	}
}

// sameInt はポインタ同士を比較する（どちらも nil なら等しい）。
func sameInt(a, b *int) bool {
	if a == nil || b == nil {
		return a == b
	}
	return *a == *b
}

// sameFloat はポインタ同士を比較する（どちらも nil なら等しい）。
func sameFloat(a, b *float64) bool {
	if a == nil || b == nil {
		return a == b
	}
	return *a == *b
}

func run(dbPath string, dryRun bool) error {
	fmt.Println("公式メニュー API からデータを取得中...")
	resp, err := fetch.Fetch(nil)
	if err != nil {
		return err
	}
	fmt.Printf("  %d 件を取得しました (lastModified=%s)\n", len(resp.Body), resp.LastModified)

	conn, err := db.Open(dbPath)
	if err != nil {
		return err
	}
	defer func() { _ = conn.Close() }()

	rows, err := db.LoadRows(conn)
	if err != nil {
		return err
	}
	fmt.Printf("  既存 DB: %d 件\n", len(rows))

	result := match.Resolve(rows, resp.Body)

	// 変更のある行だけを抽出する
	byID := make(map[int]db.Row, len(rows))
	for _, r := range rows {
		byID[r.ID] = r
	}

	type change struct {
		id     int
		name   string
		from   int
		to     int
		fields []string // 価格以外に変わった項目
	}

	// 価格・カロリー・塩分のどれかが変わっていれば書き込み対象。
	// 価格だけを見ると、栄養値だけが更新された商品を取りこぼす。
	var changes []change
	for id, v := range result.Values {
		row, ok := byID[id]
		if !ok {
			continue
		}
		var fields []string
		if !sameInt(row.Calorie, v.Calorie) {
			fields = append(fields, "kcal")
		}
		if !sameFloat(row.Salt, v.Salt) {
			fields = append(fields, "salt")
		}
		if row.PriceWithTax != v.PriceWithTax {
			fields = append([]string{"price"}, fields...)
		}
		if len(fields) > 0 {
			changes = append(changes, change{
				id: id, name: row.Name,
				from: row.PriceWithTax, to: v.PriceWithTax,
				fields: fields,
			})
		}
	}

	if len(result.Propagated) > 0 {
		fmt.Printf("  pre_id 伝播: %d 件\n", len(result.Propagated))
	}

	if len(changes) == 0 {
		fmt.Println("更新対象はありませんでした")
		return nil
	}

	// 価格が動いたものと、栄養値だけ動いたものを分けて出す
	var priceChanges []change
	for _, c := range changes {
		if c.from != c.to {
			priceChanges = append(priceChanges, c)
		}
	}

	fmt.Printf("\n更新: %d 件（うち価格変更 %d 件）\n", len(changes), len(priceChanges))
	for _, c := range priceChanges {
		fmt.Printf("  %d %s: ¥%d → ¥%d\n", c.id, c.name, c.from, c.to)
	}
	nutritionOnly := len(changes) - len(priceChanges)
	if nutritionOnly > 0 {
		fmt.Printf("  栄養値のみの更新: %d 件\n", nutritionOnly)
	}

	// 値下げは想定外のマッチングの可能性があるため目立たせる
	var decreases int
	for _, c := range priceChanges {
		if c.to < c.from {
			decreases++
		}
	}
	if decreases > 0 {
		fmt.Printf("\n⚠️  値下げが %d 件あります。マッチングを確認してください\n", decreases)
	}

	if dryRun {
		fmt.Println("\n--dry-run のため書き込みませんでした")
		return nil
	}

	tx, err := conn.Begin()
	if err != nil {
		return fmt.Errorf("トランザクションを開始できませんでした: %w", err)
	}
	for _, c := range changes {
		v := result.Values[c.id]
		row := byID[c.id]
		priceNoTax := row.PriceNoTax
		if row.PriceWithTax != v.PriceWithTax {
			// 税込価格が変わったときだけ税抜を計算し直す
			priceNoTax = int(float64(v.PriceWithTax)/1.1 + 0.5)
		}
		if err := db.UpdateNutrition(
			tx, c.id, priceNoTax, v.PriceWithTax, v.Calorie, v.Salt,
		); err != nil {
			_ = tx.Rollback()
			return err
		}
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("コミットに失敗しました: %w", err)
	}

	fmt.Printf("\n%d 件を更新しました\n", len(changes))
	return nil
}
