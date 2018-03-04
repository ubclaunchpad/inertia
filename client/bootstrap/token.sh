#!/bin/sh

set -e

RELEASE=%s

# Generates an SSH token for use with API requests.
# Generate a daemon token using CLI.
sudo docker run --rm \
    -v $HOME:/app/host \
    -e SSH_KNOWN_HOSTS='/app/host/.ssh/known_hosts' \
    -e INERTIA_DAEMON='true' \
    -e HOME=$HOME \
    --entrypoint=inertia \
    ubclaunchpad/inertia:$RELEASE daemon token
