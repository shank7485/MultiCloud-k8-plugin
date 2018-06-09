# MultiCloud-k8-plugin

[![Go Report Card](https://goreportcard.com/badge/github.com/shank7485/k8-plugin-multicloud)](https://goreportcard.com/report/github.com/shank7485/k8-plugin-multicloud)
[![GoDoc](https://godoc.org/github.com/shank7485/k8-plugin-multicloud?status.svg)](https://godoc.org/github.com/shank7485/k8-plugin-multicloud)

Multicloud Kubernetes plugin for ONAP multicloud.

# Installation
To get souce files, run the following in `$GOPATH`

`go get github.com/shank7485/k8-plugin-multicloud/...`

# Source files
After running the above installation, the binary and source files can be found in:

* Binary: `$GOPATH/bin`

* Source files: `$GOPATH/src/github.com/shank7485`

# Running tests
From the source directory there is a make file, to run unit tests use the make file by doing:

`make test`

# Building 
From the source directory there is a make file, to build use:

`make build`

This generates a binary in `$GOPATH/target/k8client`
