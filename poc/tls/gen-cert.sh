#!/bin/bash

openssl genrsa -out server.key 2048
openssl req -new \
  -x509 -sha256 -key \
  server.key -days 3650 >> server.cert
