FROM alpine:3.17

# We need protobut-dev to run protobuf compiler against startosis .proto files
RUN apk update && apk add --no-cache bash protobuf-dev sudo shadow curl xz

# Install Nix
# We need to set filter-syscalls to false to allow Nix install to work properly inside a container with cross platform emulation
# via QEMU: https://github.com/NixOS/nix/issues/5258 and use a more flexible installer https://github.com/DeterminateSystems/nix-installer
# with a workaround on the same issue: https://github.com/DeterminateSystems/nix-installer/issues/324)
RUN curl --proto '=https' --tlsv1.2 -sSf -L https://install.determinate.systems/nix | sh -s -- install linux --no-confirm --init none --extra-conf "filter-syscalls = false"

ARG TARGETARCH

WORKDIR /run

COPY ./build/api-container.$TARGETARCH ./api-container

CMD ./api-container