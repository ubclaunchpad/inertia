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

# Make Project directory
mkdir $HOME/project

# Run container with access to the host docker socket and related executables -
# this is necessary because we want the daemon to be able start
# and stop containers on the host. It's also controversial,
# https://www.lvh.io/posts/dont-expose-the-docker-socket-not-even-to-a-container.html
# It's also recommended,
# https://jpetazzo.github.io/2015/09/03/do-not-use-docker-in-docker-for-ci/
# As a result, this container has root access on the remote vps.
sudo docker run -d \
    -p "$PORT":8081 \
    -v /var/run/docker.sock:/var/run/docker.sock \
    -v /usr/bin/docker:/usr/bin/docker \
    -v $HOME:/app/host \
    -e HOME=$HOME
    -e SSH_KNOWN_HOSTS='/app/host/.ssh/known_hosts' \
    --name "$DAEMON_NAME" \
    "$IMAGE_REPOSITORY"

# -v $HOME:/app/host mounts host directory so the daemon can access it