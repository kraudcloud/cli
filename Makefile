all: kra

# go install github.com/discord-gophers/goapi-gen@latest
api/openapi.go: api/openapi.yaml
	goapi-gen --generate types --package api --out api/openapi.go api/openapi.yaml

kra: api/openapi.go .PHONY
	go build -o kra .

release:
	GOOS=linux 		GOARCH=arm64 	CGO_ENABLED=0 	go build -o kra-linux-arm64 .
	GOOS=linux 		GOARCH=arm 		CGO_ENABLED=0 	go build -o kra-linux-arm .
	GOOS=linux 		GOARCH=amd64	CGO_ENABLED=0 	go build -o kra-linux-amd64 .
	GOOS=windows 	GOARCH=amd64	CGO_ENABLED=0 	go build -o kra-windows-amd64.exe .
	GOOS=darwin 	GOARCH=amd64	CGO_ENABLED=0 	go build -o kra-darwin-amd64 .
	GOOS=darwin 	GOARCH=arm64	CGO_ENABLED=0 	go build -o kra-darwin-arm64 .

.PHONY:
