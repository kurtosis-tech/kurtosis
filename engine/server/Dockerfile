FROM  envoyproxy/envoy-alpine:v1.21.0

RUN apk update && apk add bash

COPY ./envoy.yaml /etc/envoy/envoy.yaml

RUN chmod go+r /etc/envoy/envoy.yaml

WORKDIR /run

COPY ./build/kurtosis-engine ./run-engine-and-envoy.sh ./

RUN chmod +x run-engine-and-envoy.sh

ENV ENVOY_UID=0

CMD ./run-engine-and-envoy.sh
    