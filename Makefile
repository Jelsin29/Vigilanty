.PHONY: build test lint clean

VERSION ?= dev
LDFLAGS := -ldflags "-X github.com/jelsin/vigilanty/cmd.Version=$(VERSION)"

build:
	go build $(LDFLAGS) -o vigilanty .

test:
	go test ./... -v

lint:
	go vet ./...

clean:
	rm -f vigilanty
