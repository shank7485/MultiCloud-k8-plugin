#!/bin/bash
# SPDX-license-identifier: Apache-2.0
##############################################################################
# Copyright (c) 2018
# All rights reserved. This program and the accompanying materials
# are made available under the terms of the Apache License, Version 2.0
# which accompanies this distribution, and is available at
# http://www.apache.org/licenses/LICENSE-2.0
##############################################################################

set -o nounset
set -o pipefail
set -o xtrace

function generate_binary {
    GOPATH=$(go env GOPATH)
    rm -f k8plugin
    rm -f *.so
    $GOPATH/bin/dep ensure -v
    for plugin in deployment namespace service; do
        CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -buildmode=plugin -a -tags netgo -o ./$plugin.so ../plugins/$plugin/plugin.go
    done
    CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -a -tags netgo -o ./k8plugin ../cmd/main.go
}

function build_image {
    echo "Start build docker image."
    docker-compose build --no-cache
}

generate_binary
build_image
