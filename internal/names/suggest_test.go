package names

import (
	"reflect"
	"testing"
)

func TestSuggestExamples(t *testing.T) {
	for _, tc := range []struct {
		name    string
		want    string
		wantAlt string
	}{
		{"John Rutter", "JoRu", "Jr"},
		{"Johann Sebastian Bach", "JSBa", "JoBa"},
		{"Chris Tomlin, Ed Cash, Matt Redman", "CtEcMr", ""},
		{"André J Thomas", "AnTh", ""}, // one-letter J dropped
		{"Dan Forrest", "DaFo", "Df"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s := Suggest(tc.name, nil)
			if s.Primary != tc.want {
				t.Fatalf("primary: got %q want %q (alts %#v)", s.Primary, tc.want, s.Alternatives)
			}
			if tc.wantAlt != "" {
				found := false
				for _, a := range s.Alternatives {
					if a == tc.wantAlt {
						found = true
						break
					}
				}
				if !found {
					t.Fatalf("missing alt %q in %#v", tc.wantAlt, s.Alternatives)
				}
			}
		})
	}
}

func TestSuggestUniqueness(t *testing.T) {
	existing := map[string]string{"JoRu": "John Rutter"}
	s := Suggest("John Rutter", existing)
	if s.Primary != "JoRu2" {
		t.Fatalf("got %q", s.Primary)
	}
}

func TestSimilarNames(t *testing.T) {
	existing := map[string]string{
		"JoRu":  "John Rutter",
		"DaFo":  "Dan Forrest",
		"JoSt":  "John Stainer",
		"CtEcMr": "Chris Tomlin, Ed Cash, Matt Redman",
	}
	got := SimilarNames("John Rutter", existing, "")
	if len(got) != 1 || got[0].Abbreviation != "JoRu" {
		t.Fatalf("exact: %#v", got)
	}
	// Shared given name alone must not warn.
	got = SimilarNames("John Smith", existing, "")
	if len(got) != 0 {
		t.Fatalf("shared given name should not match: %#v", got)
	}
	got = SimilarNames("Chris Tomlin, Ed Cash, Matt Redman", existing, "CtEcMr")
	if len(got) != 0 {
		t.Fatalf("exclude: %#v", got)
	}
	got = SimilarNames("Dan Forrest", existing, "")
	if !reflect.DeepEqual(got, []Entry{{"DaFo", "Dan Forrest"}}) {
		t.Fatalf("%#v", got)
	}
}
