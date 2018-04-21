#!/bin/sh

METALINTER_DIR=bin

if [ ! -d "$METALINTER_DIR" ]; then
    curl -sfL https://install.goreleaser.com/github.com/alecthomas/gometalinter.sh | bash
else
    echo "./bin directory detected - skipping gometalinter install"
fi

docker pull docker/compose:1.21.0
docker pull gliderlabs/herokuish:v0.4.0
