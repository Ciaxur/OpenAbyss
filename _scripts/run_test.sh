#!/usr/bin/env bash

PROJ_ROOT="$(dirname $0)/.."

go clean -testcache
go test -v $PROJ_ROOT/tests/**