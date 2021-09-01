FROM goreleaser/goreleaser

ADD ./login.sh /
RUN chmod +x /login.sh

# This is the KurtosisBot's private key, which is necessary for pulling down private repos
# As part of CI, it will be made available in this directory
COPY ./id_rsa /root/.ssh/id_rsa

# Use Git SSH rather than HTTP, so that we can get access to our private Go repos
# An SSH key that can read private repos will be bind-mounted inside this container
RUN git config --global url."git@github.com:".insteadOf "https://github.com/"

# We assume that the source code has been bind-mounted into this directory
WORKDIR /build

ENTRYPOINT bash
