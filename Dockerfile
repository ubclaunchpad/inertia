# This Dockerfile builds a tiny little image to ship the Inertia daemon in.

### Part 1 - Building the Web Client
FROM node:carbon AS web-build-env
ENV BUILD_HOME=/go/src/github.com/ubclaunchpad/inertia/daemon/web
# Mount source code.
ADD ./daemon/web ${BUILD_HOME}
WORKDIR ${BUILD_HOME}
# Build and minify client.
RUN npm install --production
RUN npm run build

### Part 2 - Building the Inertia daemon
FROM golang:alpine AS daemon-build-env
ARG INERTIA_VERSION
ENV BUILD_HOME=/go/src/github.com/ubclaunchpad/inertia \
    INERTIA_VERSION=${INERTIA_VERSION}
# Mount source code.
ADD . ${BUILD_HOME}
WORKDIR ${BUILD_HOME}
# Install dependencies if not already available.
RUN if [ ! -d "vendor" ]; then \
    apk add --update --no-cache git; \
    go get -u github.com/golang/dep/cmd/dep; \
    dep ensure; \
    fi
# Build daemon binary.
RUN go build -o /bin/inertiad \
    -ldflags "-w -s -X main.Version=$INERTIA_VERSION" \
    ./daemon/inertiad

### Part 3 - Copy builds into combined image
FROM alpine
LABEL maintainer "UBC Launch Pad team@ubclaunchpad.com"
RUN mkdir -p /daemon
WORKDIR /daemon
COPY --from=daemon-build-env /bin/inertiad /usr/local/bin
COPY --from=web-build-env \
    /go/src/github.com/ubclaunchpad/inertia/daemon/web/public/ \
    /daemon/inertia-web

# Directories
ENV INERTIA_PROJECT_DIR=/app/host/inertia/project/ \
    INERTIA_SSL_DIR=/app/host/inertia/config/ssl/ \
    INERTIA_DATA_DIR=/app/host/inertia/data/ \
    INERTIA_GH_KEY_PATH=/app/host/.ssh/id_rsa_inertia_deploy

# Build tool versions
ENV INERTIA_DOCKERCOMPOSE=docker/compose:1.21.0 \
    INERTIA_HEROKUISH=gliderlabs/herokuish:v0.4.0

# Serve the daemon by default.
ENTRYPOINT ["inertiad", "run"]
