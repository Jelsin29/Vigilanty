.PHONY: build test lint clean install

VERSION ?= dev
LDFLAGS := -ldflags "-X github.com/Jelsin29/Vigilanty/cmd.Version=$(VERSION)"

GOBIN := $(shell go env GOPATH)/bin
PATH_PRESENT := $(shell echo $(PATH) | grep -q "$(GOBIN)" && echo "yes" || echo "no")

build:
	@mkdir -p bin
	@go build $(LDFLAGS) -o ./bin/vigilanty .

install: setup-path build # temporary for dev
	@cp ./bin/vigilanty $(GOBIN)/vigilanty
	@echo "Successfully installed 'vigilanty' to $(GOBIN)"
	@if [ "$(PATH_PRESENT)" = "no" ]; then \
		echo "NOTE: I've added $(GOBIN) to your .bashrc. Restart your terminal or run 'source ~/.bashrc'."; \
	fi

setup-path:
ifeq ($(PATH_PRESENT),no)
	@echo "GOBIN not found in PATH. Adding to ~/.bashrc."
	@echo 'export PATH="$$PATH:$(GOBIN)"' >> ~/.bashrc
endif

test:
	go test ./... -v

lint:
	go vet ./...

clean:
	rm -f ./bin/*
