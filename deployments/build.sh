#!/bin/bash

set -x

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

function install_golang {
    local golang_version=go1.10.3.linux-amd64
    if [ ! -d /root/go ]; then
	curl -O https://dl.google.com/go/$golang_version.tar.gz
        tar -zxf $golang_version.tar.gz
        mv go /root/
        pushd /root/go
        echo GOROOT=$PWD >> /etc/environment
        echo PATH=$PATH:$PWD/bin >> /etc/environment
        popd
        rm -rf $golang_version.tar.gz
    fi
    source /etc/environment
}

function create_temp_gopath {
    mkdir temp_gopath
    pushd temp_gopath
    export GOPATH=$(pwd)
    go get github.com/shank7485/k8-plugin-multicloud/...
    popd
}

function install_dep {
    curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh
}

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

function remove_temp_gopath {
    unset GOPATH
    rm -rf temp_gopath
    rm k8plugin
}

install_golang
create_temp_gopath
install_dep
generate_binary
build_image
remove_temp_gopath
