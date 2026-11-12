package names

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestLoadWhitespaceWarnings(t *testing.T) {
	for _, tc := range []struct {
		name         string
		entries      map[string]string
		warningCount int
		fragments    []string
	}{
		{"clean and internal spaces", map[string]string{"DaFo": "Dan Forrest", "A B": "Two Names"}, 0, nil},
		{"abbreviation edges", map[string]string{" JoRu": "John Rutter", "DaFo ": "Dan Forrest"}, 2, []string{`abbreviation " JoRu"`, `abbreviation "DaFo "`}},
		{"name edges", map[string]string{"JoRu": "John Rutter ", "DaFo": " Dan Forrest"}, 2, []string{`name "John Rutter " for abbreviation "JoRu"`, `name " Dan Forrest" for abbreviation "DaFo"`}},
		{"other whitespace and both fields", map[string]string{"\tDaFo": "Dan Forrest\n", "JoRu": "\u00a0John Rutter"}, 3, []string{`abbreviation "\tDaFo"`, `name "Dan Forrest\n"`, `for abbreviation "JoRu"`}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			contents, err := json.Marshal(tc.entries)
			if err != nil {
				t.Fatal(err)
			}
			path := filepath.Join(t.TempDir(), "names.json")
			if err := os.WriteFile(path, contents, 0600); err != nil {
				t.Fatal(err)
			}
			got, err := Load(path)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(got.Entries, tc.entries) {
				t.Fatalf("dictionary values changed: got %#v, want %#v", got.Entries, tc.entries)
			}
			joined := strings.Join(got.Warnings, "\n")
			warningMarker := "WARNING: " + path
			if count := strings.Count(joined, warningMarker); count != tc.warningCount {
				t.Fatalf("got %d warnings, want %d: %s", count, tc.warningCount, joined)
			}
			if len(got.Warnings) != tc.warningCount {
				t.Fatalf("got %d warning lines, want %d: %#v", len(got.Warnings), tc.warningCount, got.Warnings)
			}
			for _, fragment := range tc.fragments {
				if !strings.Contains(joined, fragment) {
					t.Errorf("missing %q in warnings: %s", fragment, joined)
				}
			}
			after, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(after, contents) {
				t.Fatal("dictionary file changed")
			}
		})
	}
}

func TestLoadDuplicateAndCaseConflicts(t *testing.T) {
	dir := t.TempDir()
	dupPath := filepath.Join(dir, "dup.json")
	// Raw JSON with a duplicate key; encoding/json would keep the last value.
	if err := os.WriteFile(dupPath, []byte(`{"JoRu":"A","JoRu":"B"}`), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := Load(dupPath); err == nil || !strings.Contains(err.Error(), "duplicate abbreviation") {
		t.Fatalf("expected duplicate error, got %v", err)
	}

	casePath := filepath.Join(dir, "case.json")
	if err := os.WriteFile(casePath, []byte(`{"JoRu":"John Rutter","joru":"Other"}`), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := Load(casePath); err == nil || !strings.Contains(err.Error(), "conflict case-insensitively") {
		t.Fatalf("expected case conflict, got %v", err)
	}
}

func TestLoadRejectsNonObject(t *testing.T) {
	path := filepath.Join(t.TempDir(), "names.json")
	if err := os.WriteFile(path, []byte(`["JoRu"]`), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := Load(path); err == nil || !strings.Contains(err.Error(), "JSON object") {
		t.Fatalf("expected object error, got %v", err)
	}
}
