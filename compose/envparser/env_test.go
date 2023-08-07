package envparser

import (
	"bufio"
	"bytes"
	"reflect"
	"strings"
	"testing"

	"github.com/davecgh/go-spew/spew"
	"golang.org/x/exp/maps"
)

func TestFindTemplateVars(t *testing.T) {
	spconfig := spew.NewDefaultConfig()
	spconfig.SortKeys = true

	type tt struct {
		name  string
		in    string
		out   map[string]Variable
		error string
	}

	tests := []tt{
		{
			name: "short",
			in:   `$FOO`,
			out: map[string]Variable{
				"FOO": {
					Name:  "FOO",
					Short: true,
				},
			},
		},
		{
			name: "long",
			in:   `${FOO}`,
			out: map[string]Variable{
				"FOO": {
					Name: "FOO",
				},
			},
		},
		{
			name: "default",
			in:   `${FOO:-bar}`,
			out: map[string]Variable{
				"FOO": {
					Default:    "bar",
					Name:       "FOO",
					UnsetEmpty: true,
				},
			},
		},
		{
			name: "error",
			in:   `${FOO:?bar}`,
			out: map[string]Variable{
				"FOO": {
					Error:      "bar",
					Name:       "FOO",
					UnsetEmpty: true,
				},
			},
		},
		{
			name: "document",
			in:   ELKStackYAML,
			out: map[string]Variable{
				"BEATS_SYSTEM_PASSWORD": {
					Name:       "BEATS_SYSTEM_PASSWORD",
					UnsetEmpty: true,
				},
				"ELASTIC_PASSWORD": {
					Name:       "ELASTIC_PASSWORD",
					UnsetEmpty: true,
				},
				"ELASTIC_VERSION": {
					Name: "ELASTIC_VERSION",
				},
				"FILEBEAT_INTERNAL_PASSWORD": {
					Name:       "FILEBEAT_INTERNAL_PASSWORD",
					UnsetEmpty: true,
				},
				"HEARTBEAT_INTERNAL_PASSWORD": {
					Name:       "HEARTBEAT_INTERNAL_PASSWORD",
					UnsetEmpty: true,
				},
				"KIBANA_SYSTEM_PASSWORD": {
					Name:       "KIBANA_SYSTEM_PASSWORD",
					UnsetEmpty: true,
				},
				"LOGSTASH_INTERNAL_PASSWORD": {
					Name:       "LOGSTASH_INTERNAL_PASSWORD",
					UnsetEmpty: true,
				},
				"METRICBEAT_INTERNAL_PASSWORD": {
					Name:       "METRICBEAT_INTERNAL_PASSWORD",
					UnsetEmpty: true,
				},
				"MONITORING_INTERNAL_PASSWORD": {
					Name:       "MONITORING_INTERNAL_PASSWORD",
					UnsetEmpty: true,
				},
			},
		},
		{
			name: "document with escaped",
			in:   NginxGolangMYSQLYAML,
			out:  map[string]Variable{},
		},
		{
			name: "escaped and unescaped",
			in:   EscapedAndUnescaped,
			out: map[string]Variable{
				"PASSWORD": {
					Error: "password is required",
					Name:  "PASSWORD",
				},
				"INGRESS": {
					Default: "https://notes.*",
					Name:    "INGRESS",
				},
			},
		},
		{
			name: "quoted values",
			in: `
      ${INGRESS:-"https://umami.*"}
      ${INGRESS2?'https://umami.*'}
      ${INGRESS3:-https://umami.*}
      `,
			out: map[string]Variable{
				"INGRESS": {
					Default:    "https://umami.*",
					Name:       "INGRESS",
					UnsetEmpty: true,
				},
				"INGRESS2": {
					Error: "https://umami.*",
					Name:  "INGRESS2",
				},
				"INGRESS3": {
					Default:    "https://umami.*",
					Name:       "INGRESS3",
					UnsetEmpty: true,
				},
			},
		},
		{
			name:  "bad var",
			in:    `${FOO`,
			out:   map[string]Variable{},
			error: "invalid variable",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			out, err := ParseTemplateVars(strings.NewReader(tc.in))
			if err != nil {
				if tc.error == "" {
					t.Fatalf("unexpected error: %s", err)
				}

				if !strings.Contains(err.Error(), tc.error) {
					t.Fatalf("expected error to contain %s, got %s", tc.error, err)
				}
			}
			if len(out) != len(tc.out) {
				t.Errorf("expected %d vars, got %d", len(tc.out), len(out))
			}

			for k, tcv := range tc.out {
				if v, ok := out[k]; !ok || v != tcv {
					t.Errorf("expected %s: %s, got %s: %s", k, spconfig.Sdump(tcv), k, spconfig.Sdump(v))
				}
			}

			maps.DeleteFunc(out, func(k string, v Variable) bool {
				_, ok := tc.out[k]
				return ok
			})
			if len(out) > 0 {
				t.Errorf("unexpected vars: %s", spconfig.Sdump(out))
			}
		})
	}
}

func Test_splitVar(t *testing.T) {
	type args struct {
		data  []byte
		atEOF bool
	}
	tests := []struct {
		name      string
		args      args
		collected []string
		error     string
	}{
		{
			name: "simple",
			args: args{
				data:  []byte("FOO"),
				atEOF: false,
			},
			collected: []string{},
		},
		{
			name: "token",
			args: args{
				data:  []byte("$FOO"),
				atEOF: false,
			},
			collected: []string{"$FOO"},
		},
		{
			name: "token with escaped $",
			args: args{
				data:  []byte("$$FOO"),
				atEOF: false,
			},
			collected: []string{},
		},
		{
			name: "unmatched brace",
			args: args{
				data:  []byte("${FOO"),
				atEOF: false,
			},
			collected: []string{},
			error:     "unmatched brace",
		},
		{
			name: "token in the middle of word",
			args: args{
				data:  []byte("FOO$BAR"),
				atEOF: false,
			},
			collected: []string{"$BAR"},
		},
		{
			name: "token in the middle of document",
			args: args{
				data: []byte(`
        name: ABC
        image: $FOO
        `),
				atEOF: false,
			},
			collected: []string{"$FOO"},
		},
		{
			name: "long token",
			args: args{
				data: []byte(`
        name: ABC
        image: ${FOO-BAR_BAZ}
        `),
				atEOF: false,
			},
			collected: []string{"${FOO-BAR_BAZ}"},
		},
		{
			name: "long token with escaped $",
			args: args{
				data: []byte(`
        name: ABC
        image: $${FOO-BAR_BAZ}
        `),
				atEOF: false,
			},
			collected: []string{},
		},
		{
			name: "long with spaces",
			args: args{
				data:  []byte("${FOO?BAR BAZ}"),
				atEOF: false,
			},
			collected: []string{"${FOO?BAR BAZ}"},
		},
		{
			name: "quoted",
			args: args{
				data:  []byte(`"$FOO"`),
				atEOF: false,
			},
			collected: []string{"$FOO"},
		},
		{
			name: "quoted with escaped $",
			args: args{
				data:  []byte(`"$$FOO"`),
				atEOF: false,
			},
			collected: []string{},
		},
		{
			name: "single quoted",
			args: args{
				data:  []byte(`'$FOO'`),
				atEOF: false,
			},
			collected: []string{"$FOO"},
		},
		{
			name: "weird multine",
			args: args{
				data:  []byte("\nimage: ${FOO-BAR_BAZ}"),
				atEOF: false,
			},
			collected: []string{"${FOO-BAR_BAZ}"},
		},
		{
			name: "multiple",
			args: args{
				data: []byte(`
          name: $ABC
          image: ${FOO-BAR_BAZ}
        `),
				atEOF: false,
			},
			collected: []string{
				"$ABC",
				"${FOO-BAR_BAZ}",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			br := bufio.NewScanner(bytes.NewReader(tt.args.data))
			br.Split(splitVar)

			collected := []string{}
			for br.Scan() {
				collected = append(collected, br.Text())
			}
			if err := br.Err(); err != nil {
				if tt.error == "" {
					t.Errorf("unexpected error: %v", err)
				}

				if !strings.Contains(err.Error(), tt.error) {
					t.Errorf("unexpected error: %v", err)
				}
			}

			if !reflect.DeepEqual(collected, tt.collected) {
				t.Errorf("splitVarfunc() = %v, want %v", collected, tt.collected)
			}
		})
	}
}

func Test_parseVar(t *testing.T) {
	tests := []struct {
		name string
		line string
		want Variable
	}{
		{
			name: "simple",
			line: "$FOO",
			want: Variable{
				Name:  "FOO",
				Short: true,
			},
		},
		{
			name: "long",
			line: "${FOO}",
			want: Variable{
				Name: "FOO",
			},
		},
		{
			name: "error with spaces",
			line: "${FOO?BAR BAZ}",
			want: Variable{
				Name:  "FOO",
				Error: "BAR BAZ",
			},
		},
		{
			name: "default with spaces",
			line: "${FOO:-BAR BAZ}",
			want: Variable{
				Name:       "FOO",
				Default:    "BAR BAZ",
				UnsetEmpty: true,
			},
		},
		{
			name: "default no :",
			line: "${FOO-BAR BAZ}",
			want: Variable{
				Name:    "FOO",
				Default: "BAR BAZ",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseVar([]byte(tt.line))
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseVar() = %v, want %v", got, tt.want)
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

const EscapedAndUnescaped = `
version: '3.9'

volumes:
  siyuan-data:

services:
  siyuan:
    container_name: siyuan
    image: b3log/siyuan:v2.9.6
    user: 1000:1000
    labels:
      kr.ingress.6806: ${INGRESS-https://notes.*}
    volumes:
      - siyuan-data:/siyuan/workspace
    ports:
      - 6806:6806
    command: --workspace /siyuan/workspace --accessAuthCode $$PASSWORD
    environment:
      - TZ=Europe/Berlin
      - PASSWORD=${PASSWORD?password is required}
`
