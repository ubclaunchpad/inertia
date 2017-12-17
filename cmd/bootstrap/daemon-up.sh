# Basic script for bringing up the daemon.

# Args:
#    port (int): The port to run the remote daemon on.

set -e

PORT=%s
DAEMON_NAME=inertia-daemon

# Check if already running.
ALREADY_RUNNING=`sudo docker ps -q --filter "name=$DAEMON_NAME"`

# Take existing down.
if [ ! -z "$ALREADY_RUNNING" ]; then
    echo "Killing existing container"
    sudo docker rm -f $ALREADY_RUNNING
fi;

# Run container.
sudo docker run -d \
    -p "$PORT":"$PORT" \
    --name "$DAEMON_NAME" \
    ubclaunchpad/inertia
