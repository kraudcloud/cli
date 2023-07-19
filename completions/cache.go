package completions

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"hash/fnv"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"golang.org/x/crypto/hkdf"
)

// getAside gets value from a cache or loads it if it's not present.
// It's used by all the completion functions to avoid loading the same data multiple times.
//
// It's useful for more than just the load on server, but rather for UX.
// Waiting a whole http call for each flag/arg you want to complete is not great.
func getAside[T any](cacheKey string, encryptionKey string, load func() (T, error)) (T, error) {
	cd := cacheDir()
	if cd == "" {
		return load()
	}

	// encode key for fs
	cacheKey = fmt.Sprintf("%x", fnv.New32().Sum([]byte(cacheKey)))
	cache := filepath.Join(cd, cacheKey)

	if stat, err := os.Stat(cache); err == nil && time.Until(stat.ModTime().Add(time.Minute)) > 0 {
		var out T
		if err := readJSON(cache, encryptionKey, &out); err != nil {
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

	if err := writeJSON(cache, encryptionKey, out); err != nil {
		return out, err
	}

	return out, nil
}

func readJSON(path string, userToken string, data any) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	return json.NewDecoder(decryptW(userToken, file)).Decode(data)
}

func writeJSON(path string, userToken string, data any) error {
	os.MkdirAll(filepath.Dir(path), 0755)

	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	// encrypt the data itself.
	// this serves two purposes:
	// 1. We never leak the user's data because we don't remove the cache
	// 2. If the tries to use another account, the cache is reloaded.

	return json.NewEncoder(encryptW(userToken, file)).Encode(data)
}

func decryptW(tok string, reader io.Reader) io.Reader {
	keyb := make([]byte, aes.BlockSize)
	hkdf.New(sha256.New, []byte(tok), nil, nil).Read(keyb)

	block, err := aes.NewCipher(keyb)
	if err != nil {
		panic(err)
	}

	return &cipher.StreamReader{
		S: cipher.NewCTR(block, keyb[:aes.BlockSize]),
		R: reader,
	}
}

func encryptW(tok string, reader io.Writer) io.Writer {
	keyb := make([]byte, aes.BlockSize)
	hkdf.New(sha256.New, []byte(tok), nil, nil).Read(keyb)

	block, err := aes.NewCipher(keyb)
	if err != nil {
		panic(err)
	}

	return &cipher.StreamWriter{
		S: cipher.NewCTR(block, keyb[:aes.BlockSize]),
		W: reader,
	}
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
