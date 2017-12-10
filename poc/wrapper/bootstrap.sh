#!/bin/sh

# Bootstraps a machine for use with inertia.
# This is pretty alpha, it gets docker and docker-compose.
# Installs curl only if it doesn't exist.
# Tested on Ubuntu 16.04. 

DOCKER_SOURCE=get.docker.com
DOCKER_DEST='/tmp/get-docker.sh'
DOCKER_COMPOSE_SOURCE=https://github.com/docker/compose/releases/download/1.17.0/docker-compose-`uname -s`-`uname -m`
DOCKER_COMPOSE_DEST='/usr/local/bin/docker-compose'

fetchfile() {
    # Args:
    #   $1 source URL
    #   $2 destination file.
    if hash curl 2>/dev/null; then
        curl -fsSL $1 -o "$2"
    elif hash wget 2>/dev/null; then
        wget -O "$2" $1
    else
        return 1
    fi;
}

# Get docker if it doesn't exist.
if !(hash docker 2>/dev/null); then
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

# Now get docker-compose - if we've made it this
# far, we have curl or wget installed.
fetchfile $DOCKER_COMPOSE_SOURCE $DOCKER_COMPOSE_DEST
chmod +x $DOCKER_COMPOSE_DEST  # may fail without sudo :(

# Try using.
docker-compose --version
