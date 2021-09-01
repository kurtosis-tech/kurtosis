ARG GITHUB_TOKEN

FROM goreleaser/goreleaser

ADD ./login.sh /
RUN chmod +x /login.sh

# This GITHUB_TOKEN will be the KurtosisBot token, provided at build time, so that 'go get' (via goreleaser) will
# have access to private repos
RUN git config --global url."git@github.com:".insteadOf "https://${GITHUB_TOKEN}@github.com/"

# We assume that the source code has been bind-mounted into this directory
WORKDIR /build

ENTRYPOINT bash
