// Package fetch は公式サイトからメニューデータを取得する。
package fetch

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Endpoint はアレルギー情報 API の URL。
// この API はアレルゲン表示の対象商品のみを返すため、
// 酒類・ドリンクバーなど一部の商品は含まれない。
const Endpoint = "https://allergy.saizeriya.co.jp/view?lg=ja"

// Item は API が返す 1 商品分のデータ。
type Item struct {
	No           string `json:"no"`
	Name         string `json:"name"`
	Category     string `json:"category"`
	PriceWithTax string `json:"includigTaxPrice"` // API 側の綴りが "includig"
	Calorie      string `json:"calorie"`
	Salt         string `json:"salt"`
}

// Response は API のレスポンス全体。
type Response struct {
	Body         []Item `json:"body"`
	LastModified string `json:"lastModified"`
}

// Fetch はメニューデータを取得する。
// ネットワークエラーや想定外のレスポンスはすべてエラーとして返す。
func Fetch(client *http.Client) (*Response, error) {
	if client == nil {
		client = &http.Client{Timeout: 30 * time.Second}
	}

	resp, err := client.Get(Endpoint)
	if err != nil {
		return nil, fmt.Errorf("メニュー API の取得に失敗しました: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("メニュー API がステータス %d を返しました", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("レスポンスの読み込みに失敗しました: %w", err)
	}

	var parsed Response
	if err := json.Unmarshal(body, &parsed); err != nil {
		return nil, fmt.Errorf("レスポンスのパースに失敗しました: %w", err)
	}

	// 空のレスポンスで DB を壊さないよう、ここで弾く
	if len(parsed.Body) == 0 {
		return nil, fmt.Errorf("メニュー API が空のデータを返しました")
	}

	return &parsed, nil
}
