package names

import (
	"path/filepath"
	"testing"
)

func TestCRUDAndSearch(t *testing.T) {
	d := NewEmpty(filepath.Join(t.TempDir(), "names.json"))
	if err := d.Add("JoRu", "John Rutter"); err != nil {
		t.Fatal(err)
	}
	if err := d.Add("joru", "Other"); err == nil {
		t.Fatal("expected case-insensitive conflict")
	}
	if err := d.Add("Bad~", "Name"); err == nil {
		t.Fatal("expected invalid abbreviation")
	}
	if err := d.Update("JoRu", "JoRu", " John Rutter "); err != nil {
		t.Fatal(err)
	}
	if d.Entries["JoRu"] != "John Rutter" {
		t.Fatalf("name not trimmed: %q", d.Entries["JoRu"])
	}
	if err := d.Update("JoRu", "DaFo", "Dan Forrest"); err != nil {
		t.Fatal(err)
	}
	if _, ok := d.Entries["JoRu"]; ok {
		t.Fatal("old abbreviation remained")
	}
	if err := d.Add("CtEcMr", "Chris Tomlin, Ed Cash, Matt Redman"); err != nil {
		t.Fatal(err)
	}
	got := d.Search("tomlin")
	if len(got) != 1 || got[0].Abbreviation != "CtEcMr" {
		t.Fatalf("search: %#v", got)
	}
	got = d.Search("dafo")
	if len(got) != 1 || got[0].Name != "Dan Forrest" {
		t.Fatalf("search abbr: %#v", got)
	}
	if err := d.Delete("DaFo"); err != nil {
		t.Fatal(err)
	}
	if err := d.Delete("DaFo"); err == nil {
		t.Fatal("expected missing delete error")
	}
}

func TestValidateAbbreviation(t *testing.T) {
	for _, bad := range []string{"", " JoRu", "JoRu ", "Jo~Ru", "Jo_Ru", "Jo[Ru", "JoRu]", "Jo/Ru", "JöRu"} {
		if err := ValidateAbbreviation(bad); err == nil {
			t.Fatalf("expected error for %q", bad)
		}
	}
	if err := ValidateAbbreviation("JoRu"); err != nil {
		t.Fatal(err)
	}
}
