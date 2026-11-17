// Package match は既存の DB 行と公式データを突き合わせる。
//
// マッチングの方針:
//  1. 「正規化した商品名 + カテゴリ」の完全一致
//  2. 1 で漏れた行が「あとで」版デザート等で、同じ pre_id のペアが
//     解決済みなら、その値をコピーする
//
// 2 が必要な理由:
//   「ティラミス　クラシコ」は「すぐに」(3201) と「あとで」(3901) の
//   2 行あるが、公式 API の商品名は同一でカテゴリも同じため、1 の完全一致
//   では片方しかマッチしない。pre_id（DE01）が同じなら同一商品なので揃える。
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
var categoryIDs = map[string]int{
	"グランド":  1,
	"ランチ":   2,
	"モーニング": 3,
	"キッズ":   4,
	"テイクアウト": 5,
	"季節":    6,
	"季節限定":  6,
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
}

// ParsePrice は「370-」「370～」「-」などを int に変換する。
func ParsePrice(raw string) int {
	if raw == "" {
		return 0
	}
	s := strings.NewReplacer("-", "", "－", "", ",", "").Replace(raw)
	s = strings.TrimSpace(s)
	if idx := strings.Index(s, "～"); idx != -1 {
		s = strings.TrimSpace(s[:idx])
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
type Resolved struct {
	PriceWithTax int
	Calorie      *int
	Salt         *float64
}

// Result はマッチングの結果。
type Result struct {
	// Values は id → 更新後数値。
	Values map[int]Resolved
	// Propagated は pre_id 伝播で埋めた id（ログ用）。
	Propagated []int
}

// Resolve は公式データを既存行に突き合わせる。
func Resolve(rows []db.Row, items []fetch.Item) Result {
	// (正規化名, category_id) → 公式エントリ
	index := map[[2]string][]fetch.Item{}
	for _, item := range items {
		key := [2]string{Normalize(item.Name), ""}
		if id, ok := categoryIDs[item.Category]; ok {
			key[1] = string(rune('0' + id))
		}
		index[key] = append(index[key], item)
	}

	values := map[int]Resolved{}

	// 1 パス目: 完全一致
	for _, row := range rows {
		key := [2]string{Normalize(row.Name), string(rune('0' + row.CategoryID))}
		candidates := index[key]
		if len(candidates) == 0 {
			continue
		}
		item := candidates[0]
		index[key] = candidates[1:]

		price := ParsePrice(item.PriceWithTax)
		if price == 0 {
			continue // 価格が取れない行（「-」表記）は既存値を残す
		}
		values[row.ID] = Resolved{
			PriceWithTax: price,
			Calorie:      parseCalorie(item.Calorie, row.Calorie),
			Salt:         parseSalt(item.Salt, row.Salt),
		}
	}

	// 2 パス目: pre_id が同じ行へ値を伝播
	byPreID := map[string][]int{}
	for _, row := range rows {
		if row.PreID != "" {
			byPreID[row.PreID] = append(byPreID[row.PreID], row.ID)
		}
	}

	var propagated []int
	for _, ids := range byPreID {
		if len(ids) < 2 {
			continue
		}
		var done *Resolved
		for _, id := range ids {
			if v, ok := values[id]; ok {
				done = &v
				break
			}
		}
		if done == nil {
			continue
		}
		for _, id := range ids {
			if _, ok := values[id]; !ok {
				values[id] = *done
				propagated = append(propagated, id)
			}
		}
	}

	return Result{Values: values, Propagated: propagated}
}

func parseCalorie(raw string, fallback *int) *int {
	if raw == "" || raw == "-" || raw == "－" {
		return fallback
	}
	n := 0
	for _, r := range strings.TrimSpace(raw) {
		if r < '0' || r > '9' {
			return fallback
		}
		n = n*10 + int(r-'0')
	}
	return &n
}

func parseSalt(raw string, fallback *float64) *float64 {
	if raw == "" || raw == "-" || raw == "－" {
		return fallback
	}
	var whole, frac float64
	var div float64 = 1
	seenDot := false
	for _, r := range strings.TrimSpace(raw) {
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
			return fallback
		}
	}
	v := whole + frac
	return &v
}
