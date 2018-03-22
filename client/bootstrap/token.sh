#!/bin/sh

set -e

# User argument.
RELEASE=%s

# Generate a daemon token using CLI for API requests.
sudo docker run --rm \
    -v $HOME:/app/host \
    -e SSH_KNOWN_HOSTS='/app/host/.ssh/known_hosts' \
    -e HOME=$HOME \
    --entrypoint=inertia \
    ubclaunchpad/inertia:$RELEASE token
