#!/bin/bash

function start_plugin {
    pushd /plugin/
    ./k8plugin
}

start_plugin
