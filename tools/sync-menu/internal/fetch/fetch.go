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

// Timeout は API 呼び出しの制限時間。
const Timeout = 30 * time.Second

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
//
// client が nil なら Timeout 付きの既定クライアントを使う。
func Fetch(client *http.Client) (*Response, error) {
	return FetchFrom(client, Endpoint)
}

// FetchFrom は任意の URL からメニューデータを取得する。
// テストで httptest サーバーを指すために切り出している。
func FetchFrom(client *http.Client, url string) (*Response, error) {
	if client == nil {
		client = &http.Client{Timeout: Timeout}
	}

	resp, err := client.Get(url)
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

	// 空のレスポンスで DB を壊さないよう、ここで弾く。
	// json.Unmarshal は {"body": null} も成功するため、空チェックは必須。
	if len(parsed.Body) == 0 {
		return nil, fmt.Errorf("メニュー API が空のデータを返しました")
	}

	// 商品名が空のエントリは照合に使えない。混じっていれば
	// レスポンスの形が変わったサインなので失敗させる。
	for i, item := range parsed.Body {
		if item.Name == "" {
			return nil, fmt.Errorf("メニュー API の %d 件目に商品名がありません", i+1)
		}
	}

	return &parsed, nil
}
