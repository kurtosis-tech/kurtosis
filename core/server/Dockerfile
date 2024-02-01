FROM alpine:3.17

# We need protobut-dev to run protobuf compiler against startosis .proto files
RUN apk update && apk add --no-cache sudo shadow bash curl xz
RUN sh <(curl -L https://nixos.org/nix/install) --daemon --yes

ARG TARGETARCH

WORKDIR /run

COPY ./build/api-container.$TARGETARCH ./api-container

CMD ./api-container