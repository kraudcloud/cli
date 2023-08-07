package envparser

import (
	"bytes"
	"io"
	"reflect"
	"strings"
	"testing"
)

func TestLoadEnvReader(t *testing.T) {
	tests := []struct {
		name  string
		r     io.Reader
		wantV map[string]string
	}{
		{
			name:  "empty",
			r:     strings.NewReader(""),
			wantV: map[string]string{},
		},
		{
			name: "simple",
			r:    strings.NewReader("FOO=bar"),
			wantV: map[string]string{
				"FOO": "bar",
			},
		},
		{
			name:  "comment",
			r:     strings.NewReader("# FOO=bar"),
			wantV: map[string]string{},
		},
		{
			name: "multiline",
			r: strings.NewReader(`
      FOO=bar
      # Bazed on what?
      BAZ=qux
      #BAZE=QUX2
      `),
			wantV: map[string]string{
				"FOO": "bar",
				"BAZ": "qux",
			},
		},
		{
			name: "weird whitespace",
			r: strings.NewReader(`
      FOO = bar
        BAZ = qux
      `),
			wantV: map[string]string{
				"FOO": "bar",
				"BAZ": "qux",
			},
		},
		{
			name: "quoted",
			r: strings.NewReader(`
      FOO="bar"
      'BAZ'='qux'
      `),
			wantV: map[string]string{
				"FOO": "bar",
				"BAZ": "qux",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotV := LoadEnvReader(tt.r)

			for k, v := range tt.wantV {
				if got := gotV(k); got != nil && *got != v {
					t.Errorf("LoadEnvReader() = %v, want %v", got, v)
				}
			}
		})
	}
}

func TestEncodeEnv(t *testing.T) {
	tests := []struct {
		name    string
		env     map[string]string
		wantW   string
		wantErr bool
	}{
		{
			name:  "empty",
			env:   map[string]string{},
			wantW: "",
		},
		{
			name: "simple",
			env: map[string]string{
				"FOO": "bar",
			},
			wantW: "FOO=bar\n",
		},
		{
			name: "multiline",
			env: map[string]string{
				"FOO": "bar",
				"BAZ": "qux",
			},
			wantW: "BAZ=qux\nFOO=bar\n",
		},
		{
			name: "need quoting",
			env: map[string]string{
				"FOO": "bar",
				"BAZ": "bar dwq dwdw",
			},
			wantW: "BAZ=\"bar dwq dwdw\"\nFOO=bar\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &bytes.Buffer{}
			if err := EncodeEnv(w, tt.env); (err != nil) != tt.wantErr {
				t.Errorf("EncodeEnv() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotW := w.String(); gotW != tt.wantW {
				t.Errorf("EncodeEnv() = %v, want %v", gotW, tt.wantW)
			}
		})
	}
}

func TestLoadKVs(t *testing.T) {
	cpy := func(v string) *string {
		return &v
	}

	tests := []struct {
		name string
		kvs  []string

		arg     string
		wantArg *string
	}{
		{
			name: "empty",
			kvs:  []string{},
		},
		{
			name:    "simple",
			kvs:     []string{"FOO=bar"},
			arg:     "FOO",
			wantArg: cpy("bar"),
		},
		{
			name:    "not found",
			kvs:     []string{"FOO=bar"},
			arg:     "BAZ",
			wantArg: nil,
		},
		{
			name:    "multiple",
			kvs:     []string{"FOO=bar", "BAZ=qux"},
			arg:     "BAZ",
			wantArg: cpy("qux"),
		},
		{
			name: "invalid",
			kvs:  []string{"FOO=bar", "BAZ"},
			arg:  "BAZ",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := LoadKVs(tt.kvs)
			gotArg := got(tt.arg)

			if gotArg != nil && *gotArg != *tt.wantArg {
				t.Errorf("LoadKVs() = %v, want %v", got, tt.wantArg)
			}

			if gotArg == nil && tt.wantArg != nil {
				t.Errorf("LoadKVs() = %v, want %v", got, tt.wantArg)
			}
		})
	}
}

func TestLoadEnv(t *testing.T) {
	tests := []struct {
		name    string
		toLoad  map[string]Variable
		loaders []EnvLoader
		want    map[string]string
		wantErr bool
	}{
		{
			name: "simple",
			toLoad: map[string]Variable{
				"FOO": {
					Name: "FOO",
				},
			},
			loaders: []EnvLoader{
				LoadKV(map[string]string{}),
			},
			want: map[string]string{},
		},
		{
			name: "value",
			toLoad: map[string]Variable{
				"FOO": {
					Name: "FOO",
				},
			},
			loaders: []EnvLoader{LoadKV(map[string]string{
				"FOO": "bar",
			})},
			want: map[string]string{
				"FOO": "bar",
			},
		},
		{
			name: "value with default",
			toLoad: map[string]Variable{
				"FOO": {
					Name:    "FOO",
					Default: "bar",
				},
			},
			loaders: []EnvLoader{LoadKV(map[string]string{})},
			want: map[string]string{
				"FOO": "bar",
			},
		},
		{
			name: "value with default and env",
			toLoad: map[string]Variable{
				"FOO": {
					Name:    "FOO",
					Default: "bar",
				},
			},
			loaders: []EnvLoader{LoadKV(map[string]string{
				"FOO": "baz",
			})},
			want: map[string]string{
				"FOO": "baz",
			},
		},
		{
			name: "value with  error",
			toLoad: map[string]Variable{
				"FOO": {
					Name:  "FOO",
					Error: "error",
				},
			},
			loaders: []EnvLoader{LoadKV(map[string]string{})},
			wantErr: true,
		},
		{
			name: "value empty set empty",
			toLoad: map[string]Variable{
				"FOO": {
					Name:       "FOO",
					UnsetEmpty: true,
				},
			},
			loaders: []EnvLoader{LoadKV(map[string]string{})},
			want: map[string]string{
				"FOO": "",
			},
		},
		{
			name: "nil loader",
			toLoad: map[string]Variable{
				"FOO": {
					Name: "FOO",
				},
			},
			loaders: []EnvLoader{nil},
			want:    map[string]string{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := LoadEnv(tt.toLoad, tt.loaders...)
			if (err != nil) != tt.wantErr {
				t.Errorf("LoadEnv() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("LoadEnv() = %v, want %v", got, tt.want)
			}
		})
	}
}
