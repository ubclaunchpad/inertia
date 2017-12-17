# Builds a tiny little image to ship the inertia binary in.

# Build the source in a preliminary container.
FROM golang:alpine AS build-env

ENV INERTIA_BUILD_HOME=/go/src/github.com/ubclaunchpad/inertia

# Dependencies
RUN apk add --update --no-cache git
RUN go get -u github.com/golang/dep/cmd/dep

# Build the binary.
ADD . ${INERTIA_BUILD_HOME}
WORKDIR ${INERTIA_BUILD_HOME}
RUN dep ensure
RUN go build -o /bin/inertia

# Copy the binary into a smaller image.
FROM alpine
LABEL maintainer "UBC Launchpad team@ubclaunchpad.com"
WORKDIR /app
COPY --from=build-env /bin/inertia /usr/local/bin

# Container serves daemon by default.
ENTRYPOINT ["inertia", "daemon", "run"]
