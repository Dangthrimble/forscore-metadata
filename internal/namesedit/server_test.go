package namesedit

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"fn2fm/internal/names"
)

func TestEditorAddAndSuggest(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "names.json")
	dict := names.NewEmpty(path)
	if err := dict.Save(); err != nil {
		t.Fatal(err)
	}
	dict, err := names.Load(path)
	if err != nil {
		t.Fatal(err)
	}
	s := &server{dict: dict, token: "test-token"}
	handler := s.routes()

	suggestBody, _ := json.Marshal(map[string]string{"name": "John Rutter"})
	req := httptest.NewRequest(http.MethodPost, "/api/suggest?token=test-token", bytes.NewReader(suggestBody))
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != 200 {
		t.Fatalf("suggest status %d: %s", rec.Code, rec.Body.String())
	}
	var sug suggestResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &sug); err != nil {
		t.Fatal(err)
	}
	if sug.Suggestion.Primary != "JoRu" {
		t.Fatalf("suggest %#v", sug)
	}

	saveBody, _ := json.Marshal(saveRequest{Action: "add", Abbr: "JoRu", Name: "John Rutter"})
	req = httptest.NewRequest(http.MethodPost, "/api/save?token=test-token", bytes.NewReader(saveBody))
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != 200 {
		t.Fatalf("save status %d: %s", rec.Code, rec.Body.String())
	}
	var saved saveResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &saved); err != nil {
		t.Fatal(err)
	}
	if !saved.OK {
		t.Fatalf("save: %#v", saved)
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(raw, []byte("John Rutter")) {
		t.Fatalf("file content: %s", raw)
	}

	req = httptest.NewRequest(http.MethodGet, "/api/state?token=bad", nil)
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestEditorStaleSave(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "names.json")
	if err := os.WriteFile(path, []byte(`{"JoRu":"John Rutter"}`), 0600); err != nil {
		t.Fatal(err)
	}
	dict, err := names.Load(path)
	if err != nil {
		t.Fatal(err)
	}
	s := &server{dict: dict, token: "tok"}
	if err := os.WriteFile(path, []byte(`{"DaFo":"Dan Forrest"}`), 0600); err != nil {
		t.Fatal(err)
	}
	body, _ := json.Marshal(saveRequest{Action: "add", Abbr: "KaJe", Name: "Karl Jenkins"})
	req := httptest.NewRequest(http.MethodPost, "/api/save?token=tok", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	s.routes().ServeHTTP(rec, req)
	var saved saveResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &saved)
	if !saved.Stale {
		t.Fatalf("expected stale, got %#v", saved)
	}
	if _, ok := dict.Entries["KaJe"]; ok {
		t.Fatal("stale save should not keep in-memory add")
	}
}
