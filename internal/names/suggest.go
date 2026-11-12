package names

import (
	"fmt"
	"strings"
	"unicode"

	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

// Suggestion holds a primary abbreviation and optional alternatives.
type Suggestion struct {
	Primary      string
	Alternatives []string
}

// Suggest builds abbreviation candidates for name using the documented heuristic.
// Existing abbreviations are never rewritten; only new-entry suggestions collide
// with the provided map (case-insensitive).
func Suggest(name string, existing map[string]string) Suggestion {
	name = strings.TrimSpace(name)
	if name == "" {
		return Suggestion{}
	}
	people := splitPeople(name)
	var candidates []string
	if len(people) > 1 {
		candidates = appendUnique(candidates, abbreviateMany(people))
	} else if len(people) == 1 {
		words := people[0]
		candidates = appendUnique(candidates, abbreviateOne(words))
		if alt := abbreviateOneAlt(words); alt != "" {
			candidates = appendUnique(candidates, alt)
		}
	}
	if len(candidates) == 0 {
		return Suggestion{}
	}
	primary := uniquify(candidates[0], existing)
	var alts []string
	for _, c := range candidates[1:] {
		u := uniquify(c, existing)
		if u != primary {
			alts = appendUnique(alts, u)
		}
	}
	// If primary was bumped for uniqueness, also keep the raw form's next options.
	return Suggestion{Primary: primary, Alternatives: alts}
}

// SimilarNames returns existing entries that look like name. Shared given names
// alone do not trigger a match; the full normalized string or substantial token
// overlap is required.
func SimilarNames(name string, existing map[string]string, excludeAbbr string) []Entry {
	name = strings.TrimSpace(name)
	if name == "" || existing == nil {
		return nil
	}
	normName := normalizeMatch(name)
	tokens := significantTokens(name)
	var out []Entry
	for abbr, existingName := range existing {
		if excludeAbbr != "" && abbr == excludeAbbr {
			continue
		}
		if normalizeMatch(existingName) == normName {
			out = append(out, Entry{Abbreviation: abbr, Name: existingName})
			continue
		}
		if substantialOverlap(tokens, significantTokens(existingName)) {
			out = append(out, Entry{Abbreviation: abbr, Name: existingName})
		}
	}
	// Stable order by abbreviation.
	if len(out) > 1 {
		for i := 0; i < len(out); i++ {
			for j := i + 1; j < len(out); j++ {
				if out[j].Abbreviation < out[i].Abbreviation {
					out[i], out[j] = out[j], out[i]
				}
			}
		}
	}
	return out
}

func splitPeople(name string) [][]string {
	parts := strings.Split(name, ",")
	var people [][]string
	for _, p := range parts {
		w := wordsOf(p)
		if len(w) > 0 {
			people = append(people, w)
		}
	}
	return people
}

func wordsOf(s string) []string {
	fields := strings.Fields(s)
	var out []string
	for _, f := range fields {
		f = strings.Trim(f, ".")
		if f == "" {
			continue
		}
		// Drop one-letter initials (including accented letters folded later).
		runes := []rune(f)
		if len(runes) == 1 {
			continue
		}
		out = append(out, f)
	}
	return out
}

func abbreviateOne(words []string) string {
	if len(words) == 0 {
		return ""
	}
	if len(words) == 1 {
		return takeFolded(words[0], 4)
	}
	if len(words) == 2 {
		return takeFolded(words[0], 2) + takeFolded(words[1], 2)
	}
	// Three or more: initial of each leading word + first two of last.
	var b strings.Builder
	for _, w := range words[:len(words)-1] {
		b.WriteString(takeFolded(w, 1))
	}
	b.WriteString(takeFolded(words[len(words)-1], 2))
	return b.String()
}

func abbreviateOneAlt(words []string) string {
	if len(words) < 2 {
		return ""
	}
	if len(words) == 2 {
		// Alternate: first letter of each word as Upper+lower (Jr).
		alt := initialPair(words[0], words[1])
		primary := abbreviateOne(words)
		if alt == primary {
			return ""
		}
		return alt
	}
	// For 3+: offer 2+2 using first and last word.
	alt := takeFolded(words[0], 2) + takeFolded(words[len(words)-1], 2)
	if alt == abbreviateOne(words) {
		return ""
	}
	return alt
}

func abbreviateMany(people [][]string) string {
	var b strings.Builder
	for _, person := range people {
		if len(person) == 0 {
			continue
		}
		b.WriteString(initialPair(person[0], person[len(person)-1]))
	}
	return b.String()
}

// initialPair returns the first letter of first (upper) and last (lower).
func initialPair(firstWord, lastWord string) string {
	first := foldASCII(firstWord)
	last := foldASCII(lastWord)
	if first == "" || last == "" {
		return ""
	}
	fr := []rune(first)
	lr := []rune(last)
	return string([]rune{unicode.ToUpper(fr[0]), unicode.ToLower(lr[0])})
}

func takeFolded(word string, n int) string {
	folded := foldASCII(word)
	if folded == "" {
		return ""
	}
	r := []rune(folded)
	if len(r) > n {
		r = r[:n]
	}
	if len(r) == 0 {
		return ""
	}
	// Title-style: first rune upper, rest lower — matching JoRu / JSBa examples.
	out := make([]rune, len(r))
	for i, ch := range r {
		if i == 0 {
			out[i] = unicode.ToUpper(ch)
		} else {
			out[i] = unicode.ToLower(ch)
		}
	}
	return string(out)
}

func foldASCII(s string) string {
	t := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)
	result, _, err := transform.String(t, s)
	if err != nil {
		result = s
	}
	var b strings.Builder
	for _, r := range result {
		if r > unicode.MaxASCII {
			continue
		}
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(r)
		}
	}
	return b.String()
}

func uniquify(abbr string, existing map[string]string) string {
	if abbr == "" {
		return ""
	}
	if !AbbreviationTaken(existing, abbr, "") {
		return abbr
	}
	for i := 2; i < 1000; i++ {
		candidate := fmt.Sprintf("%s%d", abbr, i)
		if !AbbreviationTaken(existing, candidate, "") {
			return candidate
		}
	}
	return abbr + "X"
}

func appendUnique(list []string, s string) []string {
	if s == "" {
		return list
	}
	for _, x := range list {
		if x == s {
			return list
		}
	}
	return append(list, s)
}

func normalizeMatch(s string) string {
	s = strings.ToLower(foldASCII(s))
	s = strings.Join(strings.Fields(s), " ")
	return s
}

func significantTokens(s string) []string {
	var out []string
	for _, w := range strings.Fields(foldASCII(s)) {
		w = strings.ToLower(strings.Trim(w, "."))
		if len(w) <= 1 {
			continue
		}
		out = append(out, w)
	}
	return out
}

func substantialOverlap(a, b []string) bool {
	if len(a) == 0 || len(b) == 0 {
		return false
	}
	set := make(map[string]struct{}, len(b))
	for _, t := range b {
		set[t] = struct{}{}
	}
	shared := 0
	for _, t := range a {
		if _, ok := set[t]; ok {
			shared++
		}
	}
	// Require at least two shared tokens, or all tokens of the shorter side when
	// that side has two or more tokens. A single shared given name is ignored.
	if shared < 2 {
		return false
	}
	minLen := len(a)
	if len(b) < minLen {
		minLen = len(b)
	}
	return shared >= minLen || shared >= 2
}
