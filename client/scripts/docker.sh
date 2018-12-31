#!/bin/sh

# Bootstraps a machine for docker.

set -e

DOCKER_SOURCE=https://get.docker.com
DOCKER_DEST="/tmp/get-docker.sh"

startDockerd() {
    # Start dockerd if it is not online
    if ! sudo docker stats --no-stream >/dev/null 2>&1 ; then
        # Fall back to systemctl if service doesn"t work, otherwise just run
        # dockerd in background
        echo "dockerd is offline - starting dockerd..."
        sudo service docker start >/dev/null 2>&1 \
            || sudo systemctl start docker >/dev/null 2>&1 \
            || sudo nohup dockerd >/dev/null 2>&1 &
        echo "dockerd started"
        # Poll until dockerd is running
        while ! sudo docker stats --no-stream >/dev/null 2>&1 ; do
            echo "Waiting for dockerd to come online..."
            sleep 1
        done
    fi;
    echo "dockerd is online"
}

# Skip installation if Docker is already installed.
if hash docker >/dev/null 2>&1; then
    echo "Docker installation detected - skipping install"
    startDockerd
    exit 0
fi;

fetchfile() {
    # Args:
    #   $1 source URL
    #   $2 destination file.
    echo "Saving $1 to $2"
    if hash curl 2>/dev/null; then
        sudo curl -fsSL "$1" -o "$2"
    elif hash wget 2>/dev/null; then
        sudo wget -O "$2" "$1"
    else
        return 1
    fi;
}

echo "Installing docker..."

# Amazon ECS instances require custom install
if grep -q Amazon /etc/system-release >/dev/null 2>&1; then
    echo "AmazonOS detected"
    sudo yum install -y docker
else
    # Try to download using curl or wget,
    # before resorting to installing curl.
    if fetchfile $DOCKER_SOURCE $DOCKER_DEST; then
        sh $DOCKER_DEST
    else
        apt-get update && apt-get -y install curl
        fetchfile $DOCKER_SOURCE $DOCKER_DEST
        sh $DOCKER_DEST
    fi;
fi;

startDockerd

echo "Docker installation complete"

exit 0
