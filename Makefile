.PHONY: *

build:
	@ mkdir -p bin
	@ go build -ldflags '-s -w' -o bin/kgrep main.go

test: build
	@ go test ./...
