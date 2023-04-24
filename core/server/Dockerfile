FROM --platform=linux/amd64 alpine:3.17

# We need protobut-dev to run protobuf compiler against startosis .proto files
RUN apk update && apk add --no-cache bash protobuf-dev && apk add musl

WORKDIR /run

COPY ./build/api-container ./

CMD ./api-container