#!/bin/sh

METALINTER_DIR=bin

echo "Installing linter"
if [ ! -d "$METALINTER_DIR" ]; then
    curl -sfL https://install.goreleaser.com/github.com/alecthomas/gometalinter.sh | bash
else
    echo "./bin directory detected - skipping gometalinter install"
fi
