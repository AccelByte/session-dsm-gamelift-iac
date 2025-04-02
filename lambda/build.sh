#!/bin/bash

set -e

sudo apt update && sudo apt install -y zip

rm -rf build
mkdir -p build

GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -mod=mod -o build/bootstrap main.go

cd build
zip -j gamelift-event-processor.zip bootstrap
cd ..

echo "Build complete: build/gamelift-event-processor.zip"
