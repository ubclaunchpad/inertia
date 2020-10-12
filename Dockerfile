# This Dockerfile builds a tiny little image to ship the Inertia daemon in.

### Setting up daemon build dependencies
FROM golang:alpine AS daemon-build
ARG INERTIA_VERSION
ENV BUILD_HOME=/go/src/github.com/ubclaunchpad/inertia \
    INERTIA_VERSION=${INERTIA_VERSION}
WORKDIR ${BUILD_HOME}

# Install dependencies and cache them
RUN apk add --update --no-cache git=2.26.2-r0
COPY go.mod .
COPY go.sum .
RUN go mod download

# Mount source code
COPY . .
# Build daemon binary.
RUN go build -o /bin/inertiad \
    -ldflags "-w -s -X main.Version=$INERTIA_VERSION" \
    ./daemon/inertiad

### Copy builds into combined image for distribution in a smaller image
FROM alpine:3.12
LABEL maintainer "UBC Launch Pad team@ubclaunchpad.com"
RUN mkdir -p /daemon
WORKDIR /daemon
COPY --from=daemon-build /bin/inertiad /usr/local/bin

# Directories
ENV INERTIA_PROJECT_DIR=/app/host/inertia/project/ \
    INERTIA_DATA_DIR=/app/host/inertia/data/ \
    INERTIA_PERSIST_DIR=/app/host/inertia/persist \
    INERTIA_SECRETS_DIR=/app/host/.inertia/ \
    INERTIA_GH_KEY_PATH=/app/host/.ssh/id_rsa_inertia_deploy

# Serve the daemon by default.
ENTRYPOINT ["inertiad", "run"]
