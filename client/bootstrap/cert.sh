#!/bin/sh

set -e

RELEASE=%s

KEY='daemon.key'
CERT='daemon.cert'

CERTDIR='/app/ssl/'

# Set up SSL certificate
sudo docker run --rm \
    -v $HOME:/app/host \
    -e HOME=$HOME \
    --entrypoint='/bin/sh' \
    ubclaunchpad/inertia:$RELEASE \
    #!/bin/sh \
    /usr/bin/openssl genrsa -out $KEY 2048; \
    /usr/bin/openssl ecparam -genkey -name secp384r1 -out $KEY; \
    /usr/bin/openssl req -new -x509 -sha256 -key $KEY -out $CERT -days 3650; \
    sudo mkdir -p $CERTDIR; sudo mv $KEY $CERTDIR; sudo mv $CERT $CERTDIR;
