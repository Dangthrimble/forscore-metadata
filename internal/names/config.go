package names

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	defaultNamesFile  = "names.json"
	defaultGenresFile = "genres.json"
	libraryConfigFile = "fn2fm.json"
	configDirName     = "fn2fm"
	configFileName    = "config.json"
	// ConfigDirEnv overrides the directory that holds config.json. Used by tests
	// so Windows runners (where HOME/XDG do not affect UserConfigDir) stay isolated.
	ConfigDirEnv = "FN2FM_CONFIG_DIR"

	// SubjectArrangers writes the filename slot after "_" into PDF Subject as
	// arranger names. It is the default because fn2fm was built for Cambrensis,
	// where arrangers matter and genres do not.
	SubjectArrangers = "arrangers"
	// SubjectGenres writes that same slot into PDF Subject as genre labels.
	SubjectGenres = "genres"

	SubjectFromFlag    = "flag"
	SubjectFromLibrary = "library"
	SubjectFromUser    = "user"
	SubjectFromDefault = "default"
)

// Config holds optional user preferences. An absent file means no saved path
// and the default subject mode.
type Config struct {
	NamesPath  string `json:"namesPath,omitempty"`
	GenresPath string `json:"genresPath,omitempty"`
	Subject    string `json:"subject,omitempty"`
}

// SubjectChoice is the subject mode that applies to a run, and where it came from.
type SubjectChoice struct {
	Mode   string
	Source string
}

// Label describes the mode and its source for command output.
func (c SubjectChoice) Label() string {
	where := map[string]string{
		SubjectFromFlag:    "flag --subject",
		SubjectFromLibrary: "library ./fn2fm.json",
		SubjectFromUser:    "user config",
		SubjectFromDefault: "default",
	}[c.Source]
	if where == "" {
		where = c.Source
	}
	return c.Mode + " (" + where + ")"
}

// ConfigPath returns the absolute path of the user config file.
func ConfigPath() (string, error) {
	if override := strings.TrimSpace(os.Getenv(ConfigDirEnv)); override != "" {
		return filepath.Join(override, configFileName), nil
	}
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, configDirName, configFileName), nil
}

// LoadConfig reads the user config. Missing file yields an empty Config.
func LoadConfig() (*Config, error) {
	path, err := ConfigPath()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &Config{}, nil
		}
		return nil, err
	}
	var c Config
	if err := json.Unmarshal(data, &c); err != nil {
		return nil, fmt.Errorf("invalid config %s: %w", path, err)
	}
	return &c, nil
}

// SaveConfig writes the user config, creating the directory when needed.
func SaveConfig(c *Config) error {
	if c == nil {
		return fmt.Errorf("nil config")
	}
	path, err := ConfigPath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	temp, err := os.CreateTemp(filepath.Dir(path), ".config-*.tmp")
	if err != nil {
		return err
	}
	tempName := temp.Name()
	defer func() { _ = os.Remove(tempName) }()
	if _, err := temp.Write(data); err != nil {
		_ = temp.Close()
		return err
	}
	if err := temp.Close(); err != nil {
		return err
	}
	return replacePath(tempName, path)
}

// ClearNamesPath removes a saved dictionary path from the config.
func ClearNamesPath() error {
	c, err := LoadConfig()
	if err != nil {
		return err
	}
	c.NamesPath = ""
	return SaveConfig(c)
}

// SetNamesPath stores an absolute names dictionary path for future runs.
// Other saved preferences are kept.
func SetNamesPath(path string) error {
	return setConfigPath(path, func(c *Config, abs string) { c.NamesPath = abs })
}

// SetGenresPath stores an absolute genres dictionary path for future runs.
// Other saved preferences are kept.
func SetGenresPath(path string) error {
	return setConfigPath(path, func(c *Config, abs string) { c.GenresPath = abs })
}

func setConfigPath(path string, assign func(*Config, string)) error {
	abs, err := filepath.Abs(path)
	if err != nil {
		return err
	}
	c, err := LoadConfig()
	if err != nil {
		return err
	}
	assign(c, abs)
	return SaveConfig(c)
}

// SetUserSubject saves arrangers or genres as the per-user default.
func SetUserSubject(mode string) error {
	parsed, err := ParseSubject(mode)
	if err != nil {
		return err
	}
	c, err := LoadConfig()
	if err != nil {
		return err
	}
	c.Subject = parsed
	return SaveConfig(c)
}

// SetLibrarySubject writes ./fn2fm.json with only the subject pin, so a shared
// scores folder can force arrangers or genres over the per-user default.
func SetLibrarySubject(mode string) error {
	parsed, err := ParseSubject(mode)
	if err != nil {
		return err
	}
	path, err := filepath.Abs(libraryConfigFile)
	if err != nil {
		return err
	}
	data, err := json.MarshalIndent(struct {
		Subject string `json:"subject"`
	}{Subject: parsed}, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return writeAtomic(path, data)
}

// ParseSubject accepts arrangers or genres. Empty and unknown values are errors.
func ParseSubject(mode string) (string, error) {
	switch strings.TrimSpace(mode) {
	case SubjectArrangers, SubjectGenres:
		return strings.TrimSpace(mode), nil
	default:
		return "", fmt.Errorf("invalid subject %q (use %q or %q)", strings.TrimSpace(mode), SubjectArrangers, SubjectGenres)
	}
}

// EffectiveSubject resolves the subject mode. Flag wins, then ./fn2fm.json when
// it sets subject, then the per-user config, then arrangers.
func EffectiveSubject(flag string) (SubjectChoice, error) {
	if strings.TrimSpace(flag) != "" {
		mode, err := ParseSubject(flag)
		if err != nil {
			return SubjectChoice{}, err
		}
		return SubjectChoice{Mode: mode, Source: SubjectFromFlag}, nil
	}
	lib, ok, err := loadLibrarySubject()
	if err != nil {
		return SubjectChoice{}, err
	}
	if ok {
		return SubjectChoice{Mode: lib, Source: SubjectFromLibrary}, nil
	}
	c, err := LoadConfig()
	if err != nil {
		return SubjectChoice{}, err
	}
	if strings.TrimSpace(c.Subject) != "" {
		mode, err := ParseSubject(c.Subject)
		if err != nil {
			return SubjectChoice{}, fmt.Errorf("config: %w", err)
		}
		return SubjectChoice{Mode: mode, Source: SubjectFromUser}, nil
	}
	return SubjectChoice{Mode: SubjectArrangers, Source: SubjectFromDefault}, nil
}

func loadLibrarySubject() (string, bool, error) {
	path, err := filepath.Abs(libraryConfigFile)
	if err != nil {
		return "", false, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", false, nil
		}
		return "", false, err
	}
	var c struct {
		Subject *string `json:"subject"`
	}
	if err := json.Unmarshal(data, &c); err != nil {
		return "", false, fmt.Errorf("invalid library config %s: %w", path, err)
	}
	if c.Subject == nil || strings.TrimSpace(*c.Subject) == "" {
		return "", false, nil
	}
	mode, err := ParseSubject(*c.Subject)
	if err != nil {
		return "", false, fmt.Errorf("%s: %w", path, err)
	}
	return mode, true, nil
}

func writeAtomic(path string, data []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	temp, err := os.CreateTemp(filepath.Dir(path), ".config-*.tmp")
	if err != nil {
		return err
	}
	tempName := temp.Name()
	defer func() { _ = os.Remove(tempName) }()
	if _, err := temp.Write(data); err != nil {
		_ = temp.Close()
		return err
	}
	if err := temp.Close(); err != nil {
		return err
	}
	return replacePath(tempName, path)
}

// ResolvePath picks the names dictionary path: flagPath if set, else the configured
// path, else names.json in the current working directory. A configured path
// that is missing is returned as-is so callers can report an error; it is not
// silently replaced by the default.
func ResolvePath(flagPath string) (string, error) {
	return resolveDictionaryPath(flagPath, func(c *Config) string { return c.NamesPath }, defaultNamesFile)
}

// ResolveGenresPath picks the genres dictionary path: flagPath if set, else the
// configured path, else genres.json in the current working directory.
func ResolveGenresPath(flagPath string) (string, error) {
	return resolveDictionaryPath(flagPath, func(c *Config) string { return c.GenresPath }, defaultGenresFile)
}

func resolveDictionaryPath(flagPath string, configured func(*Config) string, defaultFile string) (string, error) {
	if flagPath = strings.TrimSpace(flagPath); flagPath != "" {
		return filepath.Abs(flagPath)
	}
	c, err := LoadConfig()
	if err != nil {
		return "", err
	}
	if path := configured(c); path != "" {
		return path, nil
	}
	return filepath.Abs(defaultFile)
}

// DefaultPath returns names.json in the current working directory.
func DefaultPath() (string, error) {
	return filepath.Abs(defaultNamesFile)
}
