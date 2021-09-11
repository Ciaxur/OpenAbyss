#!/usr/bin/env bash

PROJ_ROOT="$(dirname $0)/.."

# set -x
pushd $PROJ_ROOT
protoc \
  --go_out=./ \
  --go-grpc_out=./ \
  --proto_path=./proto/server/ \
  ./proto/server/server.proto
popd