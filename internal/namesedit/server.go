package namesedit

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"fn2fm/internal/names"
)

type server struct {
	mu   sync.Mutex
	dict *names.Dictionary
	token string
}

func (s *server) routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/", s.handleIndex)
	mux.HandleFunc("/api/state", s.withAuth(s.handleState))
	mux.HandleFunc("/api/suggest", s.withAuth(s.handleSuggest))
	mux.HandleFunc("/api/save", s.withAuth(s.handleSave))
	mux.HandleFunc("/api/reload", s.withAuth(s.handleReload))
	mux.HandleFunc("/api/open", s.withAuth(s.handleOpen))
	mux.HandleFunc("/api/new", s.withAuth(s.handleNew))
	mux.HandleFunc("/api/use", s.withAuth(s.handleUse))
	mux.HandleFunc("/api/pick-open", s.withAuth(s.handlePickOpen))
	mux.HandleFunc("/api/pick-new", s.withAuth(s.handlePickNew))
	return mux
}

func (s *server) withAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tok := r.URL.Query().Get("token")
		if tok == "" {
			tok = r.Header.Get("X-FN2FM-Token")
		}
		if tok != s.token {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		next(w, r)
	}
}

func (s *server) handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	if r.URL.Query().Get("token") != s.token {
		http.Error(w, "unauthorized: open the URL printed in the terminal", http.StatusUnauthorized)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write(indexHTML)
}

type stateResponse struct {
	Path       string        `json:"path"`
	Entries    []names.Entry `json:"entries"`
	Warnings   []string      `json:"warnings"`
	Conflicts  []string      `json:"conflicts"`
	UsingSaved bool          `json:"usingSaved"`
	Stale      bool          `json:"stale"`
}

func (s *server) handleState(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()
	stale, _ := s.dict.IsStale()
	conflicts, _ := names.ConflictedCopies(s.dict.Path)
	cfg, _ := names.LoadConfig()
	using := cfg != nil && cfg.NamesPath != "" && sameFilePath(cfg.NamesPath, s.dict.Path)
	writeJSON(w, stateResponse{
		Path:       s.dict.Path,
		Entries:    s.dict.EntriesSorted(),
		Warnings:   s.dict.Warnings,
		Conflicts:  conflicts,
		UsingSaved: using,
		Stale:      stale,
	})
}

type suggestRequest struct {
	Name         string `json:"name"`
	ExcludeAbbr  string `json:"excludeAbbr"`
}

type suggestResponse struct {
	Suggestion names.Suggestion `json:"suggestion"`
	Similar    []names.Entry    `json:"similar"`
	Taken      bool             `json:"taken"`
	Abbr       string           `json:"abbr"`
}

func (s *server) handleSuggest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req suggestRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	sug := names.Suggest(req.Name, s.dict.Entries)
	similar := names.SimilarNames(req.Name, s.dict.Entries, req.ExcludeAbbr)
	writeJSON(w, suggestResponse{
		Suggestion: sug,
		Similar:    similar,
		Taken:      names.AbbreviationTaken(s.dict.Entries, sug.Primary, req.ExcludeAbbr),
		Abbr:       sug.Primary,
	})
}

type saveRequest struct {
	Action string `json:"action"` // add | update | delete
	OldAbbr string `json:"oldAbbr"`
	Abbr   string `json:"abbr"`
	Name   string `json:"name"`
}

type saveResponse struct {
	OK      bool   `json:"ok"`
	Error   string `json:"error,omitempty"`
	Stale   bool   `json:"stale,omitempty"`
	Message string `json:"message,omitempty"`
}

func (s *server) handleSave(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req saveRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	stale, err := s.dict.IsStale()
	if err != nil {
		writeJSON(w, saveResponse{Error: err.Error()})
		return
	}
	if stale {
		writeJSON(w, saveResponse{Stale: true, Error: "The dictionary file changed on disk. Reload before saving."})
		return
	}

	snapshot := cloneEntries(s.dict.Entries)
	switch req.Action {
	case "add":
		if err := s.dict.Add(req.Abbr, req.Name); err != nil {
			writeJSON(w, saveResponse{Error: err.Error()})
			return
		}
	case "update":
		if err := s.dict.Update(req.OldAbbr, req.Abbr, req.Name); err != nil {
			writeJSON(w, saveResponse{Error: err.Error()})
			return
		}
	case "delete":
		if err := s.dict.Delete(req.OldAbbr); err != nil {
			writeJSON(w, saveResponse{Error: err.Error()})
			return
		}
	default:
		writeJSON(w, saveResponse{Error: "unknown action"})
		return
	}

	if err := s.dict.Save(); err != nil {
		s.dict.Entries = snapshot
		if strings.Contains(err.Error(), "changed on disk") {
			writeJSON(w, saveResponse{Stale: true, Error: err.Error()})
			return
		}
		writeJSON(w, saveResponse{Error: err.Error()})
		return
	}
	msg := "Saved."
	if req.Action == "update" && req.OldAbbr != strings.TrimSpace(req.Abbr) {
		msg = "Saved. Existing PDF filenames that used the old abbreviation will not match until renamed."
	}
	if req.Action == "delete" {
		msg = "Deleted. Existing PDF filenames that used this abbreviation will not match until renamed."
	}
	writeJSON(w, saveResponse{OK: true, Message: msg})
}

func (s *server) handleReload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, err := os.Stat(s.dict.Path); os.IsNotExist(err) {
		s.dict = names.NewEmpty(s.dict.Path)
		writeJSON(w, map[string]any{"ok": true})
		return
	}
	if err := s.dict.Reload(); err != nil {
		writeJSON(w, map[string]any{"ok": false, "error": err.Error()})
		return
	}
	writeJSON(w, map[string]any{"ok": true})
}

type pathRequest struct {
	Path string `json:"path"`
}

func (s *server) handleOpen(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req pathRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	path := strings.TrimSpace(req.Path)
	if path == "" {
		writeJSON(w, map[string]any{"ok": false, "error": "path required"})
		return
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		writeJSON(w, map[string]any{"ok": false, "error": err.Error()})
		return
	}
	dict, err := names.Load(abs)
	if err != nil {
		writeJSON(w, map[string]any{"ok": false, "error": err.Error()})
		return
	}
	s.mu.Lock()
	s.dict = dict
	s.mu.Unlock()
	writeJSON(w, map[string]any{"ok": true})
}

func (s *server) handleNew(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req pathRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	path := strings.TrimSpace(req.Path)
	if path == "" {
		writeJSON(w, map[string]any{"ok": false, "error": "path required"})
		return
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		writeJSON(w, map[string]any{"ok": false, "error": err.Error()})
		return
	}
	if _, err := os.Stat(abs); err == nil {
		writeJSON(w, map[string]any{"ok": false, "error": "file already exists; open it instead"})
		return
	}
	dict := names.NewEmpty(abs)
	if err := dict.Save(); err != nil {
		writeJSON(w, map[string]any{"ok": false, "error": err.Error()})
		return
	}
	s.mu.Lock()
	s.dict = dict
	s.mu.Unlock()
	writeJSON(w, map[string]any{"ok": true})
}

func (s *server) handleUse(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	s.mu.Lock()
	path := s.dict.Path
	s.mu.Unlock()
	if err := names.SetNamesPath(path); err != nil {
		writeJSON(w, map[string]any{"ok": false, "error": err.Error()})
		return
	}
	writeJSON(w, map[string]any{"ok": true, "path": path})
}

func (s *server) handlePickOpen(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	path, err := pickDictionaryPath()
	if err != nil {
		writeJSON(w, map[string]any{"ok": false, "error": err.Error(), "needTypedPath": true})
		return
	}
	if path == "" {
		writeJSON(w, map[string]any{"ok": false, "cancelled": true})
		return
	}
	writeJSON(w, map[string]any{"ok": true, "path": path})
}

func (s *server) handlePickNew(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	path, err := pickSaveDictionaryPath()
	if err != nil {
		writeJSON(w, map[string]any{"ok": false, "error": err.Error(), "needTypedPath": true})
		return
	}
	if path == "" {
		writeJSON(w, map[string]any{"ok": false, "cancelled": true})
		return
	}
	writeJSON(w, map[string]any{"ok": true, "path": path})
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(true)
	_ = enc.Encode(v)
}

func sameFilePath(a, b string) bool {
	a = filepath.Clean(a)
	b = filepath.Clean(b)
	return strings.EqualFold(a, b) || a == b
}

func cloneEntries(in map[string]string) map[string]string {
	out := make(map[string]string, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}
