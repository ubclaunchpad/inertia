#!/bin/sh

# Basic script for bringing down the daemon.

set -e

DAEMON_NAME=inertia-daemon

# Get daemon container and take it down if it is running.
ALREADY_RUNNING=`sudo docker ps -q --filter "name=$DAEMON_NAME"`
if [ ! -z "$ALREADY_RUNNING" ]; then
    sudo docker rm -f $ALREADY_RUNNING
fi;
