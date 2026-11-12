package names

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"
)

// Entry is one abbreviation-to-name mapping.
type Entry struct {
	Abbreviation string
	Name         string
}

// Dictionary is an in-memory names.json with the path and content hash used for
// stale-file detection before saves.
type Dictionary struct {
	Path     string
	Entries  map[string]string
	Hash     string
	Warnings []string
}

// EntriesSorted returns entries ordered by abbreviation.
func (d *Dictionary) EntriesSorted() []Entry {
	if d == nil || len(d.Entries) == 0 {
		return nil
	}
	keys := make([]string, 0, len(d.Entries))
	for k := range d.Entries {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	out := make([]Entry, 0, len(keys))
	for _, k := range keys {
		out = append(out, Entry{Abbreviation: k, Name: d.Entries[k]})
	}
	return out
}

// Search returns entries whose abbreviation or name contains query
// (case-insensitive). An empty query returns all entries sorted by abbreviation.
func (d *Dictionary) Search(query string) []Entry {
	all := d.EntriesSorted()
	q := strings.ToLower(strings.TrimSpace(query))
	if q == "" {
		return all
	}
	var out []Entry
	for _, e := range all {
		if strings.Contains(strings.ToLower(e.Abbreviation), q) ||
			strings.Contains(strings.ToLower(e.Name), q) {
			out = append(out, e)
		}
	}
	return out
}

// Lookup returns the name for an exact abbreviation key.
func (d *Dictionary) Lookup(abbreviation string) (string, bool) {
	if d == nil {
		return "", false
	}
	name, ok := d.Entries[abbreviation]
	return name, ok
}

func contentHash(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

func label(path string) string {
	if path == "" {
		return "names.json"
	}
	return path
}

func formatPathWarning(path, format string, args ...any) string {
	msg := fmt.Sprintf(format, args...)
	return fmt.Sprintf("WARNING: %s %s", label(path), msg)
}
