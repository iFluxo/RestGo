package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"restgo/internal/store"
)

func newTestHandler() http.Handler {
	return New(store.New())
}

func do(t *testing.T, h http.Handler, method, path, body string) *httptest.ResponseRecorder {
	t.Helper()

	req := httptest.NewRequest(method, path, strings.NewReader(body))
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func TestCRUDFlow(t *testing.T) {
	h := newTestHandler()

	rec := do(t, h, http.MethodPost, "/items", `{"name":"Sweet Tea","price":5000}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create: status = %d, want %d (body: %s)", rec.Code, http.StatusCreated, rec.Body)
	}

	var created store.Item
	if err := json.NewDecoder(rec.Body).Decode(&created); err != nil {
		t.Fatalf("decode create: %v", err)
	}
	if created.ID == 0 || created.Name != "Sweet Tea" || created.Price != 5000 {
		t.Fatalf("create: unexpected item: %+v", created)
	}
	if got := rec.Header().Get("Location"); got != "/items/1" {
		t.Fatalf("create: Location = %q, want /items/1", got)
	}

	rec = do(t, h, http.MethodGet, "/items/1", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("get: status = %d, want %d", rec.Code, http.StatusOK)
	}

	rec = do(t, h, http.MethodPut, "/items/1", `{"name":"Plain Tea","price":4000}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("update: status = %d, want %d", rec.Code, http.StatusOK)
	}
	var updated store.Item
	if err := json.NewDecoder(rec.Body).Decode(&updated); err != nil {
		t.Fatalf("decode update: %v", err)
	}
	if updated.Name != "Plain Tea" || updated.Price != 4000 {
		t.Fatalf("update: unexpected item: %+v", updated)
	}

	rec = do(t, h, http.MethodGet, "/items", "")
	var list []store.Item
	if err := json.NewDecoder(rec.Body).Decode(&list); err != nil {
		t.Fatalf("decode list: %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("list: got %d items, want 1", len(list))
	}

	rec = do(t, h, http.MethodDelete, "/items/1", "")
	if rec.Code != http.StatusNoContent {
		t.Fatalf("delete: status = %d, want %d", rec.Code, http.StatusNoContent)
	}

	rec = do(t, h, http.MethodGet, "/items/1", "")
	if rec.Code != http.StatusNotFound {
		t.Fatalf("get after delete: status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}

func TestValidation(t *testing.T) {
	h := newTestHandler()

	tests := []struct {
		name string
		body string
	}{
		{"empty name", `{"name":"  ","price":1000}`},
		{"name too long", `{"name":"` + strings.Repeat("a", 101) + `","price":1000}`},
		{"negative price", `{"name":"Tea","price":-1}`},
		{"unknown field", `{"name":"Tea","price":1000,"foo":"bar"}`},
		{"not JSON", `not json`},
		{"empty body", ``},
		{"two JSON objects", `{"name":"Tea","price":1000}{"name":"Coffee","price":2000}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := do(t, h, http.MethodPost, "/items", tt.body)
			if rec.Code != http.StatusBadRequest {
				t.Fatalf("status = %d, want %d (body: %s)", rec.Code, http.StatusBadRequest, rec.Body)
			}
		})
	}
}

func TestInvalidID(t *testing.T) {
	h := newTestHandler()

	for _, id := range []string{"abc", "0", "-5"} {
		rec := do(t, h, http.MethodGet, "/items/"+id, "")
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("id %q: status = %d, want %d", id, rec.Code, http.StatusBadRequest)
		}
	}
}

func TestRoot(t *testing.T) {
	h := newTestHandler()

	rec := do(t, h, http.MethodGet, "/", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if got := rec.Body.String(); !strings.Contains(got, `"status":"ok"`) {
		t.Fatalf("body = %q, want status ok", got)
	}

	rec = do(t, h, http.MethodGet, "/unknown", "")
	if rec.Code != http.StatusNotFound {
		t.Fatalf("unknown path: status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}

func TestHealth(t *testing.T) {
	h := newTestHandler()

	rec := do(t, h, http.MethodGet, "/health", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	var body struct {
		Status string `json:"status"`
		Uptime int64  `json:"uptime"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode health: %v", err)
	}
	if body.Status != "ok" || body.Uptime < 0 {
		t.Fatalf("health = %+v, want status ok and non-negative uptime", body)
	}
}
