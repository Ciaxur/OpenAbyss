#!/usr/bin/env bash

SCRIPT_DIR=`dirname $0`
pushd "$SCRIPT_DIR/.."
mkdir build
go build -o build/server ./server
go build -o build/client ./client
popd