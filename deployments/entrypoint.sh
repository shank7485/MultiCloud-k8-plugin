#!/bin/bash

set -x

function start_plugin {
    pushd /plugin/
    ./k8plugin
}

start_plugin
