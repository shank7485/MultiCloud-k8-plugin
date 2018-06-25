GOPATH := $(shell realpath "$(CURDIR)"/../../../../../)

export GOPATH ...

all: build run_tests
ci: run_tests

build:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -tags netgo -ldflags '-w' -o $(GOPATH)/target/k8plugin $(CURDIR)/../cmd/main.go

run_tests:
	cd ../ && go test -v ./... -cover

format:
	cd ../ && go fmt ./...
