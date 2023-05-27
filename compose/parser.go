package compose

import (
	"os"
	"strings"

	"github.com/compose-spec/compose-go/loader"
	"github.com/compose-spec/compose-go/types"
)

func ParseFile(filename string) (*types.Project, error) {
	e := os.Environ()
	envMap := make(map[string]string, len(e))
	for _, v := range e {
		b, a, ok := strings.Cut(v, "=")
		if !ok {
			continue
		}

		envMap[b] = a
	}

	return loader.Load(types.ConfigDetails{
		WorkingDir:  ".",
		ConfigFiles: []types.ConfigFile{{Filename: filename}},
		Version:     "3.8",
		Environment: envMap,
	})
}
