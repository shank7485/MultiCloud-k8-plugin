#!/bin/bash

BUILD_ARGS="--no-cache"
ORG="onap"
VERSION="1.0.0"
PROJECT="multicloud"
IMAGE="k8plugin"
DOCKER_REPOSITORY="nexus3.onap.org:10003"
IMAGE_NAME="${DOCKER_REPOSITORY}/${ORG}/${PROJECT}/${IMAGE}"
TIMESTAMP=$(date +"%Y%m%dT%H%M%S")

if [ $HTTP_PROXY ]; then
    BUILD_ARGS+=" --build-arg HTTP_PROXY=${HTTP_PROXY}"
fi
if [ $HTTPS_PROXY ]; then
    BUILD_ARGS+=" --build-arg HTTPS_PROXY=${HTTPS_PROXY}"
fi

function generate_binary {
    pushd ../makefiles
    make build
    export GOPATH="$(pwd)"/../../../../../
    popd
    cp $(GOPATH)/target/k8plugin .  
}

function build_image {
    echo "Start build docker image."
    docker build ${BUILD_ARGS} -t ${IMAGE_NAME}:latest .
}

function remove_binary {
    rm k8plugin
}

generate_binary
build_image
remove_binary
