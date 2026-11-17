package fetch

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// newServer は任意のレスポンスを返すテストサーバーを立てる。
func newServer(t *testing.T, status int, body string) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		if _, err := w.Write([]byte(body)); err != nil {
			t.Errorf("write: %v", err)
		}
	}))
	t.Cleanup(srv.Close)
	return srv
}

func TestFetchSuccess(t *testing.T) {
	const payload = `{
		"lastModified": "2026-11-17",
		"body": [
			{"no": "1", "name": "小エビのサラダ", "category": "グランド",
			 "includigTaxPrice": "370-", "calorie": "196", "salt": "1.4"}
		]
	}`
	srv := newServer(t, http.StatusOK, payload)

	got, err := FetchFrom(nil, srv.URL)
	if err != nil {
		t.Fatalf("FetchFrom() error = %v", err)
	}
	if len(got.Body) != 1 {
		t.Fatalf("len(Body) = %d, want 1", len(got.Body))
	}
	if got.LastModified != "2026-11-17" {
		t.Errorf("LastModified = %q, want 2026-11-17", got.LastModified)
	}
	item := got.Body[0]
	// API 側の綴り（includigTaxPrice）を正しく読めていること
	if item.PriceWithTax != "370-" {
		t.Errorf("PriceWithTax = %q, want 370-", item.PriceWithTax)
	}
	if item.Name != "小エビのサラダ" {
		t.Errorf("Name = %q", item.Name)
	}
	if item.Calorie != "196" || item.Salt != "1.4" {
		t.Errorf("nutrition = %q / %q", item.Calorie, item.Salt)
	}
}

func TestFetchNonOKStatus(t *testing.T) {
	tests := []struct {
		name   string
		status int
	}{
		{"500", http.StatusInternalServerError},
		{"404", http.StatusNotFound},
		{"503", http.StatusServiceUnavailable},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := newServer(t, tt.status, `{"body": []}`)
			_, err := FetchFrom(nil, srv.URL)
			if err == nil {
				t.Fatal("want error, got nil")
			}
			if !strings.Contains(err.Error(), "ステータス") {
				t.Errorf("error = %v, want status message", err)
			}
		})
	}
}

func TestFetchInvalidJSON(t *testing.T) {
	srv := newServer(t, http.StatusOK, `{"body": [`)
	_, err := FetchFrom(nil, srv.URL)
	if err == nil {
		t.Fatal("want error, got nil")
	}
	if !strings.Contains(err.Error(), "パース") {
		t.Errorf("error = %v, want parse message", err)
	}
}

// {"body": null} は json.Unmarshal に成功するため、
// 明示的な空チェックが無いと「0 件で正常終了」してしまう。
func TestFetchEmptyBody(t *testing.T) {
	tests := []struct {
		name    string
		payload string
	}{
		{"empty array", `{"body": []}`},
		{"null body", `{"body": null}`},
		{"missing body", `{"lastModified": "2026-11-17"}`},
		{"empty object", `{}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := newServer(t, http.StatusOK, tt.payload)
			_, err := FetchFrom(nil, srv.URL)
			if err == nil {
				t.Fatal("want error, got nil")
			}
			if !strings.Contains(err.Error(), "空") {
				t.Errorf("error = %v, want empty-data message", err)
			}
		})
	}
}

// 商品名が空のエントリは照合できない。混じっていれば形が変わったサイン。
func TestFetchRejectsItemWithoutName(t *testing.T) {
	const payload = `{"body": [{"name": "小エビのサラダ"}, {"name": ""}]}`
	srv := newServer(t, http.StatusOK, payload)
	_, err := FetchFrom(nil, srv.URL)
	if err == nil {
		t.Fatal("want error, got nil")
	}
	if !strings.Contains(err.Error(), "商品名") {
		t.Errorf("error = %v, want name message", err)
	}
}

func TestFetchTimeout(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-r.Context().Done():
		case <-time.After(2 * time.Second):
		}
	}))
	t.Cleanup(srv.Close)

	client := &http.Client{Timeout: 50 * time.Millisecond}
	_, err := FetchFrom(client, srv.URL)
	if err == nil {
		t.Fatal("want timeout error, got nil")
	}
}

func TestFetchConnectionError(t *testing.T) {
	// 閉じたサーバーを指すと接続エラーになる
	srv := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	url := srv.URL
	srv.Close()

	_, err := FetchFrom(nil, url)
	if err == nil {
		t.Fatal("want connection error, got nil")
	}
}

// 既定クライアントにタイムアウトが入っていることを確認する。
func TestDefaultClientHasTimeout(t *testing.T) {
	if Timeout <= 0 {
		t.Errorf("Timeout = %v, want positive", Timeout)
	}
}
