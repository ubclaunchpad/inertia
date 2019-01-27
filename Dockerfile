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

### Part 2 - Seting up daemon build dependencies
FROM golang:alpine AS daemon-build-base
ARG INERTIA_VERSION
ENV BUILD_HOME=/go/src/github.com/ubclaunchpad/inertia \
    INERTIA_VERSION=${INERTIA_VERSION}
WORKDIR ${BUILD_HOME}
COPY Gopkg.toml .
COPY Gopkg.lock .
RUN apk add --update --no-cache git
RUN go get -u github.com/golang/dep/cmd/dep
RUN dep ensure -v -vendor-only

### Part 3 - Building the Inertia daemon
FROM daemon-build-base AS daemon-build-env
# Mount source code.
ADD . .
# Build daemon binary.
RUN go build -o /bin/inertiad \
    -ldflags "-w -s -X main.Version=$INERTIA_VERSION" \
    ./daemon/inertiad

### Part 4 - Copy builds into combined image for distribution
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
    INERTIA_DATA_DIR=/app/host/inertia/data/ \
    INERTIA_SECRETS_DIR=/app/host/.inertia/ \
    INERTIA_GH_KEY_PATH=/app/host/.ssh/id_rsa_inertia_deploy

# Build tool versions
ENV INERTIA_DOCKERCOMPOSE=docker/compose:1.23.2

# Serve the daemon by default.
ENTRYPOINT ["inertiad", "run"]
