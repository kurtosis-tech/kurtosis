FROM  alpine:3.16.2

# We need protobut-dev to run protobuf compiler against startosis .proto files
RUN apk update && apk add --no-cache bash protobuf-dev

WORKDIR /run

COPY ./build/api-container ./

CMD ./api-container