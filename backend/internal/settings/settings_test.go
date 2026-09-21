package settings

import (
	"os"
	"path/filepath"
	"testing"
)

func env(m map[string]string) func(string) string { return func(k string) string { return m[k] } }

func write(t *testing.T, body string) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), "s.json")
	if err := os.WriteFile(p, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
	return p
}

func TestLoadFileAndDefaults(t *testing.T) {
	p := write(t, `{"wclClientId":"id","wclClientSecret":"sec"}`)
	s, loaded, err := Load(p, env(nil))
	if err != nil {
		t.Fatal(err)
	}
	if loaded != p || !s.HasCredentials() || s.Port != "8080" || s.CacheDir != ".cache" {
		t.Errorf("unexpected: %+v loaded=%q", s, loaded)
	}
}

func TestEnvOverridesFile(t *testing.T) {
	p := write(t, `{"port":"9000","wclClientId":"file","wclClientSecret":"file"}`)
	s, _, err := Load(p, env(map[string]string{"PORT": "7000", "WCL_CLIENT_ID": "env"}))
	if err != nil {
		t.Fatal(err)
	}
	if s.Port != "7000" || s.WCLClientID != "env" || s.WCLClientSecret != "file" {
		t.Errorf("unexpected: %+v", s)
	}
}

func TestMissingDefaultFileIsFine(t *testing.T) {
	dir := t.TempDir()
	old, _ := os.Getwd()
	defer os.Chdir(old)
	os.Chdir(dir)
	s, loaded, err := Load("", env(map[string]string{"WCL_CLIENT_ID": "a", "WCL_CLIENT_SECRET": "b"}))
	if err != nil || loaded != "" || !s.HasCredentials() {
		t.Errorf("err=%v loaded=%q s=%+v", err, loaded, s)
	}
}

func TestMissingExplicitFileFails(t *testing.T) {
	if _, _, err := Load(filepath.Join(t.TempDir(), "nope.json"), env(nil)); err == nil {
		t.Error("an explicit path that does not exist must fail")
	}
}

func TestUnknownFieldFails(t *testing.T) {
	p := write(t, `{"wclClientSecrt":"typo"}`)
	if _, _, err := Load(p, env(nil)); err == nil {
		t.Error("unknown field should be rejected")
	}
}
