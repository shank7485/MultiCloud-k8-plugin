GOPATH := $(shell realpath "$(CURDIR)"/../../)

export GOPATH ...

all: build test
test: run_tests

build: 


	go build -o $(GOPATH)/target/k8client $(CURDIR)/cmd/client/*.go

run_tests:
	go test $(CURDIR)/cmd/client/*_test.go -cover