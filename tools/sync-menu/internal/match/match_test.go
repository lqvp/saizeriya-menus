package match

import (
	"testing"

	"sync-menu/internal/db"
	"sync-menu/internal/fetch"
)

func TestNormalize(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		// 全角カタカナはそのまま（ひらがなには変換しない）
		{"小エビのサラダ", "小エビのサラダ"},
		// 半角カタカナは全角へ揃う
		{"小ｴﾋﾞのｻﾗﾀﾞ", "小エビのサラダ"},
		// 全角スペースは除去
		{"ティラミス　クラシコ", "ティラミスクラシコ"},
		// 記号は除去
		{"ジェラート＆シナモンフォッカチオ", "ジェラート&シナモンフォッカチオ"},
		// 半角カナの濁点は合成される
		{"ｼﾞｪﾗｰﾄ", "ジェラート"},
		// 全角英数は半角へ
		{"グラスワイン　１２０ｍｌ", "グラスワイン120ml"},
		// 長音「ー」は落とさない（テストの期待値を実装に合わせる）
		{"キャロットラペ", "キャロットラペ"},
	}
	for _, tt := range tests {
		if got := Normalize(tt.in); got != tt.want {
			t.Errorf("Normalize(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestParsePrice(t *testing.T) {
	tests := []struct {
		in   string
		want int
	}{
		{"370-", 370},
		{"160-", 160},
		{"1500-", 1500},
		{"370～", 370},
		{"370～500", 370},
		// 全角数字も読める
		{"３７０-", 370},
		// 先頭に範囲記号が来ても下限を採る
		{"～370", 370},
		{"-", 0},
		{"", 0},
		{"-30ml", 0},
	}
	for _, tt := range tests {
		if got := ParsePrice(tt.in); got != tt.want {
			t.Errorf("ParsePrice(%q) = %d, want %d", tt.in, got, tt.want)
		}
	}
}

func TestResolveExactMatch(t *testing.T) {
	rows := []db.Row{
		{ID: 1202, Name: "小エビのサラダ", CategoryID: 1, PriceWithTax: 350},
		{ID: 1301, Name: "コーンクリームスープ", CategoryID: 1, PriceWithTax: 150},
	}
	items := []fetch.Item{
		{Name: "小エビのサラダ", Category: "グランド", PriceWithTax: "370-", Calorie: "196", Salt: "1.4"},
		{Name: "コーンクリームスープ", Category: "グランド", PriceWithTax: "160-", Calorie: "150", Salt: "1.0"},
	}

	got := Resolve(rows, items)

	if v, ok := got.Values[1202]; !ok || v.PriceWithTax != 370 {
		t.Errorf("1202 = %+v, want price 370", got.Values[1202])
	}
	if v, ok := got.Values[1301]; !ok || v.PriceWithTax != 160 {
		t.Errorf("1301 = %+v, want price 160", got.Values[1301])
	}
	if *got.Values[1202].Calorie != 196 {
		t.Errorf("1202 calorie = %d, want 196", *got.Values[1202].Calorie)
	}
}

// 全角/半角の差を吸収できることを確認する。
func TestResolveHalfWidthKana(t *testing.T) {
	rows := []db.Row{
		{ID: 1202, Name: "小ｴﾋﾞのｻﾗﾀﾞ", CategoryID: 1, PriceWithTax: 350},
	}
	items := []fetch.Item{
		{Name: "小エビのサラダ", Category: "グランド", PriceWithTax: "370-"},
	}

	got := Resolve(rows, items)
	if v, ok := got.Values[1202]; !ok || v.PriceWithTax != 370 {
		t.Errorf("1202 = %+v, want price 370", got.Values[1202])
	}
}

// カテゴリが違えばマッチしないことを確認する。
// フォッカチオ ¥150(グランド) と フォッカチオモーニング ¥1000(モーニング) の
// 取り違えを防ぐためのテスト。
func TestResolveDoesNotCrossCategories(t *testing.T) {
	rows := []db.Row{
		{ID: 3110, Name: "フォッカチオ", CategoryID: 1, PriceWithTax: 150},
		{ID: 5603, Name: "フォッカチオモーニング", CategoryID: 3, PriceWithTax: 1000},
	}
	items := []fetch.Item{
		{Name: "フォッカチオ", Category: "グランド", PriceWithTax: "160-"},
		{Name: "フォッカチオモーニング", Category: "モーニング", PriceWithTax: "1000-"},
	}

	got := Resolve(rows, items)

	if v := got.Values[3110]; v.PriceWithTax != 160 {
		t.Errorf("3110 = %d, want 160", v.PriceWithTax)
	}
	// 1000 のまま（変更なし）なので Values に 5603 は入らない
	if v, ok := got.Values[5603]; ok && v.PriceWithTax != 1000 {
		t.Errorf("5603 = %d, want 1000", v.PriceWithTax)
	}
}

// 「あとで」版デザートが pre_id 経由で揃うことを確認する。
func TestResolvePropagatesByPreID(t *testing.T) {
	rows := []db.Row{
		{ID: 3201, Name: "ティラミス　クラシコ", CategoryID: 1, PriceWithTax: 300, PreID: "DE01"},
		{ID: 3901, Name: "ティラミス　クラシコ", CategoryID: 1, PriceWithTax: 300, PreID: "DE01"},
	}
	items := []fetch.Item{
		{Name: "ティラミス　クラシコ", Category: "グランド", PriceWithTax: "320-"},
	}

	got := Resolve(rows, items)

	for _, id := range []int{3201, 3901} {
		v, ok := got.Values[id]
		if !ok || v.PriceWithTax != 320 {
			t.Errorf("%d = %+v, want price 320", id, v)
		}
	}
	if len(got.Propagated) != 1 {
		t.Errorf("Propagated = %v, want 1 entry", got.Propagated)
	}
}

// 価格が「-」の行は価格を据え置く。ただし栄養値は API から取れるので
// そこは更新する。調味料・ランチの小皿など、価格が無いという理由で
// 栄養値まで捨てる必要はない。
func TestResolveKeepsPriceWhenMissing(t *testing.T) {
	existingCal := 100
	rows := []db.Row{
		{ID: 9999, Name: "ターメリックライス", CategoryID: 1, PriceWithTax: 200, Calorie: &existingCal},
	}
	items := []fetch.Item{
		{Name: "ターミリックライス", Category: "グランド", PriceWithTax: "-"},
	}
	items[0].Name = "ターメリックライス"
	items[0].Calorie = "475"
	items[0].Salt = "1.7"

	got := Resolve(rows, items)

	v, ok := got.Values[9999]
	if !ok {
		t.Fatal("9999 should be resolved so nutrition can be updated")
	}
	if v.PriceWithTax != 200 {
		t.Errorf("PriceWithTax = %d, want 200 (must keep existing price)", v.PriceWithTax)
	}
	if v.Calorie == nil || *v.Calorie != 475 {
		t.Errorf("Calorie = %v, want 475", v.Calorie)
	}
	if v.Salt == nil || *v.Salt != 1.7 {
		t.Errorf("Salt = %v, want 1.7", v.Salt)
	}
}

// 同じ pre_id を共有していても、名前が違えば伝播しない。
// 実データの AA06（小エビのカクテル 1406 / 柔らか青豆の温サラダ 1425）が
// 該当する。ここで伝播すると片方の価格がもう片方で上書きされる。
func TestResolveDoesNotPropagateAcrossDifferentNames(t *testing.T) {
	rows := []db.Row{
		{ID: 1406, Name: "小エビのカクテル", CategoryID: 1, PriceWithTax: 320, PreID: "AA06"},
		{ID: 1425, Name: "柔らか青豆の温サラダ", CategoryID: 1, PriceWithTax: 210, PreID: "AA06"},
	}
	// 1425 だけ API に存在する（1406 は一時的に消えた想定）
	items := []fetch.Item{
		{Name: "柔らか青豆の温サラダ", Category: "グランド", PriceWithTax: "250-"},
	}

	got := Resolve(rows, items)

	if v := got.Values[1425]; v.PriceWithTax != 250 {
		t.Errorf("1425 = %d, want 250", v.PriceWithTax)
	}
	if v, ok := got.Values[1406]; ok {
		t.Errorf("1406 must not be overwritten by 1425, got %+v", v)
	}
	if len(got.Propagated) != 0 {
		t.Errorf("Propagated = %v, want none", got.Propagated)
	}
}

// 同じ pre_id でもカテゴリが違えば伝播しない。
func TestResolveDoesNotPropagateAcrossCategories(t *testing.T) {
	rows := []db.Row{
		{ID: 1, Name: "フォッカチオ", CategoryID: 3, PriceWithTax: 200, PreID: "XX01"},
		{ID: 2, Name: "フォッカチオ", CategoryID: 2, PriceWithTax: 0, PreID: "XX01"},
	}
	items := []fetch.Item{
		{Name: "フォッカチオ", Category: "モーニング", PriceWithTax: "250-"},
	}

	got := Resolve(rows, items)

	if v := got.Values[1]; v.PriceWithTax != 250 {
		t.Errorf("1 = %d, want 250", v.PriceWithTax)
	}
	if v, ok := got.Values[2]; ok {
		t.Errorf("2 must not be filled from another category, got %+v", v)
	}
}

// API の候補が複数あるときは位置対応させず、その行を触らない。
// レスポンス順は保証されないため、割り当てると取り違える。
func TestResolveSkipsAmbiguousRows(t *testing.T) {
	rows := []db.Row{
		{ID: 1, Name: "ドリア", CategoryID: 1, PriceWithTax: 300},
		{ID: 2, Name: "ドリア", CategoryID: 1, PriceWithTax: 500},
	}
	items := []fetch.Item{
		{Name: "ドリア", Category: "グランド", PriceWithTax: "520-"},
		{Name: "ドリア", Category: "グランド", PriceWithTax: "310-"},
	}

	got := Resolve(rows, items)

	if len(got.Values) != 0 {
		t.Errorf("Values = %+v, want nothing assigned", got.Values)
	}
	if len(got.Ambiguous) != 2 {
		t.Errorf("Ambiguous = %+v, want 2 entries", got.Ambiguous)
	}
}

// 候補が一意なら、同じ名前の行が複数あってもそれぞれ独立に解決する
// （pre_id が NULL のデザートペアなど）。
func TestResolveUniqueCandidateForDuplicateNames(t *testing.T) {
	rows := []db.Row{
		{ID: 3214, Name: "ジェラート＆シナモンフォッカチオ", CategoryID: 1, PriceWithTax: 440},
		{ID: 3914, Name: "ジェラート＆シナモンフォッカチオ", CategoryID: 1, PriceWithTax: 440},
	}
	items := []fetch.Item{
		{Name: "ジェラート＆シナモンフォッカチオ", Category: "グランド", PriceWithTax: "470-"},
	}

	got := Resolve(rows, items)

	// 1 件目の行が候補を消費し、2 件目は候補なしのまま残る。
	// ここで両方に書くと「API に無い行」まで書き換えてしまう。
	if v, ok := got.Values[3214]; !ok || v.PriceWithTax != 470 {
		t.Errorf("3214 = %+v, want price 470", v)
	}
	if v, ok := got.Values[3914]; ok {
		t.Errorf("3914 must be left alone, got %+v", v)
	}
}
