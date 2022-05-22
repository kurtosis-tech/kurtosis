FROM  envoyproxy/envoy-alpine:v1.21.0

RUN apk update && apk add bash tar

WORKDIR /run

COPY ./build/files-artifacts-expander ./run-files-artifacts-expander.sh ./

CMD bash ./run-files-artifacts-expander.sh