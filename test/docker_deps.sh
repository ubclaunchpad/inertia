#!/bin/sh

echo "Installing build images"
docker pull docker/compose:1.21.0
docker pull gliderlabs/herokuish:v0.4.0
