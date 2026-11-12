package names

import (
	"os"
	"path/filepath"
	"testing"
)

func TestConflictedCopies(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "names.json")
	if err := os.WriteFile(path, []byte(`{}`), 0600); err != nil {
		t.Fatal(err)
	}
	conflict := filepath.Join(dir, "names (Jonathan's conflicted copy).json")
	if err := os.WriteFile(conflict, []byte(`{}`), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path+".bak", []byte(`{}`), 0600); err != nil {
		t.Fatal(err)
	}
	got, err := ConflictedCopies(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0] != conflict {
		t.Fatalf("got %#v", got)
	}
}
