# Builds a tiny little image to ship the inertia binary in.

# Build the source in a preliminary container.
FROM golang:alpine AS build-env

ENV INERTIA_BUILD_HOME=/go/src/github.com/ubclaunchpad/inertia \
    INERTIA_DAEMON='true'

# Mount source code.
ADD . ${INERTIA_BUILD_HOME}
WORKDIR ${INERTIA_BUILD_HOME}

# Install dependencies if not already available.
RUN apk add --update --no-cache git
RUN if [ ! -d "vendor" ]; then \
    go get -u github.com/golang/dep/cmd/dep; \
    dep ensure; \
    fi

# Build Inertia.
RUN go build -o /bin/inertia -ldflags "-X main.Version=$(git describe --tags)"

# Copy the binary into a smaller image.
FROM alpine
LABEL maintainer "UBC Launchpad team@ubclaunchpad.com"
WORKDIR /app
COPY --from=build-env /bin/inertia /usr/local/bin

# Allow daemon container to generate SSL certificates instead of
# installing it on the host
RUN apk add --update --no-cache openssl

# Container serves daemon by default.
ENTRYPOINT ["inertia", "daemon", "run"]
