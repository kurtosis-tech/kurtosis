FROM goreleaser/goreleaser

ADD ./login.sh /

# We assume that the source code has been bind-mounted into this directory
WORKDIR /build

ENTRYPOINT bash
