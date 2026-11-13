// Package names loads abbreviation dictionaries and the per-user fn2fm config,
// including whether PDF Subject is arrangers or genres.
package names

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
)

// Load reads a JSON abbreviation map from path. Values are returned unchanged.
// Whitespace warnings are returned on Dictionary.Warnings (no trailing newline);
// the dictionary file is never modified. Duplicate exact keys and
// case-insensitive abbreviation conflicts are errors.
func Load(path string) (*Dictionary, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	entries, err := decodeEntries(data, path)
	if err != nil {
		return nil, err
	}
	warnings := whitespaceWarnings(path, entries)
	return &Dictionary{
		Path:     path,
		Entries:  entries,
		Hash:     contentHash(data),
		Warnings: warnings,
	}, nil
}

// LoadMap is a compatibility helper that returns the abbreviation map and
// warnings from Load.
func LoadMap(path string) (map[string]string, []string, error) {
	d, err := Load(path)
	if err != nil {
		return nil, nil, err
	}
	return d.Entries, d.Warnings, nil
}

func decodeEntries(data []byte, path string) (map[string]string, error) {
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.UseNumber()

	tok, err := dec.Token()
	if err != nil {
		return nil, fmt.Errorf("invalid %s: %w", label(path), err)
	}
	delim, ok := tok.(json.Delim)
	if !ok || delim != '{' {
		return nil, fmt.Errorf("invalid %s: top-level value must be a JSON object", label(path))
	}

	entries := make(map[string]string)
	lowerSeen := make(map[string]string) // lower(abbr) -> first exact abbr

	for dec.More() {
		keyTok, err := dec.Token()
		if err != nil {
			return nil, fmt.Errorf("invalid %s: %w", label(path), err)
		}
		key, ok := keyTok.(string)
		if !ok {
			return nil, fmt.Errorf("invalid %s: object keys must be strings", label(path))
		}
		var value string
		if err := dec.Decode(&value); err != nil {
			return nil, fmt.Errorf("invalid %s: value for %q must be a string: %w", label(path), key, err)
		}
		if _, exists := entries[key]; exists {
			return nil, fmt.Errorf("invalid %s: duplicate abbreviation %q", label(path), key)
		}
		lower := strings.ToLower(key)
		if earlier, exists := lowerSeen[lower]; exists {
			return nil, fmt.Errorf("invalid %s: abbreviations %q and %q conflict case-insensitively", label(path), earlier, key)
		}
		entries[key] = value
		lowerSeen[lower] = key
	}

	closing, err := dec.Token()
	if err != nil {
		return nil, fmt.Errorf("invalid %s: %w", label(path), err)
	}
	if closing != json.Delim('}') {
		return nil, fmt.Errorf("invalid %s: expected end of object", label(path))
	}
	// Reject trailing junk.
	if _, err := dec.Token(); err != io.EOF {
		if err == nil {
			return nil, fmt.Errorf("invalid %s: trailing data after JSON object", label(path))
		}
		return nil, fmt.Errorf("invalid %s: %w", label(path), err)
	}
	return entries, nil
}

func whitespaceWarnings(path string, entries map[string]string) []string {
	keys := make([]string, 0, len(entries))
	for k := range entries {
		keys = append(keys, k)
	}
	// Stable warning order matches previous Load behaviour (sorted abbreviations).
	sort.Strings(keys)
	var warnings []string
	for _, abbreviation := range keys {
		if abbreviation != strings.TrimSpace(abbreviation) {
			warnings = append(warnings, formatPathWarning(path, "abbreviation %q has leading or trailing whitespace", abbreviation))
		}
		name := entries[abbreviation]
		if name != strings.TrimSpace(name) {
			warnings = append(warnings, formatPathWarning(path, "name %q for abbreviation %q has leading or trailing whitespace", name, abbreviation))
		}
	}
	return warnings
}
