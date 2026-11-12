package names

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

// Save writes the dictionary atomically when the OS rename allows it, keeps a
// single rolling .bak of the previous bytes, and refuses to overwrite when the
// file on disk no longer matches Hash.
func (d *Dictionary) Save() error {
	if d == nil {
		return fmt.Errorf("nil dictionary")
	}
	if d.Path == "" {
		return fmt.Errorf("dictionary path is empty")
	}
	if err := d.checkNotStale(); err != nil {
		return err
	}
	payload, err := encodeEntries(d.Entries)
	if err != nil {
		return err
	}
	return d.replaceFile(payload)
}

// ForceSave writes without the stale-hash check. Used after an explicit reload
// decision when the caller already reconciled state.
func (d *Dictionary) ForceSave() error {
	if d == nil {
		return fmt.Errorf("nil dictionary")
	}
	if d.Path == "" {
		return fmt.Errorf("dictionary path is empty")
	}
	payload, err := encodeEntries(d.Entries)
	if err != nil {
		return err
	}
	return d.replaceFile(payload)
}

// IsStale reports whether the file on disk differs from Hash.
func (d *Dictionary) IsStale() (bool, error) {
	if d == nil || d.Path == "" {
		return false, fmt.Errorf("nil dictionary")
	}
	data, err := os.ReadFile(d.Path)
	if err != nil {
		if os.IsNotExist(err) {
			// Missing file is not "stale from another editor"; treat as ready to create.
			return false, nil
		}
		return false, err
	}
	return contentHash(data) != d.Hash, nil
}

func (d *Dictionary) checkNotStale() error {
	stale, err := d.IsStale()
	if err != nil {
		return err
	}
	if stale {
		return fmt.Errorf("dictionary at %s changed on disk since it was loaded; reload before saving", d.Path)
	}
	return nil
}

func encodeEntries(entries map[string]string) ([]byte, error) {
	if entries == nil {
		entries = map[string]string{}
	}
	keys := make([]string, 0, len(entries))
	for k := range entries {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var buf bytes.Buffer
	buf.WriteString("{\n")
	for i, k := range keys {
		keyJSON, err := marshalNoHTML(k)
		if err != nil {
			return nil, err
		}
		valJSON, err := marshalNoHTML(entries[k])
		if err != nil {
			return nil, err
		}
		buf.WriteString("    ")
		buf.Write(keyJSON)
		buf.WriteString(": ")
		buf.Write(valJSON)
		if i < len(keys)-1 {
			buf.WriteByte(',')
		}
		buf.WriteByte('\n')
	}
	buf.WriteString("}\n")
	return buf.Bytes(), nil
}

func marshalNoHTML(v string) ([]byte, error) {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(v); err != nil {
		return nil, err
	}
	// Encode appends a newline.
	b := buf.Bytes()
	if len(b) > 0 && b[len(b)-1] == '\n' {
		b = b[:len(b)-1]
	}
	return b, nil
}

func (d *Dictionary) replaceFile(payload []byte) error {
	dir := filepath.Dir(d.Path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	var previous []byte
	hadPrevious := false
	if data, err := os.ReadFile(d.Path); err == nil {
		previous = append([]byte(nil), data...)
		hadPrevious = true
	} else if !os.IsNotExist(err) {
		return err
	}

	temp, err := os.CreateTemp(dir, ".names-*.tmp")
	if err != nil {
		return err
	}
	tempName := temp.Name()
	cleanupTemp := true
	defer func() {
		if cleanupTemp {
			_ = os.Remove(tempName)
		}
	}()

	if _, err := temp.Write(payload); err != nil {
		_ = temp.Close()
		return err
	}
	if err := temp.Sync(); err != nil {
		_ = temp.Close()
		return err
	}
	if err := temp.Close(); err != nil {
		return err
	}

	if err := replacePath(tempName, d.Path); err != nil {
		return err
	}
	cleanupTemp = false

	if hadPrevious {
		bak := d.Path + ".bak"
		if err := os.WriteFile(bak, previous, 0o600); err != nil {
			d.Hash = contentHash(payload)
			return fmt.Errorf("saved %s but failed to write backup: %w", d.Path, err)
		}
	}
	d.Hash = contentHash(payload)
	return nil
}

func replacePath(src, dst string) error {
	err := os.Rename(src, dst)
	if err == nil {
		return nil
	}
	// Windows cannot always rename over an existing destination.
	if _, statErr := os.Lstat(dst); statErr != nil {
		return fmt.Errorf("replace %s: %w", dst, err)
	}
	aside := dst + ".replace-aside"
	_ = os.Remove(aside)
	if err := os.Rename(dst, aside); err != nil {
		return err
	}
	if err := os.Rename(src, dst); err != nil {
		_ = os.Rename(aside, dst)
		return err
	}
	_ = os.Remove(aside)
	return nil
}
