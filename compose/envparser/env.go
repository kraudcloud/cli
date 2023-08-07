package envparser

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"strings"
	"unicode"
)

// Variable is the right hand side of an environment variable expression
type Variable struct {
	// Name of the variable
	Name string

	// UnsetEmpty is true if the variable should be empty if it is not set,
	// rather than non-existent
	UnsetEmpty bool

	// Short form is the form without braces
	Short bool

	// If Default is empty, and the value is not set, then the value is defaulted
	Default string
	// If Error is not empty, and the value is not set, then the error is returned
	Error string
}

func splitVar(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}

	if len(data) < 2 {
		return len(data), nil, nil
	}

	if data[0] != '$' {
		v := bytes.IndexByte(data, '$')
		if v == -1 {
			return len(data), nil, nil
		}

		ad, tok, err := splitVar(data[v:], atEOF)
		return v + ad, tok, err
	}

	// escaped
	if data[1] == '$' {
		return 2, nil, nil
	}

	// long form
	if data[1] == '{' {
		i := bytes.IndexAny(data, "}\n")
		if i == -1 || data[i] != '}' {
			return len(data), nil, fmt.Errorf("unmatched brace in expression %q", data)
		}

		return i + 1, data[:i+1], nil
	}

	// short form
	i := bytes.IndexFunc(data[1:], func(r rune) bool {
		return !unicode.IsDigit(r) && !unicode.IsLetter(r) && r != '_'
	})
	if i == -1 {
		return len(data), data, nil
	}

	return i + 1, data[:i+1], nil
}

func parseVar(line []byte) Variable {
	if line[1] != '{' {
		return Variable{
			Name:  string(line[1:]),
			Short: true,
		}
	}

	line = bytes.Trim(line, "${}")

	// string of the form `var:-default` or `var:?error` or `var` or no `:`
	found := bytes.IndexAny(line, "?-:")
	if found == -1 {
		return Variable{Name: string(line)}
	}

	v := Variable{Name: string(line[:found])}
	line = line[found:]

	// ignore :
	if len(line) > 1 && line[0] == ':' {
		v.UnsetEmpty = true
		line = line[1:]
	}

	if line[0] == '-' {
		v.Default = strings.Trim(string(line[1:]), `"' `)
	}

	if line[0] == '?' {
		v.Error = strings.Trim(string(line[1:]), `"' `)
	}

	return v
}

// ParseTemplateVars returns a map of environment variables that are referenced in the given reader
//
// The map key is the variable name, and the value is the right hand side of the expression
func ParseTemplateVars(r io.Reader) (map[string]Variable, error) {
	vars := map[string]Variable{}

	br := bufio.NewScanner(r)
	br.Split(splitVar)

	for br.Scan() {
		v := parseVar(br.Bytes())
		vars[v.Name] = v
	}

	if err := br.Err(); err != nil {
		return nil, fmt.Errorf("invalid variable: %s", err)
	}

	return vars, nil

}
