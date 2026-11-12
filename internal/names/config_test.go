package names

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolvePathOrder(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	// UserConfigDir on macOS uses HOME/Library/Application Support.
	// On Linux it may use XDG_CONFIG_HOME.
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(tmp, "config"))

	cwd := t.TempDir()
	if err := os.Chdir(cwd); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(tmp) })

	got, err := ResolvePath("")
	if err != nil {
		t.Fatal(err)
	}
	want, err := filepath.Abs("names.json")
	if err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Fatalf("default: got %q want %q", got, want)
	}

	shared := filepath.Join(tmp, "shared", "names.json")
	if err := os.MkdirAll(filepath.Dir(shared), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := SetNamesPath(shared); err != nil {
		t.Fatal(err)
	}
	got, err = ResolvePath("")
	if err != nil {
		t.Fatal(err)
	}
	if got != shared {
		t.Fatalf("config path: got %q want %q", got, shared)
	}

	flag := filepath.Join(tmp, "flag-names.json")
	got, err = ResolvePath(flag)
	if err != nil {
		t.Fatal(err)
	}
	absFlag, _ := filepath.Abs(flag)
	if got != absFlag {
		t.Fatalf("flag path: got %q want %q", got, absFlag)
	}
}
