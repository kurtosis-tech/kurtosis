FROM goreleaser/goreleaser

ADD ./login.sh /
RUN chmod +x /login.sh

# Use Git SSH rather than HTTP, so that we can get access to our private Go repos
# An SSH key that can read private repos will be bind-mounted inside this container
RUN git config --global url."git@github.com:".insteadOf "https://github.com/"

# We assume that the source code has been bind-mounted into this directory
WORKDIR /build

ENTRYPOINT bash
