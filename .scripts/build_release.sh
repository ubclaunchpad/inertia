#! /bin/bash

# Specify platforms and release version
PLATFORMS="linux/amd64 linux/386 darwin/amd64 windows/amd64 windows/386"
echo "Building release $RELEASE"

# Build, tag and push Inertia Docker image
make daemon-release RELEASE="$RELEASE"

# Build Inertia Go binaries for specified platforms
OUTDIR=${OUTDIR:-"dist"}
go run github.com/mitchellh/gox -output="${OUTDIR}/inertia.$(git describe --tags).{{.OS}}.{{.Arch}}" \
    -ldflags "-w -s -X main.Version=$RELEASE" \
    -osarch="$PLATFORMS" \
