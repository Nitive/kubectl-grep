.PHONY: *

build:
	@ go build -ldflags '-s -w' -o kubectl-grep main.go

test: build
	@ go test ./...

demo: build
	@ echo 'spec:\n  image: nginx' | ./kubectl-grep image
