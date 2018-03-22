# This Dockerfile builds a tiny little image to ship the Inertia daemon in.

### Part 1 - Building the Web Client
FROM node:carbon AS web-build-env
ENV BUILD_HOME=/go/src/github.com/ubclaunchpad/inertia/daemon/web
# Mount source code.
ADD ./daemon/web ${BUILD_HOME}
WORKDIR ${BUILD_HOME}
# Build and minify client.
RUN if [ ! -d "node_modules" ]; then \
    npm install --production; \
    fi
RUN npm run build

### Part 2 - Building the Inertia daemon
FROM golang:alpine AS daemon-build-env
ENV BUILD_HOME=/go/src/github.com/ubclaunchpad/inertia
# Mount source code.
ADD . ${BUILD_HOME}
WORKDIR ${BUILD_HOME}
# Install dependencies if not already available.
RUN apk add --update --no-cache git
RUN if [ ! -d "vendor" ]; then \
    go get -u github.com/golang/dep/cmd/dep; \
    dep ensure; \
    fi
# Build daemon binary.
RUN go build -o /bin/inertia \
    -ldflags "-X main.Version=$(git describe --tags)" \
    ./daemon/inertia

### Part 3 - Copy builds into combined image
FROM alpine
LABEL maintainer "UBC Launchpad team@ubclaunchpad.com"
WORKDIR /app
COPY --from=daemon-build-env /bin/inertia /usr/local/bin
COPY --from=web-build-env \
    /go/src/github.com/ubclaunchpad/inertia/daemon/web/public/ \
    /app/inertia-web

# Serve the daemon by default.
ENTRYPOINT ["inertia", "run"]
