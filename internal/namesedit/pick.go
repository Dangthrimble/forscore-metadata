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
func pickDictionaryPath(title string) (string, error) {
	if title == "" {
		title = "Choose names.json"
	}
	switch runtime.GOOS {
	case "darwin":
		out, err := exec.Command("osascript", "-e",
			`POSIX path of (choose file with prompt "`+title+`" of type {"public.json", "json"})`).CombinedOutput()
		if err != nil {
			msg := strings.TrimSpace(string(out))
			if strings.Contains(msg, "User canceled") || strings.Contains(msg, "User cancelled") {
				return "", nil
			}
			return "", fmt.Errorf("%s", or(msg, err.Error()))
		}
		return strings.TrimSpace(string(out)), nil
	case "windows":
		ps := `Add-Type -AssemblyName System.Windows.Forms; $f = New-Object System.Windows.Forms.OpenFileDialog; $f.Filter = 'JSON (*.json)|*.json|All files (*.*)|*.*'; $f.Title = '` + title + `'; if ($f.ShowDialog() -eq 'OK') { $f.FileName }`
		out, err := exec.Command("powershell", "-NoProfile", "-Command", ps).CombinedOutput()
		if err != nil {
			return "", fmt.Errorf("%s", or(strings.TrimSpace(string(out)), err.Error()))
		}
		return strings.TrimSpace(string(out)), nil
	default:
		if _, err := exec.LookPath("zenity"); err == nil {
			out, err := exec.Command("zenity", "--file-selection", "--title="+title, "--file-filter=*.json").CombinedOutput()
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

func pickSaveDictionaryPath(title, defaultName string) (string, error) {
	if title == "" {
		title = "Create names.json"
	}
	if defaultName == "" {
		defaultName = "names.json"
	}
	switch runtime.GOOS {
	case "darwin":
		out, err := exec.Command("osascript", "-e",
			`POSIX path of (choose file name with prompt "`+title+`" default name "`+defaultName+`")`).CombinedOutput()
		if err != nil {
			msg := strings.TrimSpace(string(out))
			if strings.Contains(msg, "User canceled") || strings.Contains(msg, "User cancelled") {
				return "", nil
			}
			return "", fmt.Errorf("%s", or(msg, err.Error()))
		}
		return strings.TrimSpace(string(out)), nil
	case "windows":
		ps := `Add-Type -AssemblyName System.Windows.Forms; $f = New-Object System.Windows.Forms.SaveFileDialog; $f.Filter = 'JSON (*.json)|*.json|All files (*.*)|*.*'; $f.FileName = '` + defaultName + `'; $f.Title = '` + title + `'; if ($f.ShowDialog() -eq 'OK') { $f.FileName }`
		out, err := exec.Command("powershell", "-NoProfile", "-Command", ps).CombinedOutput()
		if err != nil {
			return "", fmt.Errorf("%s", or(strings.TrimSpace(string(out)), err.Error()))
		}
		return strings.TrimSpace(string(out)), nil
	default:
		if _, err := exec.LookPath("zenity"); err == nil {
			out, err := exec.Command("zenity", "--file-selection", "--save", "--confirm-overwrite", "--title="+title, "--filename="+defaultName).CombinedOutput()
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
