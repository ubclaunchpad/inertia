#! /bin/bash

# Specify platforms and release version
echo "Building daemon release $RELEASE"

# Build, tag and push Inertia Docker image
make daemon-release RELEASE="$RELEASE"
