// sync-menu は公式アレルギー情報 API から価格・栄養値を取得し、
// 既存の DB 行へ反映する。
//
// このコマンドは行の追加・削除をしない。API にはアレルゲン表示の対象商品
// しか含まれず（酒類・ドリンクバー等が欠ける）、そのまま再生成すると
// 既存行が大量に消えるため。
package main

import (
	"database/sql"
	"flag"
	"fmt"
	"os"

	"sync-menu/internal/db"
	"sync-menu/internal/fetch"
	"sync-menu/internal/match"
)

// defaultMatchRateThreshold は「これ未満なら異常」とみなす突合率（%）。
//
// API は全商品を返さない（酒類・ドリンクバー等が無い）ため 100% にはならない。
// 2026-11 時点の実測では 161 行中 85 行（52%）が突き合う。これを下回るのは、
// API の商品名が変わった・カテゴリが増えたなど、取りこぼしのサインである。
//
// なお、この閾値は「DB 行数に対する割合」しか見ていない。API 側の一部の
// 商品が消えただけでは数ポイントしか動かないため、突合率だけに頼らず、
// 曖昧で見送った行（Result.Ambiguous）も必ずログに出す。
const defaultMatchRateThreshold = 50

type options struct {
	dbPath       string
	dryRun       bool
	minMatchRate int
}

// change は 1 行分の更新内容。
type change struct {
	id     int
	name   string
	from   int
	to     int
	fields []string // 変わった項目（price / kcal / salt）
}

func main() {
	dbPath := flag.String("db", "../../db/saizeriya.db", "SQLite データベースのパス")
	dryRun := flag.Bool("dry-run", false, "更新内容を表示するだけで書き込まない")
	minMatchRate := flag.Int(
		"min-match-rate", defaultMatchRateThreshold,
		"突合率の下限（%）。下回ると異常として終了コード 1 で止まる",
	)
	flag.Parse()

	opts := options{
		dbPath:       *dbPath,
		dryRun:       *dryRun,
		minMatchRate: *minMatchRate,
	}
	if err := run(opts); err != nil {
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

func run(opts options) error {
	fmt.Println("公式メニュー API からデータを取得中...")
	resp, err := fetch.Fetch(nil)
	if err != nil {
		return err
	}
	fmt.Printf("  %d 件を取得しました (lastModified=%s)\n", len(resp.Body), resp.LastModified)

	conn, err := db.Open(opts.dbPath)
	if err != nil {
		return fmt.Errorf(
			"%w\nヒント: リポジトリ直下から実行する場合は -db db/saizeriya.db を指定してください"+
				"（DB が無い場合は make restore で復元できます）",
			err,
		)
	}
	defer func() { _ = conn.Close() }()

	rows, err := db.LoadRows(conn)
	if err != nil {
		return err
	}
	fmt.Printf("  既存 DB: %d 件\n", len(rows))

	result := match.Resolve(rows, resp.Body)

	// 突合率を出す。マッチしなかった行は更新されないため、
	// ここが下がったら API 側の変化を疑う。
	rate := 0
	if len(rows) > 0 {
		rate = result.Matched * 100 / len(rows)
	}
	fmt.Printf("\n突合: %d / %d 件（%d%%）\n", result.Matched, len(rows), rate)

	if len(result.Ambiguous) > 0 {
		fmt.Printf("\n⚠️  API 候補が複数あり、一意に決められないため見送った行: %d 件\n",
			len(result.Ambiguous))
		for _, a := range result.Ambiguous {
			fmt.Printf("  %s (category_id=%d)\n", a.Name, a.CategoryID)
		}
	}

	if len(result.Propagated) > 0 {
		fmt.Printf("  pre_id 伝播: %d 件\n", len(result.Propagated))
	}

	// 突合率が下限を下回るなら、データを書く前に止める。
	if rate < opts.minMatchRate {
		return fmt.Errorf(
			"突合率 %d%% が下限 %d%% を下回りました。API の商品名・カテゴリが変わっていないか確認してください",
			rate, opts.minMatchRate,
		)
	}

	changes, byID := collectChanges(rows, result.Values)
	if len(changes) == 0 {
		fmt.Println("\n更新対象はありませんでした")
		return nil
	}

	report(changes)

	if opts.dryRun {
		fmt.Println("\n--dry-run のため書き込みませんでした")
		return nil
	}

	if err := apply(conn, changes, byID, result.Values); err != nil {
		return err
	}

	fmt.Printf("\n%d 件を更新しました\n", len(changes))
	return nil
}

// collectChanges は「実際に値が変わる行」だけを抽出する。
//
// 価格・カロリー・塩分のどれかが変わっていれば書き込み対象にする。
// 価格だけを見ると、栄養値だけが更新された商品を取りこぼす。
func collectChanges(rows []db.Row, values map[int]match.Resolved) ([]change, map[int]db.Row) {
	byID := make(map[int]db.Row, len(rows))
	for _, r := range rows {
		byID[r.ID] = r
	}

	var changes []change
	for id, v := range values {
		row, ok := byID[id]
		if !ok {
			continue
		}
		// nil は「API に値が無い」を意味する。書き込み時は既存値を保つ。
		calorie := v.Calorie
		if calorie == nil {
			calorie = row.Calorie
		}
		salt := v.Salt
		if salt == nil {
			salt = row.Salt
		}

		var fields []string
		if row.PriceWithTax != v.PriceWithTax {
			fields = append(fields, "price")
		}
		if !sameInt(row.Calorie, calorie) {
			fields = append(fields, "kcal")
		}
		if !sameFloat(row.Salt, salt) {
			fields = append(fields, "salt")
		}
		if len(fields) > 0 {
			changes = append(changes, change{
				id: id, name: row.Name,
				from: row.PriceWithTax, to: v.PriceWithTax,
				fields: fields,
			})
		}
	}
	return changes, byID
}

// report は変更内容を人間向けに出力する。
func report(changes []change) {
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
	if nutritionOnly := len(changes) - len(priceChanges); nutritionOnly > 0 {
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
}

// apply は変更をまとめて書き込む。
//
// 途中で失敗したらロールバックし、何件目で落ちたかを返す。
// CI のログだけを見て「どこまで進んだか」が分かるようにするため。
func apply(
	conn *sql.DB,
	changes []change,
	byID map[int]db.Row,
	values map[int]match.Resolved,
) error {
	tx, err := conn.Begin()
	if err != nil {
		return fmt.Errorf("トランザクションを開始できませんでした: %w", err)
	}

	for i, c := range changes {
		v := values[c.id]
		row := byID[c.id]
		priceNoTax := row.PriceNoTax
		if row.PriceWithTax != v.PriceWithTax {
			// 税込価格が変わったときだけ税抜を計算し直す
			priceNoTax = int(float64(v.PriceWithTax)/1.1 + 0.5)
		}
		// nil は「API に値が無い」。既存値をそのまま書き戻す。
		calorie := v.Calorie
		if calorie == nil {
			calorie = row.Calorie
		}
		salt := v.Salt
		if salt == nil {
			salt = row.Salt
		}
		if err := db.UpdateNutrition(
			tx, c.id, priceNoTax, v.PriceWithTax, calorie, salt,
		); err != nil {
			rollbackErr := tx.Rollback()
			msg := fmt.Sprintf(
				"%d 件目（id=%d %s）で失敗しました",
				i+1, c.id, c.name,
			)
			if rollbackErr != nil {
				return fmt.Errorf("%s。ロールバックにも失敗しました: %w（元のエラー: %w）",
					msg, rollbackErr, err)
			}
			return fmt.Errorf("%s。ここまでの更新はロールバック済みです: %w", msg, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("コミットに失敗しました（更新は反映されていません）: %w", err)
	}
	return nil
}
