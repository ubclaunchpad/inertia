#!/bin/sh

# Produces a public-private key-pair and outputs the public key.

set -e

ID_DESTINATION=$HOME/.ssh/id_rsa_inertia_deploy
PUB_ID_DESTINATION=$HOME/.ssh/id_rsa_inertia_deploy.pub

# Install openssh if ssh-keygen is not available
if ! hash ssh-keygen 2>/dev/null ; then
    sudo apt-get install openssh-client || sudo apt install openssh-client
fi;

# Check if destination file already exists
if [ -f "$ID_DESTINATION" ]; then
    if [ ! -f "$PUB_ID_DESTINATION" ]; then
        # If public key doesnt exist, make it.
        ssh-keygen -y -f "$ID_DESTINATION" > "$PUB_ID_DESTINATION"
    fi;
else
    # Generate key with no password.
    ssh-keygen -f "$ID_DESTINATION" -t rsa -N ''
fi

ssh-keyscan github.com >> ~/.ssh/known_hosts

cat "$PUB_ID_DESTINATION"
