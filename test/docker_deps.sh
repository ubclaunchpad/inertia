#!/bin/sh

echo "Installing build images"
docker pull docker/compose:1.22.0
docker pull gliderlabs/herokuish:v0.4.3
