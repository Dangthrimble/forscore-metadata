package names

import (
	"fmt"
	"strings"
)

// Add inserts a new abbreviation. abbr and name are trimmed; existing
// abbreviations are never altered automatically.
func (d *Dictionary) Add(abbr, name string) error {
	if d == nil {
		return fmt.Errorf("nil dictionary")
	}
	if d.Entries == nil {
		d.Entries = map[string]string{}
	}
	abbr = strings.TrimSpace(abbr)
	name = strings.TrimSpace(name)
	if err := ValidateAbbreviation(abbr); err != nil {
		return err
	}
	if err := ValidateName(name); err != nil {
		return err
	}
	if AbbreviationTaken(d.Entries, abbr, "") {
		return fmt.Errorf("abbreviation %q is already used", abbr)
	}
	d.Entries[abbr] = name
	return nil
}

// Update changes an existing entry. When newAbbr differs from oldAbbr, the old
// key is removed. Existing PDF filenames are not rewritten.
func (d *Dictionary) Update(oldAbbr, newAbbr, name string) error {
	if d == nil {
		return fmt.Errorf("nil dictionary")
	}
	if d.Entries == nil {
		return fmt.Errorf("abbreviation %q not found", oldAbbr)
	}
	if _, ok := d.Entries[oldAbbr]; !ok {
		return fmt.Errorf("abbreviation %q not found", oldAbbr)
	}
	newAbbr = strings.TrimSpace(newAbbr)
	name = strings.TrimSpace(name)
	if err := ValidateAbbreviation(newAbbr); err != nil {
		return err
	}
	if err := ValidateName(name); err != nil {
		return err
	}
	if AbbreviationTaken(d.Entries, newAbbr, oldAbbr) {
		return fmt.Errorf("abbreviation %q is already used", newAbbr)
	}
	if newAbbr != oldAbbr {
		delete(d.Entries, oldAbbr)
	}
	d.Entries[newAbbr] = name
	return nil
}

// Delete removes an abbreviation. Existing PDF filenames are not rewritten.
func (d *Dictionary) Delete(abbr string) error {
	if d == nil {
		return fmt.Errorf("nil dictionary")
	}
	if _, ok := d.Entries[abbr]; !ok {
		return fmt.Errorf("abbreviation %q not found", abbr)
	}
	delete(d.Entries, abbr)
	return nil
}

// Reload re-reads Path from disk, replacing Entries, Hash and Warnings.
func (d *Dictionary) Reload() error {
	if d == nil || d.Path == "" {
		return fmt.Errorf("nil dictionary")
	}
	fresh, err := Load(d.Path)
	if err != nil {
		return err
	}
	d.Entries = fresh.Entries
	d.Hash = fresh.Hash
	d.Warnings = fresh.Warnings
	return nil
}

// NewEmpty returns an empty dictionary that will be created at path on Save.
func NewEmpty(path string) *Dictionary {
	return &Dictionary{
		Path:    path,
		Entries: map[string]string{},
		Hash:    contentHash(nil),
	}
}
