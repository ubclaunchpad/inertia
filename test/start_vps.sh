#!/bin/sh

# Script for starting up testvps and neatly outputting 
# information about the testvps

set -e

SSH_PORT=$1 # argument 1: port to map the VPS's SSH port to
IMAGE=$2    # argument 2: VPS image to build

docker run --rm -d \
    -p $SSH_PORT:22 \
    -p 4303:4303 \
    -p 8000:8000 \
    --name testvps \
    --privileged \
    $IMAGE

sleep 2 # pause to see if container crashes
RUNNING=$(docker inspect --format="{{.State.Running}}" testvps 2> /dev/null)

if [ "$RUNNING" == "false" ]; then
  echo "Test VPS failed to start, oh no!"; exit 1
fi

CONTAINER_IP=$(docker inspect -f '{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}' testvps)

echo ""
echo "Test VPS is online (kill using 'docker kill testvps')"
echo "SSH port:     " $(docker port testvps 22)
echo "Daemon port:  " $(docker port testvps 4303)
echo "Test key:     " $GOPATH/src/github.com/ubclaunchpad/inertia/test/keys/id_rsa
