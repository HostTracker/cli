// Package config reads and writes the CLI's profile file.
//
// The file lives under the OS configuration directory (
// $XDG_CONFIG_HOME/ht/config.yaml on Linux, ~/Library/Application
// Support/ht/config.yaml on macOS, %AppData%\ht\config.yaml on Windows)
// and is written 0600, because it holds API tokens.
package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

// DefaultProfile is the profile used when none is named.
const DefaultProfile = "default"

// EnvConfigDir overrides where the configuration file is looked for.
const EnvConfigDir = "HT_CONFIG_DIR"

// Profile is one named set of settings.
type Profile struct {
	Token   string `yaml:"token,omitempty"`
	BaseURL string `yaml:"base-url,omitempty"`
	Output  string `yaml:"output,omitempty"`
}

// File is the whole configuration document.
type File struct {
	// Current is the profile used when --profile is not given.
	Current string `yaml:"current,omitempty"`
	// Profiles is the named profiles, keyed by name.
	Profiles map[string]Profile `yaml:"profiles,omitempty"`

	path string
}

// Settable is the list of keys `ht config set` accepts, in help order.
var Settable = []string{"token", "base-url", "output"}

// Dir is the directory the configuration file lives in.
func Dir() (string, error) {
	if custom := strings.TrimSpace(os.Getenv(EnvConfigDir)); custom != "" {
		return custom, nil
	}
	base, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, "ht"), nil
}

// Path is the full path of the configuration file.
func Path() (string, error) {
	dir, err := Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.yaml"), nil
}

// Load reads the configuration file. A missing file is not an error: it
// yields an empty document that Save will create.
func Load() (*File, error) {
	path, err := Path()
	if err != nil {
		return nil, err
	}
	f := &File{Profiles: map[string]Profile{}, path: path}
	raw, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return f, nil
	}
	if err != nil {
		return nil, err
	}
	if err := yaml.Unmarshal(raw, f); err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	if f.Profiles == nil {
		f.Profiles = map[string]Profile{}
	}
	f.path = path
	return f, nil
}

// Save writes the configuration file, creating its directory and keeping
// both 0700 and 0600 so a token is never world-readable.
func (f *File) Save() error {
	path := f.path
	if path == "" {
		var err error
		if path, err = Path(); err != nil {
			return err
		}
		f.path = path
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	raw, err := yaml.Marshal(f)
	if err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, raw, 0o600); err != nil {
		return err
	}
	if err := os.Rename(tmp, path); err != nil {
		_ = os.Remove(tmp)
		return err
	}
	return os.Chmod(path, 0o600)
}

// FilePath is where this document was read from or will be written to.
func (f *File) FilePath() string { return f.path }

// Resolve returns the named profile, or the current one when name is
// empty. A profile that does not exist yet reads as empty, so a fresh
// installation works before `ht auth login` has run.
func (f *File) Resolve(name string) (string, Profile) {
	if name == "" {
		name = f.Current
	}
	if name == "" {
		name = DefaultProfile
	}
	return name, f.Profiles[name]
}

// Put stores a profile under name and makes it current when the file had
// no current profile yet.
func (f *File) Put(name string, p Profile) {
	if name == "" {
		name = DefaultProfile
	}
	if f.Profiles == nil {
		f.Profiles = map[string]Profile{}
	}
	f.Profiles[name] = p
	if f.Current == "" {
		f.Current = name
	}
}

// Remove deletes a profile and reports whether it existed.
func (f *File) Remove(name string) bool {
	if name == "" {
		name = DefaultProfile
	}
	if _, ok := f.Profiles[name]; !ok {
		return false
	}
	delete(f.Profiles, name)
	if f.Current == name {
		f.Current = ""
		for _, other := range f.Names() {
			f.Current = other
			break
		}
	}
	return true
}

// Names lists the profile names in alphabetical order.
func (f *File) Names() []string {
	out := make([]string, 0, len(f.Profiles))
	for name := range f.Profiles {
		out = append(out, name)
	}
	sort.Strings(out)
	return out
}

// Get reads one settable key off a profile.
func (p Profile) Get(key string) (string, error) {
	switch key {
	case "token":
		return p.Token, nil
	case "base-url":
		return p.BaseURL, nil
	case "output":
		return p.Output, nil
	default:
		return "", fmt.Errorf("unknown key %q (known: %s)", key, strings.Join(Settable, ", "))
	}
}

// Set writes one settable key on a profile.
func (p *Profile) Set(key, value string) error {
	switch key {
	case "token":
		p.Token = value
	case "base-url":
		p.BaseURL = value
	case "output":
		p.Output = value
	default:
		return fmt.Errorf("unknown key %q (known: %s)", key, strings.Join(Settable, ", "))
	}
	return nil
}

// Redact replaces all but the last four characters of a token, so a
// status line can prove which credential is in use without printing it.
func Redact(token string) string {
	token = strings.TrimSpace(token)
	if token == "" {
		return ""
	}
	if len(token) <= 4 {
		return strings.Repeat("*", len(token))
	}
	return strings.Repeat("*", 8) + token[len(token)-4:]
}
