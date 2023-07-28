package compose

import (
	"bufio"
	"fmt"
	"io"
	"strings"

	"mvdan.cc/sh/v3/syntax"
)

type EnvLoader func(string) *string

// LoadEnv loads the environment variables from the given reader
// The loaders function is used to load the value of an environment variable
//
// first-non-nil loader wins
// if no loader is non-nil, then the default is used, or an error is returned
func LoadEnv(toLoad map[string]EnvExprRhs, loaders ...EnvLoader) (map[string]string, error) {
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

func LoadKVSlice(kvs []string) EnvLoader {
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

	return LoadKV(vars)
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

// EnvExprRhs is the right hand side of an environment variable expression
type EnvExprRhs struct {
	// If Default is empty, and the value is not set, then the value is defaulted
	Default string
	// If Error is not empty, and the value is not set, then the error is returned
	Error string
}

// GetTemplateVars returns a map of environment variables that are referenced in the given reader
// The map key is the variable name, and the value is the right hand side of the expression
func GetTemplateVars(r io.Reader) map[string]EnvExprRhs {
	w, err := syntax.NewParser().Parse(r, "")
	if err != nil {
		return nil
	}

	vars := make(map[string]EnvExprRhs)
	syntax.Walk(w, func(node syntax.Node) bool {
		switch x := node.(type) {
		case *syntax.ParamExp:
			ev := EnvExprRhs{}
			defer func() {
				vars[x.Param.Value] = ev
			}()

			if x.Exp == nil || x.Exp.Word == nil {
				return true
			}

			switch x.Exp.Op {
			case syntax.DefaultUnset, syntax.DefaultUnsetOrNull:
				ev.Default = x.Exp.Word.Lit()
			case syntax.ErrorUnset, syntax.ErrorUnsetOrNull:
				ev.Error = x.Exp.Word.Lit()
			}

		}

		return true
	})

	return vars
}

const toEscape = " \t\n\"'\\$"

func EncodeEnv(w io.Writer, env map[string]string) error {
	var replacer []string
	for _, c := range toEscape {
		replacer = append(replacer, string(c), "\\"+string(c))
	}

	r := strings.NewReplacer(replacer...)
	sw := bufio.NewWriter(w)

	for k, v := range env {
		sw.WriteString(k)
		sw.WriteString("=")
		sw.WriteString(r.Replace(v))
		sw.WriteString("\n")
	}

	return sw.Flush()
}
