package config

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestLoadOfAMissingFileIsEmpty(t *testing.T) {
	t.Setenv(EnvConfigDir, filepath.Join(t.TempDir(), "absent"))
	file, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if len(file.Profiles) != 0 {
		t.Errorf("got %d profiles, want none", len(file.Profiles))
	}
	name, profile := file.Resolve("")
	if name != DefaultProfile || profile.Token != "" {
		t.Errorf("Resolve = %q, %+v", name, profile)
	}
}

func TestSaveIsPrivateAndRoundTrips(t *testing.T) {
	dir := t.TempDir()
	t.Setenv(EnvConfigDir, dir)

	file, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	file.Put("default", Profile{Token: "tok-1", BaseURL: "https://example.com", Output: "json"})
	file.Put("staging", Profile{Token: "tok-2"})
	if err := file.Save(); err != nil {
		t.Fatal(err)
	}

	info, err := os.Stat(filepath.Join(dir, "config.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	// The file holds API tokens, so it must not be readable by anyone else. Windows has no POSIX mode bits;
	// the file inherits the user profile ACL there.
	if runtime.GOOS != "windows" {
		if mode := info.Mode().Perm(); mode != 0o600 {
			t.Errorf("mode = %o, want 600", mode)
		}
	}

	reread, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if got := reread.Current; got != "default" {
		t.Errorf("current = %q", got)
	}
	if _, profile := reread.Resolve("staging"); profile.Token != "tok-2" {
		t.Errorf("staging = %+v", profile)
	}
	if got := reread.Names(); len(got) != 2 || got[0] != "default" {
		t.Errorf("Names = %v", got)
	}
}

func TestRemoveMovesTheCurrentProfile(t *testing.T) {
	t.Setenv(EnvConfigDir, t.TempDir())
	file, _ := Load()
	file.Put("default", Profile{Token: "a"})
	file.Put("other", Profile{Token: "b"})
	if !file.Remove("default") {
		t.Fatal("Remove reported nothing was removed")
	}
	if file.Current != "other" {
		t.Errorf("current = %q, want the surviving profile", file.Current)
	}
	if file.Remove("default") {
		t.Error("Remove reported a second removal")
	}
}

func TestGetAndSet(t *testing.T) {
	var profile Profile
	for key, value := range map[string]string{"token": "t", "base-url": "u", "output": "yaml"} {
		if err := profile.Set(key, value); err != nil {
			t.Fatal(err)
		}
		got, err := profile.Get(key)
		if err != nil {
			t.Fatal(err)
		}
		if got != value {
			t.Errorf("%s = %q, want %q", key, got, value)
		}
	}
	if err := profile.Set("nonsense", "x"); err == nil {
		t.Error("an unknown key was accepted")
	}
}

func TestRedact(t *testing.T) {
	cases := map[string]string{
		"":                 "",
		"abc":              "***",
		"secret-token-123": "********-123",
	}
	for input, want := range cases {
		if got := Redact(input); got != want {
			t.Errorf("Redact(%q) = %q, want %q", input, got, want)
		}
	}
}
