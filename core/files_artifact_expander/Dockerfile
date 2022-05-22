FROM  envoyproxy/envoy-alpine:v1.21.0

RUN apk update && apk add bash tar

WORKDIR /run

COPY ./build/files-artifact-expander ./run-files-artifact-expander.sh ./

CMD bash ./run-files-artifact-expander.sh