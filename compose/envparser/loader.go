package envparser

import (
	"bufio"
	"fmt"
	"io"
	"sort"
	"strings"
	"sync"

	"golang.org/x/exp/maps"
)

type EnvLoader func(string) *string

// LoadEnv loads the environment variables from the given reader
// The loaders function is used to load the value of an environment variable
//
// first-non-nil loader wins
// if no loader is non-nil, then the default is used, or an error is returned
func LoadEnv(toLoad map[string]Variable, loaders ...EnvLoader) (map[string]string, error) {
	var errs []string

	vars := make(map[string]string)
	for k, v := range toLoad {
		outV := ""
		for _, loader := range loaders {
			if loader == nil {
				continue
			}

			if v := loader(k); v != nil {
				outV = *v
				break
			}
		}

		if v.UnsetEmpty && outV == "" {
			vars[k] = outV
			continue
		}

		if outV != "" {
			vars[k] = outV
			continue
		}

		if v.Default != "" {
			vars[k] = v.Default
			continue
		}

		if v.Error != "" {
			errs = append(errs, v.Error)
			continue
		}
	}

	if len(errs) > 0 {
		return nil, fmt.Errorf("missing environment variables: %s", strings.Join(errs, ", "))
	}

	return vars, nil
}

// LoadKV returns a loader function that loads the given key/value pairs from the map
func LoadKV(kv map[string]string) EnvLoader {
	return func(key string) *string {
		if v, ok := kv[key]; ok {
			return &v
		}
		return nil
	}
}

func LoadKVs(kvs []string) EnvLoader {
	m := make(map[string]string)
	for _, kv := range kvs {
		k, v, err := ParseKV(kv)
		if err != nil {
			continue
		}

		m[k] = v
	}

	return LoadKV(m)
}

func ParseKV(kv string) (string, string, error) {
	k, v, ok := strings.Cut(kv, "=")
	if !ok {
		return "", "", fmt.Errorf("invalid key/value pair: %s", kv)
	}

	k = unescapeStringVar(k)
	v = unescapeStringVar(v)

	return k, v, nil
}

// LoadEnvReader reads a .env formatted file from the given path, and returns a loader function
func LoadEnvReader(r io.Reader) EnvLoader {
	return LoadKV(EnvMapFromReader(r))
}

// EnvMapFromReader reads a .env formatted file from the given path, and returns a map of key/value pairs
func EnvMapFromReader(r io.Reader) map[string]string {
	sr := bufio.NewScanner(r)
	sr.Split(bufio.ScanLines)

	vars := make(map[string]string)
	for sr.Scan() {
		line := sr.Text()
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "#") {
			continue
		}

		k, v, err := ParseKV(line)
		if err != nil {
			continue
		}

		vars[k] = v
	}

	return vars
}

func unescapeStringVar(s string) string {
	s = strings.TrimSpace(s)
	if strings.HasPrefix(s, "'") && strings.HasSuffix(s, "'") {
		s = strings.Trim(s, "'")
	}

	if strings.HasPrefix(s, "\"") && strings.HasSuffix(s, "\"") {
		s = strings.Trim(s, "\"")
		s = strings.ReplaceAll(s, "\\\"", "\"")
	}

	return strings.TrimSpace(s)
}

const toEscape = "\t\n\"'\\$"
const toQuote = " "

var escapeReplacer interface {
	Replace(string) string
}
var onceReplacer sync.Once

func EncodeEnv(w io.Writer, env map[string]string) error {
	onceReplacer.Do(func() {
		var replacer []string
		for _, c := range toEscape {
			replacer = append(replacer, string(c), "\\"+string(c))
		}

		escapeReplacer = strings.NewReplacer(replacer...)
	})

	sw := bufio.NewWriter(w)

	sorted := maps.Keys(env)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i] < sorted[j]
	})

	for _, key := range sorted {
		value := env[key]

		sw.WriteString(key)
		sw.WriteString("=")
		quote := strings.ContainsAny(value, toQuote)
		if quote {
			sw.WriteByte('"')
		}
		sw.WriteString(escapeReplacer.Replace(value))
		if quote {
			sw.WriteByte('"')
		}
		sw.WriteString("\n")
	}

	return sw.Flush()
}
