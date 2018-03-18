# Basic script for bringing up the daemon.

set -e

DAEMON_RELEASE=%[1]s
DAEMON_PORT=%[2]s
HOST_ADDRESS=%[3]s

DAEMON_NAME=inertia-daemon
IMAGE=ubclaunchpad/inertia:$DAEMON_RELEASE
CONTAINER_PORT=8081

# Set up directories
mkdir -p $HOME/project
mkdir -p $HOME/ssl

# Check if already running and take down existing daemon.
ALREADY_RUNNING=`sudo docker ps -q --filter "name=$DAEMON_NAME"`
if [ ! -z "$ALREADY_RUNNING" ]; then
    echo "Killing existing container..."
    sudo docker rm -f $ALREADY_RUNNING
fi;

# Prepare appropriate daemon image.
if [ "$DAEMON_RELEASE" != "test" ]; then
    # Pull the inertia daemon.
    echo "Pulling Inertia daemon..."
    sudo docker pull $IMAGE
else
    echo "Launching existing Inertia daemon image..."
    sudo docker load -i /daemon-image
fi

# Run container with access to the host docker socket and relevant directories -
# this is necessary because we want the daemon to be able start
# and stop containers on the host. It's also controversial,
# https://www.lvh.io/posts/dont-expose-the-docker-socket-not-even-to-a-container.html
# It's also recommended,
# https://jpetazzo.github.io/2015/09/03/do-not-use-docker-in-docker-for-ci/
# As a result, this container has root access on the remote vps.
sudo docker run --rm \
    -p "$DAEMON_PORT":"$CONTAINER_PORT" \
    -v /var/run/docker.sock:/var/run/docker.sock \
    -v "$HOME":/app/host \
    -e HOME="$HOME" \
    -e SSH_KNOWN_HOSTS='/app/host/.ssh/known_hosts' \
    --name "$DAEMON_NAME" \
    "$IMAGE" "$HOST_ADDRESS"
