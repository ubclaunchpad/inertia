#!/bin/sh

# Basic script for setting up Inertia requirements (directories, etc)
# and brining the daemon online.

set -e

# User arguments.
DAEMON_RELEASE="%[1]s"
DAEMON_PORT="%[2]s"
HOST_ADDRESS="%[3]s"
WEBHOOK_SECRET="%[4]s"

# Inertia image details.
DAEMON_NAME=inertia-daemon
IMAGE=ubclaunchpad/inertia:$DAEMON_RELEASE

# It doesn't matter what port the daemon runs on in the container
# as long as it is mapped to the correct DAEMON_PORT.
CONTAINER_PORT=4303

# User project
mkdir -p "$HOME"/inertia/project

# Inertia data
mkdir -p "$HOME"/inertia/data

# Configuration
mkdir -p "$HOME"/inertia/config

# Persistent data
mkdir -p "$HOME"/inertia/persist

# Inertia secrets
mkdir -p "$HOME"/.inertia
mkdir -p "$HOME"/.inertia/ssl

# Check if already running and take down existing daemon.
ALREADY_RUNNING=$(sudo docker ps -q --filter "name=$DAEMON_NAME")
if [ ! -z "$ALREADY_RUNNING" ]; then
    echo "Putting existing Inertia daemon to sleep"
    sudo docker rm -f "$ALREADY_RUNNING" > /dev/null 2>&1
fi;

if [ "$DAEMON_RELEASE" != "test" ]; then
    # Download requested daemon image.
    echo "Downloading $IMAGE"
    sudo docker pull "$IMAGE" > /dev/null 2>&1
else
    # Load test build that should have been scp'd into
    # the VPS at /daemon-image.
    echo "Loading $IMAGE"
    sudo docker load -i /daemon-image > /dev/null 2>&1
fi

# Run container with access to the host docker socket and 
# relevant host directories to allow for container control.
# See the README for more details on how this works:
# https://github.com/ubclaunchpad/inertia#how-it-works
echo "Running daemon on port $DAEMON_PORT"
sudo docker run -d \
    --restart unless-stopped \
    -p "$DAEMON_PORT":"$CONTAINER_PORT" \
    -v /var/run/docker.sock:/var/run/docker.sock \
    -v "$HOME":/app/host \
    -e HOME="$HOME" \
    -e SSH_KNOWN_HOSTS='/app/host/.ssh/known_hosts' \
    --name "$DAEMON_NAME" \
    "$IMAGE" "$HOST_ADDRESS --webhook.secret $WEBHOOK_SECRET" > /dev/null # 2>&1
