package names

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	defaultNamesFile = "names.json"
	configDirName    = "fn2fm"
	configFileName   = "config.json"
	// ConfigDirEnv overrides the directory that holds config.json. Used by tests
	// so Windows runners (where HOME/XDG do not affect UserConfigDir) stay isolated.
	ConfigDirEnv = "FN2FM_CONFIG_DIR"
)

// Config holds optional user preferences. An absent file means no saved path.
type Config struct {
	NamesPath string `json:"namesPath"`
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

// SetNamesPath stores an absolute dictionary path for future runs.
func SetNamesPath(path string) error {
	abs, err := filepath.Abs(path)
	if err != nil {
		return err
	}
	return SaveConfig(&Config{NamesPath: abs})
}

// ResolvePath picks the dictionary path: flagPath if set, else the configured
// path, else names.json in the current working directory. A configured path
// that is missing is returned as-is so callers can report an error; it is not
// silently replaced by the default.
func ResolvePath(flagPath string) (string, error) {
	if flagPath = strings.TrimSpace(flagPath); flagPath != "" {
		return filepath.Abs(flagPath)
	}
	c, err := LoadConfig()
	if err != nil {
		return "", err
	}
	if c.NamesPath != "" {
		return c.NamesPath, nil
	}
	return filepath.Abs(defaultNamesFile)
}

// DefaultPath returns names.json in the current working directory.
func DefaultPath() (string, error) {
	return filepath.Abs(defaultNamesFile)
}
