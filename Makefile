.PHONY: build build-all clean install lint

VERSION ?= dev
LDFLAGS := -s -w -X github.com/sonmezerekrem/atrisos/cmd.Version=$(VERSION)

build:
	go build -ldflags "$(LDFLAGS)" -o atrisos .

build-all:
	GOOS=linux  GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o dist/atrisos-linux-amd64  .
	GOOS=linux  GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o dist/atrisos-linux-arm64  .
	GOOS=darwin GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o dist/atrisos-darwin-amd64 .
	GOOS=darwin GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o dist/atrisos-darwin-arm64 .

install: build
	install -m 0755 atrisos /usr/local/bin/atrisos

clean:
	rm -f atrisos
	rm -rf dist/

lint:
	go vet ./...
