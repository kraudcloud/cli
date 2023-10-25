package compose_test

import (
	"strings"
	"testing"

	"github.com/kraudcloud/cli/compose"
	"gopkg.in/yaml.v3"
)

const s = `version: '3.9'
services:
  a:
    image: abc:latest
    container_name: abc
    labels:
      abc: d
    volumes:
      - ./script.sh:mount/it/there
      - volume:/var/data
`

func TestRewrite(t *testing.T) {
	file := &compose.File{}
	err := yaml.NewDecoder(strings.NewReader(s)).Decode(file)
	if err != nil {
		t.Errorf("error parsing file: %s", err)
		return
	}

	for i := range file.Services {
		s, _, err := file.Services[i].Rewrite("magic")
		if err != nil {
			t.Errorf("error rewriting service: %s", err)
		}
		file.Services[i] = s
	}

	out, err := yaml.Marshal(file)
	if err != nil {
		t.Errorf("error rewriting service: %s", err)
	}

	t.Log(string(out))
}
