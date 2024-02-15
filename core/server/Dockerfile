FROM alpine:3.17

# We need protobut-dev to run protobuf compiler against startosis .proto files
RUN apk update && apk add --no-cache bash protobuf-dev sudo shadow curl xz

# Install Nix
# We need to set filter-syscalls to false to allow Nix to work properly inside a container: https://github.com/NixOS/nix/issues/5258
ENV NIX_CONFIG=$'filter-syscalls = false\nexperimental-features = nix-command flakes'
RUN sh <(curl -L https://nixos.org/nix/install) --daemon --yes

ARG TARGETARCH

WORKDIR /run

COPY ./build/api-container.$TARGETARCH ./api-container

CMD ./api-container