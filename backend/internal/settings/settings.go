// Package settings loads the server settings (port, Warcraft Logs credentials,
// cache folder). Values come from a JSON file and can be overridden by
// environment variables, which always win.
package settings

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
)

// DefaultFile is read from the working directory when no -config flag is given.
const DefaultFile = "config.json"

type Settings struct {
	Port            string `json:"port"`
	WCLClientID     string `json:"wclClientId"`
	WCLClientSecret string `json:"wclClientSecret"`
	CacheDir        string `json:"cacheDir"`
	// AnalysisConfig is an optional path to an analysis config JSON (defensives,
	// avoidable abilities, thresholds) that replaces the embedded default.json.
	AnalysisConfig string `json:"analysisConfig"`
}

// HasCredentials reports whether both Warcraft Logs credentials are set.
func (s Settings) HasCredentials() bool { return s.WCLClientID != "" && s.WCLClientSecret != "" }

// Load reads the settings file at path (or DefaultFile when path is empty),
// applies environment overrides through getenv and fills in defaults.
//
// A missing default file is not an error; a missing explicit path is.
// The returned string is the file that was loaded ("" if none).
func Load(path string, getenv func(string) string) (Settings, string, error) {
	var s Settings
	loaded := ""

	file, explicit := path, path != ""
	if !explicit {
		file = DefaultFile
	}
	b, err := os.ReadFile(file)
	switch {
	case err == nil:
		dec := json.NewDecoder(bytes.NewReader(b))
		dec.DisallowUnknownFields() // catch typos such as "wclClientSecrt"
		if err := dec.Decode(&s); err != nil {
			return s, "", fmt.Errorf("%s: %w", file, err)
		}
		loaded = file
	case errors.Is(err, fs.ErrNotExist) && !explicit:
		// no settings file: fine, env vars and defaults apply
	default:
		return s, "", fmt.Errorf("reading %s: %w", file, err)
	}

	override := func(dst *string, key string) {
		if v := getenv(key); v != "" {
			*dst = v
		}
	}
	override(&s.Port, "PORT")
	override(&s.WCLClientID, "WCL_CLIENT_ID")
	override(&s.WCLClientSecret, "WCL_CLIENT_SECRET")
	override(&s.CacheDir, "GUILDLOGS_CACHE")
	override(&s.AnalysisConfig, "GUILDLOGS_CONFIG")

	if s.Port == "" {
		s.Port = "8080"
	}
	if s.CacheDir == "" {
		s.CacheDir = ".cache"
	}
	return s, loaded, nil
}
