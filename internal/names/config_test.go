package names

import (
	"os"
	"path/filepath"
	"strings"
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

func TestPathUpdatesPreserveOtherFields(t *testing.T) {
	t.Setenv(ConfigDirEnv, filepath.Join(t.TempDir(), "config"))
	if err := SetUserSubject(SubjectGenres); err != nil {
		t.Fatal(err)
	}
	namesPath := filepath.Join(t.TempDir(), "names.json")
	genresPath := filepath.Join(t.TempDir(), "genres.json")
	if err := SetNamesPath(namesPath); err != nil {
		t.Fatal(err)
	}
	if err := SetGenresPath(genresPath); err != nil {
		t.Fatal(err)
	}
	cfg, err := LoadConfig()
	if err != nil {
		t.Fatal(err)
	}
	absNames, _ := filepath.Abs(namesPath)
	absGenres, _ := filepath.Abs(genresPath)
	if cfg.NamesPath != absNames || cfg.GenresPath != absGenres || cfg.Subject != SubjectGenres {
		t.Fatalf("config = %#v", cfg)
	}

	got, err := ResolveGenresPath("")
	if err != nil {
		t.Fatal(err)
	}
	if got != absGenres {
		t.Fatalf("genres path: got %q want %q", got, absGenres)
	}
	flag := filepath.Join(t.TempDir(), "flag-genres.json")
	got, err = ResolveGenresPath(flag)
	if err != nil {
		t.Fatal(err)
	}
	absFlag, _ := filepath.Abs(flag)
	if got != absFlag {
		t.Fatalf("genres flag: got %q want %q", got, absFlag)
	}
}

func TestEffectiveSubjectOrder(t *testing.T) {
	t.Setenv(ConfigDirEnv, filepath.Join(t.TempDir(), "config"))
	cwd := t.TempDir()
	t.Chdir(cwd)

	got, err := EffectiveSubject("")
	if err != nil {
		t.Fatal(err)
	}
	if got.Mode != SubjectArrangers || got.Source != SubjectFromDefault {
		t.Fatalf("default: %#v", got)
	}

	if err := SetUserSubject(SubjectGenres); err != nil {
		t.Fatal(err)
	}
	got, err = EffectiveSubject("")
	if err != nil {
		t.Fatal(err)
	}
	if got.Mode != SubjectGenres || got.Source != SubjectFromUser {
		t.Fatalf("user: %#v", got)
	}

	if err := SetLibrarySubject(SubjectArrangers); err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(filepath.Join(cwd, libraryConfigFile))
	if err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(string(raw)) != "{\n  \"subject\": \"arrangers\"\n}" {
		t.Fatalf("library file: %s", raw)
	}
	if strings.Contains(string(raw), "genresPath") || strings.Contains(string(raw), "namesPath") {
		t.Fatalf("library file should pin subject only: %s", raw)
	}
	got, err = EffectiveSubject("")
	if err != nil {
		t.Fatal(err)
	}
	if got.Mode != SubjectArrangers || got.Source != SubjectFromLibrary {
		t.Fatalf("library: %#v", got)
	}

	got, err = EffectiveSubject(SubjectGenres)
	if err != nil {
		t.Fatal(err)
	}
	if got.Mode != SubjectGenres || got.Source != SubjectFromFlag {
		t.Fatalf("flag: %#v", got)
	}

	if err := os.WriteFile(filepath.Join(cwd, libraryConfigFile), []byte(`{"subject":"nope"}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := EffectiveSubject(""); err == nil {
		t.Fatal("expected invalid library subject to fail")
	}
	if _, err := ParseSubject("both"); err == nil {
		t.Fatal("expected invalid subject to fail")
	}
}
