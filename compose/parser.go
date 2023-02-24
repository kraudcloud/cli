package compose

import (
	"gopkg.in/yaml.v3"
	"os"
)

func ParseFile(filename string) (*File, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}

	var file = &File{}
	err = yaml.NewDecoder(f).Decode(file)
	if err != nil {
		return nil, err
	}

	return file, nil
}
