#!/bin/bash
# SPDX-license-identifier: Apache-2.0
##############################################################################
# Copyright (c) 2018
# All rights reserved. This program and the accompanying materials
# are made available under the terms of the Apache License, Version 2.0
# which accompanies this distribution, and is available at
# http://www.apache.org/licenses/LICENSE-2.0
##############################################################################

set -x

IMAGE_NAME="${DOCKER_REPOSITORY:-"nexus3.onap.org:10003"}/${ORG:-"onap"}/${PROJECT:-"multicloud"}/${IMAGE:-"k8plugin"}"

BUILD_ARGS="--no-cache"
if [ $HTTP_PROXY ]; then
    BUILD_ARGS+=" --build-arg HTTP_PROXY=${HTTP_PROXY}"
fi
if [ $HTTPS_PROXY ]; then
    BUILD_ARGS+=" --build-arg HTTPS_PROXY=${HTTPS_PROXY}"
fi

function generate_binary {
    pushd $GOPATH/src/github.com/shank7485/k8-plugin-multicloud
    $GOPATH/bin/dep ensure -v
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -tags netgo -ldflags '-w' -o $GOPATH/k8plugin cmd/main.go
    popd
    mv $GOPATH/k8plugin .
}

function build_image {
    echo "Start build docker image."
    docker build ${BUILD_ARGS} -t ${IMAGE_NAME}:latest .
}

generate_binary
build_image
