package names

import (
	"path/filepath"
	"testing"
)

func TestResolvePathOrder(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv(ConfigDirEnv, filepath.Join(tmp, "config"))

	cwd := t.TempDir()
	t.Chdir(cwd)

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
	if err := SetNamesPath(shared); err != nil {
		t.Fatal(err)
	}
	got, err = ResolvePath("")
	if err != nil {
		t.Fatal(err)
	}
	absShared, err := filepath.Abs(shared)
	if err != nil {
		t.Fatal(err)
	}
	if got != absShared {
		t.Fatalf("config path: got %q want %q", got, absShared)
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
