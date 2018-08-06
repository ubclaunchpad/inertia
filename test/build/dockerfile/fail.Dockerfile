# This dockerfile should fail to build
FROM alpine
RUN exit 1
