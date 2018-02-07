# Script for neatly outputting information about the test VPS

SSH_PORT=$1 # argument 1: port to map the VPS's SSH port to
IMAGE=$2    # argument 2: VPS image to build

docker run --rm -d \
    -p $SSH_PORT:22 -p 8081:8081 \
    --name testvps \
    --privileged \
    $IMAGE

echo ""
echo "Test VPS is online (kill using 'docker kill testvps')"
echo "Test key:" $GOPATH/src/github.com/ubclaunchpad/inertia/test_env/test_key
