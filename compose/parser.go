package compose

import (
	"io"
	"os"

	"gopkg.in/yaml.v3"
)

func ParseFile(filename string) (*File, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return Parse(f)
}

func Parse(r io.Reader) (*File, error) {
	file := &File{}
	err := yaml.NewDecoder(r).Decode(file)
	if err != nil {
		return nil, err
	}

	return file, nil
}
