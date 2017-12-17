# Basic script for bringing up the daemon.

set -e

PORT=%s
DAEMON_NAME=inertia-daemon
CONTAINER_PORT=8081
IMAGE_REPOSITORY=ubclaunchpad/inertia

# Check if already running.
ALREADY_RUNNING=`sudo docker ps -q --filter "name=$DAEMON_NAME"`

# Take existing down.
if [ ! -z "$ALREADY_RUNNING" ]; then
    echo "Killing existing container"
    sudo docker rm -f $ALREADY_RUNNING
fi;

# Pull the latest inertia daemon.
sudo docker pull $IMAGE_REPOSITORY

# Run container.
sudo docker run -d \
    -p "$PORT":8081 \
    --name "$DAEMON_NAME" \
    "$IMAGE_REPOSITORY"
