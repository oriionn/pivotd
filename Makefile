BINARY := bin/pivotd
VERSION ?= 0.1.0

GOFLAGS ?= -trimpath
LDFLAGS ?= -s -w

SOURCES := $(shell find src -name '*.go') go.mod

.PHONY: all build run install uninstall clean fmt vet check help

all: build

build: $(BINARY)


$(BINARY): $(SOURCES)
	CGO_ENABLED=1 go build $(GOFLAGS) -ldflags '$(LDFLAGS)' -o $@ ./src

run: build
	./$(BINARY)

clean:
	rm -f $(BINARY)
