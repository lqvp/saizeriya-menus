// Package match は既存の DB 行と公式データを突き合わせる。
//
// マッチングの方針:
//  1. 「正規化した商品名 + カテゴリ」の完全一致
//  2. 1 で漏れた行が「あとで」版デザート等で、同じ pre_id のペアが
//     解決済みなら、その値をコピーする
//
// 2 が必要な理由:
//
//	「ティラミス　クラシコ」は「すぐに」(3201) と「あとで」(3901) の
//	2 行あるが、公式 API の商品名は同一でカテゴリも同じため、1 の完全一致
//	では片方しかマッチしない。pre_id（DE01）が同じなら同一商品なので揃える。
//
// 接尾辞を落とす曖昧な一致は行わない。モーニングのセット/単品を
// 取り違えるため（例: フォッカチオ ¥150 と フォッカチオモーニング ¥1000）。
package match

import (
	"strings"
	"unicode"

	"sync-menu/internal/db"
	"sync-menu/internal/fetch"
)

// categoryIDs は公式カテゴリ名から DB の category_id への対応。
// API が返すのは「グランド」「ランチ」「モーニング」「テイクアウト」「季節」。
// 「キッズ」は現状 API に無いが、追加されたときに無視しないよう残す。
var categoryIDs = map[string]int{
	"グランド":   1,
	"ランチ":    2,
	"モーニング":  3,
	"キッズ":    4,
	"テイクアウト": 5,
	"季節":     6,
}

// Normalize は商品名を照合用に正規化する。
// 全角/半角・記号・空白の差を吸収する。
func Normalize(name string) string {
	var b strings.Builder
	for _, r := range name {
		switch {
		// 全角英数記号 → 半角
		case r >= 0xFF01 && r <= 0xFF5E:
			r -= 0xFEE0
		// 半角カタカナ → 全角カタカナ
		case r >= 0xFF61 && r <= 0xFF9F:
			if full, ok := halfToFullKana[r]; ok {
				r = full
			}
		}
		switch r {
		case '（', '）', '(', ')', '「', '」', '・', '，', ',', '／', '/', '　',
			' ', '-', '－', '〜', '～', '※':
			continue
		}
		b.WriteRune(unicode.ToLower(r))
	}
	return strings.TrimSpace(composeDakuten(b.String()))
}

// composeDakuten は「カ゛」を「ガ」のように濁点を合成する。
// 半角カナを全角へ変換すると濁点が分離するため、元の 1 文字に戻す。
func composeDakuten(s string) string {
	var b strings.Builder
	for _, r := range s {
		switch r {
		case '゛':
			// 直前の文字を濁音化する
			prev := []rune(b.String())
			if len(prev) == 0 {
				continue
			}
			base := prev[len(prev)-1]
			if voiced, ok := dakuten[base]; ok {
				b.Reset()
				b.WriteString(string(prev[:len(prev)-1]))
				b.WriteRune(voiced)
			}
			continue
		case '゜':
			prev := []rune(b.String())
			if len(prev) == 0 {
				continue
			}
			base := prev[len(prev)-1]
			if voiced, ok := handakuten[base]; ok {
				b.Reset()
				b.WriteString(string(prev[:len(prev)-1]))
				b.WriteRune(voiced)
			}
			continue
		}
		b.WriteRune(r)
	}
	return b.String()
}

var dakuten = map[rune]rune{
	'カ': 'ガ', 'キ': 'ギ', 'ク': 'グ', 'ケ': 'ゲ', 'コ': 'ゴ',
	'サ': 'ザ', 'シ': 'ジ', 'ス': 'ズ', 'セ': 'ゼ', 'ソ': 'ゾ',
	'タ': 'ダ', 'チ': 'ヂ', 'ツ': 'ヅ', 'テ': 'デ', 'ト': 'ド',
	'ハ': 'バ', 'ヒ': 'ビ', 'フ': 'ブ', 'ヘ': 'ベ', 'ホ': 'ボ',
	'ウ': 'ヴ', 'ワ': 'ヷ', 'ヲ': 'ヺ',
}

var handakuten = map[rune]rune{
	'ハ': 'パ', 'ヒ': 'ピ', 'フ': 'プ', 'ヘ': 'ペ', 'ホ': 'ポ',
}

// halfToFullKana は半角カタカナを全角カタカナへ対応付ける。
// 濁点・半濁点付き（ｶﾞ 等）は 2 文字で 1 文字になるため、
// ここでは単体の清音のみを扱い、濁点は直前の文字に合成する。
var halfToFullKana = map[rune]rune{
	'ｦ': 'ヲ', 'ｧ': 'ァ', 'ｨ': 'ィ', 'ｩ': 'ゥ', 'ｪ': 'ェ', 'ｫ': 'ォ',
	'ｬ': 'ャ', 'ｭ': 'ュ', 'ｮ': 'ョ', 'ｯ': 'ッ',
	'ｱ': 'ア', 'ｲ': 'イ', 'ｳ': 'ウ', 'ｴ': 'エ', 'ｵ': 'オ',
	'ｶ': 'カ', 'ｷ': 'キ', 'ｸ': 'ク', 'ｹ': 'ケ', 'ｺ': 'コ',
	'ｻ': 'サ', 'ｼ': 'シ', 'ｽ': 'ス', 'ｾ': 'セ', 'ｿ': 'ソ',
	'ﾀ': 'タ', 'ﾁ': 'チ', 'ﾂ': 'ツ', 'ﾃ': 'テ', 'ﾄ': 'ト',
	'ﾅ': 'ナ', 'ﾆ': 'ニ', 'ﾇ': 'ヌ', 'ﾈ': 'ネ', 'ﾉ': 'ノ',
	'ﾊ': 'ハ', 'ﾋ': 'ヒ', 'ﾌ': 'フ', 'ﾍ': 'ヘ', 'ﾎ': 'ホ',
	'ﾏ': 'マ', 'ﾐ': 'ミ', 'ﾑ': 'ム', 'ﾒ': 'メ', 'ﾓ': 'モ',
	'ﾔ': 'ヤ', 'ﾕ': 'ユ', 'ﾖ': 'ヨ',
	'ﾗ': 'ラ', 'ﾘ': 'リ', 'ﾙ': 'ル', 'ﾚ': 'レ', 'ﾛ': 'ロ',
	'ﾜ': 'ワ', 'ﾝ': 'ン', 'ﾞ': '゛', 'ﾟ': '゜',
	// 半角の長音記号。これを落とすと「ｼﾞｪﾗｰﾄ」が「ジェラｰト」になり、
	// 全角の「ジェラート」と一致しなくなる。
	'ｰ': 'ー',
}

// ParsePrice は「370-」「370～」「-」などを int に変換する。
//
// 解釈できない値は 0 を返す。呼び出し側は 0 を「価格が取れない行」として
// 既存値を残す扱いにするため、ここで error を返しても情報は増えない。
// ただし全角数字や先頭の「～」のような、意味が読めるのに 0 になってしまう
// 入力は、API 仕様変更のサインなので取りこぼさないよう吸収する。
func ParsePrice(raw string) int {
	s := strings.TrimSpace(raw)
	if s == "" {
		return 0
	}
	// 全角数字を半角へ寄せ、区切り記号を落とす
	var b strings.Builder
	for _, r := range s {
		switch {
		case r >= '０' && r <= '９':
			b.WriteRune(r - '０' + '0')
		case r == '-', r == '－', r == ',', r == '，':
			continue
		default:
			b.WriteRune(r)
		}
	}
	s = strings.TrimSpace(b.String())

	// 「370～500」のような範囲は下限を採る。
	// 「～370」のように先頭に来た場合は、以降の数字をそのまま読む。
	if idx := strings.IndexAny(s, "～〜"); idx > 0 {
		s = strings.TrimSpace(s[:idx])
	} else if idx == 0 {
		s = strings.TrimSpace(s[len("～"):])
	}
	if s == "" {
		return 0
	}

	n := 0
	for _, r := range s {
		if r < '0' || r > '9' {
			return 0
		}
		n = n*10 + int(r-'0')
	}
	return n
}

// Resolved は 1 行分の更新後数値。
//
// Calorie / Salt は「API から取れた値」だけを入れる。API が空や「-」を
// 返した項目は nil のままにして、書き込み時に既存値を保つ。
type Resolved struct {
	PriceWithTax int
	Calorie      *int
	Salt         *float64
}

// apiNutrition は API から実際に取れた値。
// 「API に値があったか」を保持するため、既存値でのフォールバック前の状態で持つ。
// 伝播はこちらだけを元にすることで、既存値のコピーで伝播先を壊さない。
type apiNutrition struct {
	// priceWithTax は API の税込価格。価格が「-」の行は 0。
	priceWithTax int
	calorie      *int
	salt         *float64
}

// Result はマッチングの結果。
type Result struct {
	// Values は id → 更新後数値。
	Values map[int]Resolved
	// Propagated は pre_id 伝播で埋めた id（ログ用）。
	Propagated []int
	// Ambiguous は API 候補が複数あり、一意に決められないため
	// 書き込みを見送った (正規化名, category_id) の組。
	Ambiguous []AmbiguousKey
	// Matched は API の商品と突き合わせられた DB 行数。
	Matched int
}

// AmbiguousKey は候補が複数あった突き合わせキー。
type AmbiguousKey struct {
	Name       string
	CategoryID int
}

// matchKey は突き合わせのキー。
// category_id を int のまま持つのは、文字列表現への変換で
// 2 桁以上になったときに衝突しないようにするため。
type matchKey struct {
	name       string
	categoryID int
}

// Resolve は公式データを既存行に突き合わせる。
//
// 破壊的な取り違えを避けるため、次の 2 点を守る:
//   - API 候補が複数ある行は、レスポンス順に割り当てない。
//     位置対応は順序が保証されないため、取り違えを検知できない。
//   - pre_id 伝播は「同じカテゴリかつ同じ正規化名」の行に限る。
//     pre_id はテーブルをまたいで意味が揺れており、実データでも
//     別商品が同じ pre_id を共有している（AA06: 1406 と 1425）。
func Resolve(rows []db.Row, items []fetch.Item) Result {
	// 正規化名 + カテゴリ → API エントリ（重複を保持する）
	index := map[matchKey][]fetch.Item{}
	for _, item := range items {
		id, ok := categoryIDs[item.Category]
		if !ok {
			// 未知カテゴリは 0 にして、既知のカテゴリと混ざらないようにする
			id = 0
		}
		key := matchKey{name: Normalize(item.Name), categoryID: id}
		index[key] = append(index[key], item)
	}

	values := map[int]Resolved{}
	// apiValues は「API から実際に取れた栄養値」。
	// 伝播ではこちらだけを使う（既存値のコピーで伝播先を壊さないため）。
	apiValues := map[int]apiNutrition{}
	var ambiguous []AmbiguousKey

	// 1 パス目: 完全一致
	//
	// キーごとの DB 行数を先に数えておく。DB 側に同じキーの行が複数あり
	// API 候補が 1 件しかない場合、ID 順に割り当てるのは「たまたま
	// 先頭だった行を選ぶ」だけの推測になる。
	//
	// ただし「同じ pre_id を持つ同一商品のペア」（3201/3901 など）は
	// 2 パス目で正しく揃えられるので、ここでは弾かない。弾くと
	// 「あとで」版デザートが永久に更新されなくなる。
	rowCount := map[matchKey]int{}
	for _, row := range rows {
		key := matchKey{name: Normalize(row.Name), categoryID: row.CategoryID}
		rowCount[key]++
	}

	for _, row := range rows {
		key := matchKey{name: Normalize(row.Name), categoryID: row.CategoryID}
		candidates := index[key]
		if len(candidates) == 0 {
			continue
		}
		// 候補が複数ある場合と、DB 側に同じキーの行が複数あって
		// pre_id のペアとして救済できない場合は、どれを選ぶか
		// 決められないので触らない。
		if len(candidates) > 1 || (rowCount[key] > 1 && !isPairedByPreID(rows, key)) {
			ambiguous = append(ambiguous, AmbiguousKey{
				Name:       row.Name,
				CategoryID: row.CategoryID,
			})
			continue
		}

		item := candidates[0]
		delete(index, key)

		apiPrice := ParsePrice(item.PriceWithTax)
		price := apiPrice
		if price == 0 {
			// 価格が「-」の行は価格を据え置く。ただし栄養値は
			// API から取れているので、そこは更新する。
			// 調味料・ランチの小皿・モーニングの一部が該当し、
			// 価格が無いという理由で栄養値まで捨てる理由は無い。
			price = row.PriceWithTax
		}

		calorie := parseCalorie(item.Calorie)
		salt := parseSalt(item.Salt)
		apiValues[row.ID] = apiNutrition{
			priceWithTax: apiPrice,
			calorie:      calorie,
			salt:         salt,
		}
		values[row.ID] = Resolved{
			PriceWithTax: price,
			Calorie:      calorie,
			Salt:         salt,
		}
	}

	// 2 パス目: pre_id が同じ行へ値を伝播
	//
	// 伝播は「同じ pre_id・同じカテゴリ・同じ正規化名」の行に限る。
	// pre_id だけを手がかりにすると、別商品へ値が写る（AA06 など）。
	byPreID := map[string][]db.Row{}
	for _, row := range rows {
		if row.PreID != "" {
			byPreID[row.PreID] = append(byPreID[row.PreID], row)
		}
	}

	var propagated []int
	for _, group := range byPreID {
		if len(group) < 2 {
			continue
		}
		// 同一商品とみなせる単位（カテゴリ + 正規化名）で分ける
		units := map[matchKey][]db.Row{}
		for _, row := range group {
			key := matchKey{name: Normalize(row.Name), categoryID: row.CategoryID}
			units[key] = append(units[key], row)
		}
		for _, unit := range units {
			if len(unit) < 2 {
				continue
			}
			var done *apiNutrition
			for _, row := range unit {
				if v, ok := apiValues[row.ID]; ok {
					done = &v
					break
				}
			}
			if done == nil {
				continue
			}
			for _, row := range unit {
				if _, ok := apiValues[row.ID]; !ok {
					// 伝播するのは API から取れた値だけ。
					// 既存値でフォールバックした値までコピーすると、
					// 伝播先の既存の栄養値を壊す。
					price := done.priceWithTax
					if price == 0 {
						// API に価格が無いなら、伝播先の既存価格を保つ
						price = row.PriceWithTax
					}
					values[row.ID] = Resolved{
						PriceWithTax: price,
						Calorie:      done.calorie,
						Salt:         done.salt,
					}
					propagated = append(propagated, row.ID)
				}
			}
		}
	}

	return Result{
		Values:     values,
		Propagated: propagated,
		Ambiguous:  ambiguous,
		Matched:    len(values),
	}
}

// parseCalorie は API のカロリー表記を int に変換する。
// 解釈できなければ nil を返す（呼び出し側で既存値を保つ）。
func parseCalorie(raw string) *int {
	s := strings.TrimSpace(raw)
	if s == "" || s == "-" || s == "－" {
		return nil
	}
	n := 0
	for _, r := range s {
		if r < '0' || r > '9' {
			return nil
		}
		n = n*10 + int(r-'0')
	}
	return &n
}

// parseSalt は API の食塩相当量を float64 に変換する。
// 解釈できなければ nil を返す（呼び出し側で既存値を保つ）。
func parseSalt(raw string) *float64 {
	s := strings.TrimSpace(raw)
	if s == "" || s == "-" || s == "－" {
		return nil
	}
	var whole, frac float64
	var div float64 = 1
	seenDot := false
	for _, r := range s {
		switch {
		case r == '.':
			seenDot = true
		case r >= '0' && r <= '9':
			if seenDot {
				div *= 10
				frac += float64(r-'0') / div
			} else {
				whole = whole*10 + float64(r-'0')
			}
		default:
			return nil
		}
	}
	v := whole + frac
	return &v
}

// isPairedByPreID は、同じキーの行が「同一 pre_id を持つペア」として
// 成立しているかを返す。成立していれば 2 パス目で揃えられるので、
// 1 パス目でどちらかに割り当てなくてよい。
//
// 成立条件:
//   - すべての行が空でない pre_id を持つ
//   - pre_id がすべて同じ
func isPairedByPreID(rows []db.Row, key matchKey) bool {
	var preID string
	for _, row := range rows {
		if Normalize(row.Name) != key.name || row.CategoryID != key.categoryID {
			continue
		}
		if row.PreID == "" {
			return false
		}
		if preID == "" {
			preID = row.PreID
		} else if preID != row.PreID {
			return false
		}
	}
	return preID != ""
}
