#!/usr/bin/env bash

SCRIPT_DIR=`dirname $0`
pushd "$SCRIPT_DIR/.."
mkdir build

platforms=( 'darwin' 'linux' 'windows')
archs=('amd64' 'arm64')
# GOOS=darwin GOARCH=amd64 go build -o hello_world_macOS

# Release Build
if [ "$1" == "--release" ]; then
  for platform in "${platforms[@]}"; do
    for arch in "${archs[@]}"; do
      # Build client binary for declared archs & platforms
      client_binary_name="client-$platform-$arch"
      echo "Building to '$client_binary_name'"
      GOOS=$platform GOARCH=$arch go build -o build/$client_binary_name ./client

      # Build client binary for declared archs & platforms
      server_binary_name="server-$platform-$arch"
      echo "Building to '$server_binary_name'"
      GOOS=$platform GOARCH=$arch go build -o build/$server_binary_name ./server
    done
  done

else
  # Build based on current platform/architecture
  go build -o build/server ./server
  go build -o build/client ./client
fi

popd