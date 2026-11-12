package names

import (
	"os"
	"path/filepath"
	"strings"
)

// ConflictedCopies looks for common cloud-sync conflict siblings beside path.
// It does not merge them; callers should warn the user.
func ConflictedCopies(path string) ([]string, error) {
	dir := filepath.Dir(path)
	base := filepath.Base(path)
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	stem := strings.TrimSuffix(base, filepath.Ext(base))
	ext := filepath.Ext(base)
	var found []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if name == base || name == base+".bak" {
			continue
		}
		lower := strings.ToLower(name)
		if isConflictName(lower, strings.ToLower(stem), strings.ToLower(ext)) {
			found = append(found, filepath.Join(dir, name))
		}
	}
	return found, nil
}

func isConflictName(name, stem, ext string) bool {
	// Dropbox: "name (Conflicted copy).json" or "name (Jonathan's conflicted copy).json"
	if strings.Contains(name, "conflicted copy") && strings.HasPrefix(name, stem) {
		return true
	}
	// OneDrive: "name-ComputerName.json" is too ambiguous; look for "conflict"
	if strings.Contains(name, "conflict") && strings.HasPrefix(name, stem) {
		return true
	}
	// Box sometimes uses "name (N).json" — skip generic numbered copies.
	// Google Drive / misc: "name (conflict).ext"
	if ext != "" && strings.HasSuffix(name, ext) {
		body := strings.TrimSuffix(name, ext)
		if strings.HasPrefix(body, stem+" ") || strings.HasPrefix(body, stem+"(") || strings.HasPrefix(body, stem+"-") {
			if strings.Contains(body, "conflict") || strings.Contains(body, "copy") {
				return true
			}
		}
	}
	return false
}
