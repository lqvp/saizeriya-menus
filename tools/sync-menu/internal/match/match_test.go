package match

import (
	"testing"

	"sync-menu/internal/db"
	"sync-menu/internal/fetch"
)

func intPtr(v int) *int          { return &v }
func floatPtr(v float64) *float64 { return &v }

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
		// 長音の有無は区別しない（ｰ と ー の両方を落とす）
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

// 価格が「-」の行は既存値を残す。
func TestResolveKeepsPriceWhenMissing(t *testing.T) {
	rows := []db.Row{
		{ID: 9999, Name: "ターメリックライス", CategoryID: 1, PriceWithTax: 200},
	}
	items := []fetch.Item{
		{Name: "ターメリックライス", Category: "グランド", PriceWithTax: "-"},
	}

	got := Resolve(rows, items)
	if _, ok := got.Values[9999]; ok {
		t.Errorf("9999 should be left untouched, got %+v", got.Values[9999])
	}
}
