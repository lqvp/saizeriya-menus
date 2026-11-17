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
		id   int
		name string
		from int
		to   int
	}
	var changes []change
	for id, v := range result.Values {
		row, ok := byID[id]
		if !ok {
			continue
		}
		if row.PriceWithTax != v.PriceWithTax {
			changes = append(changes, change{id, row.Name, row.PriceWithTax, v.PriceWithTax})
		}
	}

	if len(result.Propagated) > 0 {
		fmt.Printf("  pre_id 伝播: %d 件\n", len(result.Propagated))
	}

	if len(changes) == 0 {
		fmt.Println("更新対象はありませんでした")
		return nil
	}

	fmt.Printf("\n価格変更: %d 件\n", len(changes))
	for _, c := range changes {
		fmt.Printf("  %d %s: ¥%d → ¥%d\n", c.id, c.name, c.from, c.to)
	}

	// 値下げは想定外のマッチングの可能性があるため目立たせる
	var decreases int
	for _, c := range changes {
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
	for id, v := range result.Values {
		if err := db.UpdateNutrition(tx, id, v.PriceWithTax, v.Calorie, v.Salt); err != nil {
			_ = tx.Rollback()
			return err
		}
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("コミットに失敗しました: %w", err)
	}

	fmt.Printf("\n%d 件を更新しました\n", len(result.Values))
	return nil
}
