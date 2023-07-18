package completions

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"
)

// getAside gets value from a cache or loads it if it's not present.
// It's used by all the completion functions to avoid loading the same data multiple times.
//
// It's useful for more than just the load on server, but rather for UX.
// Waiting a whole http call for each flag/arg you want to complete is not great.
func getAside[T any](key string, load func() (T, error)) (T, error) {
	cd := cacheDir()
	if cd == "" {
		return load()
	}

	// encode key for fs
	key = fmt.Sprintf("%x", sha256.Sum256([]byte(key)))
	cache := filepath.Join(cd, key)

	if stat, err := os.Stat(cache); err == nil && time.Until(stat.ModTime().Add(time.Minute)) > 0 {
		var out T
		if err := readJSON(cache, &out); err != nil {
			os.WriteFile(cache+".debug", []byte(err.Error()), 0644)
			return load()
		}

		return out, nil
	}

	// cache doesn't exist or is outdated
	out, err := load()
	if err != nil {
		return out, err
	}

	if err := writeJSON(cache, out); err != nil {
		return out, err
	}

	return out, nil
}

func readJSON(path string, out any) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	return json.NewDecoder(file).Decode(out)
}

func writeJSON(path string, in any) error {
	os.MkdirAll(filepath.Dir(path), 0755)

	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	return json.NewEncoder(file).Encode(in)
}

// override: $KR_CACHE_DIR
// on linux: $XDG_CACHE_HOME/kra or $HOME/.cache/kra
// on windows: %LOCALAPPDATA%/kra
// on mac: $HOME/Library/Caches/kra
// Other OSes can be added as needed.
func cacheDir() string {
	if cd := os.Getenv("KR_CACHE_DIR"); cd != "" {
		return cd
	}

	switch runtime.GOOS {
	case "windows":
		return filepath.Join(os.Getenv("LOCALAPPDATA") + "/kra")
	case "darwin":
		return filepath.Join(os.Getenv("HOME") + "/Library/Caches/kra")
	case "linux":
		if os.Getenv("XDG_CACHE_HOME") != "" {
			return filepath.Join(os.Getenv("XDG_CACHE_HOME") + "/kra")
		}

		return filepath.Join(os.Getenv("HOME") + "/.cache/kra")
	}

	return ""
}
