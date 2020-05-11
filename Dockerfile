# This Dockerfile builds a tiny little image to ship the Inertia daemon in.

### Setting up daemon build dependencies
FROM golang:alpine AS daemon-build-base
ARG INERTIA_VERSION
ENV BUILD_HOME=/go/src/github.com/ubclaunchpad/inertia \
    INERTIA_VERSION=${INERTIA_VERSION}
WORKDIR ${BUILD_HOME}
COPY go.mod .
COPY go.mod .
RUN apk add --update --no-cache git
RUN go mod download

### Building the Inertia daemon
FROM daemon-build-base AS daemon-build-env
# Mount source code.
ADD . .
# Build daemon binary.
RUN go build -o /bin/inertiad \
    -ldflags "-w -s -X main.Version=$INERTIA_VERSION" \
    ./daemon/inertiad

### Copy builds into combined image for distribution
FROM alpine
LABEL maintainer "UBC Launch Pad team@ubclaunchpad.com"
RUN mkdir -p /daemon
WORKDIR /daemon
COPY --from=daemon-build-env /bin/inertiad /usr/local/bin

# Directories
ENV INERTIA_PROJECT_DIR=/app/host/inertia/project/ \
    INERTIA_DATA_DIR=/app/host/inertia/data/ \
    INERTIA_PERSIST_DIR=/app/host/inertia/persist \
    INERTIA_SECRETS_DIR=/app/host/.inertia/ \
    INERTIA_GH_KEY_PATH=/app/host/.ssh/id_rsa_inertia_deploy

# Serve the daemon by default.
ENTRYPOINT ["inertiad", "run"]
