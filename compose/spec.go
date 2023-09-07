package compose

import (
	"context"
	"fmt"
	"log"
	"path"
	"strings"
)

type File struct {
	Version    string
	Services   map[string]Service `yaml:"services"`
	Volumes    map[string]any     `yaml:"volumes"`
	Properties map[string]any     `yaml:",inline"`
}

type Service struct {
	Image      string
	Volumes    []string
	Properties map[string]any `yaml:",inline"`
}

// VolumeDirective
type VolumeDirective struct {
	Source, Destination, Options string
	local                        bool
}

// IsLocal reports whether the path is a local path
// the matching depends on . being the first character.
// TODO: see how docker does it
func (vd VolumeDirective) IsLocal() bool {
	return strings.HasPrefix(vd.Source, ".") || vd.local
}

func (vd VolumeDirective) String() string {
	b := strings.Builder{}

	b.WriteString(vd.Source)
	b.WriteString(":")
	b.WriteString(vd.Destination)
	if vd.Options != "" {
		b.WriteString(":")
		b.WriteString(vd.Options)
	}

	return b.String()
}

func (vd VolumeDirective) withSource(newSource string) VolumeDirective {
	vd.Source = newSource
	return vd
}

func (s Service) parseVolumes() ([]VolumeDirective, error) {
	out := make([]VolumeDirective, 0, len(s.Volumes))
	for _, v := range s.Volumes {
		vd, err := parseVolume(v)
		if err != nil {
			return nil, err
		}
		out = append(out, vd)
	}

	return out, nil
}

// Rewrite rewrites the local paths in the compose spec to remote paths,
// based on the volume name passed in
func (s Service) Rewrite(volPrefix string) (Service, []VolumeDirective, error) {
	v, err := s.parseVolumes()
	if err != nil {
		return s, nil, err
	}

	newV := v
	for i := range v {
		if !v[i].IsLocal() {
			continue
		}

		newV[i] = v[i].withSource(path.Join(volPrefix, v[i].Source))
	}

	newVol := make([]string, len(newV))
	for i := range newV {
		newVol[i] = newV[i].String()
	}

	s.Volumes = newVol
	return s, v, nil
}

func parseVolume(s string) (VolumeDirective, error) {
	parts := strings.Split(s, ":")
	if len(parts) < 2 {
		return VolumeDirective{}, fmt.Errorf("could not parse volume directive, only got 1 part, at least 2 are required")
	}

	v := VolumeDirective{
		Source:      parts[0],
		Destination: parts[1],
	}

	if len(parts) > 2 {
		opt := strings.Join(parts[2:], ":")
		v.Options = opt
	}

	v.local = v.IsLocal()
	return v, nil
}

// rewriteComposeLocal takes in a compose file,
// it parses the volumes section.
// It generates an application function and a new spec
// that must be handled *after* applying.
func (f *File) Rewrite(magicVolume string) (func(ctx context.Context) error, *File, error) {
	volumePaths := []VolumeDirective{}
	for k := range f.Services {
		s, p, err := f.Services[k].Rewrite(path.Join(magicVolume, k))
		if err != nil {
			return nil, nil, fmt.Errorf("error rewriting service %q: %w", k, err)
		}

		for _, path := range p {
			if path.IsLocal() {
				volumePaths = append(volumePaths, path)
			}
		}

		f.Services[k] = s
	}

	if len(volumePaths) > 0 {
		f.Volumes[magicVolume] = struct{}{}
	}

	return func(ctx context.Context) error {
		log.Println("create volume, mount files here", volumePaths)

		return nil
	}, f, nil
}
