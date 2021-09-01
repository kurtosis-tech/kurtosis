FROM goreleaser/goreleaser
ARG GITHUB_TOKEN

ADD ./login.sh /
RUN chmod +x /login.sh

# This GITHUB_TOKEN will be the KurtosisBot token, provided at build time, so that 'go get' (via goreleaser) will
# have access to private repos
RUN git config --global url."https://${GITHUB_TOKEN}@github.com/".insteadOf "https://github.com/"

# We need to set GOPRIVATE to tell Go not to try and checksum our modules against the public database
RUN go env -w GOPRIVATE=github.com/kurtosis-tech/*

# We assume that the source code has been bind-mounted into this directory
WORKDIR /build

ENTRYPOINT bash
