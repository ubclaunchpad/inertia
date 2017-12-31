# Script for bringing project online from the daemon

# Uses a technique similar to the one in daemon-up.sh to get
# around docker-compose installation difficulties by using the
# docker-compose image
# See https://cloud.google.com/community/tutorials/docker-compose-on-container-optimized-os

docker run -d \
    -v /var/run/docker.sock:/var/run/docker.sock \
    -v $HOME:/build \
    -w="/build/project" \
    docker/compose:1.18.0 up --build
