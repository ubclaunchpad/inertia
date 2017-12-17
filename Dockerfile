# Builds a tiny little dockerfile to ship the inertia binary in.
# Useful for running the deamon remotely.

# Build the source in a preliminary container.
FROM golang:alpine AS build-env

ENV INERTIA_BUILD_HOME=/go/src/github.com/ubclaunchpad/inertia

RUN apk add --update --no-cache git
ADD . ${INERTIA_BUILD_HOME}
RUN go get -u github.com/golang/dep/cmd/dep
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
