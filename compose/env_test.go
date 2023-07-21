package compose

import (
	"io"
	"strings"
	"testing"

	"github.com/davecgh/go-spew/spew"
	"golang.org/x/exp/maps"
)

func TestFindTemplateVars(t *testing.T) {
	spconfig := spew.NewDefaultConfig()
	spconfig.SortKeys = true

	type tt struct {
		name string
		in   string
		out  map[string]EnvExprRhs
	}

	tests := []tt{
		{
			name: "short",
			in:   `$FOO`,
			out: map[string]EnvExprRhs{
				"FOO": {},
			},
		},
		{
			name: "long",
			in:   `${FOO}`,
			out: map[string]EnvExprRhs{
				"FOO": {},
			},
		},
		{
			name: "default",
			in:   `${FOO:-bar}`,
			out: map[string]EnvExprRhs{
				"FOO": {
					Default: "bar",
				},
			},
		},
		{
			name: "error",
			in:   `${FOO:?bar}`,
			out: map[string]EnvExprRhs{
				"FOO": {
					Error: "bar",
				},
			},
		},
		{
			name: "document",
			in:   ELKStackYAML,
			out: map[string]EnvExprRhs{
				"BEATS_SYSTEM_PASSWORD":        {},
				"ELASTIC_PASSWORD":             {},
				"ELASTIC_VERSION":              {},
				"FILEBEAT_INTERNAL_PASSWORD":   {},
				"HEARTBEAT_INTERNAL_PASSWORD":  {},
				"KIBANA_SYSTEM_PASSWORD":       {},
				"LOGSTASH_INTERNAL_PASSWORD":   {},
				"METRICBEAT_INTERNAL_PASSWORD": {},
				"MONITORING_INTERNAL_PASSWORD": {},
			},
		},
		{
			name: "document with escaped",
			in:   NginxGolangMYSQLYAML,
			out:  map[string]EnvExprRhs{},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			out := GetTemplateVars(strings.NewReader(tc.in))
			if len(out) != len(tc.out) {
				t.Errorf("expected %d vars, got %d", len(tc.out), len(out))
			}

			for k, tcv := range tc.out {
				if v, ok := out[k]; !ok || v != tcv {
					t.Errorf("expected %s: %s, got %s: %s", k, spconfig.Sdump(tcv), k, spconfig.Sdump(v))
				}
			}

			maps.DeleteFunc(out, func(k string, v EnvExprRhs) bool {
				_, ok := tc.out[k]
				return ok
			})
			if len(out) > 0 {
				t.Errorf("unexpected vars: %s", spconfig.Sdump(out))
			}
		})
	}
}

// https://raw.githubusercontent.com/deviantony/docker-elk/43716a84adb7b5e9a9bb4ff1169e81075d8120a7/docker-compose.yml
const ELKStackYAML = `
version: '3.7'

services:

  # The 'setup' service runs a one-off script which initializes users inside
  # Elasticsearch — such as 'logstash_internal' and 'kibana_system' — with the
  # values of the passwords defined in the '.env' file. It also creates the
  # roles required by some of these users.
  #
  # This task only needs to be performed once, during the *initial* startup of
  # the stack. Any subsequent run will reset the passwords of existing users to
  # the values defined inside the '.env' file, and the built-in roles to their
  # default permissions.
  #
  # By default, it is excluded from the services started by 'docker compose up'
  # due to the non-default profile it belongs to. To run it, either provide the
  # '--profile=setup' CLI flag to Compose commands, or "up" the service by name
  # such as 'docker compose up setup'.
  setup:
    profiles:
      - setup
    build:
      context: setup/
      args:
        ELASTIC_VERSION: ${ELASTIC_VERSION}
    init: true
    volumes:
      - ./setup/entrypoint.sh:/entrypoint.sh:ro,Z
      - ./setup/lib.sh:/lib.sh:ro,Z
      - ./setup/roles:/roles:ro,Z
    environment:
      ELASTIC_PASSWORD: ${ELASTIC_PASSWORD:-}
      LOGSTASH_INTERNAL_PASSWORD: ${LOGSTASH_INTERNAL_PASSWORD:-}
      KIBANA_SYSTEM_PASSWORD: ${KIBANA_SYSTEM_PASSWORD:-}
      METRICBEAT_INTERNAL_PASSWORD: ${METRICBEAT_INTERNAL_PASSWORD:-}
      FILEBEAT_INTERNAL_PASSWORD: ${FILEBEAT_INTERNAL_PASSWORD:-}
      HEARTBEAT_INTERNAL_PASSWORD: ${HEARTBEAT_INTERNAL_PASSWORD:-}
      MONITORING_INTERNAL_PASSWORD: ${MONITORING_INTERNAL_PASSWORD:-}
      BEATS_SYSTEM_PASSWORD: ${BEATS_SYSTEM_PASSWORD:-}
    networks:
      - elk
    depends_on:
      - elasticsearch

  elasticsearch:
    build:
      context: elasticsearch/
      args:
        ELASTIC_VERSION: ${ELASTIC_VERSION}
    volumes:
      - ./elasticsearch/config/elasticsearch.yml:/usr/share/elasticsearch/config/elasticsearch.yml:ro,Z
      - elasticsearch:/usr/share/elasticsearch/data:Z
    ports:
      - 9200:9200
      - 9300:9300
    environment:
      node.name: elasticsearch
      ES_JAVA_OPTS: -Xms512m -Xmx512m
      # Bootstrap password.
      # Used to initialize the keystore during the initial startup of
      # Elasticsearch. Ignored on subsequent runs.
      ELASTIC_PASSWORD: ${ELASTIC_PASSWORD:-}
      # Use single node discovery in order to disable production mode and avoid bootstrap checks.
      # see: https://www.elastic.co/guide/en/elasticsearch/reference/current/bootstrap-checks.html
      discovery.type: single-node
    networks:
      - elk
    restart: unless-stopped

  logstash:
    build:
      context: logstash/
      args:
        ELASTIC_VERSION: ${ELASTIC_VERSION}
    volumes:
      - ./logstash/config/logstash.yml:/usr/share/logstash/config/logstash.yml:ro,Z
      - ./logstash/pipeline:/usr/share/logstash/pipeline:ro,Z
    ports:
      - 5044:5044
      - 50000:50000/tcp
      - 50000:50000/udp
      - 9600:9600
    environment:
      LS_JAVA_OPTS: -Xms256m -Xmx256m
      LOGSTASH_INTERNAL_PASSWORD: ${LOGSTASH_INTERNAL_PASSWORD:-}
    networks:
      - elk
    depends_on:
      - elasticsearch
    restart: unless-stopped

  kibana:
    build:
      context: kibana/
      args:
        ELASTIC_VERSION: ${ELASTIC_VERSION}
    volumes:
      - ./kibana/config/kibana.yml:/usr/share/kibana/config/kibana.yml:ro,Z
    ports:
      - 5601:5601
    environment:
      KIBANA_SYSTEM_PASSWORD: ${KIBANA_SYSTEM_PASSWORD:-}
    networks:
      - elk
    depends_on:
      - elasticsearch
    restart: unless-stopped

networks:
  elk:
    driver: bridge

volumes:
  elasticsearch:
`

const NginxGolangMYSQLYAML = `
services:
  backend:
    build:
      context: backend
      target: builder
    secrets:
      - db-password
    depends_on:
      db:
        condition: service_healthy

  db:
    # We use a mariadb image which supports both amd64 & arm64 architecture
    image: mariadb:10-focal
    # If you really want to use MySQL, uncomment the following line
    #image: mysql:8
    command: '--default-authentication-plugin=mysql_native_password'
    restart: always
    healthcheck:
      test: ['CMD-SHELL', 'mysqladmin ping -h 127.0.0.1 --password="$$(cat /run/secrets/db-password)" --silent']
      interval: 3s
      retries: 5
      start_period: 30s
    secrets:
      - db-password
    volumes:
      - db-data:/var/lib/mysql
    environment:
      - MYSQL_DATABASE=example
      - MYSQL_ROOT_PASSWORD_FILE=/run/secrets/db-password
    expose:
      - 3306

  proxy:
    image: nginx
    volumes:
      - type: bind
        source: ./proxy/nginx.conf
        target: /etc/nginx/conf.d/default.conf
        read_only: true
    ports:
      - 80:80
    depends_on: 
      - backend

volumes:
  db-data:

secrets:
  db-password:
    file: db/password.txt
`

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
