FROM goreleaser/goreleaser

ADD ./login.sh /
RUN chmod +x /login.sh

# We assume that the source code has been bind-mounted into this directory
WORKDIR /build

ENTRYPOINT bash
