package names

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

const (
	minPrintASCII = '\u0020'
	maxPrintASCII = '\u007E'
	// Forbidden in abbreviations: filename metadata separators plus cloud-storage
	// characters that fn2fm already rejects in score filenames.
	abbrForbidden = "\"*./:;<>?\\|~_[]"
)

// ValidateAbbreviation reports whether abbr is a safe, non-empty filename token.
func ValidateAbbreviation(abbr string) error {
	if abbr == "" {
		return fmt.Errorf("abbreviation must not be empty")
	}
	if abbr != strings.TrimSpace(abbr) {
		return fmt.Errorf("abbreviation must not have leading or trailing whitespace")
	}
	for i, w := 0, 0; i < len(abbr); i += w {
		r, width := utf8.DecodeRuneInString(abbr[i:])
		if r == utf8.RuneError && width == 1 {
			return fmt.Errorf("abbreviation contains invalid UTF-8")
		}
		if r < minPrintASCII || r > maxPrintASCII {
			return fmt.Errorf("abbreviation must use printable ASCII only (invalid %q)", r)
		}
		if strings.ContainsRune(abbrForbidden, r) {
			return fmt.Errorf("abbreviation must not contain %q", r)
		}
		w = width
	}
	return nil
}

// AbbreviationTaken reports whether abbr collides case-insensitively with an
// existing entry, ignoring exclude (exact key) when editing.
func AbbreviationTaken(entries map[string]string, abbr, exclude string) bool {
	want := strings.ToLower(abbr)
	excludeLower := strings.ToLower(exclude)
	for existing := range entries {
		if strings.ToLower(existing) == want {
			if exclude != "" && strings.ToLower(existing) == excludeLower {
				continue
			}
			return true
		}
	}
	return false
}

// ValidateName reports basic name issues. Names may contain Unicode.
func ValidateName(name string) error {
	if strings.TrimSpace(name) == "" {
		return fmt.Errorf("name must not be empty")
	}
	return nil
}
