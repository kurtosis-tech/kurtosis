FROM alpine:3.17

# We need protobut-dev to run protobuf compiler against startosis .proto files
RUN apk update && apk add --no-cache bash protobuf-dev

ARG TARGETARCH

WORKDIR /run

COPY ./build/api-container.$TARGETARCH ./api-container

CMD ./api-container