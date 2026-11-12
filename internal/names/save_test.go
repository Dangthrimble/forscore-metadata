package names

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSaveAtomicAndBackup(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "names.json")
	if err := os.WriteFile(path, []byte("{\n    \"JoRu\": \"John Rutter\"\n}\n"), 0600); err != nil {
		t.Fatal(err)
	}
	d, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := d.Add("DaFo", "Dan Forrest"); err != nil {
		t.Fatal(err)
	}
	if err := d.Save(); err != nil {
		t.Fatal(err)
	}
	bak, err := os.ReadFile(path + ".bak")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(bak), "JoRu") || strings.Contains(string(bak), "DaFo") {
		t.Fatalf("backup should be previous content: %s", bak)
	}
	fresh, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if fresh.Entries["DaFo"] != "Dan Forrest" || fresh.Entries["JoRu"] != "John Rutter" {
		t.Fatalf("unexpected entries: %#v", fresh.Entries)
	}
	// Keys should be sorted in the file.
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(raw), "André") && strings.Contains(string(raw), `\u`) {
		t.Fatalf("unexpected unicode escapes: %s", raw)
	}
	dafo := strings.Index(string(raw), `"DaFo"`)
	joru := strings.Index(string(raw), `"JoRu"`)
	if dafo < 0 || joru < 0 || dafo > joru {
		t.Fatalf("keys not sorted: %s", raw)
	}
}

func TestSaveUnicodeNoEscape(t *testing.T) {
	path := filepath.Join(t.TempDir(), "names.json")
	d := NewEmpty(path)
	if err := d.Add("AnTh", "André J Thomas"); err != nil {
		t.Fatal(err)
	}
	if err := d.Save(); err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(raw), "André") {
		t.Fatalf("expected raw unicode name, got %s", raw)
	}
	if strings.Contains(string(raw), `\u00`) {
		t.Fatalf("unexpected escape: %s", raw)
	}
}

func TestSaveStaleRefused(t *testing.T) {
	path := filepath.Join(t.TempDir(), "names.json")
	if err := os.WriteFile(path, []byte(`{"JoRu":"John Rutter"}`), 0600); err != nil {
		t.Fatal(err)
	}
	d, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(`{"DaFo":"Dan Forrest"}`), 0600); err != nil {
		t.Fatal(err)
	}
	if err := d.Add("KaJe", "Karl Jenkins"); err != nil {
		t.Fatal(err)
	}
	err = d.Save()
	if err == nil || !strings.Contains(err.Error(), "changed on disk") {
		t.Fatalf("expected stale error, got %v", err)
	}
}

func TestReplacePathOverExisting(t *testing.T) {
	dir := t.TempDir()
	dst := filepath.Join(dir, "names.json")
	src := filepath.Join(dir, "new.json")
	if err := os.WriteFile(dst, []byte("old"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(src, []byte("new"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := replacePath(src, dst); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(dst)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "new" {
		t.Fatalf("got %q", got)
	}
}
