# Script for taking down all Docker containers except the Inertia daemon

DAEMON_NAME=inertia-daemon
docker rm $(docker ps -a -q | grep -v "$DAEMON_NAME")