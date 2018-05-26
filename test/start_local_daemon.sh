#!/bin/sh

# Script for running an Inertia daemon locally.
go build ./daemon/inertiad

echo "Daemon Token:" $(./inertiad token $(pwd)/test/keys/id_rsa)

mkdir inertia_local

sudo ./inertiad run 127.0.0.1 \
    $(pwd)/test/keys/id_rsa \
    $(pwd)/test/certs/ \
    $(pwd)/inertia_local/users.db
