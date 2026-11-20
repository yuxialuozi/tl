BINARY := tl
VERSION ?= 0.1.0
LDFLAGS := -s -w -X main.version=$(VERSION)

.PHONY: build test lint clean

build:
	mkdir -p bin
	go build -trimpath -ldflags "$(LDFLAGS)" -o bin/$(BINARY) ./cmd/tl

test:
	go test ./...

lint:
	go vet ./...

clean:
	rm -rf bin dist coverage.out
