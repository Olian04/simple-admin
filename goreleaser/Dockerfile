# syntax=docker/dockerfile:1
#
# Release image for GoReleaser dockers_v2: copies the pre-built Linux binary from the
# build context (see https://goreleaser.com/customization/package/dockers_v2/).

FROM alpine:3.22

RUN apk add --no-cache ca-certificates \
    && addgroup -S app && adduser -S -G app app

USER app

WORKDIR /home/app

ARG VERSION=unknown
ARG REVISION=unknown
ARG BUILD_TIME=unknown
ARG TARGETPLATFORM

LABEL org.opencontainers.image.title="echo-template"
LABEL org.opencontainers.image.description="Sample architecture-template service (replace when forking)."
LABEL org.opencontainers.image.version="${VERSION}"
LABEL org.opencontainers.image.revision="${REVISION}"
LABEL org.opencontainers.image.created="${BUILD_TIME}"

COPY --chmod=0755 ${TARGETPLATFORM}/echo /usr/local/bin/echo

ENTRYPOINT ["/usr/local/bin/echo"]
