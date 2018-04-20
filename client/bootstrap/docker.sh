#!/bin/sh

# Bootstraps a machine for docker and docker-compose.

set -e

DOCKER_SOURCE=get.docker.com
DOCKER_DEST='/tmp/get-docker.sh'

# Skip installation if Docker is already installed.
if hash docker 2>/dev/null; then
    exit 0
fi;

fetchfile() {
    # Args:
    #   $1 source URL
    #   $2 destination file.
    if hash curl 2>/dev/null; then
        sudo curl -fsSL $1 -o "$2"
    elif hash wget 2>/dev/null; then
        sudo wget -O "$2" $1
    else
        return 1
    fi;
}

# Amazon ECS instances require custom install
if grep -q Amazon /etc/system-release; then
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
    fi
fi

sudo service docker start
