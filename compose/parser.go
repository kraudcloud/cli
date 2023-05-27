package compose

import (
	"github.com/compose-spec/compose-go/loader"
	"github.com/compose-spec/compose-go/types"
)

func ParseFile(filename string) (*types.Project, error) {
	return loader.Load(types.ConfigDetails{
		WorkingDir:  ".",
		ConfigFiles: []types.ConfigFile{{Filename: filename}},
		Version:     "3.8",
	})
}
