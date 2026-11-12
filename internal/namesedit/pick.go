package namesedit

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"
)

// pickDictionaryPath asks the OS for a file path. Returns empty string when the
// user cancels. Falls back to an error that callers can use to prompt for a
// typed path when no dialog helper is available.
func pickDictionaryPath() (string, error) {
	switch runtime.GOOS {
	case "darwin":
		out, err := exec.Command("osascript", "-e",
			`POSIX path of (choose file with prompt "Choose names.json" of type {"public.json", "json"})`).CombinedOutput()
		if err != nil {
			msg := strings.TrimSpace(string(out))
			if strings.Contains(msg, "User canceled") || strings.Contains(msg, "User cancelled") {
				return "", nil
			}
			return "", fmt.Errorf("%s", or(msg, err.Error()))
		}
		return strings.TrimSpace(string(out)), nil
	case "windows":
		ps := `Add-Type -AssemblyName System.Windows.Forms; $f = New-Object System.Windows.Forms.OpenFileDialog; $f.Filter = 'JSON (*.json)|*.json|All files (*.*)|*.*'; $f.Title = 'Choose names.json'; if ($f.ShowDialog() -eq 'OK') { $f.FileName }`
		out, err := exec.Command("powershell", "-NoProfile", "-Command", ps).CombinedOutput()
		if err != nil {
			return "", fmt.Errorf("%s", or(strings.TrimSpace(string(out)), err.Error()))
		}
		return strings.TrimSpace(string(out)), nil
	default:
		if _, err := exec.LookPath("zenity"); err == nil {
			out, err := exec.Command("zenity", "--file-selection", "--title=Choose names.json", "--file-filter=*.json").CombinedOutput()
			if err != nil {
				// zenity returns 1 on cancel
				if strings.TrimSpace(string(out)) == "" {
					return "", nil
				}
				return "", fmt.Errorf("%s", or(strings.TrimSpace(string(out)), err.Error()))
			}
			return strings.TrimSpace(string(out)), nil
		}
		return "", fmt.Errorf("no file dialog available; type the path instead")
	}
}

func pickSaveDictionaryPath() (string, error) {
	switch runtime.GOOS {
	case "darwin":
		out, err := exec.Command("osascript", "-e",
			`POSIX path of (choose file name with prompt "Create names.json" default name "names.json")`).CombinedOutput()
		if err != nil {
			msg := strings.TrimSpace(string(out))
			if strings.Contains(msg, "User canceled") || strings.Contains(msg, "User cancelled") {
				return "", nil
			}
			return "", fmt.Errorf("%s", or(msg, err.Error()))
		}
		return strings.TrimSpace(string(out)), nil
	case "windows":
		ps := `Add-Type -AssemblyName System.Windows.Forms; $f = New-Object System.Windows.Forms.SaveFileDialog; $f.Filter = 'JSON (*.json)|*.json|All files (*.*)|*.*'; $f.FileName = 'names.json'; $f.Title = 'Create names.json'; if ($f.ShowDialog() -eq 'OK') { $f.FileName }`
		out, err := exec.Command("powershell", "-NoProfile", "-Command", ps).CombinedOutput()
		if err != nil {
			return "", fmt.Errorf("%s", or(strings.TrimSpace(string(out)), err.Error()))
		}
		return strings.TrimSpace(string(out)), nil
	default:
		if _, err := exec.LookPath("zenity"); err == nil {
			out, err := exec.Command("zenity", "--file-selection", "--save", "--confirm-overwrite", "--title=Create names.json", "--filename=names.json").CombinedOutput()
			if err != nil {
				if strings.TrimSpace(string(out)) == "" {
					return "", nil
				}
				return "", fmt.Errorf("%s", or(strings.TrimSpace(string(out)), err.Error()))
			}
			return strings.TrimSpace(string(out)), nil
		}
		return "", fmt.Errorf("no file dialog available; type the path instead")
	}
}

func or(a, b string) string {
	if a != "" {
		return a
	}
	return b
}
