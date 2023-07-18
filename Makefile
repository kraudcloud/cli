tag = $(shell git describe --tags)
ldflags = -X 'github.com/kraudcloud/cli/api.Version=$(tag)'

all: kra

# go install github.com/discord-gophers/goapi-gen@latest
api/openapi.go: api/openapi.yaml
	goapi-gen --generate types --package api --out api/openapi.go api/openapi.yaml

kra: api/openapi.go .PHONY
	go build -ldflags="$(ldflags)" -o kra .

release:
	mkdir -p bin
	GOOS=linux 		GOARCH=arm64 	CGO_ENABLED=0 	go build -ldflags="$(ldflags)" -o bin/kra-linux-arm64 .
	GOOS=linux 		GOARCH=arm 		CGO_ENABLED=0 	go build -ldflags="$(ldflags)" -o bin/kra-linux-arm .
	GOOS=linux 		GOARCH=amd64	CGO_ENABLED=0 	go build -ldflags="$(ldflags)" -o bin/kra-linux-amd64 .
	GOOS=windows 	GOARCH=amd64	CGO_ENABLED=0 	go build -ldflags="$(ldflags)" -o bin/kra-windows-amd64.exe .
	GOOS=darwin 	GOARCH=amd64	CGO_ENABLED=0 	go build -ldflags="$(ldflags)" -o bin/kra-darwin-amd64 .
	GOOS=darwin 	GOARCH=arm64	CGO_ENABLED=0 	go build -ldflags="$(ldflags)" -o bin/kra-darwin-arm64 .

.PHONY:
