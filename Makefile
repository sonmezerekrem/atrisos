.PHONY: build build-all clean install lint docs docs-serve

VERSION ?= dev
LDFLAGS := -s -w -X github.com/sonmezerekrem/atrisos/cmd.Version=$(VERSION)

build:
	cd app && go build -ldflags "$(LDFLAGS)" -o ../atrisos .

build-all:
	cd app && GOOS=linux  GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o ../dist/atrisos-linux-amd64  .
	cd app && GOOS=linux  GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o ../dist/atrisos-linux-arm64  .
	cd app && GOOS=darwin GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o ../dist/atrisos-darwin-amd64 .
	cd app && GOOS=darwin GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o ../dist/atrisos-darwin-arm64 .

install: build
	install -m 0755 atrisos /usr/local/bin/atrisos

clean:
	rm -f atrisos
	rm -rf dist/

lint:
	cd app && go vet ./...

docs:
	node docs/build.mjs

docs-serve: docs
	cd docs && python3 -m http.server 8080
