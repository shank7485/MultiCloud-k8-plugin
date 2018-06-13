GOPATH := $(shell realpath "$(CURDIR)"/../../../../../)

export GOPATH ...

all: build run_tests
ci: run_tests

build:
	go build -o $(GOPATH)/target/k8plugin $(CURDIR)/../cmd/main.go

run_tests:
	cd ../ && go test ./... -cover